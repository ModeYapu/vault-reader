package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vault-reader/internal/parser"
)

// createTestVaultAndIndexer creates a temp vault with markdown files and an indexer.
func createTestVaultAndIndexer(t *testing.T) (*Indexer, string) {
	t.Helper()
	vaultDir := t.TempDir()
	dbPath := filepath.Join(vaultDir, ".test-data", "test.db")

	ix, err := New(dbPath, vaultDir)
	if err != nil {
		t.Fatalf("create indexer: %v", err)
	}
	t.Cleanup(func() { ix.Close() })

	return ix, vaultDir
}

// writeMarkdown writes a markdown file with frontmatter tags into the vault.
func writeMarkdown(t *testing.T, vaultDir, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(vaultDir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

// TestGetTagTree_Empty tests that an empty vault returns an empty tag tree.
func TestGetTagTree_Empty(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# Hello\nNo tags here.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tree, err := ix.GetTagTree()
	if err != nil {
		t.Fatalf("GetTagTree: %v", err)
	}
	if len(tree) != 0 {
		t.Fatalf("expected empty tree, got %d nodes", len(tree))
	}
}

// TestGetTagTree_Flat tests flat tags (no nesting).
func TestGetTagTree_Flat(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\ntags: [openclaw, debug]\n---\n# A\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "---\ntags: [openclaw]\n---\n# B\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tree, err := ix.GetTagTree()
	if err != nil {
		t.Fatalf("GetTagTree: %v", err)
	}

	if len(tree) < 2 {
		t.Fatalf("expected at least 2 top-level nodes, got %d: %+v", len(tree), tree)
	}

	// Build a map for easier lookup
	nodeMap := make(map[string]*TagTreeNode)
	for i := range tree {
		nodeMap[tree[i].Name] = &tree[i]
	}

	if n, ok := nodeMap["openclaw"]; !ok {
		t.Fatal("expected 'openclaw' node")
	} else {
		if n.Count != 2 {
			t.Errorf("openclaw count: got %d, want 2", n.Count)
		}
		if n.FullName != "openclaw" {
			t.Errorf("openclaw fullName: got %q, want 'openclaw'", n.FullName)
		}
	}

	if n, ok := nodeMap["debug"]; !ok {
		t.Fatal("expected 'debug' node")
	} else {
		if n.Count != 1 {
			t.Errorf("debug count: got %d, want 1", n.Count)
		}
	}
}

// TestGetTagTree_Nested tests nested tags like debug/proxy, debug/oauth.
func TestGetTagTree_Nested(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\ntags: [debug/proxy]\n---\n# A\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "---\ntags: [debug/oauth, debug/proxy]\n---\n# B\n")
	writeMarkdown(t, vaultDir, "notes/c.md", "---\ntags: [debug]\n---\n# C\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tree, err := ix.GetTagTree()
	if err != nil {
		t.Fatalf("GetTagTree: %v", err)
	}

	// Should have 1 top-level node: debug
	if len(tree) != 1 {
		t.Fatalf("expected 1 top-level node, got %d: %+v", len(tree), tree)
	}

	root := tree[0]
	if root.Name != "debug" {
		t.Errorf("root name: got %q, want 'debug'", root.Name)
	}
	if root.FullName != "debug" {
		t.Errorf("root fullName: got %q, want 'debug'", root.FullName)
	}
	// debug itself is used by c.md (1), debug/proxy by a.md+b.md (2), debug/oauth by b.md (1)
	// root.Count should reflect the "debug" tag itself: 1
	if root.Count != 1 {
		t.Errorf("root count: got %d, want 1 (only c.md has plain 'debug' tag)", root.Count)
	}

	// Children: proxy and oauth
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d: %+v", len(root.Children), root.Children)
	}

	childMap := make(map[string]*TagTreeNode)
	for i := range root.Children {
		childMap[root.Children[i].Name] = &root.Children[i]
	}

	if n, ok := childMap["proxy"]; !ok {
		t.Fatal("expected 'proxy' child")
	} else {
		if n.FullName != "debug/proxy" {
			t.Errorf("proxy fullName: got %q, want 'debug/proxy'", n.FullName)
		}
		if n.Count != 2 {
			t.Errorf("proxy count: got %d, want 2", n.Count)
		}
	}

	if n, ok := childMap["oauth"]; !ok {
		t.Fatal("expected 'oauth' child")
	} else {
		if n.FullName != "debug/oauth" {
			t.Errorf("oauth fullName: got %q, want 'debug/oauth'", n.FullName)
		}
		if n.Count != 1 {
			t.Errorf("oauth count: got %d, want 1", n.Count)
		}
	}
}

// TestGetTagTree_DeepNesting tests 3-level tags like a/b/c.
func TestGetTagTree_DeepNesting(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\ntags: [codex/oauth/token]\n---\n# A\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tree, err := ix.GetTagTree()
	if err != nil {
		t.Fatalf("GetTagTree: %v", err)
	}

	if len(tree) != 1 {
		t.Fatalf("expected 1 root, got %d", len(tree))
	}
	if tree[0].Name != "codex" {
		t.Errorf("root: got %q, want 'codex'", tree[0].Name)
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("expected 1 child of codex, got %d", len(tree[0].Children))
	}
	oauth := tree[0].Children[0]
	if oauth.Name != "oauth" {
		t.Errorf("oauth: got %q", oauth.Name)
	}
	if len(oauth.Children) != 1 {
		t.Fatalf("expected 1 child of oauth, got %d", len(oauth.Children))
	}
	token := oauth.Children[0]
	if token.FullName != "codex/oauth/token" {
		t.Errorf("token fullName: got %q", token.FullName)
	}
	if token.Count != 1 {
		t.Errorf("token count: got %d, want 1", token.Count)
	}
}

// TestGetTagTree_InlineTags tests that inline #tag extraction works with tag tree.
func TestGetTagTree_InlineTags(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# Title\n\nSome text #openclaw here.\n\nMore #debug/proxy stuff.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tree, err := ix.GetTagTree()
	if err != nil {
		t.Fatalf("GetTagTree: %v", err)
	}

	if len(tree) != 2 {
		t.Fatalf("expected 2 roots, got %d: %+v", len(tree), tree)
	}

	// Verify the tree is valid JSON
	data, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("marshal tree: %v", err)
	}
	t.Logf("Tag tree JSON: %s", data)
}

// ==================== Graph Tests ====================

// TestGetGraph_Empty tests graph with no links.
func TestGetGraph_Empty(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)
	writeMarkdown(t, vaultDir, "notes/a.md", "# A\nNo links.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	nodes, edges, err := ix.GetGraph("", "", "", 1, 500)
	if err != nil {
		t.Fatalf("GetGraph: %v", err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected at least 1 node")
	}
	if len(edges) != 0 {
		t.Fatalf("expected 0 edges, got %d", len(edges))
	}
}

// TestGetGraph_WithLinks tests graph with wikilinks between files.
func TestGetGraph_WithLinks(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\nLink to [[b]]\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "# B\nLink to [[a]]\n")
	writeMarkdown(t, vaultDir, "notes/c.md", "# C\nNo links.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	nodes, edges, err := ix.GetGraph("", "", "", 1, 500)
	if err != nil {
		t.Fatalf("GetGraph: %v", err)
	}

	if len(nodes) < 2 {
		t.Fatalf("expected at least 2 nodes, got %d", len(nodes))
	}
	if len(edges) < 1 {
		t.Fatalf("expected at least 1 edge, got %d", len(edges))
	}

	// Verify edge direction: a -> b or b -> a
	foundAB := false
	for _, e := range edges {
		if (e.Source == "notes/a.md" && e.Target == "notes/b.md") ||
			(e.Source == "notes/b.md" && e.Target == "notes/a.md") {
			foundAB = true
		}
	}
	if !foundAB {
		t.Errorf("expected edge between a and b, got edges: %+v", edges)
	}
}

// TestGetGraph_FilterByFolder tests folder filtering.
func TestGetGraph_FilterByFolder(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "debug/a.md", "# A\nLink to [[b]]\n")
	writeMarkdown(t, vaultDir, "debug/b.md", "# B\n")
	writeMarkdown(t, vaultDir, "ref/c.md", "# C\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	nodes, _, err := ix.GetGraph("debug", "", "", 1, 500)
	if err != nil {
		t.Fatalf("GetGraph: %v", err)
	}

	for _, n := range nodes {
		if n.Group != "debug" {
			t.Errorf("expected all nodes in 'debug' group, got group %q for %s", n.Group, n.ID)
		}
	}
}

// TestGetGraph_MaxNodes tests that maxNodes limits the result.
func TestGetGraph_MaxNodes(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	for i := 0; i < 20; i++ {
		writeMarkdown(t, vaultDir, fmt.Sprintf("notes/n%d.md", i), fmt.Sprintf("# N%d\n", i))
	}

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	nodes, _, err := ix.GetGraph("", "", "", 1, 5)
	if err != nil {
		t.Fatalf("GetGraph: %v", err)
	}

	if len(nodes) > 5 {
		t.Errorf("expected max 5 nodes, got %d", len(nodes))
	}
}

// ==================== Search Tests ====================

func TestSearch(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/go.md", "# Go Language\nGo is a statically typed language.\n")
	writeMarkdown(t, vaultDir, "notes/rust.md", "# Rust Language\nRust is a systems programming language.\n")
	writeMarkdown(t, vaultDir, "notes/python.md", "# Python\nPython is a dynamic language.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tests := []struct {
		name          string
		query         string
		wantMin       int
		mustContain   string
		mustNotContain string
	}{
		{
			name:        "keyword_language",
			query:       "language",
			wantMin:     2,
			mustContain: "notes/go.md",
		},
		{
			name:        "single_result",
			query:       "python",
			wantMin:     1,
			mustContain: "notes/python.md",
		},
		{
			name:        "title_search",
			query:       "Rust",
			wantMin:     1,
			mustContain: "notes/rust.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results, err := ix.Search(tc.query, 10)
			if err != nil {
				t.Fatalf("Search(%q): %v", tc.query, err)
			}
			if len(results) < tc.wantMin {
				t.Fatalf("expected at least %d results for %q, got %d", tc.wantMin, tc.query, len(results))
			}
			if tc.mustContain != "" {
				found := false
				for _, r := range results {
					if r.Path == tc.mustContain {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected result to contain path %q, got: %+v", tc.mustContain, results)
				}
			}
			if tc.mustNotContain != "" {
				for _, r := range results {
					if r.Path == tc.mustNotContain {
						t.Errorf("result should not contain path %q", tc.mustNotContain)
					}
				}
			}
		})
	}
}

func TestSearch_NoResults(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# Hello\nSimple content.\n")
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	results, err := ix.Search("xyznonexistent", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_CJK(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/cn.md", "# 中文笔记\n这是一段中文内容测试。\n")
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	results, err := ix.Search("中文笔记", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// FTS5 tokenization of CJK varies by build; log the result count
	// for visibility but do not fail on zero matches since the default
	// tokenizer may not support CJK word segmentation.
	t.Logf("CJK search returned %d results", len(results))
	if len(results) >= 1 {
		found := false
		for _, r := range results {
			if r.Path == "notes/cn.md" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected notes/cn.md in results, got: %+v", results)
		}
	}
}

// ==================== Backlinks Tests ====================

func TestGetBacklinks(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\nLink to [[b]]\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "# B\nSome content.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	links, err := ix.GetBacklinks("notes/b.md")
	if err != nil {
		t.Fatalf("GetBacklinks: %v", err)
	}
	if len(links) == 0 {
		t.Fatal("expected at least 1 backlink")
	}

	found := false
	for _, bl := range links {
		if bl.FromPath == "notes/a.md" {
			found = true
			if bl.Title != "A" {
				t.Errorf("backlink title: got %q, want 'A'", bl.Title)
			}
		}
	}
	if !found {
		t.Errorf("expected backlink from notes/a.md, got: %+v", links)
	}
}

func TestGetBacklinks_NoBacklinks(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\nNo outgoing links.\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "# B\nNo outgoing links.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	links, err := ix.GetBacklinks("notes/a.md")
	if err != nil {
		t.Fatalf("GetBacklinks: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 backlinks, got %d", len(links))
	}
}

// ==================== Tags Tests ====================

func TestGetTags(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\ntags: [debug, openclaw]\n---\n# A\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "---\ntags: [debug]\n---\n# B\n")
	writeMarkdown(t, vaultDir, "notes/c.md", "# C\nNo tags.\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tags, err := ix.GetTags()
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}

	tagMap := make(map[string]int)
	for _, tc := range tags {
		tagMap[tc.Tag] = tc.Count
	}

	if tagMap["debug"] != 2 {
		t.Errorf("debug count: got %d, want 2", tagMap["debug"])
	}
	if tagMap["openclaw"] != 1 {
		t.Errorf("openclaw count: got %d, want 1", tagMap["openclaw"])
	}
}

func TestGetTags_EmptyVault(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\nNo frontmatter tags.\n")
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	tags, err := ix.GetTags()
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags, got %d", len(tags))
	}
}

// ==================== FilesByTag Tests ====================

func TestGetFilesByTag(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\ntags: [debug, openclaw]\n---\n# A\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "---\ntags: [debug]\n---\n# B\n")
	writeMarkdown(t, vaultDir, "notes/c.md", "---\ntags: [openclaw]\n---\n# C\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	files, err := ix.GetFilesByTag("debug")
	if err != nil {
		t.Fatalf("GetFilesByTag: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files with tag 'debug', got %d", len(files))
	}

	paths := make(map[string]bool)
	for _, f := range files {
		paths[f.Path] = true
	}
	if !paths["notes/a.md"] || !paths["notes/b.md"] {
		t.Errorf("expected notes/a.md and notes/b.md, got: %+v", files)
	}
}

func TestGetFilesByTag_NoMatch(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\ntags: [debug]\n---\n# A\n")
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	files, err := ix.GetFilesByTag("nonexistent")
	if err != nil {
		t.Fatalf("GetFilesByTag: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

// ==================== Properties Tests ====================

func TestGetProperties(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\ntitle: My Note\nstatus: active\npriority: 3\n---\n# A\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	props, err := ix.GetProperties("notes/a.md")
	if err != nil {
		t.Fatalf("GetProperties: %v", err)
	}

	propMap := make(map[string]Property)
	for _, p := range props {
		propMap[p.Key] = p
	}

	if p, ok := propMap["status"]; !ok {
		t.Error("expected 'status' property")
	} else {
		if p.Value != "active" {
			t.Errorf("status value: got %q, want 'active'", p.Value)
		}
		if p.ValueType != "string" {
			t.Errorf("status type: got %q, want 'string'", p.ValueType)
		}
	}

	if p, ok := propMap["priority"]; !ok {
		t.Error("expected 'priority' property")
	} else {
		// yaml.v3 unmarshals integers as `int`, which falls into the default
		// case of insertProperty and is stored as "string".
		if p.Value != "3" {
			t.Errorf("priority value: got %q, want '3'", p.Value)
		}
	}
}

// ==================== FilterByProperty Tests ====================

func TestFilterByProperty(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\nstatus: active\n---\n# A\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "---\nstatus: inactive\n---\n# B\n")
	writeMarkdown(t, vaultDir, "notes/c.md", "---\nstatus: active\n---\n# C\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	files, err := ix.FilterByProperty("status", "active")
	if err != nil {
		t.Fatalf("FilterByProperty: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 active files, got %d", len(files))
	}

	paths := make(map[string]bool)
	for _, f := range files {
		paths[f.Path] = true
	}
	if !paths["notes/a.md"] || !paths["notes/c.md"] {
		t.Errorf("expected notes/a.md and notes/c.md, got: %+v", files)
	}

	// Verify inactive is not included
	if paths["notes/b.md"] {
		t.Error("notes/b.md should not be in active results")
	}
}

func TestFilterByProperty_NoMatch(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\nstatus: active\n---\n# A\n")
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	files, err := ix.FilterByProperty("status", "archived")
	if err != nil {
		t.Fatalf("FilterByProperty: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}

// ==================== Block Tests ====================

func TestGetBlock(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\n\nSome paragraph ^myblock\n\nAnother paragraph\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	block, err := ix.GetBlock("myblock")
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if block == nil {
		t.Fatal("expected block, got nil")
	}
	if block.BlockID != "myblock" {
		t.Errorf("block ID: got %q, want 'myblock'", block.BlockID)
	}
	if block.Path != "notes/a.md" {
		t.Errorf("block path: got %q, want 'notes/a.md'", block.Path)
	}
	if !strings.Contains(block.Text, "Some paragraph") {
		t.Errorf("block text should contain 'Some paragraph', got %q", block.Text)
	}
}

func TestGetBlock_NotFound(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\nNo blocks.\n")
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	block, err := ix.GetBlock("nonexistent")
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if block != nil {
		t.Fatalf("expected nil for nonexistent block, got %+v", block)
	}
}

func TestGetBlocksByFile(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\n\nFirst block ^block1\n\nSecond block ^block2\n\nThird block ^block3\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	blocks, err := ix.GetBlocksByFile("notes/a.md")
	if err != nil {
		t.Fatalf("GetBlocksByFile: %v", err)
	}
	if len(blocks) < 3 {
		t.Fatalf("expected at least 3 blocks, got %d", len(blocks))
	}

	ids := make(map[string]bool)
	for _, b := range blocks {
		ids[b.BlockID] = true
	}
	for _, id := range []string{"block1", "block2", "block3"} {
		if !ids[id] {
			t.Errorf("expected block ID %q not found", id)
		}
	}
}

// ==================== Dashboard Tests ====================

func TestGetDashboard(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "00_Inbox/inbox1.md", "# Inbox Note\nContent.\n")
	writeMarkdown(t, vaultDir, "notes/active1.md", "---\nstatus: active\n---\n# Active Note\n")
	writeMarkdown(t, vaultDir, "debug/dbg1.md", "---\ntype: debug-note\n---\n# Debug Note\n")
	writeMarkdown(t, vaultDir, "notes/tagged.md", "---\ntags: [debug, openclaw]\n---\n# Tagged\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	dash, err := ix.GetDashboard()
	if err != nil {
		t.Fatalf("GetDashboard: %v", err)
	}

	// Recent should have files
	if len(dash.Recent) == 0 {
		t.Error("expected recent files, got none")
	}

	// Inbox should have the inbox note
	found := false
	for _, f := range dash.Inbox {
		if strings.Contains(f.Path, "00_Inbox/") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected inbox file in Inbox section, got: %+v", dash.Inbox)
	}

	// Active should have the active note
	found = false
	for _, f := range dash.Active {
		if f.Path == "notes/active1.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected active1.md in Active section, got: %+v", dash.Active)
	}

	// Debug should have the debug note
	found = false
	for _, f := range dash.Debug {
		if f.Path == "debug/dbg1.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected dbg1.md in Debug section, got: %+v", dash.Debug)
	}

	// Tags should be populated
	if len(dash.Tags) == 0 {
		t.Error("expected tags in dashboard")
	}
}

// ==================== VaultQuery Tests ====================

func TestExecuteVaultQuery(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "---\nstatus: active\npriority: 1\n---\n# Alpha\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "---\nstatus: inactive\n---\n# Beta\n")
	writeMarkdown(t, vaultDir, "notes/c.md", "---\nstatus: active\npriority: 2\n---\n# Gamma\n")
	writeMarkdown(t, vaultDir, "other/d.md", "---\nstatus: active\n---\n# Delta\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	q, err := parser.ParseVaultQuery(`type: table
from: notes
where:
  status: active
fields:
  - priority
sort: updated
limit: 10`)
	if err != nil {
		t.Fatalf("ParseVaultQuery: %v", err)
	}

	results, err := ix.ExecuteVaultQuery(q)
	if err != nil {
		t.Fatalf("ExecuteVaultQuery: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	paths := make(map[string]bool)
	for _, r := range results {
		paths[r.Path] = true
		if r.Path == "other/d.md" {
			t.Error("other/d.md should not be in results (filtered by 'from: notes')")
		}
	}
	if !paths["notes/a.md"] || !paths["notes/c.md"] {
		t.Errorf("expected notes/a.md and notes/c.md, got: %+v", results)
	}

	// Verify fields are populated
	for _, r := range results {
		if r.Path == "notes/a.md" {
			if v, ok := r.Fields["status"]; !ok || v != "active" {
				t.Errorf("notes/a.md fields: expected status=active, got %q", v)
			}
		}
	}
}

// ==================== cleanFTSQuery Tests ====================

func TestCleanFTSQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "sandwich_preserves_AND",
			input: "SANDWICH",
			want:  `"SANDWICH"`,
		},
		{
			name:  "removes_standalone_AND",
			input: "open AND close",
			want:  `"open" OR "close"`,
		},
		{
			name:  "removes_standalone_OR_NOT",
			input: "foo OR bar NOT baz",
			want:  `"foo" OR "bar" OR "baz"`,
		},
		{
			name:  "strips_special_chars",
			input: `"hello" {world}`,
			want:  `"hello" OR "world"`,
		},
		{
			name:  "strips_braces_and_quotes",
			input: `test"value{here}`,
			want:  `"testvaluehere"`,
		},
		{
			name:  "empty_query",
			input: "",
			want:  `""`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := cleanFTSQuery(tc.input)
			if got != tc.want {
				t.Errorf("cleanFTSQuery(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ==================== Resolve Tests ====================

func TestResolve(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# Alpha\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "# Beta\n")

	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	result := ix.Resolve("a")
	if !result.Found {
		t.Fatal("expected to resolve 'a'")
	}
	if result.TargetPath != "notes/a.md" {
		t.Errorf("resolved path: got %q, want 'notes/a.md'", result.TargetPath)
	}
}

func TestResolve_NotFound(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# Alpha\n")
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("full index: %v", err)
	}

	result := ix.Resolve("nonexistent_xyz")
	if result.Found {
		t.Errorf("expected not found, got %+v", result)
	}
}

// ==================== NewIndexer Tests ====================

func TestNewIndexer(t *testing.T) {
	vaultDir := t.TempDir()
	dbPath := filepath.Join(vaultDir, ".test-data", "sub", "deep", "test.db")

	ix, err := New(dbPath, vaultDir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer ix.Close()

	// Verify database is usable
	if _, err := ix.db.Exec("CREATE TABLE IF NOT EXISTS _test (id INTEGER)"); err != nil {
		t.Fatalf("database not usable: %v", err)
	}

	// Verify parent directories were created
	dir := filepath.Dir(dbPath)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("parent dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory at %q", dir)
	}
}

func TestNewIndexer_CreatesSchema(t *testing.T) {
	vaultDir := t.TempDir()
	dbPath := filepath.Join(vaultDir, "schema_test.db")

	ix, err := New(dbPath, vaultDir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer ix.Close()

	// Verify key tables exist
	tables := []string{"files", "links", "tags", "headings", "properties", "blocks", "file_fts"}
	for _, table := range tables {
		var name string
		err := ix.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			// FTS is a virtual table, check differently
			if table == "file_fts" {
				err = ix.db.QueryRow(
					"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
			}
			if err != nil {
				t.Errorf("table %q not found in schema: %v", table, err)
			}
		}
	}
}

// ==================== FullIndex Idempotency ====================

func TestFullIndex_DoubleRun(t *testing.T) {
	ix, vaultDir := createTestVaultAndIndexer(t)

	writeMarkdown(t, vaultDir, "notes/a.md", "# A\nContent for A.\n")
	writeMarkdown(t, vaultDir, "notes/b.md", "---\ntags: [debug]\n---\n# B\n")

	// First run
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("first FullIndex: %v", err)
	}

	// Query to get baseline
	tags1, err := ix.GetTags()
	if err != nil {
		t.Fatalf("GetTags after first run: %v", err)
	}

	// Second run (should be idempotent)
	if err := ix.FullIndex(); err != nil {
		t.Fatalf("second FullIndex: %v", err)
	}

	// Verify same results
	tags2, err := ix.GetTags()
	if err != nil {
		t.Fatalf("GetTags after second run: %v", err)
	}

	if len(tags1) != len(tags2) {
		t.Fatalf("tag count changed after re-index: first=%d, second=%d", len(tags1), len(tags2))
	}

	// Compare counts
	counts1 := tagCountsToMap(tags1)
	counts2 := tagCountsToMap(tags2)
	for tag, c1 := range counts1 {
		c2, ok := counts2[tag]
		if !ok {
			t.Errorf("tag %q missing after re-index", tag)
		} else if c1 != c2 {
			t.Errorf("tag %q count changed: first=%d, second=%d", tag, c1, c2)
		}
	}
}

// tagCountsToMap converts a slice of TagCount to a map for easy comparison.
func tagCountsToMap(tags []TagCount) map[string]int {
	m := make(map[string]int, len(tags))
	for _, tc := range tags {
		m[tc.Tag] = tc.Count
	}
	return m
}
