package parser

import (
	"strings"
	"testing"
)

func TestParseDocumentWithMermaid(t *testing.T) {
	input := "# Mermaid Test\n\n" + "```mermaid\ngraph TD\n  A --> B\n```" + "\n\nSome text after."

	doc, err := ParseDocument(input, "mermaid-test.md")
	if err != nil {
		t.Fatalf("ParseDocument failed: %v", err)
	}

	// Mermaid code block should have language-mermaid class
	if !strings.Contains(doc.HTML, "language-mermaid") {
		t.Errorf("expected language-mermaid class in HTML, got: %s", doc.HTML)
	}
	if !strings.Contains(doc.HTML, "graph TD") {
		t.Errorf("mermaid content should be preserved, got: %s", doc.HTML)
	}
}

func TestParseDocumentWithMermaidSequenceDiagram(t *testing.T) {
	input := "```mermaid\nsequenceDiagram\n  Alice->>Bob: Hello\n```"

	doc, err := ParseDocument(input, "test.md")
	if err != nil {
		t.Fatalf("ParseDocument failed: %v", err)
	}

	if !strings.Contains(doc.HTML, "language-mermaid") {
		t.Errorf("expected language-mermaid class, got: %s", doc.HTML)
	}
	if !strings.Contains(doc.HTML, "sequenceDiagram") {
		t.Errorf("sequenceDiagram content should be preserved, got: %s", doc.HTML)
	}
}

func TestParseDocumentMermaidAndNormalCode(t *testing.T) {
	input := "```go\nfmt.Println(\"hello\")\n```\n\n```mermaid\ngraph TD\n  A --> B\n```"

	doc, err := ParseDocument(input, "test.md")
	if err != nil {
		t.Fatalf("ParseDocument failed: %v", err)
	}

	// Both should have their language classes
	if !strings.Contains(doc.HTML, "language-go") {
		t.Error("expected language-go class")
	}
	if !strings.Contains(doc.HTML, "language-mermaid") {
		t.Error("expected language-mermaid class")
	}
}
