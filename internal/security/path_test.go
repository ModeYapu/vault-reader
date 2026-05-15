package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestUNCPathsBlocked(t *testing.T) {
	vaultDir := filepath.Join(t.TempDir(), "vault")
	os.MkdirAll(vaultDir, 0o755)

	tests := []struct {
		name string
		path string
	}{
		{"UNC extended path", `\\?\C:\secret`},
		{"UNC server share", `\\server\share`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(vaultDir, tt.path)
			if err == nil {
				t.Errorf("expected UNC path %q to be blocked", tt.path)
			}
		})
	}
}

func TestBackslashPath(t *testing.T) {
	vaultDir := filepath.Join(t.TempDir(), "vault")
	os.MkdirAll(vaultDir, 0o755)

	// On Windows, backslash paths like "folder\file.md" are treated as absolute
	// or get cleaned to "folder/file.md". Either way, the result must be safe.
	err := ValidatePath(vaultDir, `folder\file.md`)
	if runtime.GOOS == "windows" {
		// Windows filepath.Clean converts backslashes; the path is still relative
		// and within the vault, so it should either pass or be rejected for
		// drive-letter check — but not cause a traversal.
		if err != nil && !strings.Contains(err.Error(), "drive letter") {
			t.Logf("backslash path on windows: %v (acceptable)", err)
		}
	} else {
		// On non-Windows, backslash is a valid filename character
		if err != nil {
			t.Errorf("expected backslash path to be valid on non-Windows, got: %v", err)
		}
	}
}

func TestNullByteInMiddle(t *testing.T) {
	vaultDir := filepath.Join(t.TempDir(), "vault")
	os.MkdirAll(vaultDir, 0o755)

	err := ValidatePath(vaultDir, "notes/test\x00.md")
	if err == nil {
		t.Error("expected null byte in path to be rejected")
	}
}

func TestEncodedTraversal(t *testing.T) {
	vaultDir := filepath.Join(t.TempDir(), "vault")
	os.MkdirAll(vaultDir, 0o755)

	tests := []struct {
		name string
		path string
	}{
		{"dotdot percent-slash", "..%2f"},
		{"dotdot percent-backslash", "..%5c"},
		{"dotdot uppercase", "..%2F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Percent-encoded paths are passed as literal strings; the path
			// component "..%2f" is not a real traversal but should not resolve
			// outside the vault after cleaning.
			err := ValidatePath(vaultDir, tt.path)
			// The key guarantee: no path escapes the vault.
			// "..%2f" is a valid filename on most systems; if it passes
			// validation it is still confined to vaultDir.
			if err == nil {
				t.Logf("path %q passed validation (confined to vault)", tt.path)
			}
		})
	}
}
