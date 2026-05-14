package indexer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vault-reader/internal/parser"
	"vault-reader/internal/resolver"
	"vault-reader/internal/scanner"

	_ "modernc.org/sqlite"
)

// Indexer manages the SQLite index for a vault.
type Indexer struct {
	db       *sql.DB
	vaultDir string
	resolver *resolver.Resolver
}

// New creates or opens an indexer database at the given path.
func New(dbPath, vaultDir string) (*Indexer, error) {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &Indexer{
		db:       db,
		vaultDir: vaultDir,
	}, nil
}

// Close closes the database.
func (ix *Indexer) Close() error {
	return ix.db.Close()
}

// DB returns the underlying database connection.
func (ix *Indexer) DB() *sql.DB {
	return ix.db
}

// FullIndex scans the entire vault and rebuilds the index.
func (ix *Indexer) FullIndex() error {
	start := time.Now()
	slog.Info("starting full index", "vault", ix.vaultDir)

	files, err := scanner.Scan(ix.vaultDir)
	if err != nil {
		return fmt.Errorf("scan vault: %w", err)
	}

	// First pass: parse all markdown files to get titles and aliases
	metas := make([]resolver.FileMeta, 0, len(files))
	for _, f := range files {
		if !f.IsMarkdown {
			continue
		}
		meta := ix.parseFileMeta(f)
		metas = append(metas, meta)
	}
	ix.resolver = resolver.New(metas)

	// Start transaction
	tx, err := ix.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing data
	for _, table := range []string{"files", "links", "tags", "headings", "properties", "blocks", "file_fts"} {
		if _, err := tx.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("clear table %s: %w", table, err)
		}
	}

	// Second pass: index all markdown files
	now := time.Now().Unix()
	for _, f := range files {
		if !f.IsMarkdown {
			continue
		}
		if err := ix.indexFile(tx, f, now); err != nil {
			slog.Error("failed to index file", "path", f.Path, "error", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	elapsed := time.Since(start)
	slog.Info("full index complete", "files", len(metas), "elapsed", elapsed)
	return nil
}

func (ix *Indexer) parseFileMeta(f scanner.VaultFile) resolver.FileMeta {
	fullPath := filepath.Join(ix.vaultDir, filepath.FromSlash(f.Path))
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return resolver.BuildFileMeta(f.Path, "")
	}
	doc, err := parser.ParseDocument(string(content), f.Path)
	if err != nil {
		return resolver.BuildFileMeta(f.Path, "")
	}

	// Extract aliases from frontmatter
	var aliases []string
	if a, ok := doc.Frontmatter["aliases"]; ok {
		switch v := a.(type) {
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					aliases = append(aliases, s)
				}
			}
		case string:
			aliases = []string{v}
		}
	}

	meta := resolver.BuildFileMeta(f.Path, doc.Title)
	meta.Aliases = aliases
	return meta
}

func (ix *Indexer) indexFile(tx *sql.Tx, f scanner.VaultFile, now int64) error {
	fullPath := filepath.Join(ix.vaultDir, filepath.FromSlash(f.Path))
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	doc, err := parser.ParseDocument(string(content), f.Path)
	if err != nil {
		return err
	}

	// Serialize frontmatter
	fmJSON, _ := json.Marshal(doc.Frontmatter)

	// Insert into files
	_, err = tx.Exec(`INSERT INTO files (path, title, ext, size, mtime, content, html, frontmatter_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.Path, doc.Title, f.Ext, f.Size, f.ModTime.Unix(),
		doc.Content, doc.HTML, string(fmJSON), now, now)
	if err != nil {
		return fmt.Errorf("insert file: %w", err)
	}

	// Insert into FTS
	_, err = tx.Exec(`INSERT INTO file_fts (title, path, content) VALUES (?, ?, ?)`,
		doc.Title, f.Path, doc.PlainText)
	if err != nil {
		return fmt.Errorf("insert fts: %w", err)
	}

	// Insert links
	for _, link := range doc.Links {
		resolvedPath := ""
		resolved := 0
		if !link.IsAsset {
			result := ix.resolver.Resolve(link.Target)
			if result.Found {
				resolvedPath = result.TargetPath
				resolved = 1
			}
		}

		isEmbed := 0
		if link.IsEmbed {
			isEmbed = 1
		}
		isAsset := 0
		if link.IsAsset {
			isAsset = 1
		}

		_, err = tx.Exec(`INSERT INTO links (from_path, raw, target, target_path, alias, heading, is_embed, is_asset, resolved)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			f.Path, link.Raw, link.Target, resolvedPath, link.Alias, link.Heading,
			isEmbed, isAsset, resolved)
		if err != nil {
			return fmt.Errorf("insert link: %w", err)
		}
	}

	// Insert tags
	for _, tag := range doc.Tags {
		_, err = tx.Exec(`INSERT INTO tags (file_path, tag) VALUES (?, ?)`,
			f.Path, tag)
		if err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}

	// Insert headings
	for _, h := range doc.Headings {
		_, err = tx.Exec(`INSERT INTO headings (file_path, level, text, slug) VALUES (?, ?, ?, ?)`,
			f.Path, h.Level, h.Text, h.Slug)
		if err != nil {
			return fmt.Errorf("insert heading: %w", err)
		}
	}

	// Insert properties from frontmatter
	for key, val := range doc.Frontmatter {
		if err := insertProperty(tx, f.Path, key, val); err != nil {
			return fmt.Errorf("insert property %s: %w", key, err)
		}
	}

		// Insert blocks
		for _, b := range doc.Blocks {
			_, err = tx.Exec(`INSERT INTO blocks (file_path, block_id, text, line) VALUES (?, ?, ?, ?)`,
				f.Path, b.ID, b.Text, b.Line)
			if err != nil {
				return fmt.Errorf("insert block: %w", err)
			}
		}

	return nil
}

// insertProperty inserts a frontmatter key-value pair into the properties table.
func insertProperty(tx *sql.Tx, filePath, key string, val any) error {
	switch v := val.(type) {
	case []interface{}:
		for _, item := range v {
			s := formatPropertyValue(item)
			if _, err := tx.Exec(`INSERT INTO properties (file_path, key, value, value_type) VALUES (?, ?, ?, 'array')`,
				filePath, key, s); err != nil {
				return err
			}
		}
	case string:
		if _, err := tx.Exec(`INSERT INTO properties (file_path, key, value, value_type) VALUES (?, ?, ?, 'string')`,
			filePath, key, v); err != nil {
			return err
		}
	case float64:
		// YAML floats that are whole numbers (like dates parsed as timestamps) — format nicely
		if v == math.Trunc(v) && !math.IsInf(v, 0) && !math.IsNaN(v) && math.Abs(v) < 1e15 {
			if _, err := tx.Exec(`INSERT INTO properties (file_path, key, value, value_type) VALUES (?, ?, ?, 'number')`,
				filePath, key, fmt.Sprintf("%d", int64(v))); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(`INSERT INTO properties (file_path, key, value, value_type) VALUES (?, ?, ?, 'number')`,
				filePath, key, fmt.Sprintf("%v", v)); err != nil {
				return err
			}
		}
	case bool:
		if _, err := tx.Exec(`INSERT INTO properties (file_path, key, value, value_type) VALUES (?, ?, ?, 'boolean')`,
			filePath, key, fmt.Sprintf("%v", v)); err != nil {
			return err
		}
	case time.Time:
		// YAML parses bare dates like 2026-05-14 as time.Time
		dateStr := v.Format("2006-01-02")
		if _, err := tx.Exec(`INSERT INTO properties (file_path, key, value, value_type) VALUES (?, ?, ?, 'date')`,
			filePath, key, dateStr); err != nil {
			return err
		}
	default:
		if v != nil {
			if _, err := tx.Exec(`INSERT INTO properties (file_path, key, value, value_type) VALUES (?, ?, ?, 'string')`,
				filePath, key, formatPropertyValue(v)); err != nil {
				return err
			}
		}
	}
	return nil
}

// formatPropertyValue formats a property value for storage.
func formatPropertyValue(v any) string {
	switch val := v.(type) {
	case time.Time:
		return val.Format("2006-01-02")
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Resolve resolves a wikilink target using the resolver.
func (ix *Indexer) Resolve(target string) resolver.ResolveResult {
	if ix.resolver == nil {
		return resolver.ResolveResult{}
	}
	return ix.resolver.Resolve(target)
}

// Search performs a full-text search.
func (ix *Indexer) Search(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}

	// Escape special FTS5 characters
	cleanQuery := cleanFTSQuery(query)

	rows, err := ix.db.Query(`
		SELECT f.path, f.title, snippet(file_fts, 2, '<mark>', '</mark>', '...', 10) as snippet
		FROM file_fts
		JOIN files f ON file_fts.path = f.path
		WHERE file_fts MATCH ?
		ORDER BY rank
		LIMIT ?`,
		cleanQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Path, &r.Title, &r.Snippet); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// SearchResult represents a single search result.
type SearchResult struct {
	Path    string `json:"path"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// GetBacklinks returns all files that link to the given path.
func (ix *Indexer) GetBacklinks(path string) ([]Backlink, error) {
	rows, err := ix.db.Query(`
		SELECT l.from_path, f.title, l.raw, l.alias
		FROM links l
		LEFT JOIN files f ON l.from_path = f.path
		WHERE l.target_path = ?
		ORDER BY l.from_path`,
		path)
	if err != nil {
		return nil, fmt.Errorf("backlinks query: %w", err)
	}
	defer rows.Close()

	var links []Backlink
	for rows.Next() {
		var bl Backlink
		if err := rows.Scan(&bl.FromPath, &bl.Title, &bl.Raw, &bl.Alias); err != nil {
			return nil, err
		}
		links = append(links, bl)
	}
	return links, rows.Err()
}

// Backlink represents a backlink reference.
type Backlink struct {
	FromPath string `json:"fromPath"`
	Title    string `json:"title"`
	Raw      string `json:"raw"`
	Alias    string `json:"alias,omitempty"`
}

// GetTags returns all tags with their counts.
func (ix *Indexer) GetTags() ([]TagCount, error) {
	rows, err := ix.db.Query(`
		SELECT tag, COUNT(DISTINCT file_path) as count
		FROM tags
		GROUP BY tag
		ORDER BY count DESC, tag ASC`)
	if err != nil {
		return nil, fmt.Errorf("tags query: %w", err)
	}
	defer rows.Close()

	var tags []TagCount
	for rows.Next() {
		var t TagCount
		if err := rows.Scan(&t.Tag, &t.Count); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// TagCount represents a tag and how many files use it.
type TagCount struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// GetFilesByTag returns file paths that have the given tag.
func (ix *Indexer) GetFilesByTag(tag string) ([]TagFile, error) {
	rows, err := ix.db.Query(`
		SELECT t.file_path, f.title
		FROM tags t
		LEFT JOIN files f ON t.file_path = f.path
		WHERE t.tag = ?
		ORDER BY t.file_path`,
		tag)
	if err != nil {
		return nil, fmt.Errorf("tag files query: %w", err)
	}
	defer rows.Close()

	var files []TagFile
	for rows.Next() {
		var tf TagFile
		if err := rows.Scan(&tf.Path, &tf.Title); err != nil {
			return nil, err
		}
		files = append(files, tf)
	}
	return files, rows.Err()
}

// TagFile represents a file associated with a tag.
type TagFile struct {
	Path  string `json:"path"`
	Title string `json:"title"`
}

// GetProperties returns all properties for a given file path.
func (ix *Indexer) GetProperties(filePath string) ([]Property, error) {
	rows, err := ix.db.Query(`
		SELECT key, value, value_type
		FROM properties
		WHERE file_path = ?
		ORDER BY id`,
		filePath)
	if err != nil {
		return nil, fmt.Errorf("properties query: %w", err)
	}
	defer rows.Close()

	var props []Property
	for rows.Next() {
		var p Property
		if err := rows.Scan(&p.Key, &p.Value, &p.ValueType); err != nil {
			return nil, err
		}
		props = append(props, p)
	}
	return props, rows.Err()
}

// Property represents a single frontmatter property.
type Property struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	ValueType string `json:"valueType"`
}

// FilterByProperty returns files that match a property key-value pair.
func (ix *Indexer) FilterByProperty(key, value string) ([]TagFile, error) {
	rows, err := ix.db.Query(`
		SELECT DISTINCT p.file_path, f.title
		FROM properties p
		LEFT JOIN files f ON p.file_path = f.path
		WHERE p.key = ? AND p.value = ?
		ORDER BY p.file_path`,
		key, value)
	if err != nil {
		return nil, fmt.Errorf("filter query: %w", err)
	}
	defer rows.Close()

	var files []TagFile
	for rows.Next() {
		var tf TagFile
		if err := rows.Scan(&tf.Path, &tf.Title); err != nil {
			return nil, err
		}
		files = append(files, tf)
	}
	return files, rows.Err()
}

// BlockRef represents a block reference stored in the index.
type BlockRef struct {
	Path     string `json:"path"`
	BlockID  string `json:"blockId"`
	Text     string `json:"text"`
	Line     int    `json:"line"`
}

// GetBlock returns the block reference for a given block ID.
func (ix *Indexer) GetBlock(blockID string) (*BlockRef, error) {
	row := ix.db.QueryRow(`
		SELECT b.file_path, b.block_id, b.text, b.line
		FROM blocks b
		WHERE b.block_id = ?`,
		blockID)

	var br BlockRef
	if err := row.Scan(&br.Path, &br.BlockID, &br.Text, &br.Line); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("block query: %w", err)
	}
	return &br, nil
}

// GetBlocksByFile returns all block references for a given file path.
func (ix *Indexer) GetBlocksByFile(filePath string) ([]BlockRef, error) {
	rows, err := ix.db.Query(`
		SELECT file_path, block_id, text, line
		FROM blocks
		WHERE file_path = ?
		ORDER BY line`,
		filePath)
	if err != nil {
		return nil, fmt.Errorf("blocks query: %w", err)
	}
	defer rows.Close()

	var blocks []BlockRef
	for rows.Next() {
		var b BlockRef
		if err := rows.Scan(&b.Path, &b.BlockID, &b.Text, &b.Line); err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

// cleanFTSQuery sanitizes a search query for FTS5 MATCH.
func cleanFTSQuery(query string) string {
	// Remove characters that have special meaning in FTS5
	replacer := strings.NewReplacer(
		`"`, ``, `{`, ``, `}`, ``, `(`, ``, `)`, ``,
		`:`, ``, `^`, ``, `+`, ``, `*`, ``,
		`~`, ``, `AND`, ``, `OR`, ``, `NOT`, ``,
	)
	cleaned := replacer.Replace(query)
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return `""`
	}

	// Wrap each word in quotes for exact matching (better for CJK)
	words := strings.Fields(cleaned)
	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = `"` + w + `"`
	}
	return strings.Join(quoted, " OR ")
}
