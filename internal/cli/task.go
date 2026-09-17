package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/trues/qbs/internal/git"
)

var validTaskName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func createTask(dir, name string, stdout, stderr io.Writer) error {
	if err := validateTaskName(name); err != nil {
		return err
	}
	repo, err := git.Detect(dir)
	if err != nil {
		return err
	}
	task, err := newTaskIdentity(repo.Root, name)
	if err != nil {
		return err
	}
	if err := requireTmux(); err != nil {
		return err
	}

	exists, err := git.BranchExists(repo.Root, task.branch)
	if err != nil {
		return fmt.Errorf("check task branch %q: %w", task.branch, err)
	}
	if exists {
		return fmt.Errorf("task branch %q already exists", task.branch)
	}
	if _, err := os.Lstat(task.worktree); err == nil {
		return fmt.Errorf("task worktree path %q already exists", task.worktree)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check task worktree path %q: %w", task.worktree, err)
	}
	if err := prepareProvisioning(repo, task.worktree); err != nil {
		return fmt.Errorf("cannot provision task workspace: %w", err)
	}

	if err := git.AddWorktree(repo.Root, task.worktree, task.branch); err != nil {
		return fmt.Errorf("create task worktree %q: %w", task.worktree, err)
	}
	if err := provisionWorkspace(task.worktree, stderr); err != nil {
		return partialTaskError(task, "provision local AI workspace", err)
	}
	if err := openTaskInTmux(task.worktree, stdout, stderr); err != nil {
		return partialTaskError(task, "open tmux workspace", err)
	}
	_, err = fmt.Fprintf(stdout, "Created task %s at %s\n", task.branch, task.worktree)
	return err
}

type taskIdentity struct {
	name     string
	branch   string
	worktree string
}

func newTaskIdentity(repoRoot, name string) (taskIdentity, error) {
	branch := "qbs/" + name
	if err := git.ValidateBranchName(branch); err != nil {
		return taskIdentity{}, fmt.Errorf("invalid task name %q: %w", name, err)
	}
	return taskIdentity{
		name:     name,
		branch:   branch,
		worktree: filepath.Join(filepath.Dir(repoRoot), filepath.Base(repoRoot)+"-"+name),
	}, nil
}

func partialTaskError(task taskIdentity, operation string, err error) error {
	return fmt.Errorf("%s for task %q: %w; partial setup remains: worktree %s and branch %s exist; remove the task with `qbs task remove %s --force` when safe", operation, task.branch, err, task.worktree, task.branch, task.name)
}

func validateTaskName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid task name: name must not be empty")
	}
	if name == "." || name == ".." || !validTaskName.MatchString(name) {
		return fmt.Errorf("invalid task name %q: use letters, numbers, dots, underscores, and hyphens; start with a letter or number", name)
	}
	if strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("invalid task name %q: names ending in .lock are not allowed", name)
	}
	return nil
}

func listTasks(dir string, stdout io.Writer) error {
	repo, err := git.Detect(dir)
	if err != nil {
		return err
	}
	worktrees, err := git.ListWorktrees(repo.Root)
	if err != nil {
		return err
	}
	for _, worktree := range worktrees {
		if strings.HasPrefix(worktree.Branch, "qbs/") {
			if _, err := fmt.Fprintf(stdout, "%s\t%s\n", worktree.Branch, worktree.Path); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeTask(dir, name string, force bool, stdout, stderr io.Writer) error {
	if err := validateTaskName(name); err != nil {
		return err
	}
	repo, err := git.Detect(dir)
	if err != nil {
		return err
	}
	branch := "qbs/" + name
	worktrees, err := git.ListWorktrees(repo.Root)
	if err != nil {
		return err
	}
	var worktreePath string
	for _, worktree := range worktrees {
		if worktree.Branch == branch {
			worktreePath = worktree.Path
			break
		}
	}
	if worktreePath == "" {
		return fmt.Errorf("task %q does not exist", name)
	}
	modified, err := git.WorktreeStatus(worktreePath)
	if err != nil {
		return fmt.Errorf("inspect task %q: %w", name, err)
	}
	if modified && !force {
		return fmt.Errorf("task %q has modified tracked files or untracked files; use --force to remove it", name)
	}
	merged, err := git.BranchIsMerged(repo.Root, branch)
	if err != nil {
		return fmt.Errorf("inspect task branch %q: %w", branch, err)
	}
	if !merged && !force {
		return fmt.Errorf("task %q has unmerged commits; use --force to remove it", name)
	}
	if _, err := fmt.Fprintf(stderr, "warning: removing %s will delete ignored local AI files, specs, and research notes in %s\n", branch, worktreePath); err != nil {
		return err
	}
	if err := git.RemoveWorktree(worktreePath, true); err != nil {
		return fmt.Errorf("remove task %q: %w; worktree remains at %s and branch %s was not deleted", name, err, worktreePath, branch)
	}
	if err := git.DeleteBranch(repo.Root, branch, force); err != nil {
		return fmt.Errorf("remove task branch %q: %w; worktree was removed, but branch %s remains", branch, err, branch)
	}
	_, err = fmt.Fprintf(stdout, "Removed task %s\n", branch)
	return err
}
