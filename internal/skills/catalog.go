package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

const managedMarker = ".qbs-managed"

var validName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Catalog owns one canonical set of curated skills and synchronizes it to
// harness-specific global skill directories.
type Catalog struct {
	Root    string
	Targets []string
}

// ConfirmReplacement decides whether an unmanaged target skill may be
// replaced. The target is the harness directory containing the skill.
type ConfirmReplacement func(name, target string) (bool, error)

// DefaultCatalog resolves the per-user catalog and harness destinations.
// Environment overrides keep the interface portable and testable.
func DefaultCatalog() (Catalog, error) {
	root := os.Getenv("QBS_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Catalog{}, fmt.Errorf("resolve user home: %w", err)
		}
		root = filepath.Join(home, ".qbs")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return Catalog{}, fmt.Errorf("resolve QBS home: %w", err)
	}

	targets, err := defaultTargets()
	if err != nil {
		return Catalog{}, err
	}
	catalogRoot := filepath.Join(root, "skills")
	for _, target := range targets {
		if filepath.Clean(target) == filepath.Clean(catalogRoot) {
			return Catalog{}, fmt.Errorf("skill target %s must differ from the canonical catalog", target)
		}
	}
	return Catalog{Root: catalogRoot, Targets: targets}, nil
}

func defaultTargets() ([]string, error) {
	if configured := os.Getenv("QBS_SKILL_TARGETS"); configured != "" {
		var targets []string
		for _, target := range filepath.SplitList(configured) {
			if target == "" {
				continue
			}
			absolute, err := filepath.Abs(target)
			if err != nil {
				return nil, fmt.Errorf("resolve skill target %q: %w", target, err)
			}
			targets = append(targets, absolute)
		}
		if len(targets) == 0 {
			return nil, errors.New("QBS_SKILL_TARGETS does not contain a usable path")
		}
		return uniquePaths(targets), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home: %w", err)
	}
	targets := []string{
		filepath.Join(home, ".agents", "skills"),
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".config", "opencode", "skills"),
	}

	return uniquePaths(targets), nil
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]bool)
	var unique []string
	for _, path := range paths {
		clean := filepath.Clean(path)
		if !seen[clean] {
			seen[clean] = true
			unique = append(unique, clean)
		}
	}
	return unique
}

// Import copies one skill, or every immediate child skill in a collection,
// into the canonical catalog and synchronizes the imported skills.
func (catalog Catalog) Import(source string, force bool) ([]string, error) {
	return catalog.importWithConfirmation(source, force, nil)
}

// ImportInteractive imports skills and asks before replacing unmanaged target
// skills. A declined replacement leaves that target untouched.
func (catalog Catalog) ImportInteractive(source string, force bool, confirm ConfirmReplacement) ([]string, error) {
	return catalog.importWithConfirmation(source, force, confirm)
}

func (catalog Catalog) importWithConfirmation(source string, force bool, confirm ConfirmReplacement) ([]string, error) {
	sources, err := discoverSources(source)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(catalog.Root, 0o755); err != nil {
		return nil, fmt.Errorf("create skill catalog: %w", err)
	}

	for _, item := range sources {
		if err := validateSource(item.path); err != nil {
			return nil, fmt.Errorf("validate skill %q: %w", item.name, err)
		}
		destination := filepath.Join(catalog.Root, item.name)
		if _, err := os.Lstat(destination); err == nil && !force {
			return nil, fmt.Errorf("skill %q already exists in the catalog; use --force to replace it", item.name)
		} else if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect catalog skill %q: %w", item.name, err)
		}
	}
	names := make([]string, 0, len(sources))
	for _, item := range sources {
		names = append(names, item.name)
	}
	if err := catalog.validateTargets(names, force, confirm); err != nil {
		return nil, err
	}

	for _, item := range sources {
		if err := replaceDirectory(item.path, filepath.Join(catalog.Root, item.name), false); err != nil {
			return names, fmt.Errorf("import skill %q: %w", item.name, err)
		}
	}
	if err := catalog.syncNames(names, force, confirm); err != nil {
		return names, err
	}
	return names, nil
}

// List returns catalogued skill names in lexical order.
func (catalog Catalog) List() ([]string, error) {
	entries, err := os.ReadDir(catalog.Root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read skill catalog: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(catalog.Root, entry.Name(), "SKILL.md")); err == nil {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// Sync copies every catalogued skill to each configured harness destination.
func (catalog Catalog) Sync() ([]string, error) {
	names, err := catalog.List()
	if err != nil {
		return nil, err
	}
	return names, catalog.SyncNames(names)
}

// SyncInteractive synchronizes skills and asks before replacing unmanaged
// target skills. A declined replacement leaves that target untouched.
func (catalog Catalog) SyncInteractive(confirm ConfirmReplacement) ([]string, error) {
	names, err := catalog.List()
	if err != nil {
		return nil, err
	}
	return names, catalog.syncNames(names, false, confirm)
}

// SyncNames synchronizes selected catalogued skills. QBS only replaces
// destinations carrying its ownership marker.
func (catalog Catalog) SyncNames(names []string) error {
	return catalog.syncNames(names, false, nil)
}

func (catalog Catalog) syncNames(names []string, force bool, confirm ConfirmReplacement) error {
	if err := catalog.validateTargets(names, force, confirm); err != nil {
		return err
	}
	for _, target := range catalog.Targets {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("create skill target %s: %w", target, err)
		}
		for _, name := range names {
			if err := validateName(name); err != nil {
				return err
			}
			source := filepath.Join(catalog.Root, name)
			if _, err := os.Stat(filepath.Join(source, "SKILL.md")); err != nil {
				return fmt.Errorf("catalog skill %q is invalid: SKILL.md is required", name)
			}
			destination := filepath.Join(target, name)
			if _, err := os.Lstat(destination); err == nil && !force && !isManaged(destination) && confirm != nil {
				ok, err := confirm(name, target)
				if err != nil {
					return err
				}
				if !ok {
					continue
				}
			} else if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("inspect skill target %s: %w", destination, err)
			}
			if err := replaceDirectory(source, destination, true); err != nil {
				return fmt.Errorf("sync skill %q to %s: %w", name, target, err)
			}
		}
	}
	return nil
}

func (catalog Catalog) validateTargets(names []string, force bool, confirm ConfirmReplacement) error {
	for _, name := range names {
		if err := validateName(name); err != nil {
			return err
		}
		for _, target := range catalog.Targets {
			destination := filepath.Join(target, name)
			if info, err := os.Lstat(destination); err == nil {
				if !info.IsDir() {
					return fmt.Errorf("preserving unmanaged skill target at %s", destination)
				}
				if !isManaged(destination) && !force && confirm == nil {
					return fmt.Errorf("preserving unmanaged skill at %s; move it or import it into QBS before syncing", destination)
				}
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("inspect skill target %s: %w", destination, err)
			}
		}
	}
	return nil
}

// Remove deletes a catalogued skill and only the synchronized copies marked
// as QBS-managed.
func (catalog Catalog) Remove(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	canonical := filepath.Join(catalog.Root, name)
	if _, err := os.Stat(filepath.Join(canonical, "SKILL.md")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q is not in the catalog", name)
		}
		return err
	}
	for _, target := range catalog.Targets {
		destination := filepath.Join(target, name)
		if isManaged(destination) {
			if err := os.RemoveAll(destination); err != nil {
				return fmt.Errorf("remove synchronized skill %s: %w", destination, err)
			}
		}
	}
	if err := os.RemoveAll(canonical); err != nil {
		return fmt.Errorf("remove catalog skill %q: %w", name, err)
	}
	return nil
}

func isManaged(directory string) bool {
	data, err := os.ReadFile(filepath.Join(directory, managedMarker))
	return err == nil && string(data) == "managed by qbs\n"
}

type sourceSkill struct {
	name string
	path string
}

func discoverSources(source string) ([]sourceSkill, error) {
	absolute, err := filepath.Abs(source)
	if err != nil {
		return nil, fmt.Errorf("resolve skill source: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, fmt.Errorf("inspect skill source: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("skill source %s is not a directory", absolute)
	}
	if _, err := os.Stat(filepath.Join(absolute, "SKILL.md")); err == nil {
		name := filepath.Base(absolute)
		if err := validateName(name); err != nil {
			return nil, err
		}
		return []sourceSkill{{name: name, path: absolute}}, nil
	}

	entries, err := os.ReadDir(absolute)
	if err != nil {
		return nil, err
	}
	var skills []sourceSkill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(absolute, entry.Name())
		if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err == nil {
			if err := validateName(entry.Name()); err != nil {
				return nil, err
			}
			skills = append(skills, sourceSkill{name: entry.Name(), path: path})
		}
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].name < skills[j].name })
	if len(skills) == 0 {
		return nil, fmt.Errorf("%s contains no skill directories with SKILL.md", absolute)
	}
	return skills, nil
}

func validateName(name string) error {
	if !validName.MatchString(name) || name == "." || name == ".." {
		return fmt.Errorf("invalid skill name %q", name)
	}
	return nil
}

func replaceDirectory(source, destination string, markManaged bool) error {
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp(parent, ".qbs-skill-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	staged := filepath.Join(temporary, filepath.Base(destination))
	if err := copyDirectory(source, staged); err != nil {
		return err
	}
	if markManaged {
		if err := os.WriteFile(filepath.Join(staged, managedMarker), []byte("managed by qbs\n"), 0o644); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(destination); err != nil {
		return err
	}
	return os.Rename(staged, destination)
}

func copyDirectory(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not supported: %s", path)
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

func validateSource(source string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not supported: %s", path)
		}
		return nil
	})
}
