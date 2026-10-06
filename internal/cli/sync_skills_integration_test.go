package cli_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/cli"
	"github.com/trues/qbs/internal/templates"
)

func TestSyncInstallsSkillsShippedWithThisRelease(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(repoRoot, ".agents", "skills", "goal-ticket")
	if _, err := os.Stat(filepath.Join(source, "SKILL.md")); err != nil {
		t.Fatalf("release skill source is missing: %v", err)
	}

	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("CODEX_HOME", "")
	catalog := filepath.Join(root, ".qbs")
	codexSkills := filepath.Join(root, "codex", "skills")
	claudeSkills := filepath.Join(root, "claude", "skills")
	t.Setenv("QBS_HOME", catalog)
	t.Setenv("QBS_SKILL_TARGETS", codexSkills+string(os.PathListSeparator)+claudeSkills)

	var stdout, stderr bytes.Buffer
	if err := cli.Run([]string{"sync"}, &stdout, &stderr); err != nil {
		t.Fatalf("initial qbs sync failed: %v\nstderr: %s", err, stderr.String())
	}

	destinations := []string{
		filepath.Join(catalog, "skills", "goal-ticket"),
		filepath.Join(codexSkills, "goal-ticket"),
		filepath.Join(claudeSkills, "goal-ticket"),
	}
	for _, destination := range destinations {
		assertDirectoryFilesMatch(t, source, destination)
	}

	for _, destination := range destinations {
		if err := os.WriteFile(filepath.Join(destination, "SKILL.md"), []byte("corrupted contents\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stdout.Reset()
	stderr.Reset()
	if err := cli.Run([]string{"sync"}, &stdout, &stderr); err != nil {
		t.Fatalf("qbs sync failed to restore release skill: %v\nstderr: %s", err, stderr.String())
	}
	for _, destination := range destinations {
		assertDirectoryFilesMatch(t, source, destination)
	}
}

func TestSyncInstallsSharedInstructionsGlobally(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)
	t.Setenv("QBS_HOME", filepath.Join(root, ".qbs"))
	t.Setenv("QBS_SKILL_TARGETS", filepath.Join(root, "skills"))
	codexHome := filepath.Join(root, "codex-home")
	t.Setenv("CODEX_HOME", codexHome)
	instructions, err := templates.Read("instructions.md")
	if err != nil {
		t.Fatal(err)
	}
	block := "<!-- qbs:begin -->\n" + string(instructions) + "<!-- qbs:end -->\n"

	codex := filepath.Join(codexHome, "AGENTS.md")
	if err := os.MkdirAll(codexHome, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codex, []byte("My rules.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"sync"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	stale := strings.Replace(block, "ponytail", "stale", 1)
	if err := os.WriteFile(codex, []byte("My rules.\n\n"+stale+"More rules.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"sync"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	for path, want := range map[string]string{
		filepath.Join(root, ".claude", "rules", "qbs.md"): string(instructions),
		codex: "My rules.\n\n" + block + "More rules.\n",
		filepath.Join(root, ".config", "opencode", "AGENTS.md"): block,
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Errorf("%s = %q, want %q", path, data, want)
		}
	}
}

func assertDirectoryFilesMatch(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		want, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		got, err := os.ReadFile(filepath.Join(destination, relative))
		if err != nil {
			return err
		}
		if string(got) != string(want) {
			t.Errorf("%s differs from skill source", filepath.Join(destination, relative))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("compare %s with skill source: %v", destination, err)
	}
}
