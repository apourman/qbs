package backlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func do(t *testing.T, h http.Handler, method, path, body string) (int, Item) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	var item Item
	_ = json.Unmarshal(rec.Body.Bytes(), &item)
	return rec.Code, item
}

func TestCreateAndUpdateKeepItemFilesCompatible(t *testing.T) {
	dir := t.TempDir()
	existing := "---\ntitle: Old\nkind: bug\narea: chat\nstatus: open\nsource: play-test\ncreated: 2026-10-21\nticket: .specs/chat/tickets/01-links.md\n---\n\nNotes.\n"
	if err := os.WriteFile(filepath.Join(dir, "012-old.md"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	h := Handler(dir)

	code, item := do(t, h, "POST", "/api/items", `{"title":"Chat fades: instantly!","kind":"polish","body":"Like retail."}`)
	if code != http.StatusCreated || item["id"] != "013-chat-fades-instantly" || item["status"] != "open" {
		t.Fatalf("POST = %d %v", code, item)
	}

	code, item = do(t, h, "PUT", "/api/items/012-old", `{"status":"done"}`)
	if code != http.StatusOK || item["status"] != "done" {
		t.Fatalf("PUT = %d %v", code, item)
	}
	data, err := os.ReadFile(filepath.Join(dir, "012-old.md"))
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Replace(existing, "status: open", "status: done", 1); string(data) != want {
		t.Errorf("updated file =\n%s\nwant\n%s", data, want)
	}

	stale := item["version"]
	if code, _ := do(t, h, "PUT", "/api/items/012-old", `{"status":"open","version":"`+stale+`"}`); code != http.StatusOK {
		t.Fatalf("PUT with current version = %d", code)
	}
	if code, _ := do(t, h, "PUT", "/api/items/012-old", `{"status":"dropped","version":"`+stale+`"}`); code != http.StatusConflict {
		t.Errorf("PUT with stale version = %d, want 409", code)
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "012-old.md")); string(data) != existing {
		t.Errorf("stale save changed the file:\n%s", data)
	}

	if code, _ := do(t, h, "PUT", "/api/items/999-missing", `{}`); code != http.StatusNotFound {
		t.Errorf("PUT missing = %d", code)
	}
	if code, _ := do(t, h, "POST", "/api/items", `{"title":" "}`); code != http.StatusBadRequest {
		t.Errorf("POST without title = %d", code)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/items", nil))
	var items []Item
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil || len(items) != 2 {
		t.Fatalf("GET = %s (%v)", rec.Body, err)
	}
}

func TestServesEmbeddedImages(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "012-1.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := Handler(dir)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/assets/012-1.png", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "png" {
		t.Fatalf("GET asset = %d %q", rec.Code, rec.Body.String())
	}
}

func TestConcurrentCreatesGetDistinctNumbers(t *testing.T) {
	dir := t.TempDir()
	const n = 20
	ids := make(chan string, n)
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			item, err := Create(dir, Item{"title": "Same title"})
			ids <- item["id"]
			errs <- err
		}()
	}
	seen := map[string]bool{}
	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
		num, _, _ := strings.Cut(<-ids, "-")
		if seen[num] {
			t.Fatalf("number %s used twice", num)
		}
		seen[num] = true
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, ".*")); len(matches) != 0 {
		t.Errorf("claims left behind: %v", matches)
	}
}

func TestListsAndServesAreaSpecs(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".backlog", "items")
	for _, p := range []string{dir, filepath.Join(root, ".specs", "chat"), filepath.Join(root, ".specs", "notes")} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".specs", "chat", "spec.md"), []byte("# Chat"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := Handler(dir)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/specs", nil))
	if got := strings.TrimSpace(rec.Body.String()); got != `["chat"]` {
		t.Errorf("GET /api/specs = %s", got)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/specs/chat/spec.md", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "# Chat" {
		t.Errorf("GET spec = %d %q", rec.Code, rec.Body.String())
	}
}

func TestEmptyOptionalFieldsAreNotWritten(t *testing.T) {
	dir := t.TempDir()
	item, err := Create(dir, Item{"title": "A", "blocked-by": ""})
	if err != nil {
		t.Fatal(err)
	}
	h := Handler(dir)
	if code, _ := do(t, h, "PUT", "/api/items/"+item["id"], `{"blocked-by":"001, 002"}`); code != http.StatusOK {
		t.Fatalf("PUT = %d", code)
	}
	data, _ := os.ReadFile(filepath.Join(dir, item["id"]+".md"))
	if !strings.Contains(string(data), "\nblocked-by: 001, 002\n") {
		t.Errorf("file =\n%s", data)
	}
	do(t, h, "PUT", "/api/items/"+item["id"], `{"blocked-by":""}`)
	if data, _ := os.ReadFile(filepath.Join(dir, item["id"]+".md")); strings.Contains(string(data), "blocked-by") {
		t.Errorf("empty blocked-by written:\n%s", data)
	}
}

func TestServesOnlySpecFiles(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".backlog", "items")
	if err := os.MkdirAll(filepath.Join(root, ".specs", "chat"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".specs", "chat", "notes.md"), []byte("private"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := Handler(dir)
	for _, path := range []string{"/specs/chat/notes.md", "/specs/", "/specs/chat/"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code == http.StatusOK {
			t.Errorf("GET %s = 200 %q, want refused", path, rec.Body.String())
		}
	}
}
