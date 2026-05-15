package parser

import (
	"strings"
	"testing"
)

func TestExtractBlockRefs(t *testing.T) {
	input := `# Title

Some paragraph text. ^abc123

Another paragraph.

## Section

Important conclusion here. ^def456
`
	blocks := ExtractBlockRefs(input)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	if blocks[0].ID != "abc123" {
		t.Errorf("blocks[0].ID = %q, want abc123", blocks[0].ID)
	}
	if blocks[0].Line != 3 {
		t.Errorf("blocks[0].Line = %d, want 3", blocks[0].Line)
	}

	if blocks[1].ID != "def456" {
		t.Errorf("blocks[1].ID = %q, want def456", blocks[1].ID)
	}
	if blocks[1].Line != 9 {
		t.Errorf("blocks[1].Line = %d, want 9", blocks[1].Line)
	}
}

func TestExtractBlockRefsEmpty(t *testing.T) {
	input := "No block refs here.\nJust normal text."
	blocks := ExtractBlockRefs(input)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(blocks))
	}
}

func TestExtractBlockRefsWithHyphen(t *testing.T) {
	input := "Some text ^my-block-id"
	blocks := ExtractBlockRefs(input)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].ID != "my-block-id" {
		t.Errorf("ID = %q, want my-block-id", blocks[0].ID)
	}
}

func TestExtractBlockRefsInCodeBlock(t *testing.T) {
	input := "Normal text ^abc123\n\n```\ncode ^notablock\n```\n"
	blocks := ExtractBlockRefs(input)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block (skip code block), got %d", len(blocks))
	}
	if blocks[0].ID != "abc123" {
		t.Errorf("ID = %q, want abc123", blocks[0].ID)
	}
}

func TestExtractBlockRefsBlockIDOnlyLine(t *testing.T) {
	input := "Some paragraph\n\n^standalone\n"
	blocks := ExtractBlockRefs(input)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].ID != "standalone" {
		t.Errorf("ID = %q, want standalone", blocks[0].ID)
	}
}

func TestProcessBlockRefsInHTML(t *testing.T) {
	html := `<p>Some paragraph text. ^abc123</p><p>Another paragraph.</p>`
	result := ProcessBlockRefsInHTML(html)

	if !strings.Contains(result, `id="block-abc123"`) {
		t.Errorf("expected block id attribute, got: %s", result)
	}
	if strings.Contains(result, "^abc123") {
		t.Errorf("block ID marker should be removed, got: %s", result)
	}
	if !strings.Contains(result, "Some paragraph text.") {
		t.Errorf("original text should be preserved, got: %s", result)
	}
}

func TestProcessBlockRefsNoMatch(t *testing.T) {
	html := `<p>No block refs here.</p>`
	result := ProcessBlockRefsInHTML(html)
	if result != html {
		t.Errorf("expected unchanged HTML, got: %s", result)
	}
}

func TestProcessBlockRefsMultiple(t *testing.T) {
	html := `<p>Text A ^aaa111</p><p>Text B</p><p>Text C ^ccc333</p>`
	result := ProcessBlockRefsInHTML(html)

	if !strings.Contains(result, `id="block-aaa111"`) {
		t.Error("expected block-aaa111 id")
	}
	if !strings.Contains(result, `id="block-ccc333"`) {
		t.Error("expected block-ccc333 id")
	}
	if strings.Contains(result, "^aaa111") || strings.Contains(result, "^ccc333") {
		t.Errorf("markers should be removed, got: %s", result)
	}
}

func TestWikiLinkBlockRefHeading(t *testing.T) {
	links := ExtractWikiLinks("[[OpenClaw#^abc123]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "OpenClaw" {
		t.Errorf("target = %q, want OpenClaw", l.Target)
	}
	if l.Heading != "^abc123" {
		t.Errorf("heading = %q, want ^abc123", l.Heading)
	}
}

func TestWikiLinkBlockRefWithAlias(t *testing.T) {
	links := ExtractWikiLinks("[[OpenClaw#^abc123|重要结论]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Heading != "^abc123" {
		t.Errorf("heading = %q, want ^abc123", l.Heading)
	}
	if l.Alias != "重要结论" {
		t.Errorf("alias = %q, want 重要结论", l.Alias)
	}
}

func TestRenderWikiLinkBlockRef(t *testing.T) {
	html := `[[OpenClaw#^abc123]]`
	resolve := func(target string) (string, bool) {
		if target == "OpenClaw" {
			return "20_Debug/OpenClaw.md", true
		}
		return "", false
	}

	result := RenderWikiLinksInHTML(html, resolve, "")
	if !strings.Contains(result, "#block-abc123") {
		t.Errorf("expected #block-abc123 in href, got: %s", result)
	}
	if !strings.Contains(result, `data-path="20_Debug/OpenClaw.md"`) {
		t.Errorf("expected data-path, got: %s", result)
	}
}

func TestRenderWikiLinkBlockRefBroken(t *testing.T) {
	html := `[[NotFound#^xyz789]]`
	resolve := func(target string) (string, bool) {
		return "", false
	}

	result := RenderWikiLinksInHTML(html, resolve, "")
	if !strings.Contains(result, "broken-link") {
		t.Errorf("expected broken-link class, got: %s", result)
	}
}
