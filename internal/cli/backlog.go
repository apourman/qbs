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

// serveBacklog serves the primary checkout's .backlog.
func serveBacklog(dir string, port int, stdout io.Writer) error {
	items, err := backlogItems(dir)
	if err != nil {
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

// newBacklogItem creates an item stub with the next free number and prints
// its path, so agents claim numbers without racing each other.
func newBacklogItem(dir, title string, stdout io.Writer) error {
	items, err := backlogItems(dir)
	if err != nil {
		return err
	}
	item, err := backlog.Create(items, backlog.Item{"title": title})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, filepath.Join(items, item["id"]+".md"))
	return err
}

// backlogItems returns the primary checkout's .backlog/items, so linked
// worktrees share one backlog with it.
func backlogItems(dir string) (string, error) {
	repo, err := git.Detect(dir)
	if err != nil {
		return "", err
	}
	root := repo.Root
	if filepath.Base(repo.CommonDir) == ".git" {
		root = filepath.Dir(repo.CommonDir)
	}
	if info, err := os.Stat(filepath.Join(root, ".backlog")); err != nil || !info.IsDir() {
		return "", fmt.Errorf("no backlog in %s; run qbs init first", root)
	}
	items := filepath.Join(root, ".backlog", "items")
	return items, os.MkdirAll(items, 0o755)
}
