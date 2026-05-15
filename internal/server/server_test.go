package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"vault-reader/internal/indexer"
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

// createIndexedVault creates a vault with an indexer for testing API endpoints.
func createIndexedVault(t *testing.T) (*Server, string) {
	t.Helper()
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, "notes"), 0o755)
	os.WriteFile(filepath.Join(vaultDir, "notes", "tagged.md"),
		[]byte("---\ntags: [debug/proxy, openclaw]\n---\n# Tagged Note\nContent here.\n"), 0o644)
	os.WriteFile(filepath.Join(vaultDir, "notes", "plain.md"),
		[]byte("# Plain Note\nNo tags.\n"), 0o644)

	dbPath := filepath.Join(vaultDir, ".test-data", "test.db")
	ix, err := indexer.New(dbPath, vaultDir)
	if err != nil {
		t.Fatalf("create indexer: %v", err)
	}
	t.Cleanup(func() { ix.Close() })

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	srv := New(vaultDir, WithIndexer(ix))
	return srv, vaultDir
}

func TestTagTreeEndpoint(t *testing.T) {
	srv, _ := createIndexedVault(t)

	req := httptest.NewRequest("GET", "/api/tag-tree", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatal("expected 'items' array in response")
	}
	if len(items) == 0 {
		t.Error("expected at least one tag tree node")
	}
}

func TestTagTreeEndpointWithoutIndexer(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir) // no indexer

	req := httptest.NewRequest("GET", "/api/tag-tree", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestCanvasEndpoint(t *testing.T) {
	vaultDir := t.TempDir()
	canvasJSON := `{"nodes":[{"id":"n1","type":"text","text":"Hello","x":0,"y":0,"width":300,"height":200}],"edges":[]}`
	os.WriteFile(filepath.Join(vaultDir, "test.canvas"), []byte(canvasJSON), 0o644)

	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/canvas?path=test.canvas", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]interface{})
	if !ok || len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %v", result["nodes"])
	}
}

func TestCanvasEndpointPathTraversal(t *testing.T) {
	vaultDir := t.TempDir()
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/canvas?path=../../etc/passwd", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestCanvasEndpointNotFound(t *testing.T) {
	vaultDir := t.TempDir()
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/canvas?path=missing.canvas", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGraphEndpoint(t *testing.T) {
	srv, _ := createIndexedVault(t)

	req := httptest.NewRequest("GET", "/api/graph", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	nodes, ok := result["nodes"].([]interface{})
	if !ok {
		t.Fatal("expected 'nodes' array")
	}
	if len(nodes) == 0 {
		t.Error("expected at least 1 node")
	}
}

func TestGraphEndpointWithFolder(t *testing.T) {
	srv, _ := createIndexedVault(t)

	req := httptest.NewRequest("GET", "/api/graph?folder=notes", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGraphEndpointWithoutIndexer(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/graph", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

// ==================== Dashboard Tests ====================

func createDashboardVault(t *testing.T) (*Server, string) {
	t.Helper()
	vaultDir := t.TempDir()

	dirs := []string{
		"00_Inbox",
		"10_Reference",
		"20_Debug",
		"30_Dashboard",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(vaultDir, d), 0o755)
	}

	files := map[string]string{
		"00_Inbox/未整理笔记.md":        "---\ntags: [inbox]\n---\n# 未整理笔记\n内容",
		"10_Reference/参考资料.md":       "---\ntags: [ref]\nstatus: active\n---\n# 参考资料\n内容",
		"20_Debug/bug1.md":            "---\ntags: [debug]\nstatus: active\ntype: debug-note\n---\n# Bug 1\n内容",
		"20_Debug/bug2.md":            "---\ntags: [debug]\ntype: debug-note\n---\n# Bug 2\n内容",
		"30_Dashboard/首页.md":         "# 首页\n欢迎",
	}
	for path, content := range files {
		fullPath := filepath.Join(vaultDir, filepath.FromSlash(path))
		os.MkdirAll(filepath.Dir(fullPath), 0o755)
		os.WriteFile(fullPath, []byte(content), 0o644)
	}

	// Add a canvas file
	canvasJSON := `{"nodes":[{"id":"n1","type":"text","text":"Hello","x":0,"y":0,"width":300,"height":200}],"edges":[]}`
	os.WriteFile(filepath.Join(vaultDir, "30_Dashboard", "知识地图.canvas"), []byte(canvasJSON), 0o644)

	dbPath := filepath.Join(vaultDir, ".test-data", "test.db")
	ix, err := indexer.New(dbPath, vaultDir)
	if err != nil {
		t.Fatalf("create indexer: %v", err)
	}
	t.Cleanup(func() { ix.Close() })

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	srv := New(vaultDir, WithIndexer(ix))
	return srv, vaultDir
}

func TestDashboardEndpoint(t *testing.T) {
	srv, _ := createDashboardVault(t)

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check all expected sections exist
	for _, key := range []string{"recent", "inbox", "active", "debug", "tags", "canvas"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected '%s' in dashboard response", key)
		}
	}

	// Inbox should have 1 file
	inbox, _ := result["inbox"].([]interface{})
	if len(inbox) < 1 {
		t.Errorf("expected at least 1 inbox item, got %d", len(inbox))
	}

	// Active should have 2 files (bug1 + 参考资料)
	active, _ := result["active"].([]interface{})
	if len(active) < 1 {
		t.Errorf("expected at least 1 active item, got %d", len(active))
	}

	// Debug should have 2 files
	debug, _ := result["debug"].([]interface{})
	if len(debug) < 1 {
		t.Errorf("expected at least 1 debug item, got %d", len(debug))
	}

	// Tags should exist
	tags, _ := result["tags"].([]interface{})
	if len(tags) < 1 {
		t.Errorf("expected at least 1 tag, got %d", len(tags))
	}
}

func TestDashboardEndpointWithoutIndexer(t *testing.T) {
	vaultDir := createMinimalVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
