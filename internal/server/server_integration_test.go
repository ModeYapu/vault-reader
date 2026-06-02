package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func createFullTestVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create directories
	dirs := []string{
		"00_Inbox",
		"10_Reference",
		"20_Debug",
		"30_Dashboard",
		"90_Templates",
		"attachments",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(dir, d), 0o755)
	}

	// Create markdown files with rich content
	files := map[string]string{
		"00_Inbox/临时记录.md": `---
title: 临时记录
tags:
  - inbox
---
# 临时记录

这是一条临时记录。

参考 [[OpenClaw]]
查看 [[Docker|Docker 排查]]

## 代理配置

HTTP_PROXY 和 HTTPS_PROXY 设置。

![[attachments/架构图.png]]
`,
		"10_Reference/官方文档.md": "# 官方文档\n\n参考内容 [[OpenClaw]]",
		"20_Debug/OpenClaw.md":   "# OpenClaw\n\n这是 OpenClaw 的内容。\n\n[[Codex]]",
		"20_Debug/Codex.md":      "# Codex\n\n[[OpenClaw]] [[Docker|Docker 链接]]",
		"20_Debug/Docker.md":     "# Docker\n\nDocker 内容",
		"30_Dashboard/常用系统入口.md": "# 常用系统入口\n\n- [[OpenClaw]]\n- [[Codex]]",
		"90_Templates/Web Clip 模板.md": "# Web Clip 模板\n\n模板内容",
	}
	for path, content := range files {
		fullPath := filepath.Join(dir, filepath.FromSlash(path))
		os.MkdirAll(filepath.Dir(fullPath), 0o755)
		os.WriteFile(fullPath, []byte(content), 0o644)
	}

	// Create fake asset
	os.WriteFile(filepath.Join(dir, "attachments", "架构图.png"), []byte("fake-png-data"), 0o644)

	return dir
}

func TestIntegrationFullVaultTree(t *testing.T) {
	vaultDir := createFullTestVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/tree", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	children := result["children"].([]interface{})
	// Should have: 00_Inbox, 10_Reference, 20_Debug, 30_Dashboard, 90_Templates, attachments
	if len(children) < 5 {
		t.Errorf("expected at least 5 top-level items, got %d", len(children))
	}
}

func TestIntegrationNoteWithFrontmatter(t *testing.T) {
	vaultDir := createFullTestVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/note?path=00_Inbox/临时记录.md", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	// Title from frontmatter
	if result["title"] != "临时记录" {
		t.Errorf("expected title '临时记录', got %v", result["title"])
	}

	// Should have frontmatter
	fm, ok := result["frontmatter"].(map[string]interface{})
	if !ok {
		t.Fatal("expected frontmatter object")
	}
	if fm["title"] != "临时记录" {
		t.Errorf("frontmatter title = %v", fm["title"])
	}

	// Should have tags
	tags, ok := result["tags"].([]interface{})
	if !ok {
		t.Fatal("expected tags array")
	}
	if len(tags) == 0 {
		t.Error("expected at least one tag")
	}

	// Should have links
	links, ok := result["links"].([]interface{})
	if !ok {
		t.Fatal("expected links array")
	}
	if len(links) < 2 {
		t.Errorf("expected at least 2 links, got %d", len(links))
	}

	// Should have HTML
	html, ok := result["html"].(string)
	if !ok || len(html) == 0 {
		t.Error("expected non-empty HTML")
	}
}

func TestIntegrationAssetEndpoint(t *testing.T) {
	vaultDir := createFullTestVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/assets?path=attachments/架构图.png", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("expected image/png content type, got %s", ct)
	}
}

func TestIntegrationAssetPathTraversal(t *testing.T) {
	vaultDir := createFullTestVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/assets?path=../../etc/passwd", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestIntegrationAssetNotFound(t *testing.T) {
	vaultDir := createFullTestVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/assets?path=attachments/nonexistent.png", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestIntegrationChinesePathNote(t *testing.T) {
	vaultDir := createFullTestVault(t)
	srv := New(vaultDir)

	req := httptest.NewRequest("GET", "/api/note?path=00_Inbox/临时记录.md", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
