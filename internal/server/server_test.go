package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func createMinimalVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "notes"), 0o755)
	os.WriteFile(filepath.Join(dir, "notes", "hello.md"), []byte("# Hello\nWorld"), 0o644)
	return dir
}

func TestTreeEndpoint(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/tree", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	items, ok := result["children"].([]interface{})
	if !ok {
		t.Fatal("expected 'children' array in response")
	}
	if len(items) == 0 {
		t.Error("expected at least one item in tree")
	}
}

func TestStaticHTML(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty HTML response")
	}
}

func TestNoteEndpoint(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/note?path=notes/hello.md", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", w.Body.String())
	}

	if result["title"] != "Hello" {
		t.Errorf("expected title 'Hello', got %v", result["title"])
	}
	html, ok := result["html"].(string)
	if !ok || len(html) == 0 {
		t.Error("expected non-empty html field")
	}
}

func TestNoteEndpointNotFound(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/note?path=nonexistent.md", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestNoteEndpointPathTraversal(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/note?path=../../../etc/passwd", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
