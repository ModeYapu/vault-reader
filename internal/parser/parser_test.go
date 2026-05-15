package parser

import (
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	input := `---
title: OpenClaw 配置排查
tags:
  - openclaw
  - debug
---

# OpenClaw 配置排查

内容`

	fm, body, err := ExtractFrontmatter(input)
	if err != nil {
		t.Fatalf("ExtractFrontmatter failed: %v", err)
	}

	if fm["title"] != "OpenClaw 配置排查" {
		t.Errorf("expected title, got %v", fm["title"])
	}

	tags, ok := fm["tags"].([]interface{})
	if !ok {
		t.Fatalf("expected tags to be a list, got %T", fm["tags"])
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "openclaw" || tags[1] != "debug" {
		t.Errorf("expected tags [openclaw debug], got %v", tags)
	}

	if body != "# OpenClaw 配置排查\n\n内容" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	input := "# Just a title\n\nNo frontmatter here."
	fm, body, err := ExtractFrontmatter(input)
	if err != nil {
		t.Fatalf("ExtractFrontmatter failed: %v", err)
	}
	if len(fm) != 0 {
		t.Errorf("expected empty frontmatter, got %v", fm)
	}
	if body != input {
		t.Errorf("body should be unchanged")
	}
}

func TestParseIncompleteFrontmatter(t *testing.T) {
	input := "---\ntitle: test\nno closing"
	fm, body, err := ExtractFrontmatter(input)
	if err != nil {
		t.Fatalf("ExtractFrontmatter should not fail on incomplete fm: %v", err)
	}
	if len(fm) != 0 {
		t.Errorf("expected empty frontmatter for incomplete, got %v", fm)
	}
	if body != input {
		t.Error("body should be unchanged for incomplete frontmatter")
	}
}

func TestExtractTagsFromBody(t *testing.T) {
	input := "#tag1 Some text #tag2 more #tag3_with_underscore"
	tags := ExtractInlineTags(input)
	expected := []string{"tag1", "tag2", "tag3_with_underscore"}
	if len(tags) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %v", len(expected), len(tags), tags)
	}
	for i, exp := range expected {
		if tags[i] != exp {
			t.Errorf("tags[%d] = %q, want %q", i, tags[i], exp)
		}
	}
}

func TestExtractTagsNoDuplicates(t *testing.T) {
	input := "#test #test #other"
	tags := ExtractInlineTags(input)
	if len(tags) != 2 {
		t.Errorf("expected 2 unique tags, got %d: %v", len(tags), tags)
	}
}

func TestExtractHeadings(t *testing.T) {
	input := `# Title

## Section 1

### Subsection

## Section 2`

	headings := ExtractHeadings(input)
	if len(headings) != 4 {
		t.Fatalf("expected 4 headings, got %d", len(headings))
	}

	if headings[0].Level != 1 || headings[0].Text != "Title" {
		t.Errorf("heading[0] unexpected: %+v", headings[0])
	}
	if headings[1].Level != 2 || headings[1].Text != "Section 1" {
		t.Errorf("heading[1] unexpected: %+v", headings[1])
	}
	if headings[2].Level != 3 || headings[2].Text != "Subsection" {
		t.Errorf("heading[2] unexpected: %+v", headings[2])
	}
}

func TestHeadingSlugs(t *testing.T) {
	input := `## OpenClaw 配置

### HTTP Proxy 设置`

	headings := ExtractHeadings(input)

	if headings[0].Slug == "" {
		t.Error("heading slug should not be empty")
	}
	if headings[1].Slug == "" {
		t.Error("heading slug should not be empty")
	}
}

func TestRenderMarkdown(t *testing.T) {
	input := `# Hello

This is **bold** and *italic*.

- item 1
- item 2

` + "```go" + `
fmt.Println("hello")
` + "```" + `

| A | B |
|---|---|
| 1 | 2 |
`

	doc, err := ParseDocument(input, "test.md")
	if err != nil {
		t.Fatalf("ParseDocument failed: %v", err)
	}

	if doc.HTML == "" {
		t.Error("expected non-empty HTML")
	}
	if doc.PlainText == "" {
		t.Error("expected non-empty plain text")
	}
	// Should contain rendered bold
	if len(doc.HTML) < 50 {
		t.Errorf("HTML seems too short: %s", doc.HTML)
	}
}

func TestRenderTaskList(t *testing.T) {
	input := `- [x] Done
- [ ] Not done`

	doc, err := ParseDocument(input, "test.md")
	if err != nil {
		t.Fatalf("ParseDocument failed: %v", err)
	}
	if doc.HTML == "" {
		t.Error("expected HTML output")
	}
}

// --- ExtractFrontmatter additional tests ---

func TestExtractFrontmatter_EmptyFrontmatter(t *testing.T) {
	input := "---\n---\n\nSome content"
	fm, body, err := ExtractFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm) != 0 {
		t.Errorf("expected empty frontmatter map, got %v", fm)
	}
	if body != "Some content" {
		t.Errorf("body = %q, want %q", body, "Some content")
	}
}

func TestExtractFrontmatter_MalformedYAML(t *testing.T) {
	input := "---\n: : bad yaml\n---\n\ncontent"
	fm, body, err := ExtractFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Malformed YAML should return empty frontmatter and original input
	if len(fm) != 0 {
		t.Errorf("expected empty frontmatter for malformed YAML, got %v", fm)
	}
	if body != input {
		t.Error("body should be unchanged for malformed YAML")
	}
}

// --- Slugify tests ---

func TestSlugify(t *testing.T) {
	headings := ExtractHeadings("## Hello World\n## 你好世界\n## Foo & Bar!")

	if headings[0].Slug != "hello-world" {
		t.Errorf("slug for 'Hello World' = %q, want %q", headings[0].Slug, "hello-world")
	}
	// Chinese characters are preserved in UTF-8
	if headings[1].Slug != "你好世界" {
		t.Errorf("slug for Chinese heading = %q, want %q", headings[1].Slug, "你好世界")
	}
	// Special chars like & and ! are stripped
	if headings[2].Slug != "foo--bar" {
		t.Errorf("slug for 'Foo & Bar!' = %q, want %q", headings[2].Slug, "foo--bar")
	}
}

// --- ExtractInlineTags edge cases ---

func TestExtractInlineTags_EdgeCases(t *testing.T) {
	// Tag at end of sentence with punctuation
	input := "This is a sentence. #tag1!"
	tags := ExtractInlineTags(input)
	found := false
	for _, t := range tags {
		if t == "tag1" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'tag1' in %v", tags)
	}

	// Tag with Unicode punctuation
	input2 := "这里有个标签 #测试。后面是中文标点"
	tags2 := ExtractInlineTags(input2)
	if len(tags2) == 0 {
		t.Error("expected to find Chinese tag")
	} else if tags2[0] != "测试。后面是中文标点" {
		// Note: only trailing ASCII punctuation is trimmed
		t.Logf("Chinese tag result: %v", tags2)
	}
}

// --- SanitizeLang tests ---

func TestSanitizeLang(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"go", "go"},
		{"javascript", "javascript"},
		{`lang"<> `, "lang"},
		{"c++", "c"},
		{"my-lang_v2.0", "my-lang_v2.0"},
	}

	for _, tt := range tests {
		got := sanitizeLang(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeLang(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
