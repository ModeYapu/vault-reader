package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func createTestVault(t *testing.T) string {
	t.Helper()
	vaultDir := t.TempDir()

	dirs := []string{
		"00_Inbox",
		"10_Reference",
		"20_Debug",
		"20_Debug/OpenClaw",
		"30_Dashboard",
		"attachments",
		"90_Templates",
		".obsidian",
		".git",
		"node_modules",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(vaultDir, d), 0o755)
	}

	files := map[string]string{
		"00_Inbox/临时记录.md":                 "# 临时记录\n一些内容",
		"10_Reference/官方文档.md":              "# 官方文档\n参考内容",
		"20_Debug/OpenClaw.md":              "# OpenClaw\n内容",
		"20_Debug/Codex.md":                 "# Codex\n内容",
		"20_Debug/Docker.md":                "# Docker\n内容",
		"30_Dashboard/常用系统入口.md":           "# 常用系统入口",
		"90_Templates/Web Clip 模板.md":       "# Web Clip 模板",
		"attachments/架构图.png":               "fake-png",
		".obsidian/config":                  "config",
		".git/HEAD":                         "ref: refs/heads/main",
		"node_modules/fake/index.js":        "module.exports = {}",
	}
	for path, content := range files {
		fullPath := filepath.Join(vaultDir, filepath.FromSlash(path))
		os.MkdirAll(filepath.Dir(fullPath), 0o755)
		os.WriteFile(fullPath, []byte(content), 0o644)
	}

	return vaultDir
}

func TestScanMarkdownFiles(t *testing.T) {
	vaultDir := createTestVault(t)

	files, err := Scan(vaultDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find 7 markdown files
	mdCount := 0
	for _, f := range files {
		if f.IsMarkdown {
			mdCount++
		}
	}
	if mdCount != 7 {
		t.Errorf("expected 7 markdown files, got %d", mdCount)
	}
}

func TestScanIgnoresObsidian(t *testing.T) {
	vaultDir := createTestVault(t)

	files, err := Scan(vaultDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	for _, f := range files {
		if containsIgnoredPrefix(f.Path, ".obsidian") {
			t.Errorf("should not include .obsidian files, got: %s", f.Path)
		}
		if containsIgnoredPrefix(f.Path, ".git") {
			t.Errorf("should not include .git files, got: %s", f.Path)
		}
		if containsIgnoredPrefix(f.Path, "node_modules") {
			t.Errorf("should not include node_modules files, got: %s", f.Path)
		}
	}
}

func TestScanDetectsAttachments(t *testing.T) {
	vaultDir := createTestVault(t)

	files, err := Scan(vaultDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	found := false
	for _, f := range files {
		if f.Name == "架构图.png" {
			found = true
			if f.IsMarkdown {
				t.Error("png should not be marked as markdown")
			}
			if f.Ext != ".png" {
				t.Errorf("expected ext .png, got %s", f.Ext)
			}
		}
	}
	if !found {
		t.Error("expected to find 架构图.png")
	}
}

func TestScanFileProperties(t *testing.T) {
	vaultDir := createTestVault(t)

	files, err := Scan(vaultDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	for _, f := range files {
		if f.Path == "" {
			t.Error("Path should not be empty")
		}
		if f.AbsPath == "" {
			t.Error("AbsPath should not be empty")
		}
		if f.Name == "" {
			t.Error("Name should not be empty")
		}
		if f.Ext == "" {
			t.Error("Ext should not be empty")
		}
		if f.Size < 0 {
			t.Error("Size should not be negative")
		}
		if f.ModTime.IsZero() {
			t.Error("ModTime should not be zero")
		}
	}
}

func TestBuildTree(t *testing.T) {
	vaultDir := createTestVault(t)

	files, _ := Scan(vaultDir)
	tree := BuildTree(files)

	if tree == nil {
		t.Fatal("tree should not be nil")
	}

	// Should have top-level directories
	names := dirNames(tree.Children)
	sort.Strings(names)

	expectedTop := []string{"00_Inbox", "10_Reference", "20_Debug", "30_Dashboard", "90_Templates", "attachments"}
	sort.Strings(expectedTop)

	if len(names) != len(expectedTop) {
		t.Errorf("expected %d top-level items, got %d: %v", len(expectedTop), len(names), names)
	}

	for _, exp := range expectedTop {
		found := false
		for _, n := range names {
			if n == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected top-level item %s not found", exp)
		}
	}
}

func TestTreeNestedFiles(t *testing.T) {
	vaultDir := createTestVault(t)

	files, _ := Scan(vaultDir)
	tree := BuildTree(files)

	// Find 20_Debug directory
	var debugDir *TreeNode
	for _, child := range tree.Children {
		if child.Name == "20_Debug" {
			debugDir = child
			break
		}
	}

	if debugDir == nil {
		t.Fatal("20_Debug directory not found in tree")
	}

	if debugDir.Type != "dir" {
		t.Error("20_Debug should be type 'dir'")
	}

	childNames := fileNames(debugDir.Children)
	expected := []string{"Codex.md", "Docker.md", "OpenClaw.md"}
	for _, exp := range expected {
		found := false
		for _, n := range childNames {
			if n == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s in 20_Debug, got children: %v", exp, childNames)
		}
	}
}

func containsIgnoredPrefix(path, prefix string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts[:len(parts)-1] {
		if p == prefix {
			return true
		}
	}
	return false
}

func dirNames(nodes []*TreeNode) []string {
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	return names
}

func fileNames(nodes []*TreeNode) []string {
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	return names
}

// Ensure VaultFile has ModTime
func TestVaultFileHasModTime(t *testing.T) {
	f := VaultFile{
		Path:       "test.md",
		AbsPath:    "/tmp/test.md",
		Name:       "test.md",
		Ext:        ".md",
		IsMarkdown: true,
		Size:       100,
		ModTime:    time.Now(),
	}
	if f.ModTime.IsZero() {
		t.Error("ModTime should be set")
	}
}

func TestScanCanvasFiles(t *testing.T) {
	vaultDir := createTestVault(t)

	// Add a .canvas file
	canvasContent := `{"nodes":[{"id":"n1","type":"text","text":"Hello","x":0,"y":0,"width":300,"height":200}],"edges":[]}`
	canvasPath := filepath.Join(vaultDir, "30_Dashboard", "知识地图.canvas")
	os.WriteFile(canvasPath, []byte(canvasContent), 0o644)

	files, err := Scan(vaultDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	found := false
	for _, f := range files {
		if f.Name == "知识地图.canvas" {
			found = true
			if !f.IsCanvas {
				t.Error("canvas file should have IsCanvas=true")
			}
			if f.IsMarkdown {
				t.Error("canvas file should not be marked as markdown")
			}
			if f.Ext != ".canvas" {
				t.Errorf("expected ext .canvas, got %s", f.Ext)
			}
		}
	}
	if !found {
		t.Error("expected to find 知识地图.canvas")
	}
}
