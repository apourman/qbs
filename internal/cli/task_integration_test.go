package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/cli"
)

func TestTaskCreatesProvisionedCleanWorktree(t *testing.T) {
	repo := newInitializedRepo(t)

	var out, errOut bytes.Buffer
	if err := runInDirectory(repo, []string{"task", "planning"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}

	worktree := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-planning")
	if got := runGit(t, worktree, "branch", "--show-current"); strings.TrimSpace(got) != "qbs/planning" {
		t.Fatalf("branch = %q", got)
	}
	for _, name := range []string{
		"AGENTS.md", "CLAUDE.md",
		filepath.Join(".codex", "agents", "implementer.toml"),
		filepath.Join(".claude", "agents", "implementer.md"),
		filepath.Join(".opencode", "agents", "implementer.md"),
		filepath.Join(".agents", "skills", "qbs", "SKILL.md"),
		filepath.Join(".claude", "skills", "qbs", "SKILL.md"),
		filepath.Join(".opencode", "skills", "qbs", "SKILL.md"),
		".research", ".specs",
	} {
		if _, err := os.Stat(filepath.Join(worktree, name)); err != nil {
			t.Errorf("task did not create %s: %v", name, err)
		}
	}
	if status := runGit(t, worktree, "status", "--short"); status != "" {
		t.Fatalf("task worktree is dirty: %q", status)
	}
	if !strings.Contains(out.String(), worktree) || !strings.Contains(out.String(), "qbs/planning") {
		t.Fatalf("output does not identify task branch and path: %q", out.String())
	}
}

func TestTaskCreatesIndependentMultipleWorktrees(t *testing.T) {
	repo := newInitializedRepo(t)
	if err := runInDirectory(repo, []string{"task", "one"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := runInDirectory(repo, []string{"task", "two"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	one := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-one")
	two := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-two")
	if one == two {
		t.Fatal("task worktree paths are not distinct")
	}
	if err := os.WriteFile(filepath.Join(one, ".research", "notes.md"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(two, ".research", "notes.md")); !os.IsNotExist(err) {
		t.Fatalf("research storage is shared: err = %v", err)
	}
}

func TestTaskRejectsInvalidNameBeforeRepositoryChanges(t *testing.T) {
	repo := newInitializedRepo(t)
	before := runGit(t, repo, "worktree", "list", "--porcelain")
	for _, name := range []string{"", ".", "..", "bad/name", "bad name", "-bad", "bad~name", "foo..bar", "trailing."} {
		if err := runInDirectory(repo, []string{"task", name}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
			t.Errorf("task %q unexpectedly succeeded", name)
		}
	}
	if after := runGit(t, repo, "worktree", "list", "--porcelain"); after != before {
		t.Fatal("invalid task names changed Git worktrees")
	}
}

func TestTaskRejectsGitInvalidNameBeforeChangingExcludes(t *testing.T) {
	installFakeTmux(t)
	repo := newCommittedRepo(t)
	exclude := filepath.Join(repo, ".git", "info", "exclude")
	before, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}

	err = runInDirectory(repo, []string{"task", "foo..bar"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "invalid task name") {
		t.Fatalf("error = %v", err)
	}
	after, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("invalid task name changed shared excludes")
	}
}

func TestTaskRejectsTrackedAIFileBeforeChangingExcludes(t *testing.T) {
	installFakeTmux(t)
	repo := newCommittedRepo(t)
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte("tracked instructions\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "AGENTS.md")
	runGit(t, repo, "commit", "-qm", "track agent instructions")
	exclude := filepath.Join(repo, ".git", "info", "exclude")
	before, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}

	err = runInDirectory(repo, []string{"task", "tracked"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "tracked") || !strings.Contains(err.Error(), "AGENTS.md") {
		t.Fatalf("error = %v", err)
	}
	after, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("task changed excludes despite tracked AI file")
	}
	if branches := runGit(t, repo, "branch", "--list", "qbs/tracked"); strings.TrimSpace(branches) != "" {
		t.Fatalf("task created branch despite tracked AI file: %q", branches)
	}
}

func TestTaskReportsBranchAndPathCollisions(t *testing.T) {
	repo := newInitializedRepo(t)
	if err := runInDirectory(repo, []string{"task", "same"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := runInDirectory(repo, []string{"task", "same"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("collision error = %v", err)
	}
	pathCollision := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-other")
	if err := os.Mkdir(pathCollision, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := runInDirectory(repo, []string{"task", "other"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "worktree path") {
		t.Fatalf("path collision error = %v", err)
	}
}

func TestTasksListsQBSTaskBranchesAndLocationsFromMainAndLinkedWorktree(t *testing.T) {
	repo := newInitializedRepo(t)
	if err := runInDirectory(repo, []string{"task", "planning"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := runInDirectory(repo, []string{"task", "review"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	wantPlanning := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-planning")
	wantReview := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-review")
	for _, dir := range []string{repo, wantPlanning} {
		var out bytes.Buffer
		if err := runInDirectory(dir, []string{"tasks"}, &out, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"qbs/planning", wantPlanning, "qbs/review", wantReview} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("tasks from %s missing %q: %q", dir, want, out.String())
			}
		}
	}
}

func TestTaskRemoveDeletesCleanWorktreeAndBranchWithIgnoredFilesWarning(t *testing.T) {
	repo := newInitializedRepo(t)
	if err := runInDirectory(repo, []string{"task", "cleanup"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-cleanup")
	if err := os.WriteFile(filepath.Join(worktree, ".research", "notes.md"), []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := runInDirectory(repo, []string{"task", "remove", "cleanup"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(worktree); !os.IsNotExist(err) {
		t.Fatalf("worktree still exists: %v", err)
	}
	if strings.Contains(runGit(t, repo, "branch", "--list", "qbs/cleanup"), "qbs/cleanup") {
		t.Fatal("task branch still exists")
	}
	if !strings.Contains(errOut.String(), "ignored") || !strings.Contains(errOut.String(), "delete") {
		t.Fatalf("removal warning = %q", errOut.String())
	}
}

func TestTaskRemoveProtectsModifiedTrackedFilesUnlessForced(t *testing.T) {
	repo := newInitializedRepo(t)
	if err := runInDirectory(repo, []string{"task", "important"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-important")
	if err := os.WriteFile(filepath.Join(worktree, "README.md"), []byte("work in progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := runInDirectory(repo, []string{"task", "remove", "important"}, &out, &errOut); err == nil || !strings.Contains(err.Error(), "modified tracked files") {
		t.Fatalf("unforced removal error = %v", err)
	}
	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("unforced removal deleted worktree: %v", err)
	}

	errOut.Reset()
	if err := runInDirectory(repo, []string{"task", "remove", "important", "--force"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(worktree); !os.IsNotExist(err) {
		t.Fatalf("forced removal left worktree: %v", err)
	}
}

func TestTaskRemoveProtectsOrdinaryUntrackedFilesUnlessForced(t *testing.T) {
	repo := newInitializedRepo(t)
	if err := runInDirectory(repo, []string{"task", "untracked"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-untracked")
	untracked := filepath.Join(worktree, "notes.txt")
	if err := os.WriteFile(untracked, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := runInDirectory(repo, []string{"task", "remove", "untracked"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "untracked files") {
		t.Fatalf("unforced removal error = %v", err)
	}
	if _, statErr := os.Stat(worktree); statErr != nil {
		t.Fatalf("unforced removal deleted worktree: %v", statErr)
	}

	if err := runInDirectory(repo, []string{"task", "remove", "untracked", "--force"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(worktree); !os.IsNotExist(statErr) {
		t.Fatalf("forced removal left worktree: %v", statErr)
	}
}

func TestTaskRemoveRequiresForceForUnmergedTaskCommit(t *testing.T) {
	repo := newInitializedRepo(t)
	if err := runInDirectory(repo, []string{"task", "unmerged"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-unmerged")
	if err := os.WriteFile(filepath.Join(worktree, "task.txt"), []byte("task\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, worktree, "add", "task.txt")
	runGit(t, worktree, "commit", "-qm", "task commit")

	var out, errOut bytes.Buffer
	err := runInDirectory(repo, []string{"task", "remove", "unmerged"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("unforced removal error = %v", err)
	}
	if _, statErr := os.Stat(worktree); statErr != nil {
		t.Fatalf("unforced removal deleted worktree: %v", statErr)
	}

	if err := runInDirectory(repo, []string{"task", "remove", "unmerged", "--force"}, &out, &errOut); err != nil {
		t.Fatal(err)
	}
}

func newInitializedRepo(t *testing.T) string {
	t.Helper()
	installFakeTmux(t)
	repo := newCommittedRepo(t)
	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	return repo
}

func newCommittedRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "qbs-tests@example.invalid")
	runGit(t, repo, "config", "user.name", "QBS Tests")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-qm", "fixture")
	return repo
}

func installFakeTmux(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	name := "tmux"
	contents := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		name = "tmux.cmd"
		contents = "@exit /b 0\r\n"
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func runInDirectory(dir string, args []string, stdout, stderr *bytes.Buffer) error {
	old, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer os.Chdir(old)
	return cli.Run(args, stdout, stderr)
}
