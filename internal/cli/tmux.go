package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const qbsTmuxSession = "qbs"

// tmuxCommand is a variable so tests can replace process execution without
// needing an interactive tmux server.
var tmuxCommand = runTmuxCommand
var tmuxAvailable = func() error {
	_, err := exec.LookPath("tmux")
	return err
}

func openTaskInTmux(worktree string, stdout, stderr io.Writer) error {
	if err := requireTmux(); err != nil {
		return err
	}

	if os.Getenv("TMUX") != "" {
		if err := tmuxCommand(stdout, stderr, "split-window", "-c", worktree); err != nil {
			return fmt.Errorf("open task workspace in a tmux pane: %w", err)
		}
		return nil
	}

	// A detached session is created first so that the task's directory is
	// correct even when this is the first task in a new QBS session. When the
	// session already exists, a new window is needed because -c on
	// new-session is ignored for an existing session.
	if err := tmuxCommand(io.Discard, stderr, "has-session", "-t", qbsTmuxSession); err != nil {
		if !isMissingTmuxSession(err) {
			return fmt.Errorf("check QBS tmux session: %w", err)
		}
		if err := tmuxCommand(stdout, stderr, "new-session", "-d", "-s", qbsTmuxSession, "-c", worktree); err != nil {
			return fmt.Errorf("create QBS tmux session: %w", err)
		}
	} else if err := tmuxCommand(stdout, stderr, "new-window", "-t", qbsTmuxSession, "-c", worktree); err != nil {
		return fmt.Errorf("open task workspace in QBS tmux session: %w", err)
	}

	if err := tmuxCommand(stdout, stderr, "attach-session", "-t", qbsTmuxSession); err != nil {
		return fmt.Errorf("attach to QBS tmux session: %w", err)
	}
	return nil
}

func requireTmux() error {
	if err := tmuxAvailable(); err != nil {
		return errors.New("tmux is required to open the task workspace; install tmux and ensure it is on PATH")
	}
	return nil
}

func isMissingTmuxSession(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 1
}

func runTmuxCommand(stdout, stderr io.Writer, args ...string) error {
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	var diagnostics bytes.Buffer
	cmd.Stderr = io.MultiWriter(stderr, &diagnostics)
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(diagnostics.String())
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return err
	}
	return nil
}
