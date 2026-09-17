package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/git"
)

func TestDetectFindsRootAndCommonDirectory(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "git", "init", "-q")
	subdir := filepath.Join(dir, "nested")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	repo, err := git.Detect(subdir)
	if err != nil {
		t.Fatal(err)
	}
	if repo.Root != dir {
		t.Errorf("Root = %q, want %q", repo.Root, dir)
	}
	if _, err := os.Stat(repo.CommonDir); err != nil {
		t.Fatalf("CommonDir %q is not readable: %v", repo.CommonDir, err)
	}
	if filepath.Base(repo.CommonDir) != ".git" {
		t.Errorf("CommonDir = %q, want a .git directory", repo.CommonDir)
	}
}

func TestDetectExplainsOutsideRepository(t *testing.T) {
	_, err := git.Detect(t.TempDir())
	if err == nil || err.Error() != "not inside a Git repository; run this command from a Git working tree" {
		t.Fatalf("error = %v", err)
	}
}

func TestDetectExplainsMissingGit(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := git.Detect(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "Git is required") || !strings.Contains(err.Error(), "install Git") {
		t.Fatalf("error = %v", err)
	}
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: %v\n%s", name, err, out)
	}
}
