package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRunReportsWorkingDirectoryFailure(t *testing.T) {
	oldGetwd := getwd
	t.Cleanup(func() { getwd = oldGetwd })
	getwd = func() (string, error) { return "", errors.New("working directory unavailable") }

	err := Run([]string{"init"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "working directory unavailable") {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "not inside a Git repository") {
		t.Fatalf("working directory failure was converted to repository error: %v", err)
	}
}
