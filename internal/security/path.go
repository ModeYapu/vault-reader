package security

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// ValidatePath ensures the requested path is safe and stays within vaultDir.
// Returns an error if path traversal is detected.
func ValidatePath(vaultDir, requestedPath string) error {
	// Reject null bytes
	if strings.ContainsRune(requestedPath, 0) {
		return fmt.Errorf("invalid path: contains null byte")
	}

	// Reject absolute paths (Unix and Windows styles)
	if strings.HasPrefix(requestedPath, "/") || filepath.IsAbs(requestedPath) {
		return fmt.Errorf("invalid path: absolute paths not allowed")
	}

	// On Windows, also reject drive-letter paths
	if runtime.GOOS == "windows" && len(requestedPath) >= 2 && requestedPath[1] == ':' {
		return fmt.Errorf("invalid path: drive letter paths not allowed")
	}

	// Clean the path
	cleaned := filepath.Clean(requestedPath)

	// After cleaning, the path must not start with ..
	if strings.HasPrefix(cleaned, "..") || cleaned == ".." {
		return fmt.Errorf("invalid path: traversal outside vault directory")
	}

	// Build the full path and verify it's within vaultDir
	fullPath := filepath.Join(vaultDir, cleaned)
	absVault, err := filepath.Abs(vaultDir)
	if err != nil {
		return fmt.Errorf("invalid vault dir: %w", err)
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if !strings.HasPrefix(absFull, absVault+string(filepath.Separator)) && absFull != absVault {
		return fmt.Errorf("invalid path: escapes vault directory")
	}

	return nil
}

// CleanPath normalizes a path by cleaning it using forward slashes (web-safe).
func CleanPath(p string) string {
	return path.Clean(p)
}
