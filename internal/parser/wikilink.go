package parser

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// WikiLink represents an Obsidian-style wikilink.
type WikiLink struct {
	Raw     string `json:"raw"`
	Target  string `json:"target"`
	Alias   string `json:"alias"`
	Heading string `json:"heading"`
	IsEmbed bool   `json:"isEmbed"`
	IsAsset bool   `json:"isAsset"`
}

var wikilinkRe = regexp.MustCompile(`(!?)\[\[([^\]]+)\]\]`)

var assetExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".svg": true, ".webp": true, ".pdf": true, ".txt": true,
	".mp3": true, ".mp4": true, ".wav": true,
}

// ExtractWikiLinks finds all wikilinks in the input text.
func ExtractWikiLinks(text string) []WikiLink {
	matches := wikilinkRe.FindAllStringSubmatch(text, -1)
	links := make([]WikiLink, 0, len(matches))

	for _, m := range matches {
		raw := m[0]
		isEmbed := m[1] == "!"
		inner := m[2]

		link := WikiLink{
			Raw:     raw,
			IsEmbed: isEmbed,
		}

		// Split by | for alias
		parts := strings.SplitN(inner, "|", 2)
		mainPart := parts[0]
		if len(parts) == 2 {
			link.Alias = parts[1]
		}

		// Split by # for heading
		hashParts := strings.SplitN(mainPart, "#", 2)
		link.Target = hashParts[0]
		if len(hashParts) == 2 {
			link.Heading = hashParts[1]
		}

		link.IsAsset = isAssetExt(link.Target)

		links = append(links, link)
	}

	return links
}

func isAssetExt(target string) bool {
	lower := strings.ToLower(target)
	for ext := range assetExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// ResolveFunc takes a wikilink target and returns (resolvedPath, found).
type ResolveFunc func(target string) (resolvedPath string, found bool)

// RenderWikiLinksInHTML replaces [[wikilinks]] in already-rendered HTML with proper HTML elements.
// It escapes wikilinks that were rendered inside <code> or <pre> blocks.
func RenderWikiLinksInHTML(html string, resolve ResolveFunc, prefix string) string {
	// Collect positions inside <code> and <pre> blocks to skip
	skipRanges := findCodeRanges(html)

	var result strings.Builder
	result.Grow(len(html))

	lastEnd := 0
	matches := wikilinkRe.FindAllStringSubmatchIndex(html, -1)

	for _, m := range matches {
		fullStart := m[0]
		fullEnd := m[1]

		// Skip if inside a code block
		if isInSkipRange(fullStart, skipRanges) {
			continue
		}

		result.WriteString(html[lastEnd:fullStart])

		isEmbed := m[2] != -1 && html[m[2]:m[3]] == "!"
		inner := html[m[4]:m[5]]

		// Parse the link
		link := parseWikiLinkInner(inner, isEmbed)

		// Render the replacement
		replacement := renderLinkHTML(link, resolve, prefix)
		result.WriteString(replacement)

		lastEnd = fullEnd
	}

	result.WriteString(html[lastEnd:])
	return result.String()
}

func parseWikiLinkInner(inner string, isEmbed bool) WikiLink {
	link := WikiLink{IsEmbed: isEmbed}

	parts := strings.SplitN(inner, "|", 2)
	mainPart := parts[0]
	if len(parts) == 2 {
		link.Alias = parts[1]
	}

	hashParts := strings.SplitN(mainPart, "#", 2)
	link.Target = hashParts[0]
	if len(hashParts) == 2 {
		link.Heading = hashParts[1]
	}

	link.IsAsset = isAssetExt(link.Target)

	// Reconstruct raw for display
	link.Raw = "[[" + inner + "]]"
	if isEmbed {
		link.Raw = "!" + link.Raw
	}

	return link
}

func renderLinkHTML(link WikiLink, resolve ResolveFunc, prefix string) string {
	if link.IsEmbed {
		return renderEmbedHTML(link, resolve, prefix)
	}

	// Regular wikilink
	displayText := link.Alias
	if displayText == "" {
		if link.Heading != "" {
			displayText = link.Target + " > " + link.Heading
		} else {
			displayText = link.Target
		}
	}

	resolvedPath, found := resolve(link.Target)
	if !found {
		return `<span class="broken-link" title="未找到: ` + escapeHTML(link.Target) + `">` + escapeHTML(displayText) + `</span>`
	}

	href := prefix + "/api/note?path=" + encodeURIComponent(resolvedPath)
	if link.Heading != "" {
		if strings.HasPrefix(link.Heading, "^") {
			// Block reference: ^abc123 -> #block-abc123
			blockID := link.Heading[1:]
			href += "#block-" + encodeURIComponent(blockID)
		} else {
			href += "#" + encodeURIComponent(link.Heading)
		}
	}

	return `<a class="wikilink" href="` + href + `" data-path="` + escapeHTML(resolvedPath) + `">` + escapeHTML(displayText) + `</a>`
}

func renderEmbedHTML(link WikiLink, resolve ResolveFunc, prefix string) string {
	if link.IsAsset {
		assetPath, found := resolve(link.Target)
		if !found {
			// Try the target as-is (might be a direct relative path)
			assetPath = link.Target
		}
		src := prefix + "/assets?path=" + encodeURIComponent(assetPath)
		lowerTarget := strings.ToLower(link.Target)

		if strings.HasSuffix(lowerTarget, ".png") || strings.HasSuffix(lowerTarget, ".jpg") ||
			strings.HasSuffix(lowerTarget, ".jpeg") || strings.HasSuffix(lowerTarget, ".gif") ||
			strings.HasSuffix(lowerTarget, ".svg") || strings.HasSuffix(lowerTarget, ".webp") {
			return `<div class="embed-image"><img src="` + src + `" alt="` + escapeHTML(link.Target) + `" loading="lazy"></div>`
		}

		if strings.HasSuffix(lowerTarget, ".pdf") {
			return `<div class="embed-pdf"><iframe src="` + src + `" width="100%" height="600px"></iframe></div>`
		}

		// Other files: download link
		return `<div class="embed-file"><a href="` + src + `">📎 ` + escapeHTML(link.Target) + `</a></div>`
	}

	// Embed a note (![[Page]])
	resolvedPath, found := resolve(link.Target)
	if !found {
		return `<div class="embed-broken">未找到: ` + escapeHTML(link.Target) + `</div>`
	}

	return `<div class="embed-note" data-path="` + escapeHTML(resolvedPath) + `"><a href="` + prefix + `/api/note?path=` + encodeURIComponent(resolvedPath) + `">🔗 ` + escapeHTML(link.Target) + `</a></div>`
}

// findCodeRanges returns byte ranges [start, end) of content inside <code>...</code> and <pre>...</pre>.
func findCodeRanges(html string) [][2]int {
	var ranges [][2]int
	lower := strings.ToLower(html)

	// Find <code>...</code>
	findTagRanges(lower, "<code", "</code>", &ranges)
	// Find <pre>...</pre>
	findTagRanges(lower, "<pre", "</pre>", &ranges)

	return ranges
}

func findTagRanges(lower string, openTag, closeTag string, ranges *[][2]int) {
	idx := 0
	for {
		start := strings.Index(lower[idx:], openTag)
		if start == -1 {
			break
		}
		start += idx
		// Find end of opening tag
		tagEnd := strings.Index(lower[start:], ">")
		if tagEnd == -1 {
			break
		}
		contentStart := start + tagEnd + 1

		end := strings.Index(lower[contentStart:], closeTag)
		if end == -1 {
			break
		}
		end += contentStart

		*ranges = append(*ranges, [2]int{contentStart, end})
		idx = end + len(closeTag)
	}
}

func isInSkipRange(pos int, ranges [][2]int) bool {
	for _, r := range ranges {
		if pos >= r[0] && pos < r[1] {
			return true
		}
	}
	return false
}

func escapeHTML(s string) string {
	return html.EscapeString(s)
}

func encodeURIComponent(s string) string {
	var result strings.Builder
	result.Grow(len(s) * 2)
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_' || ch == '.' || ch == '!' || ch == '~' || ch == '*' || ch == '\'' || ch == '(' || ch == ')' {
			result.WriteRune(ch)
		} else if ch == '/' {
			result.WriteRune(ch)
		} else {
			// Percent-encode
			for _, b := range string(ch) {
				result.WriteString(fmt.Sprintf("%%%02X", b))
			}
		}
	}
	return result.String()
}
