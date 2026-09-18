package git

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

// OriginURL returns the configured origin URL, or an empty string when the
// repository has no origin. Missing remotes are valid for local repositories.
func OriginURL(dir string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return "", nil
	}
	return "", fmt.Errorf("git config remote.origin.url: %w", err)
}

// IsGitHubURL reports whether remote identifies github.com, including SCP-like
// Git URLs such as git@github.com:owner/repository.git.
func IsGitHubURL(remote string) bool {
	remote = strings.TrimSpace(remote)
	if strings.HasPrefix(remote, "git@github.com:") {
		return true
	}
	parsed, err := url.Parse(remote)
	return err == nil && strings.EqualFold(parsed.Hostname(), "github.com")
}

// Repository describes the Git repository containing the current directory.
type Repository struct {
	Root      string
	CommonDir string
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
