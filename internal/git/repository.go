package git

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository describes the Git repository containing the current directory.
type Repository struct {
	Root      string
	CommonDir string
}

// Worktree describes one entry reported by Git's worktree list command.
type Worktree struct {
	Path   string
	Branch string
}

// Detect finds the repository root and common Git directory from dir. It
// works from both a normal worktree and a linked worktree.
func Detect(dir string) (Repository, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return Repository{}, errors.New("Git is required; install Git and ensure it is on PATH")
	}
	root, err := gitPath(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Repository{}, errors.New("not inside a Git repository; run this command from a Git working tree")
	}
	common, err := gitPath(dir, "rev-parse", "--git-common-dir")
	if err != nil {
		return Repository{}, fmt.Errorf("could not determine the Git directory: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return Repository{}, fmt.Errorf("could not resolve repository root: %w", err)
	}
	if !filepath.IsAbs(common) {
		// --git-common-dir is relative to the directory where Git was
		// invoked, which may differ from the repository root in a linked
		// worktree.
		common = filepath.Join(dir, common)
	}
	common, err = filepath.Abs(common)
	if err != nil {
		return Repository{}, fmt.Errorf("could not resolve Git directory: %w", err)
	}
	return Repository{Root: root, CommonDir: common}, nil
}

func gitPath(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// BranchExists reports whether branch is already a local branch.
func BranchExists(dir, branch string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = dir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// ValidateBranchName asks Git whether branch is a valid local branch name.
func ValidateBranchName(branch string) error {
	cmd := exec.Command("git", "check-ref-format", "--branch", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("invalid Git branch name %q: %s", branch, strings.TrimSpace(string(output)))
	}
	return nil
}

// TrackedPaths returns provisioned local-AI paths that are already in the
// repository index. Git excludes cannot make an indexed path untracked.
func TrackedPaths(dir string, paths ...string) ([]string, error) {
	args := append([]string{"ls-files", "--cached", "--"}, paths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	text := strings.TrimSpace(string(output))
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

// AddWorktree creates a new branch and linked worktree together.
func AddWorktree(dir, path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", "-b", branch, path)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// ListWorktrees returns all worktrees known to the repository, including the
// main worktree and linked worktrees.
func ListWorktrees(dir string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	var worktrees []Worktree
	var current *Worktree
	flush := func() {
		if current != nil {
			worktrees = append(worktrees, *current)
			current = nil
		}
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			current = &Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch refs/heads/") && current != nil:
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "":
			flush()
		}
	}
	flush()
	return worktrees, nil
}

// WorktreeStatus reports whether a worktree has modified tracked files or
// ordinary untracked files. Ignored files are intentionally excluded.
func WorktreeStatus(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=normal")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return len(strings.TrimSpace(string(output))) != 0, nil
}

// RemoveWorktree removes a worktree, including ignored files when force is
// true. The caller is responsible for checking any safety policy first.
func RemoveWorktree(dir string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, dir)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// BranchIsMerged reports whether branch is an ancestor of the current branch,
// which is the condition Git uses for a normal branch deletion without force.
func BranchIsMerged(dir, branch string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", branch, "HEAD")
	cmd.Dir = dir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("git merge-base: %w", err)
}

// DeleteBranch deletes a task branch after its worktree is removed.
func DeleteBranch(dir, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	cmd := exec.Command("git", "branch", flag, branch)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch delete: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
