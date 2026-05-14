package parser

import (
	"regexp"
	"strings"
)

// CalloutType maps callout type names to their visual properties.
type CalloutType struct {
	Icon  string
	Color string
}

// Known callout types with their default icons and colors.
var calloutTypes = map[string]CalloutType{
	"note":      {"ℹ️", "#448aff"},
	"abstract":  {"📝", "#00b0ff"},
	"summary":   {"📝", "#00b0ff"},
	"tldr":      {"📝", "#00b0ff"},
	"info":      {"ℹ️", "#00b0ff"},
	"todo":      {"☑️", "#448aff"},
	"tip":       {"💡", "#00c853"},
	"hint":      {"💡", "#00c853"},
	"important": {"🔥", "#00c853"},
	"success":   {"✅", "#00c853"},
	"check":     {"✅", "#00c853"},
	"done":      {"✅", "#00c853"},
	"question":  {"❓", "#ff9100"},
	"help":      {"❓", "#ff9100"},
	"faq":       {"❓", "#ff9100"},
	"warning":   {"⚠️", "#ff9100"},
	"caution":   {"⚠️", "#ff9100"},
	"attention": {"⚠️", "#ff9100"},
	"failure":   {"❌", "#ff5252"},
	"fail":      {"❌", "#ff5252"},
	"missing":   {"❌", "#ff5252"},
	"danger":    {"⚡", "#ff5252"},
	"error":     {"⚡", "#ff5252"},
	"bug":       {"🐛", "#ff5252"},
	"example":   {"📋", "#7c4dff"},
	"quote":     {"💬", "#9e9e9e"},
	"cite":      {"💬", "#9e9e9e"},
}

var calloutFirstLineRe = regexp.MustCompile(`^\[!(\w+)\]([+-]?)(.*)`)

// ProcessCallouts transforms Obsidian callout blockquotes in rendered HTML.
func ProcessCallouts(html string) string {
	var result strings.Builder
	i := 0

	for i < len(html) {
		bqStart := strings.Index(html[i:], "<blockquote>")
		if bqStart == -1 {
			result.WriteString(html[i:])
			break
		}
		bqStart += i
		result.WriteString(html[i:bqStart])

		bqEnd := findCloseBlockquote(html, bqStart+12)
		if bqEnd == -1 {
			result.WriteString(html[bqStart:])
			break
		}

		inner := html[bqStart+12 : bqEnd]
		callout := parseCalloutInner(inner)
		if callout != nil {
			result.WriteString(renderCallout(callout))
		} else {
			result.WriteString("<blockquote>")
			result.WriteString(inner)
			result.WriteString("</blockquote>")
		}

		i = bqEnd + 13
	}

	return result.String()
}

type callout struct {
	cType     string
	title     string
	content   string
	foldable  bool
	collapsed bool
}

func parseCalloutInner(inner string) *callout {
	stripped := strings.TrimLeft(inner, " \n\r")

	// Must start with <p> or be plain text starting with [!
	if !strings.HasPrefix(stripped, "<p>") && !strings.HasPrefix(stripped, "[!") {
		return nil
	}

	// Get the text content after optional <p>
	text := stripped
	if strings.HasPrefix(text, "<p>") {
		text = text[3:]
	}

	text = strings.TrimLeft(text, " ")
	if !strings.HasPrefix(text, "[!") {
		return nil
	}

	// Match [!type][+/-][title...]
	m := calloutFirstLineRe.FindStringSubmatch(text)
	if m == nil {
		return nil
	}

	cType := strings.ToLower(m[1])
	if _, ok := calloutTypes[cType]; !ok {
		cType = "note"
	}

	foldable := m[2] == "+" || m[2] == "-"
	collapsed := m[2] == "-"
	titleText := strings.TrimSpace(m[3])

	ct := calloutTypes[cType]

	// Title is only the text before </p> or <br> — strip any HTML tags that leaked in.
	if idx := strings.Index(titleText, "</p>"); idx != -1 {
		titleText = titleText[:idx]
	}
	// Clean up title: remove trailing <br /> and <br>
	titleText = strings.TrimRight(titleText, " ")
	for _, br := range []string{"<br />", "<br/>", "<br>"} {
		titleText = strings.TrimSuffix(titleText, br)
	}
	titleText = strings.TrimSpace(titleText)
	if titleText == "" {
		titleText = ct.Icon + " " + strings.Title(cType)
	}

	// Determine where the first line ends
	// The rest of the text after [!type]... includes the title and possibly body
	restAfterMatch := text[len(m[0]):]

	// Find end of first "line" — could be </p> or \n
	var bodyHTML string

	newlineIdx := strings.Index(restAfterMatch, "\n")
	pCloseIdx := strings.Index(restAfterMatch, "</p>")

	if newlineIdx != -1 && (pCloseIdx == -1 || newlineIdx < pCloseIdx) {
		// There's a newline within the <p> — split title and body
		if titleText == "" {
			// [!type] alone on first line, body starts after newline
			titleText = ct.Icon + " " + strings.Title(cType)
		}
		afterNewline := restAfterMatch[newlineIdx+1:]
		// Body is everything from the newline to </p>
		if pCloseIdx != -1 {
			bodyHTML = "<p>" + strings.TrimSpace(afterNewline[:pCloseIdx-(newlineIdx+1)]) + "</p>"
			// Plus any content after </p>
			afterBody := strings.TrimLeft(restAfterMatch[pCloseIdx+4:], "\n\r ")
			if afterBody != "" {
				bodyHTML += afterBody
			}
		} else {
			bodyHTML = "<p>" + strings.TrimSpace(afterNewline) + "</p>"
		}
	} else {
		// No newline in first <p> — title and body are separate
		if titleText == "" {
			titleText = ct.Icon + " " + strings.Title(cType)
		}
		if pCloseIdx != -1 {
			afterBody := strings.TrimLeft(restAfterMatch[pCloseIdx+4:], "\n\r ")
			bodyHTML = afterBody
		}
	}

	return &callout{
		cType:     cType,
		title:     titleText,
		content:   bodyHTML,
		foldable:  foldable,
		collapsed: collapsed,
	}
}

func renderCallout(c *callout) string {
	ct, ok := calloutTypes[c.cType]
	if !ok {
		ct = CalloutType{Icon: "ℹ️", Color: "#448aff"}
	}

	var sb strings.Builder

	classes := "callout callout-" + c.cType
	if c.foldable {
		classes += " callout-foldable"
		if c.collapsed {
			classes += " callout-collapsed"
		}
	}

	sb.WriteString(`<div class="`)
	sb.WriteString(classes)
	sb.WriteString(`" style="--callout-color:`)
	sb.WriteString(ct.Color)
	sb.WriteString(`"`)

	if c.foldable {
		sb.WriteString(` data-foldable="true"`)
		if c.collapsed {
			sb.WriteString(` data-collapsed="true"`)
		}
	}
	sb.WriteString(`>`)

	// Title
	sb.WriteString(`<div class="callout-title">`)
	sb.WriteString(`<span class="callout-icon">`)
	sb.WriteString(ct.Icon)
	sb.WriteString(`</span>`)
	sb.WriteString(`<span class="callout-title-text">`)
	sb.WriteString(escapeHTML(c.title))
	sb.WriteString(`</span>`)
	if c.foldable {
		sb.WriteString(`<span class="callout-fold-icon">`)
		sb.WriteString(`<svg width="12" height="12" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 6l4 4 4-4"/></svg>`)
		sb.WriteString(`</span>`)
	}
	sb.WriteString(`</div>`)

	// Content
	sb.WriteString(`<div class="callout-content">`)
	sb.WriteString(c.content)
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)
	return sb.String()
}

func findCloseBlockquote(html string, start int) int {
	depth := 1
	i := start
	for i < len(html) && depth > 0 {
		o := strings.Index(html[i:], "<blockquote>")
		c := strings.Index(html[i:], "</blockquote>")
		if o != -1 && (c == -1 || o < c) {
			depth++
			i += o + 12
		} else if c != -1 {
			depth--
			if depth == 0 {
				return i + c
			}
			i += c + 13
		} else {
			break
		}
	}
	return -1
}
