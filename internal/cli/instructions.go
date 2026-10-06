package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/trues/qbs/internal/git"
	"github.com/trues/qbs/internal/templates"
)

const (
	instructionBlockStart = "<!-- qbs:begin -->"
	instructionBlockEnd   = "<!-- qbs:end -->"
)

// legacyInstructions lists the exact per-repository instruction files that
// older QBS versions wrote before instructions moved to global harness files.
var legacyInstructions = map[string][]string{
	"AGENTS.md": {"legacy/qbs-agents-v1.md", "legacy/qbs-agents-v2.md"},
	"CLAUDE.md": {"legacy/qbs-claude-v1.md", "legacy/qbs-claude-v2.md"},
}

// syncGlobalInstructions installs the shared QBS instructions where each
// harness reads user-level instructions. Claude gets a QBS-owned rules file;
// Codex and OpenCode have a single global AGENTS.md, so QBS maintains only a
// marked block inside it and leaves the rest of the file untouched.
func syncGlobalInstructions() ([]string, error) {
	data, err := templates.Read("instructions.md")
	if err != nil {
		return nil, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home: %w", err)
	}
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		codexHome = filepath.Join(home, ".codex")
	}

	claude := filepath.Join(home, ".claude", "rules", "qbs.md")
	if err := writeFile(claude, data); err != nil {
		return nil, fmt.Errorf("install %s: %w", claude, err)
	}
	paths := []string{claude}
	for _, path := range []string{
		filepath.Join(codexHome, "AGENTS.md"),
		filepath.Join(home, ".config", "opencode", "AGENTS.md"),
	} {
		if err := upsertInstructionBlock(path, data); err != nil {
			return nil, fmt.Errorf("install %s: %w", path, err)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func upsertInstructionBlock(path string, data []byte) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	block := instructionBlockStart + "\n" + string(data) + instructionBlockEnd + "\n"
	start := strings.Index(content, instructionBlockStart)
	end := strings.Index(content, instructionBlockEnd)
	switch {
	case start >= 0 && end > start:
		end += len(instructionBlockEnd)
		if end < len(content) && content[end] == '\n' {
			end++
		}
		content = content[:start] + block + content[end:]
	case start >= 0 || end >= 0:
		return fmt.Errorf("found an incomplete QBS block; restore both %s and %s or remove both", instructionBlockStart, instructionBlockEnd)
	case content == "":
		content = block
	default:
		content = strings.TrimRight(content, "\n") + "\n\n" + block
	}
	if content == string(existing) {
		return nil
	}
	return writeFile(path, []byte(content))
}

func writeFile(path string, data []byte) error {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, data) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// removeLegacyInstructions deletes per-repository instruction files that are
// exact copies of what an older QBS version wrote, and drops their Git exclude
// entries once no file remains to hide. Edited copies are preserved with a
// warning so no user change is lost.
func removeLegacyInstructions(repo git.Repository, stderr io.Writer) error {
	tracked, err := git.TrackedPaths(repo.Root, "AGENTS.md", "CLAUDE.md")
	if err != nil {
		return fmt.Errorf("inspect tracked instruction files: %w", err)
	}
	var gone []string
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		if slices.Contains(tracked, name) {
			continue
		}
		path := filepath.Join(repo.Root, name)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			gone = append(gone, name)
			continue
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		switch {
		case isLegacyInstruction(name, data):
			if err := os.Remove(path); err != nil {
				return err
			}
			gone = append(gone, name)
		case bytes.HasPrefix(data, []byte("<!-- qbs-managed-instruction:")):
			_, _ = fmt.Fprintf(stderr, "warning: preserving edited QBS file %s; shared instructions now come from qbs sync, so remove it once your changes are elsewhere\n", path)
		}
	}
	return removeExcludeEntries(filepath.Join(repo.CommonDir, "info", "exclude"), gone)
}

func isLegacyInstruction(name string, data []byte) bool {
	for _, template := range legacyInstructions[name] {
		legacy, err := templates.Read(template)
		if err == nil && bytes.Equal(data, legacy) {
			return true
		}
	}
	return false
}

func removeExcludeEntries(path string, entries []string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	drop := make(map[string]bool)
	for _, entry := range entries {
		drop[entry] = true
	}
	lines := strings.SplitAfter(string(data), "\n")
	kept := lines[:0]
	for _, line := range lines {
		if !drop[strings.TrimSpace(line)] {
			kept = append(kept, line)
		}
	}
	if len(kept) == len(lines) {
		return nil
	}
	return os.WriteFile(path, []byte(strings.Join(kept, "")), 0o644)
}
