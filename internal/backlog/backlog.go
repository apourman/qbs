// Package backlog serves a project's .backlog/items as a small local web app.
// Each entry is .backlog/items/NNN-slug.md with simple `key: value`
// frontmatter and a Markdown body.
package backlog

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed index.html
var page []byte

// fields are written first, in this order; any other frontmatter keys follow.
var fields = []string{"title", "kind", "area", "status", "source", "created"}

var (
	frontmatter = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n?(.*)$`)
	idPattern   = regexp.MustCompile(`^[\w-]+$`)
	nonSlug     = regexp.MustCompile(`[^a-z0-9]+`)
)

// Item is one backlog entry: its frontmatter keys plus "id" and "body".
type Item map[string]string

func parse(id, text string) Item {
	item := Item{"body": strings.TrimSpace(text)}
	if m := frontmatter.FindStringSubmatch(text); m != nil {
		item["body"] = strings.TrimSpace(m[2])
		for _, line := range strings.Split(m[1], "\n") {
			if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) != "" {
				item[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}
	item["id"] = id
	item["version"] = version([]byte(text))
	return item
}

// version identifies a file's content, so a save made from a stale copy can
// be refused instead of overwriting newer edits.
func version(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}

func serialize(item Item) []byte {
	var extra []string
	for k := range item {
		if k != "id" && k != "body" && k != "version" && item[k] != "" && !slices.Contains(fields, k) {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	var b strings.Builder
	b.WriteString("---\n")
	for _, k := range append(append([]string{}, fields...), extra...) {
		b.WriteString(k + ": " + strings.ReplaceAll(item[k], "\n", " ") + "\n")
	}
	b.WriteString("---\n\n" + strings.TrimSpace(item["body"]) + "\n")
	return []byte(b.String())
}

// Handler serves the backlog UI and its JSON API for the items in dir.
func Handler(dir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	})
	// Items embed images as assets/<file>, which resolves here from the page.
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(filepath.Join(dir, "assets")))))
	// Specs live in the same checkout: <root>/.specs beside <root>/.backlog/items.
	specs := filepath.Join(dir, "..", "..", ".specs")
	mux.Handle("GET /specs/", http.StripPrefix("/specs/", http.FileServer(http.Dir(specs))))
	mux.HandleFunc("GET /api/specs", func(w http.ResponseWriter, r *http.Request) {
		files, _ := filepath.Glob(filepath.Join(specs, "*", "spec.md"))
		areas := []string{}
		for _, f := range files {
			areas = append(areas, filepath.Base(filepath.Dir(f)))
		}
		send(w, http.StatusOK, areas)
	})
	mux.HandleFunc("GET /api/items", func(w http.ResponseWriter, r *http.Request) {
		items, err := list(dir)
		if err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		send(w, http.StatusOK, items)
	})
	mux.HandleFunc("POST /api/items", func(w http.ResponseWriter, r *http.Request) {
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(item["title"]) == "" {
			fail(w, http.StatusBadRequest, "title required")
			return
		}
		item, err := Create(dir, item)
		if err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		send(w, http.StatusCreated, item)
	})
	mux.HandleFunc("PUT /api/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !idPattern.MatchString(id) {
			fail(w, http.StatusNotFound, "not found")
			return
		}
		file := filepath.Join(dir, id+".md")
		data, err := os.ReadFile(file)
		if os.IsNotExist(err) {
			fail(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := parse(id, string(data))
		var changes Item
		if err := json.NewDecoder(r.Body).Decode(&changes); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		if v := changes["version"]; v != "" && v != item["version"] {
			fail(w, http.StatusConflict, "this item changed since you opened it; copy your edits, reload, and save again")
			return
		}
		for k, v := range changes {
			item[k] = v
		}
		item["id"] = id
		data = serialize(item)
		if err := os.WriteFile(file, data, 0o644); err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		item["version"] = version(data)
		send(w, http.StatusOK, item)
	})
	return mux
}

func list(dir string) ([]Item, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	items := []Item{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		items = append(items, parse(strings.TrimSuffix(name, ".md"), string(data)))
	}
	return items, nil
}

// Create writes item as the next numbered file in dir and returns it with
// its id and version. Several processes may create at once: each claims its
// number with an exclusive .NNN.claim file, then checks no item already has
// that number, so two items never share one.
func Create(dir string, item Item) (Item, error) {
	if strings.TrimSpace(item["title"]) == "" {
		return nil, errors.New("title required")
	}
	if item["status"] == "" {
		item["status"] = "open"
	}
	if item["created"] == "" {
		item["created"] = time.Now().Format("2006-01-02")
	}
	n, err := highest(dir)
	if err != nil {
		return nil, err
	}
	for {
		n++
		num := fmt.Sprintf("%03d", n)
		claim := filepath.Join(dir, "."+num+".claim")
		f, err := os.OpenFile(claim, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		f.Close()
		taken, err := filepath.Glob(filepath.Join(dir, num+"-*.md"))
		if err == nil && len(taken) == 0 {
			item["id"] = num + "-" + slug(item["title"])
			data := serialize(item)
			err = os.WriteFile(filepath.Join(dir, item["id"]+".md"), data, 0o644)
			item["version"] = version(data)
		}
		os.Remove(claim)
		if err != nil {
			return nil, err
		}
		if len(taken) == 0 {
			return item, nil
		}
	}
}

func highest(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		digits := strings.SplitN(e.Name(), "-", 2)[0]
		if v, err := strconv.Atoi(digits); err == nil && v > n {
			n = v
		}
	}
	return n, nil
}

func slug(title string) string {
	s := nonSlug.ReplaceAllString(strings.ToLower(title), "-")
	if len(s) > 50 {
		s = s[:50]
	}
	if s = strings.Trim(s, "-"); s == "" {
		return "item"
	}
	return s
}

func send(w http.ResponseWriter, status int, body any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func fail(w http.ResponseWriter, status int, msg string) {
	send(w, status, map[string]string{"error": msg})
}
