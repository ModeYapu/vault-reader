package resolver

import (
	"path/filepath"
	"strings"
)

// ResolveResult is the outcome of resolving a wikilink target.
type ResolveResult struct {
	Found       bool     `json:"found"`
	TargetPath  string   `json:"targetPath,omitempty"`
	Candidates  []string `json:"candidates,omitempty"`
	IsAmbiguous bool     `json:"isAmbiguous,omitempty"`
}

// FileMeta holds minimal metadata about a vault file for resolution.
type FileMeta struct {
	Path           string   // relative path within vault, e.g. "20_Debug/OpenClaw.md"
	Name           string   // filename without extension, e.g. "OpenClaw"
	Title          string   // from frontmatter or first heading
	NormalizedPath string   // path without .md extension
	Aliases        []string // from frontmatter aliases field
}

// Resolver resolves wikilink targets to actual vault file paths.
type Resolver struct {
	files      []FileMeta
	byName     map[string][]int // name (lower) -> indices
	byTitle    map[string][]int // title (lower) -> indices
	byNormPath map[string]int   // normalized path (lower) -> index
	byRelPath  map[string]int   // exact relative path (lower) -> index
	byAlias    map[string][]int // alias (lower) -> indices
}

// New creates a Resolver from a list of file metadata.
func New(files []FileMeta) *Resolver {
	r := &Resolver{
		files:      files,
		byName:     make(map[string][]int),
		byTitle:    make(map[string][]int),
		byNormPath: make(map[string]int),
		byRelPath:  make(map[string]int),
		byAlias:    make(map[string][]int),
	}

	for i, f := range files {
		lowerName := strings.ToLower(f.Name)
		r.byName[lowerName] = append(r.byName[lowerName], i)

		if f.Title != "" {
			lowerTitle := strings.ToLower(f.Title)
			r.byTitle[lowerTitle] = append(r.byTitle[lowerTitle], i)
		}

		r.byRelPath[strings.ToLower(f.Path)] = i

		normPath := strings.ToLower(f.NormalizedPath)
		r.byNormPath[normPath] = i

		for _, alias := range f.Aliases {
			lowerAlias := strings.ToLower(alias)
			r.byAlias[lowerAlias] = append(r.byAlias[lowerAlias], i)
		}
	}

	return r
}

// Resolve attempts to resolve a wikilink target to a file path.
//
// Priority:
//  1. Exact relative path match (with extension)
//  2. Normalized path match (without .md extension)
//  3. Filename match (without extension)
//  4. Frontmatter alias match
//  5. Frontmatter title match
//  6. Ambiguous: return multiple candidates
func (r *Resolver) Resolve(target string) ResolveResult {
	if target == "" {
		return ResolveResult{}
	}

	// Normalize: strip leading/trailing whitespace
	target = strings.TrimSpace(target)

	// Strip .md/.markdown extension for matching purposes, but keep for exact path lookup
	targetNoExt := stripMdExt(target)

	// 1. Exact relative path match
	lower := strings.ToLower(target)
	if idx, ok := r.byRelPath[lower]; ok {
		return ResolveResult{
			Found:      true,
			TargetPath: r.files[idx].Path,
		}
	}

	// 2. Try adding .md extension
	if idx, ok := r.byRelPath[lower+".md"]; ok {
		return ResolveResult{
			Found:      true,
			TargetPath: r.files[idx].Path,
		}
	}

	// 3. Normalized path match (path without .md)
	lowerNoExt := strings.ToLower(targetNoExt)
	if idx, ok := r.byNormPath[lowerNoExt]; ok {
		return ResolveResult{
			Found:      true,
			TargetPath: r.files[idx].Path,
		}
	}

	// Extract filename part (last segment if path-like)
	namePart := targetNoExt
	if idx := strings.LastIndex(targetNoExt, "/"); idx >= 0 {
		namePart = targetNoExt[idx+1:]
	}
	lowerName := strings.ToLower(namePart)

	// 4. Filename match
	if indices, ok := r.byName[lowerName]; ok {
		if len(indices) == 1 {
			return ResolveResult{
				Found:      true,
				TargetPath: r.files[indices[0]].Path,
			}
		}
		// Multiple files with same name - ambiguous
		candidates := make([]string, len(indices))
		for i, idx := range indices {
			candidates[i] = r.files[idx].Path
		}
		return ResolveResult{
			Found:       true,
			TargetPath:  candidates[0],
			Candidates:  candidates,
			IsAmbiguous: true,
		}
	}

	// 5. Alias match (before title — aliases are more specific)
	if indices, ok := r.byAlias[lowerName]; ok {
		if len(indices) == 1 {
			return ResolveResult{
				Found:      true,
				TargetPath: r.files[indices[0]].Path,
			}
		}
		candidates := make([]string, len(indices))
		for i, idx := range indices {
			candidates[i] = r.files[idx].Path
		}
		return ResolveResult{
			Found:       true,
			TargetPath:  candidates[0],
			Candidates:  candidates,
			IsAmbiguous: true,
		}
	}

	// 6. Frontmatter title match
	if indices, ok := r.byTitle[lowerName]; ok {
		if len(indices) == 1 {
			return ResolveResult{
				Found:      true,
				TargetPath: r.files[indices[0]].Path,
			}
		}
		candidates := make([]string, len(indices))
		for i, idx := range indices {
			candidates[i] = r.files[idx].Path
		}
		return ResolveResult{
			Found:       true,
			TargetPath:  candidates[0],
			Candidates:  candidates,
			IsAmbiguous: true,
		}
	}

	return ResolveResult{Found: false}
}

// stripMdExt removes .md or .markdown extension from a path segment.
func stripMdExt(s string) string {
	if strings.HasSuffix(strings.ToLower(s), ".md") {
		return s[:len(s)-3]
	}
	if strings.HasSuffix(strings.ToLower(s), ".markdown") {
		return s[:len(s)-9]
	}
	return s
}

// BuildFileMeta extracts FileMeta from a vault file path and optional title.
func BuildFileMeta(relPath, title string) FileMeta {
	name := filepath.Base(relPath)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	normPath := relPath
	if strings.HasSuffix(strings.ToLower(normPath), ".md") {
		normPath = normPath[:len(normPath)-3]
	}
	if strings.HasSuffix(strings.ToLower(normPath), ".markdown") {
		normPath = normPath[:len(normPath)-9]
	}

	return FileMeta{
		Path:           relPath,
		Name:           name,
		Title:          title,
		NormalizedPath: normPath,
	}
}
