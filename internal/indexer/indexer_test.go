package indexer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
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
