package parser

import (
	"regexp"
	"strings"
)

// BlockRef represents an Obsidian block reference (^blockid).
type BlockRef struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Line int    `json:"line"`
}

var blockIDRe = regexp.MustCompile(`(?:^|\s)\^([a-zA-Z0-9_-]+)\s*$`)

// ExtractBlockRefs finds all block references (^blockid) in raw markdown text.
// It skips block IDs inside code blocks.
func ExtractBlockRefs(text string) []BlockRef {
	lines := strings.Split(text, "\n")
	var blocks []BlockRef
	inCodeBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Skip inline code lines (heuristic: if line starts with backtick)
		if strings.HasPrefix(trimmed, "`") && strings.Count(trimmed, "`") >= 2 {
			continue
		}

		m := blockIDRe.FindStringSubmatch(trimmed)
		if m != nil {
			// Extract text without the block ID marker
			cleanText := strings.TrimSpace(blockIDRe.ReplaceAllString(trimmed, ""))
			blocks = append(blocks, BlockRef{
				ID:   m[1],
				Text: cleanText,
				Line: i + 1,
			})
		}
	}

	return blocks
}

// ProcessBlockRefsInHTML transforms rendered HTML to add block IDs and remove markers.
// It finds ^blockid patterns in text nodes, removes the marker, and adds id attributes
// to the containing element.
func ProcessBlockRefsInHTML(html string) string {
	// Pattern to find ^blockid at end of text within HTML elements
	blockHTMLRe := regexp.MustCompile(`\^([a-zA-Z0-9_-]+)(\s*</)`)

	result := html

	// Find all block ID markers and their positions
	matches := blockHTMLRe.FindAllStringSubmatchIndex(result, -1)
	if len(matches) == 0 {
		return result
	}

	// Process in reverse order to maintain positions
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		blockID := result[m[2]:m[3]]
		fullStart := m[0]
		fullEnd := m[1]
		closingPart := result[m[4]:m[5]] // the "</" part

		// Find the opening tag before this position
		tagStart := findOpeningTagStart(result, fullStart)
		if tagStart == -1 {
			continue
		}

		// Find the end of the opening tag
		tagEnd := strings.Index(result[tagStart:], ">")
		if tagEnd == -1 {
			continue
		}
		tagEnd += tagStart + 1

		// Check if already has an id
		tagContent := result[tagStart:tagEnd]
		if strings.Contains(tagContent, `id="`) {
			// Already has id, just remove the marker
			result = result[:fullStart] + closingPart + result[fullEnd:]
			continue
		}

		// Add id attribute and remove marker
		newTag := result[tagStart:tagEnd-1] + ` id="block-` + blockID + `">`
		newContent := closingPart

		result = result[:tagStart] + newTag + result[tagEnd:fullStart] + newContent + result[fullEnd:]
	}

	return result
}

// findOpeningTagStart searches backwards from pos to find the start of the containing HTML tag.
func findOpeningTagStart(html string, pos int) int {
	// Walk backwards looking for '<'
	for i := pos - 1; i >= 0; i-- {
		if html[i] == '>' {
			// We passed a closing tag, keep going
			continue
		}
		if html[i] == '<' {
			// Make sure it's an opening tag (not </)
			if i+1 < len(html) && html[i+1] != '/' {
				return i
			}
		}
		// If we hit a newline, stop
		if html[i] == '\n' && i > 0 && html[i-1] == '\n' {
			break
		}
	}
	return -1
}
