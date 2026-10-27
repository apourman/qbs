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
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/assets/../012-old.md", nil))
	if rec.Code == http.StatusOK {
		t.Fatal("asset route escaped the assets directory")
	}
}
