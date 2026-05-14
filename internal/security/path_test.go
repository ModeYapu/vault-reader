package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidPaths(t *testing.T) {
	vaultDir := filepath.Join(t.TempDir(), "vault")
	os.MkdirAll(vaultDir, 0o755)

	tests := []struct {
		name string
		path string
	}{
		{"simple file", "notes/test.md"},
		{"chinese filename", "笔记/测试.md"},
		{"with spaces", "my notes/hello world.md"},
		{"image asset", "attachments/image.png"},
		{"nested path", "a/b/c/d/file.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(vaultDir, tt.path)
			if err != nil {
				t.Errorf("expected path %q to be valid, got error: %v", tt.path, err)
			}
		})
	}
}

func TestPathTraversalBlocked(t *testing.T) {
	vaultDir := filepath.Join(t.TempDir(), "vault")
	os.MkdirAll(vaultDir, 0o755)

	tests := []struct {
		name string
		path string
	}{
		{"parent traversal", "../../etc/passwd"},
		{"mixed traversal", "notes/../../../etc/passwd"},
		{"absolute path", "/etc/passwd"},
		{"windows absolute", "C:\\Windows\\System32"},
		{"null byte", "notes/test.md\x00../../../etc/passwd"},
		{"dot dot only", ".."},
		{"dot dot slash", "../"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(vaultDir, tt.path)
			if err == nil {
				t.Errorf("expected path %q to be blocked", tt.path)
			}
		})
	}
}

func TestCleanPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"notes/test.md", "notes/test.md"},
		{"notes/../test.md", "test.md"},
		{"notes/./test.md", "notes/test.md"},
		{"//notes//test.md", "/notes/test.md"},
	}

	for _, tt := range tests {
		got := CleanPath(tt.input)
		if got != tt.expected {
			t.Errorf("CleanPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
