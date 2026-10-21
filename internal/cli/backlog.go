package cli

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/trues/qbs/internal/backlog"
	"github.com/trues/qbs/internal/git"
)

const defaultBacklogPort = 4317

// serveBacklog serves the primary checkout's .backlog, so linked worktrees
// share one backlog with it.
func serveBacklog(dir string, port int, stdout io.Writer) error {
	repo, err := git.Detect(dir)
	if err != nil {
		return err
	}
	root := repo.Root
	if filepath.Base(repo.CommonDir) == ".git" {
		root = filepath.Dir(repo.CommonDir)
	}
	if info, err := os.Stat(filepath.Join(root, ".backlog")); err != nil || !info.IsDir() {
		return fmt.Errorf("no backlog in %s; run qbs init first", root)
	}
	items := filepath.Join(root, ".backlog", "items")
	if err := os.MkdirAll(items, 0o755); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "Backlog: http://localhost:%d (%s)\n", port, items); err != nil {
		return err
	}
	return http.Serve(listener, backlog.Handler(items))
}
