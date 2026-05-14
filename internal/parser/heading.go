package parser

import (
	"regexp"
	"strings"
)

// Heading represents a markdown heading.
type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	Slug  string `json:"slug"`
}

var headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// ExtractHeadings extracts all headings from markdown text.
func ExtractHeadings(text string) []Heading {
	var headings []Heading
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		m := headingRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		level := len(m[1])
		text := strings.TrimSpace(m[2])
		slug := slugify(text)
		headings = append(headings, Heading{
			Level: level,
			Text:  text,
			Slug:  slug,
		})
	}

	return headings
}

func slugify(text string) string {
	// Simple slug: lowercase, replace spaces with hyphens, remove special chars
	s := strings.ToLower(text)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric, non-hyphen, non-CJK characters would be complex
	// For now, keep it simple
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			result = append(result, ch)
		} else if ch >= 0x80 {
			// Keep UTF-8 bytes for Chinese etc
			result = append(result, s[i])
		}
	}
	return string(result)
}
