package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/trues/qbs/internal/agents"
	"github.com/trues/qbs/internal/git"
	"github.com/trues/qbs/internal/templates"
)

var excludedPaths = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".agents/skills/",
	".claude/agents/",
	".claude/skills/",
	".codex/agents/",
	".opencode/agents/",
	".opencode/skills/",
	".research/",
	".specs/",
}

var provisionedPathspecs = []string{
	"AGENTS.md",
	"CLAUDE.md",
	".agents/skills",
	".claude/agents",
	".claude/skills",
	".codex/agents",
	".opencode/agents",
	".opencode/skills",
	".research",
	".specs",
}

func initializeRepository(dir string, stdout, stderr io.Writer) error {
	return provisionRepositoryWithLabel(dir, "Initialized", stdout, stderr)
}

func provisionRepository(dir string, stdout, stderr io.Writer) error {
	return provisionRepositoryWithLabel(dir, "Provisioned", stdout, stderr)
}

func provisionRepositoryWithLabel(dir, label string, stdout, stderr io.Writer) error {
	repo, err := git.Detect(dir)
	if err != nil {
		return err
	}
	if err := prepareProvisioning(repo, repo.Root); err != nil {
		return fmt.Errorf("cannot provision local AI workspace: %w", err)
	}
	if err := provisionWorkspace(repo.Root, stderr); err != nil {
		return fmt.Errorf("initialize local AI workspace after Git excludes were configured: %w; existing exclude changes and any files already provisioned remain", err)
	}
	_, err = fmt.Fprintf(stdout, "%s Git worktree: %s\n", label, repo.Root)
	return err
}

func prepareProvisioning(repo git.Repository, target string) error {
	tracked, err := git.TrackedPaths(repo.Root, provisionedPathspecs...)
	if err != nil {
		return fmt.Errorf("inspect tracked local AI paths: %w", err)
	}
	if len(tracked) > 0 {
		return fmt.Errorf("local AI files must be untracked, but Git already tracks: %s", strings.Join(tracked, ", "))
	}
	if err := validateProvisioningTargets(target); err != nil {
		return err
	}
	if err := updateExcludeFile(filepath.Join(repo.CommonDir, "info", "exclude")); err != nil {
		return fmt.Errorf("configure Git excludes: %w", err)
	}
	return nil
}

func provisionWorkspace(root string, stderr io.Writer) error {
	if err := provisionInstructions(root, stderr); err != nil {
		return err
	}
	if err := provisionSkills(root, stderr); err != nil {
		return err
	}
	if err := provisionAgents(root, stderr); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, ".research"), 0o755); err != nil {
		return fmt.Errorf("create .research: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".specs"), 0o755); err != nil {
		return fmt.Errorf("create .specs: %w", err)
	}
	return nil
}

func updateExcludeFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(data)
	seen := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		seen[strings.TrimSpace(line)] = true
	}
	var missing []string
	for _, entry := range excludedPaths {
		if !seen[entry] {
			missing = append(missing, entry)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += strings.Join(missing, "\n") + "\n"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func provisionInstructions(root string, stderr io.Writer) error {
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		data, err := templates.Read(name)
		if err != nil {
			return err
		}
		if err := writeIfAbsent(filepath.Join(root, name), data, stderr); err != nil {
			return fmt.Errorf("provision %s: %w", name, err)
		}
	}
	return nil
}

func provisionSkills(root string, stderr io.Writer) error {
	return fs.WalkDir(templates.Files, "canonical/skills", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative := strings.TrimPrefix(path, "canonical/skills/")
		if relative == "" {
			return nil
		}
		destinations := []string{
			filepath.Join(root, ".agents", "skills", relative),
			filepath.Join(root, ".claude", "skills", relative),
			filepath.Join(root, ".opencode", "skills", relative),
		}
		for _, destination := range destinations {
			if entry.IsDir() {
				if err := os.MkdirAll(destination, 0o755); err != nil {
					return err
				}
				continue
			}
			data, err := fs.ReadFile(templates.Files, path)
			if err != nil {
				return err
			}
			if err := writeIfAbsent(destination, data, stderr); err != nil {
				return fmt.Errorf("provision skill %s: %w", relative, err)
			}
		}
		return nil
	})
}

func provisionAgents(root string, stderr io.Writer) error {
	files, err := canonicalAgentFiles()
	if err != nil {
		return err
	}
	for _, file := range files {
		path := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create agent directory for %s: %w", file.Path, err)
		}
		if existing, err := os.ReadFile(path); err == nil {
			if !agents.IsGenerated(existing) {
				if string(existing) != string(file.Data) {
					_, _ = fmt.Fprintf(stderr, "warning: preserving unmanaged local agent %s\n", path)
				}
				continue
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect agent %s: %w", file.Path, err)
		}
		if err := os.WriteFile(path, file.Data, 0o644); err != nil {
			return fmt.Errorf("generate agent %s: %w", file.Path, err)
		}
	}
	return nil
}

func canonicalAgentFiles() ([]agents.File, error) {
	catalog, err := agents.Canonical()
	if err != nil {
		return nil, err
	}
	files, err := agents.GenerateAll(catalog)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// validateProvisioningTargets checks every path that provisioning may need
// before any exclude or file changes are made. Existing regular instruction
// and skill files are preserved; managed agent files are regenerated.
func validateProvisioningTargets(root string) error {
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		if err := validateFileTarget(filepath.Join(root, name)); err != nil {
			return err
		}
	}
	if err := validateDirectoryTarget(filepath.Join(root, ".research")); err != nil {
		return err
	}
	if err := validateDirectoryTarget(filepath.Join(root, ".specs")); err != nil {
		return err
	}
	for _, base := range []string{".claude/agents", ".codex/agents", ".opencode/agents"} {
		if err := validateDirectoryTarget(filepath.Join(root, filepath.FromSlash(base))); err != nil {
			return err
		}
	}
	agentFiles, err := canonicalAgentFiles()
	if err != nil {
		return err
	}
	for _, file := range agentFiles {
		if err := validateFileTarget(filepath.Join(root, filepath.FromSlash(file.Path))); err != nil {
			return err
		}
	}
	return fs.WalkDir(templates.Files, "canonical/skills", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative := strings.TrimPrefix(path, "canonical/skills/")
		if relative == "" {
			return nil
		}
		for _, base := range []string{".agents/skills", ".claude/skills", ".opencode/skills"} {
			target := filepath.Join(root, base, filepath.FromSlash(relative))
			if entry.IsDir() {
				if err := validateDirectoryTarget(target); err != nil {
					return err
				}
			} else if err := validateFileTarget(target); err != nil {
				return err
			}
		}
		return nil
	})
}

func validateFileTarget(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file, but a file is required", path)
	}
	return nil
}

func validateDirectoryTarget(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory, but a directory is required", path)
	}
	return nil
}

func writeIfAbsent(path string, data []byte, stderr io.Writer) error {
	if existing, err := os.ReadFile(path); err == nil {
		if string(existing) != string(data) {
			_, _ = fmt.Fprintf(stderr, "warning: preserving existing local file %s\n", path)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
