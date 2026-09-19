package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/trues/qbs/internal/agents"
	"github.com/trues/qbs/internal/git"
	"github.com/trues/qbs/internal/skills"
	"github.com/trues/qbs/internal/templates"
)

const managedInstructionMarker = "<!-- qbs-managed-instruction: v2 -->"

var projectSkillDirectories = []string{
	".agents/skills",
	".claude/skills",
	".opencode/skills",
}

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
	"docs/agents/",
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
	"docs/agents",
}

func initializeRepository(dir string, stdout, stderr io.Writer) error {
	return provisionRepositoryWithLabel(dir, "Initialized", stdout, stderr)
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
	if err := provisionAgents(root, stderr); err != nil {
		return err
	}
	if err := provisionEngineeringConfig(root, stderr); err != nil {
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

func provisionEngineeringConfig(root string, stderr io.Writer) error {
	for _, name := range []string{"domain.md", "issue-tracker.md"} {
		data, err := engineeringTemplate(name)
		if err != nil {
			return err
		}
		if err := writeIfAbsent(filepath.Join(root, "docs", "agents", name), data, stderr); err != nil {
			return fmt.Errorf("provision docs/agents/%s: %w", name, err)
		}
	}
	hasTriage, err := hasTriageSkill(root)
	if err != nil {
		return err
	}
	if hasTriage {
		data, err := templates.Read("docs/agents/triage.md")
		if err != nil {
			return err
		}
		if err := writeIfAbsent(filepath.Join(root, "docs", "agents", "triage.md"), data, stderr); err != nil {
			return fmt.Errorf("provision docs/agents/triage.md: %w", err)
		}
	}
	return nil
}

func engineeringTemplate(name string) ([]byte, error) {
	if name == "issue-tracker.md" {
		name = "issue-tracker-local.md"
	}
	return templates.Read("docs/agents/" + name)
}

func hasTriageSkill(root string) (bool, error) {
	var skillRoots []string
	for _, base := range projectSkillDirectories {
		skillRoots = append(skillRoots, filepath.Join(root, base))
	}

	catalog, err := skills.DefaultCatalog()
	if err != nil {
		return false, fmt.Errorf("resolve global skill locations: %w", err)
	}
	skillRoots = append(skillRoots, catalog.Root)
	skillRoots = append(skillRoots, catalog.Targets...)

	for _, base := range skillRoots {
		if skillInstalled(base, "triage") {
			return true, nil
		}
	}
	return false, nil
}

func skillInstalled(base, name string) bool {
	info, err := os.Stat(filepath.Join(base, name, "SKILL.md"))
	return err == nil && info.Mode().IsRegular()
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
		if err := writeInstruction(filepath.Join(root, name), data, stderr); err != nil {
			return fmt.Errorf("provision %s: %w", name, err)
		}
	}
	return nil
}

// writeInstruction updates an unchanged instruction file produced by an older
// QBS version, while preserving files that a user has customized. Current
// templates carry a version marker so future migrations can identify their
// ownership without treating an edited file as disposable.
func writeInstruction(path string, data []byte, stderr io.Writer) error {
	existing, err := os.ReadFile(path)
	if err == nil {
		if bytes.Equal(existing, data) {
			return nil
		}
		if isLegacyInstruction(path, existing) {
			return os.WriteFile(path, data, 0o644)
		}
		_, _ = fmt.Fprintf(stderr, "warning: preserving existing local file %s\n", path)
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func isMarkedInstruction(data []byte) bool {
	return bytes.HasPrefix(data, []byte(managedInstructionMarker+"\n"))
}

func isLegacyInstruction(path string, data []byte) bool {
	// A marked file may be an older QBS version or a user-customized QBS file.
	// Preserve it unless a future migration explicitly adds its exact bytes to
	// the legacy templates.
	if isMarkedInstruction(data) {
		return false
	}
	name := filepath.Base(path)
	if name != "AGENTS.md" && name != "CLAUDE.md" {
		return false
	}
	legacyName := map[string]string{
		"AGENTS.md": "legacy/qbs-agents-v1.md",
		"CLAUDE.md": "legacy/qbs-claude-v1.md",
	}[name]
	legacy, err := templates.Read(legacyName)
	return err == nil && bytes.Equal(data, legacy)
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
// before any exclude or file changes are made. Existing instruction files are
// preserved; managed agent files are regenerated. Project-local skills are not
// provisioning targets.
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
	if err := validateDirectoryTarget(filepath.Join(root, "docs", "agents")); err != nil {
		return err
	}
	for _, name := range []string{"domain.md", "issue-tracker.md", "triage.md"} {
		if err := validateFileTarget(filepath.Join(root, "docs", "agents", name)); err != nil {
			return err
		}
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
	return nil
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
