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
