package parser

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// ParsedDocument is the result of parsing a markdown file.
type ParsedDocument struct {
	Path        string         `json:"path"`
	Title       string         `json:"title"`
	Frontmatter map[string]any `json:"frontmatter"`
	Content     string         `json:"-"`
	PlainText   string         `json:"plainText"`
	HTML        string         `json:"html"`
	Headings    []Heading      `json:"headings"`
	Links       []WikiLink     `json:"links"`
	Tags        []string       `json:"tags"`
	Blocks      []BlockRef     `json:"blocks"`
	Backlinks   interface{}    `json:"backlinks"`
}

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.NewTable(),
		extension.Strikethrough,
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
				chromahtml.PreventSurroundingPre(true),
			),
			highlighting.WithWrapperRenderer(func(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
				lang, hasLang := c.Language()
				if entering {
					if hasLang {
						w.WriteString(`<pre class="chroma language-` + string(lang) + `"><code>`)
					} else {
						w.WriteString(`<pre><code>`)
					}
				} else {
					w.WriteString(`</code></pre>`)
				}
			}),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

// ParseDocument parses a markdown file content into a ParsedDocument.
func ParseDocument(content string, filePath string) (*ParsedDocument, error) {
	fm, body, err := ExtractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Determine title
	title := extractTitle(fm, body, filePath)

	// Render to HTML
	var htmlBuf strings.Builder
	if err := md.Convert([]byte(body), &htmlBuf); err != nil {
		return nil, err
	}

	// Extract metadata
	headings := ExtractHeadings(body)
	links := ExtractWikiLinks(body)
	tags := extractAllTags(fm, body)
	blocks := ExtractBlockRefs(body)

	// Fix heading IDs: goldmark generates IDs differently from our slugify.
	// Replace them so TOC links always match.
	htmlStr := fixHeadingIDs(htmlBuf.String(), headings)

	// Make external links open in new tab
	htmlStr = addTargetBlank(htmlStr)

	// Process Obsidian callouts
	htmlStr = ProcessCallouts(htmlStr)

	// Process block references: add IDs, remove ^markers
	htmlStr = ProcessBlockRefsInHTML(htmlStr)

	return &ParsedDocument{
		Path:        filePath,
		Title:       title,
		Frontmatter: fm,
		Content:     body,
		PlainText:   stripHTML(htmlStr),
		HTML:        htmlStr,
		Headings:    headings,
		Links:       links,
		Tags:        tags,
		Blocks:      blocks,
	}, nil
}

func extractTitle(fm map[string]any, body string, filePath string) string {
	// 1. Frontmatter title
	if t, ok := fm["title"].(string); ok && t != "" {
		return t
	}
	// 2. First heading
	headings := ExtractHeadings(body)
	if len(headings) > 0 {
		return headings[0].Text
	}
	// 3. Filename (without extension)
	parts := strings.Split(filePath, "/")
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".md")
	return name
}

func extractAllTags(fm map[string]any, body string) []string {
	seen := map[string]bool{}
	var tags []string

	// From frontmatter
	if fmTags, ok := fm["tags"].([]interface{}); ok {
		for _, t := range fmTags {
			if s, ok := t.(string); ok && !seen[s] {
				seen[s] = true
				tags = append(tags, s)
			}
		}
	}

	// From body
	inlineTags := ExtractInlineTags(body)
	for _, t := range inlineTags {
		if !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	}

	return tags
}

// stripHTML removes HTML tags to produce plain text.
func stripHTML(html string) string {
	var b strings.Builder
	b.Grow(utf8.RuneCountInString(html))

	inTag := false
	for _, ch := range html {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteRune(ch)
		}
	}

	return strings.TrimSpace(b.String())
}

// fixHeadingIDs replaces goldmark-generated heading IDs with our own slugs
// so that TOC links in the sidebar always point to the correct element.
func fixHeadingIDs(htmlStr string, headings []Heading) string {
	for i, h := range headings {
		tagPrefix := fmt.Sprintf(`<h%d `, h.Level)
		searchFrom := 0

		for {
			tagStart := strings.Index(htmlStr[searchFrom:], tagPrefix)
			if tagStart == -1 {
				break
			}
			tagStart += searchFrom

			tagEnd := strings.Index(htmlStr[tagStart:], ">")
			if tagEnd == -1 {
				break
			}
			tagEnd += tagStart

			// Get text content
			closeTag := fmt.Sprintf(`</h%d>`, h.Level)
			closeIdx := strings.Index(htmlStr[tagEnd:], closeTag)
			if closeIdx == -1 {
				searchFrom = tagEnd + 1
				continue
			}
			inner := htmlStr[tagEnd+1 : tagEnd+closeIdx]
			plain := strings.TrimSpace(stripHTML(inner))

			if plain == h.Text {
				oldTag := htmlStr[tagStart : tagEnd+1]
				newTag := setAttr(oldTag, "id", h.Slug)
				// Only replace this specific occurrence
				htmlStr = htmlStr[:tagStart] + newTag + htmlStr[tagEnd+1:]
				// Use a marker to avoid re-matching this heading for the next heading with same text
				_ = i
				break
			}
			searchFrom = tagEnd + 1
		}
	}
	return htmlStr
}

// setAttr sets an attribute value in an HTML opening tag.
func setAttr(tag, attr, value string) string {
	prefix := attr + `="`
	idx := strings.Index(tag, prefix)
	if idx != -1 {
		// Replace existing
		valStart := idx + len(prefix)
		valEnd := strings.Index(tag[valStart:], `"`)
		if valEnd != -1 {
			return tag[:valStart] + value + tag[valStart+valEnd:]
		}
	}
	// Add new attribute before closing >
	return tag[:len(tag)-1] + ` ` + prefix + value + `">`
}

// addTargetBlank adds target="_blank" rel="noopener" to all <a> tags that are
// external links (href starts with http), excluding wikilinks.
func addTargetBlank(htmlStr string) string {
	// Simple approach: replace <a href="http with <a target="_blank" rel="noopener noreferrer" href="http
	// This won't touch wikilinks since they have href="/api/note..."
	return strings.ReplaceAll(htmlStr, `<a href="http`,
		`<a target="_blank" rel="noopener noreferrer" href="http`)
}
