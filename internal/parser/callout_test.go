package parser

import (
	"strings"
	"testing"
)

func TestProcessCallouts_Basic(t *testing.T) {
	// ProcessCallouts operates on rendered HTML (blockquote tags), not raw markdown
	input := "<blockquote>\n<p>[!note]\nThis is a note callout.</p>\n</blockquote>\n"
	result := ProcessCallouts(input)
	if !strings.Contains(result, "callout") {
		t.Errorf("expected callout class in output, got: %s", result)
	}
	if !strings.Contains(result, "This is a note callout") {
		t.Error("expected callout content in output")
	}
}

func TestProcessCallouts_WithTitle(t *testing.T) {
	input := "<blockquote>\n<p>[!warning] Attention</p>\n<p>Warning content</p>\n</blockquote>\n"
	result := ProcessCallouts(input)
	if !strings.Contains(result, "Attention") {
		t.Error("expected title 'Attention' in output")
	}
	if !strings.Contains(result, "Warning content") {
		t.Error("expected content in output")
	}
}

func TestProcessCallouts_Foldable(t *testing.T) {
	input := "<blockquote>\n<p>[!tip]- Collapsed</p>\n<p>Hidden content</p>\n</blockquote>\n"
	result := ProcessCallouts(input)
	if !strings.Contains(result, "callout-foldable") {
		t.Errorf("expected foldable class, got: %s", result)
	}
	if !strings.Contains(result, "callout-collapsed") {
		t.Error("expected collapsed class")
	}
}

func TestProcessCallouts_Expanded(t *testing.T) {
	input := "<blockquote>\n<p>[!example]+ Expanded</p>\n<p>Visible content</p>\n</blockquote>\n"
	result := ProcessCallouts(input)
	if !strings.Contains(result, "callout-foldable") {
		t.Errorf("expected foldable class, got: %s", result)
	}
	if strings.Contains(result, "callout-collapsed") {
		t.Error("should not be collapsed")
	}
}

func TestProcessCallouts_MultipleTypes(t *testing.T) {
	types := []string{"note", "warning", "tip", "important", "info", "success", "danger", "bug", "example", "quote"}
	for _, ct := range types {
		input := "<blockquote>\n<p>[!" + ct + "]\nContent</p>\n</blockquote>\n"
		result := ProcessCallouts(input)
		if !strings.Contains(result, "callout") {
			t.Errorf("callout type %s: expected callout class, got: %s", ct, result)
		}
		if !strings.Contains(result, "callout-"+ct) {
			t.Errorf("callout type %s: expected callout-%s class", ct, ct)
		}
	}
}

func TestProcessCallouts_NoCallout(t *testing.T) {
	input := "<blockquote>\n<p>Regular blockquote. Not a callout.</p>\n</blockquote>\n"
	result := ProcessCallouts(input)
	t.Logf("Result: %q", result)
	// Check if "callout" appears (not "blockquote")
	if strings.Contains(result, "callout") && !strings.Contains(result, "<blockquote>") {
		t.Error("regular blockquote should not become callout")
	}
}

func TestProcessCallouts_InternalMarkdown(t *testing.T) {
	input := "<blockquote>\n<p>[!note]\nSome <strong>bold</strong> and <code>code</code> text</p>\n</blockquote>\n"
	result := ProcessCallouts(input)
	if !strings.Contains(result, "<strong>bold</strong>") {
		t.Error("expected internal HTML preserved")
	}
}
