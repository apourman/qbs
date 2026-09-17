package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/cli"
)

func TestRunHelpAndVersion(t *testing.T) {
	for _, args := range [][]string{{}, {"help"}, {"--help"}} {
		var out bytes.Buffer
		if err := cli.Run(args, &out, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "qbs init") {
			t.Errorf("help for %v = %q", args, out.String())
		}
		if !strings.Contains(out.String(), "qbs provision") || !strings.Contains(out.String(), "qbs skills import") {
			t.Errorf("help for %v is missing current commands: %q", args, out.String())
		}
	}
	var out bytes.Buffer
	if err := cli.Run([]string{"--version"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "qbs "+cli.Version+"\n" {
		t.Errorf("version = %q", got)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	err := cli.Run([]string{"wat"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), `unknown command "wat"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestRunSkillsLifecycle(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source", "review")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "SKILL.md"), []byte("review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("QBS_HOME", filepath.Join(root, "qbs-home"))
	t.Setenv("QBS_SKILL_TARGETS", filepath.Join(root, "target"))

	var out bytes.Buffer
	if err := cli.Run([]string{"skills", "import", source}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "review") {
		t.Fatalf("import output = %q", out.String())
	}
	out.Reset()
	if err := cli.Run([]string{"skills", "list"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "review") {
		t.Fatalf("list output = %q", out.String())
	}
	if err := cli.Run([]string{"skills", "remove", "review"}, &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
}

func TestRunSkillsImportPromptsForExistingTarget(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source", "review")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "SKILL.md"), []byte("qbs review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(filepath.Join(target, "review"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "review", "SKILL.md"), []byte("existing review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("QBS_HOME", filepath.Join(root, "qbs-home"))
	t.Setenv("QBS_SKILL_TARGETS", target)

	var out bytes.Buffer
	if err := cli.RunWithInput([]string{"skills", "import", source}, strings.NewReader("y\n"), &out, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Replace it?") {
		t.Fatalf("prompt output = %q", out.String())
	}
	data, err := os.ReadFile(filepath.Join(target, "review", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "qbs review\n" {
		t.Fatalf("replacement content = %q", data)
	}
}

func TestRunInitOutsideGitRepository(t *testing.T) {
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := cli.Run([]string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "not inside a Git repository") {
		t.Fatalf("error = %v", err)
	}
}
