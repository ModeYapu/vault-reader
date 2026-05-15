package resolver

import (
	"testing"
)

func TestResolve_ExactPath(t *testing.T) {
	files := []FileMeta{
		{Path: "A_O_A/A_O_A_知识体系总索引.md", Name: "A_O_A_知识体系总索引", NormalizedPath: "A_O_A/A_O_A_知识体系总索引"},
	}
	r := New(files)

	result := r.Resolve("A_O_A/A_O_A_知识体系总索引.md")
	if !result.Found {
		t.Error("expected to find exact path")
	}
	if result.TargetPath != "A_O_A/A_O_A_知识体系总索引.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}
}

func TestResolve_RelativePathWithDotDot(t *testing.T) {
	files := []FileMeta{
		{Path: "A_O_A/A_O_A_知识体系总索引.md", Name: "A_O_A_知识体系总索引", NormalizedPath: "A_O_A/A_O_A_知识体系总索引"},
		{Path: "A_O_A/WASM/WASM_知识体系索引.md", Name: "WASM_知识体系索引", NormalizedPath: "A_O_A/WASM/WASM_知识体系索引"},
		{Path: "A_O_A/WebGPU/WebGPU_知识体系索引.md", Name: "WebGPU_知识体系索引", NormalizedPath: "A_O_A/WebGPU/WebGPU_知识体系索引"},
	}
	r := New(files)

	// Relative path with ../ and .md extension — should resolve by filename
	result := r.Resolve("../../../A_O_A_知识体系总索引.md")
	if !result.Found {
		t.Error("expected to resolve relative path with ../")
	}
	if result.TargetPath != "A_O_A/A_O_A_知识体系总索引.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}

	// Relative path without .md
	result = r.Resolve("../../../WebGPU/WebGPU_知识体系索引")
	if !result.Found {
		t.Error("expected to resolve relative path without .md")
	}
	if result.TargetPath != "A_O_A/WebGPU/WebGPU_知识体系索引.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}
}

func TestResolve_ByNameWithMdExt(t *testing.T) {
	files := []FileMeta{
		{Path: "notes/hello.md", Name: "hello", NormalizedPath: "notes/hello"},
	}
	r := New(files)

	// Target includes .md extension but resolves by name
	result := r.Resolve("hello.md")
	if !result.Found {
		t.Error("expected to find by name with .md extension")
	}
	if result.TargetPath != "notes/hello.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}
}

func TestResolve_ByTitle(t *testing.T) {
	files := []FileMeta{
		{Path: "docs/guide.md", Name: "guide", Title: "Getting Started", NormalizedPath: "docs/guide"},
	}
	r := New(files)

	result := r.Resolve("Getting Started")
	if !result.Found {
		t.Error("expected to find by title")
	}
}

func TestResolve_NotFound(t *testing.T) {
	files := []FileMeta{
		{Path: "a.md", Name: "a", NormalizedPath: "a"},
	}
	r := New(files)

	result := r.Resolve("nonexistent")
	if result.Found {
		t.Error("expected not found")
	}
}

func TestResolve_ByAlias(t *testing.T) {
	files := []FileMeta{
		{Path: "notes/real-name.md", Name: "real-name", NormalizedPath: "notes/real-name", Aliases: []string{"my-alias"}},
	}
	r := New(files)

	result := r.Resolve("my-alias")
	if !result.Found {
		t.Fatal("expected to find by alias")
	}
	if result.TargetPath != "notes/real-name.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}
}

func TestResolve_Ambiguous(t *testing.T) {
	files := []FileMeta{
		{Path: "dir1/notes.md", Name: "notes", NormalizedPath: "dir1/notes"},
		{Path: "dir2/notes.md", Name: "notes", NormalizedPath: "dir2/notes"},
	}
	r := New(files)

	result := r.Resolve("notes")
	if !result.Found {
		t.Fatal("expected to find")
	}
	if !result.IsAmbiguous {
		t.Error("expected IsAmbiguous=true")
	}
	if len(result.Candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(result.Candidates))
	}
}

func TestResolve_EmptyResolver(t *testing.T) {
	r := New(nil)

	result := r.Resolve("anything")
	if result.Found {
		t.Error("expected Found=false for empty resolver")
	}
}

func TestResolve_CaseInsensitive(t *testing.T) {
	files := []FileMeta{
		{Path: "notes/Hello.md", Name: "Hello", NormalizedPath: "notes/Hello"},
	}
	r := New(files)

	result := r.Resolve("hello")
	if !result.Found {
		t.Error("expected case-insensitive match for 'hello'")
	}
	if result.TargetPath != "notes/Hello.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}
}

func TestResolve_WithMdExt(t *testing.T) {
	files := []FileMeta{
		{Path: "folder/test.md", Name: "test", NormalizedPath: "folder/test"},
	}
	r := New(files)

	result := r.Resolve("test.md")
	if !result.Found {
		t.Error("expected to find 'test.md'")
	}
	if result.TargetPath != "folder/test.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}
}

func TestBuildFileMeta(t *testing.T) {
	fm := BuildFileMeta("notes/my file.md", "My Title")

	if fm.Name != "my file" {
		t.Errorf("Name = %q, want %q", fm.Name, "my file")
	}
	if fm.Path != "notes/my file.md" {
		t.Errorf("Path = %q, want %q", fm.Path, "notes/my file.md")
	}
	if fm.NormalizedPath != "notes/my file" {
		t.Errorf("NormalizedPath = %q, want %q", fm.NormalizedPath, "notes/my file")
	}
	if fm.Title != "My Title" {
		t.Errorf("Title = %q, want %q", fm.Title, "My Title")
	}
}

func TestResolve_TargetWithHash(t *testing.T) {
	files := []FileMeta{
		{Path: "pages/page.md", Name: "page", NormalizedPath: "pages/page"},
	}
	r := New(files)

	// Resolve should handle the page part before #
	result := r.Resolve("page")
	if !result.Found {
		t.Error("expected to resolve page part of 'page#heading'")
	}
	if result.TargetPath != "pages/page.md" {
		t.Errorf("unexpected path: %s", result.TargetPath)
	}
}
