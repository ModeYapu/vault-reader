package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// VaultFile represents a file found in the vault.
type VaultFile struct {
	Path       string    `json:"path"`
	AbsPath    string    `json:"-"`
	Name       string    `json:"name"`
	Ext        string    `json:"ext"`
	IsMarkdown bool     `json:"isMarkdown"`
	IsCanvas   bool     `json:"isCanvas"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"modTime"`
}

var (
	markdownExts = map[string]bool{
		".md":       true,
		".markdown": true,
	}

	assetExts = map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".webp": true,
		".gif": true, ".svg": true, ".pdf": true, ".txt": true,
		".json": true, ".yaml": true, ".yml": true, ".zip": true,
	}

	ignoreDirs = map[string]bool{
		".obsidian":          true,
		".trash":             true,
		".git":               true,
		".DS_Store":          true,
		".vault-reader-data": true,
		"node_modules":       true,
	}
)

// Scan recursively scans a vault directory and returns all recognized files.
func Scan(vaultDir string) ([]VaultFile, error) {
	absVault, err := filepath.Abs(vaultDir)
	if err != nil {
		return nil, fmt.Errorf("invalid vault dir: %w", err)
	}

	var files []VaultFile

	err = filepath.WalkDir(absVault, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}

		relPath, _ := filepath.Rel(absVault, path)
		relPath = filepath.ToSlash(relPath)

		// Skip dot-prefixed and ignored directories
		if d.IsDir() {
			name := d.Name()
			if len(name) > 0 && name[0] == '.' {
				return filepath.SkipDir
			}
			if ignoreDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		// Also skip dot-prefixed files
		if len(d.Name()) > 0 && d.Name()[0] == '.' {
			return nil
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !markdownExts[ext] && !assetExts[ext] {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		files = append(files, VaultFile{
			Path:       relPath,
			AbsPath:    path,
			Name:       d.Name(),
			Ext:        ext,
			IsMarkdown: markdownExts[ext],
			Size:       info.Size(),
			ModTime:    info.ModTime(),
		})

		return nil
	})

	return files, err
}

func splitPath(p string) []string {
	if p == "." {
		return nil
	}
	var parts []string
	current := ""
	for _, ch := range p {
		if ch == '/' || ch == '\\' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
