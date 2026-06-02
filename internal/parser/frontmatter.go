package parser

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// ExtractFrontmatter splits input into frontmatter map and body text.
func ExtractFrontmatter(input string) (map[string]any, string, error) {
	input = strings.TrimLeft(input, "\n\r")

	if !strings.HasPrefix(input, "---") {
		return map[string]any{}, input, nil
	}

	// Find closing ---
	afterFirst := input[3:]
	endIdx := strings.Index(afterFirst, "\n---")
	if endIdx < 0 {
		// No closing delimiter found, treat as regular content
		return map[string]any{}, input, nil
	}

	fmText := strings.TrimSpace(afterFirst[:endIdx])
	body := strings.TrimLeft(afterFirst[endIdx+4:], "\n\r")

	var fm map[string]any
	if err := yaml.Unmarshal([]byte(fmText), &fm); err != nil {
		// If YAML parsing fails, return empty frontmatter
		return map[string]any{}, input, nil
	}

	if fm == nil {
		fm = map[string]any{}
	}

	return fm, body, nil
}

// ExtractInlineTags finds #tag occurrences in text.
// Excludes heading lines (# heading) and only matches inline tags.
func ExtractInlineTags(text string) []string {
	seen := map[string]bool{}
	var tags []string

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		// Skip heading lines: one or more # followed by a space
		trimmed := strings.TrimSpace(line)
		if isHeadingLine(trimmed) {
			continue
		}
		// Find inline tags: #word (must start with letter or CJK)
		words := strings.Fields(line)
		for _, w := range words {
			if strings.HasPrefix(w, "#") && len(w) > 1 {
				tag := strings.TrimPrefix(w, "#")
				// Clean trailing punctuation
				tag = strings.TrimRight(tag, ".,;:!?)")
				// Skip empty or purely punctuation tags
				if tag == "" || !isValidTag(tag) {
					continue
				}
				if !seen[tag] {
					seen[tag] = true
					tags = append(tags, tag)
				}
			}
		}
	}

	return tags
}

// isHeadingLine checks if a line is a markdown heading (starts with 1-6 # followed by space).
func isHeadingLine(line string) bool {
	if len(line) == 0 || line[0] != '#' {
		return false
	}
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	return i <= 6 && i < len(line) && line[i] == ' '
}

// isValidTag checks if a tag contains at least one letter or CJK character.
func isValidTag(tag string) bool {
	for _, ch := range tag {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch >= 0x4e00 {
			return true
		}
	}
	return false
}
