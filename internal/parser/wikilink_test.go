package parser

import (
	"testing"
)

func TestParseSimpleWikiLink(t *testing.T) {
	links := ExtractWikiLinks("[[OpenClaw]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "OpenClaw" {
		t.Errorf("target = %q, want OpenClaw", l.Target)
	}
	if l.Alias != "" {
		t.Errorf("alias should be empty, got %q", l.Alias)
	}
	if l.Heading != "" {
		t.Errorf("heading should be empty, got %q", l.Heading)
	}
	if l.IsEmbed {
		t.Error("should not be embed")
	}
	if l.Raw != "[[OpenClaw]]" {
		t.Errorf("raw = %q", l.Raw)
	}
}

func TestParseWikiLinkWithAlias(t *testing.T) {
	links := ExtractWikiLinks("[[OpenClaw|排查文档]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "OpenClaw" {
		t.Errorf("target = %q, want OpenClaw", l.Target)
	}
	if l.Alias != "排查文档" {
		t.Errorf("alias = %q, want 排查文档", l.Alias)
	}
}

func TestParseWikiLinkWithHeading(t *testing.T) {
	links := ExtractWikiLinks("[[OpenClaw#代理配置]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "OpenClaw" {
		t.Errorf("target = %q, want OpenClaw", l.Target)
	}
	if l.Heading != "代理配置" {
		t.Errorf("heading = %q, want 代理配置", l.Heading)
	}
}

func TestParseWikiLinkWithHeadingAndAlias(t *testing.T) {
	links := ExtractWikiLinks("[[OpenClaw#代理配置|代理说明]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "OpenClaw" {
		t.Errorf("target = %q", l.Target)
	}
	if l.Heading != "代理配置" {
		t.Errorf("heading = %q, want 代理配置", l.Heading)
	}
	if l.Alias != "代理说明" {
		t.Errorf("alias = %q, want 代理说明", l.Alias)
	}
}

func TestParseEmbedImage(t *testing.T) {
	links := ExtractWikiLinks("![[架构图.png]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "架构图.png" {
		t.Errorf("target = %q", l.Target)
	}
	if !l.IsEmbed {
		t.Error("should be embed")
	}
	if !l.IsAsset {
		t.Error("png should be marked as asset")
	}
}

func TestParseEmbedFolderImage(t *testing.T) {
	links := ExtractWikiLinks("![[folder/image.png]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "folder/image.png" {
		t.Errorf("target = %q", l.Target)
	}
	if !l.IsEmbed {
		t.Error("should be embed")
	}
}

func TestParseEmbedPage(t *testing.T) {
	links := ExtractWikiLinks("![[SomePage]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	l := links[0]
	if l.Target != "SomePage" {
		t.Errorf("target = %q", l.Target)
	}
	if !l.IsEmbed {
		t.Error("should be embed")
	}
	if l.IsAsset {
		t.Error("should not be asset")
	}
}

func TestParseFolderLink(t *testing.T) {
	links := ExtractWikiLinks("[[folder/Page]]")
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Target != "folder/Page" {
		t.Errorf("target = %q", links[0].Target)
	}
}

func TestParseMultipleLinks(t *testing.T) {
	input := "See [[A]] and [[B|alias]] and ![[C.png]]"
	links := ExtractWikiLinks(input)
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d", len(links))
	}
}

func TestParseNoLinks(t *testing.T) {
	links := ExtractWikiLinks("just regular text")
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

func TestIsAsset(t *testing.T) {
	tests := []struct {
		target  string
		isAsset bool
	}{
		{"image.png", true},
		{"image.jpg", true},
		{"image.jpeg", true},
		{"image.gif", true},
		{"image.svg", true},
		{"image.webp", true},
		{"doc.pdf", true},
		{"SomePage", false},
		{"notes/test", false},
		{"file.txt", true},
	}

	for _, tt := range tests {
		got := isAssetExt(tt.target)
		if got != tt.isAsset {
			t.Errorf("isAssetExt(%q) = %v, want %v", tt.target, got, tt.isAsset)
		}
	}
}
