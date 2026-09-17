package cli

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestOpenTaskInTmuxSplitsCurrentSession(t *testing.T) {
	calls := installFakeTmuxCommands(t, nil)
	t.Setenv("TMUX", "/tmp/tmux")

	if err := openTaskInTmux("/repo/task", &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"split-window", "-c", "/repo/task"}}
	if !reflect.DeepEqual(calls(), want) {
		t.Fatalf("tmux calls = %#v, want %#v", calls(), want)
	}
}

func TestOpenTaskInTmuxCreatesAndAttachesManagedSession(t *testing.T) {
	calls := installFakeTmuxCommands(t, func(call int, _ []string) error {
		if call == 1 {
			return missingSessionError()
		}
		return nil
	})
	t.Setenv("TMUX", "")

	if err := openTaskInTmux("/repo/task", &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"has-session", "-t", "qbs"},
		{"new-session", "-d", "-s", "qbs", "-c", "/repo/task"},
		{"attach-session", "-t", "qbs"},
	}
	if !reflect.DeepEqual(calls(), want) {
		t.Fatalf("tmux calls = %#v, want %#v", calls(), want)
	}
}

func TestOpenTaskInTmuxUsesNewWindowForExistingManagedSession(t *testing.T) {
	calls := installFakeTmuxCommands(t, nil)
	t.Setenv("TMUX", "")

	if err := openTaskInTmux("/repo/task", &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"has-session", "-t", "qbs"}, {"new-window", "-t", "qbs", "-c", "/repo/task"}, {"attach-session", "-t", "qbs"}}
	if !reflect.DeepEqual(calls(), want) {
		t.Fatalf("tmux calls = %#v, want %#v", calls(), want)
	}
}

func installFakeTmuxCommands(t *testing.T, result func(int, []string) error) func() [][]string {
	t.Helper()
	oldAvailable, oldCommand := tmuxAvailable, tmuxCommand
	t.Cleanup(func() { tmuxAvailable, tmuxCommand = oldAvailable, oldCommand })
	tmuxAvailable = func() error { return nil }
	var calls [][]string
	tmuxCommand = func(_ ioWriter, _ ioWriter, args ...string) error {
		calls = append(calls, args)
		if result != nil {
			return result(len(calls), args)
		}
		return nil
	}
	return func() [][]string { return calls }
}

func TestOpenTaskInTmuxReportsMissingTmux(t *testing.T) {
	oldAvailable := tmuxAvailable
	t.Cleanup(func() { tmuxAvailable = oldAvailable })
	tmuxAvailable = func() error { return errors.New("executable file not found") }

	err := openTaskInTmux("/repo/task", &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "install tmux") {
		t.Fatalf("error = %v", err)
	}
}

// ioWriter keeps the command seam's test signatures independent of concrete
// output buffer types while matching io.Writer.
type ioWriter = io.Writer

func missingSessionError() error {
	return exec.Command("sh", "-c", "exit 1").Run()
}
