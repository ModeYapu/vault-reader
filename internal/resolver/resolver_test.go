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
