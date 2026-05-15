package indexer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"vault-reader/internal/parser"
	"vault-reader/internal/resolver"
	"vault-reader/internal/scanner"

	_ "modernc.org/sqlite"
)

// Indexer manages the SQLite index for a vault.
type Indexer struct {
	db         *sql.DB
	vaultDir   string
	resolver   *resolver.Resolver
	resolverMu sync.RWMutex
	indexMu    sync.Mutex // serializes FullIndex calls
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

// GetFileList returns all indexed file paths for building the tree view.
func (ix *Indexer) GetFileList() ([]string, error) {
	rows, err := ix.db.Query(`SELECT path FROM files ORDER BY path`)
	if err != nil {
		return nil, fmt.Errorf("file list query: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

// GetCanvasFiles returns canvas file paths from the database.
func (ix *Indexer) GetCanvasFiles() ([]TagFile, error) {
	rows, err := ix.db.Query(`SELECT path, title FROM files WHERE ext = '.canvas' ORDER BY path`)
	if err != nil {
		return nil, fmt.Errorf("canvas files query: %w", err)
	}
	defer rows.Close()

	var files []TagFile
	for rows.Next() {
		var f TagFile
		if err := rows.Scan(&f.Path, &f.Title); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// FullIndex scans the entire vault and rebuilds the index.
func (ix *Indexer) FullIndex() error {
	ix.indexMu.Lock()
	defer ix.indexMu.Unlock()

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

	// Start transaction
	tx, err := ix.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update resolver after transaction starts, keeping resolver and DB in sync
	ix.resolverMu.Lock()
	ix.resolver = resolver.New(metas)
	ix.resolverMu.Unlock()

	// Clear existing data
	for _, table := range []string{"files", "links", "tags", "headings", "properties", "blocks", "file_fts"} {
		if _, err := tx.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("clear table %s: %w", table, err)
		}
	}

	// Capture resolver snapshot for the indexing loop
	ix.resolverMu.RLock()
	localResolver := ix.resolver
	ix.resolverMu.RUnlock()

	// Second pass: index all markdown files
	now := time.Now().Unix()
	for _, f := range files {
		if !f.IsMarkdown {
			continue
		}
		if err := ix.indexFileWithResolver(tx, localResolver, f, now); err != nil {
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

func (ix *Indexer) indexFileWithResolver(tx *sql.Tx, res *resolver.Resolver, f scanner.VaultFile, now int64) error {
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
			result := res.Resolve(link.Target)
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
	ix.resolverMu.RLock()
	r := ix.resolver
	ix.resolverMu.RUnlock()
	if r == nil {
		return resolver.ResolveResult{}
	}
	return r.Resolve(target)
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

// TagTreeNode represents a node in the hierarchical tag tree.
type TagTreeNode struct {
	Name     string         `json:"name"`
	FullName string         `json:"fullName"`
	Count    int            `json:"count"`
	Children []TagTreeNode  `json:"children,omitempty"`
}

// GetTagTree returns all tags organized into a tree structure based on "/" separators.
func (ix *Indexer) GetTagTree() ([]TagTreeNode, error) {
	rows, err := ix.db.Query(`
		SELECT tag, COUNT(DISTINCT file_path) as count
		FROM tags
		GROUP BY tag
		ORDER BY tag ASC`)
	if err != nil {
		return nil, fmt.Errorf("tag tree query: %w", err)
	}
	defer rows.Close()

	type internalNode struct {
		count    int
		children map[string]*internalNode
	}

	root := make(map[string]*internalNode)

	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, err
		}

		parts := strings.Split(tag, "/")
		current := root
		for i, part := range parts {
			if _, ok := current[part]; !ok {
				current[part] = &internalNode{children: make(map[string]*internalNode)}
			}
			node := current[part]
			// If this is the last part, the count belongs to this node
			if i == len(parts)-1 {
				node.count = count
			}
			current = node.children
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert internal map to sorted []TagTreeNode
	var buildTree func(m map[string]*internalNode, prefix string) []TagTreeNode
	buildTree = func(m map[string]*internalNode, prefix string) []TagTreeNode {
		if len(m) == 0 {
			return nil
		}
		// Collect keys for deterministic ordering
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		nodes := make([]TagTreeNode, 0, len(m))
		for _, k := range keys {
			child := m[k]
			fullName := k
			if prefix != "" {
				fullName = prefix + "/" + k
			}
			node := TagTreeNode{
				Name:     k,
				FullName: fullName,
				Count:    child.count,
				Children: buildTree(child.children, fullName),
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	return buildTree(root, ""), nil
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

// GraphNode represents a node in the graph view.
type GraphNode struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
	Group string `json:"group"`
}

// GraphEdge represents a directed edge between two graph nodes.
type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// GetGraph returns graph data (nodes and edges) based on resolved links.
// folder, tag, path can be empty (no filter). depth and maxNodes limit the result.
func (ix *Indexer) GetGraph(folder, tag, path string, depth, maxNodes int) ([]GraphNode, []GraphEdge, error) {
	if maxNodes <= 0 {
		maxNodes = 500
	}
	if depth <= 0 {
		depth = 1
	}

	// Build base node set from files
	nodeQuery := `SELECT path, title FROM files WHERE 1=1`
	nodeArgs := []interface{}{}

	if folder != "" {
		nodeQuery += ` AND (path LIKE ? || '/%' OR path = ?)`
		nodeArgs = append(nodeArgs, escapeLike(folder), folder)
	}
	if tag != "" {
		nodeQuery += ` AND path IN (SELECT file_path FROM tags WHERE tag = ?)`
		nodeArgs = append(nodeArgs, tag)
	}
	if path != "" {
		nodeQuery += ` AND path = ?`
		nodeArgs = append(nodeArgs, path)
	}
	nodeQuery += ` ORDER BY path LIMIT ?`
	nodeArgs = append(nodeArgs, maxNodes)

	rows, err := ix.db.Query(nodeQuery, nodeArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("graph nodes query: %w", err)
	}
	defer rows.Close()

	nodeMap := make(map[string]GraphNode) // path -> node
	for rows.Next() {
		var p, title string
		if err := rows.Scan(&p, &title); err != nil {
			return nil, nil, err
		}
		parts := strings.Split(p, "/")
		group := ""
		if len(parts) > 1 {
			group = parts[0]
		}
		nodeMap[p] = GraphNode{
			ID:    p,
			Title: title,
			Path:  p,
			Group: group,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// If path-based local graph, expand to include linked neighbors
	if path != "" && depth > 1 {
		ix.expandGraphNeighbors(nodeMap, path, depth-1, maxNodes)
	}

	// Build edges from resolved links
	paths := make([]string, 0, len(nodeMap))
	for p := range nodeMap {
		paths = append(paths, p)
	}

	var edges []GraphEdge
	if len(paths) > 0 {
		// Filter edges by node set in SQL to avoid loading all links
		placeholders := strings.Repeat("?,", len(paths))
		placeholders = placeholders[:len(placeholders)-1]
		edgeArgs := make([]interface{}, len(paths)*2)
		for i, p := range paths {
			edgeArgs[i] = p
			edgeArgs[len(paths)+i] = p
		}
		edgeRows, err := ix.db.Query(`
			SELECT DISTINCT from_path, target_path
			FROM links
			WHERE resolved = 1 AND target_path != ''
			AND from_path != target_path
			AND from_path IN (`+placeholders+`)
			AND target_path IN (`+placeholders+`)`,
			edgeArgs...)
		if err != nil {
			return nil, nil, fmt.Errorf("graph edges query: %w", err)
		}
		defer edgeRows.Close()

		seen := make(map[string]bool)
		for edgeRows.Next() {
			var src, tgt string
			if err := edgeRows.Scan(&src, &tgt); err != nil {
				return nil, nil, err
			}
			key := src + "->" + tgt
			if !seen[key] {
				seen[key] = true
				edges = append(edges, GraphEdge{Source: src, Target: tgt})
			}
		}
		if err := edgeRows.Err(); err != nil {
			return nil, nil, err
		}
	}

	// Convert map to sorted slice
	nodes := make([]GraphNode, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}

	return nodes, edges, nil
}

// DashboardData holds all sections for the dashboard page.
type DashboardData struct {
	Recent []TagFile  `json:"recent"`
	Inbox  []TagFile  `json:"inbox"`
	Active []TagFile  `json:"active"`
	Debug  []TagFile  `json:"debug"`
	Tags   []TagCount `json:"tags"`
	Canvas []TagFile  `json:"canvas"`
}

// GetDashboard returns aggregated dashboard data.
func (ix *Indexer) GetDashboard() (*DashboardData, error) {
	data := &DashboardData{}

	// Recent: files ordered by mtime desc, limit 10
	recentRows, err := ix.db.Query(`
		SELECT path, title FROM files
		ORDER BY mtime DESC LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("dashboard recent: %w", err)
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var f TagFile
		if err := recentRows.Scan(&f.Path, &f.Title); err != nil {
			slog.Warn("dashboard: scan recent error", "error", err)
			continue
		}
		data.Recent = append(data.Recent, f)
	}

	// Inbox: files in 00_Inbox
	inboxRows, err := ix.db.Query(`
		SELECT path, title FROM files
		WHERE path LIKE '00_Inbox/%'
		ORDER BY mtime DESC LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("dashboard inbox: %w", err)
	}
	defer inboxRows.Close()
	for inboxRows.Next() {
		var f TagFile
		if err := inboxRows.Scan(&f.Path, &f.Title); err != nil {
			slog.Warn("dashboard: scan inbox error", "error", err)
			continue
		}
		data.Inbox = append(data.Inbox, f)
	}

	// Active: status = active
	activeRows, err := ix.db.Query(`
		SELECT DISTINCT p.file_path, f.title
		FROM properties p
		LEFT JOIN files f ON p.file_path = f.path
		WHERE p.key = 'status' AND p.value = 'active'
		ORDER BY p.file_path LIMIT 20`)
	if err != nil {
		return nil, fmt.Errorf("dashboard active: %w", err)
	}
	defer activeRows.Close()
	for activeRows.Next() {
		var f TagFile
		if err := activeRows.Scan(&f.Path, &f.Title); err != nil {
			slog.Warn("dashboard: scan active error", "error", err)
			continue
		}
		data.Active = append(data.Active, f)
	}

	// Debug: type = debug-note
	debugRows, err := ix.db.Query(`
		SELECT DISTINCT p.file_path, f.title
		FROM properties p
		LEFT JOIN files f ON p.file_path = f.path
		WHERE p.key = 'type' AND p.value = 'debug-note'
		ORDER BY p.file_path LIMIT 20`)
	if err != nil {
		return nil, fmt.Errorf("dashboard debug: %w", err)
	}
	defer debugRows.Close()
	for debugRows.Next() {
		var f TagFile
		if err := debugRows.Scan(&f.Path, &f.Title); err != nil {
			slog.Warn("dashboard: scan debug error", "error", err)
			continue
		}
		data.Debug = append(data.Debug, f)
	}

	// Tags: top tags
	data.Tags, err = ix.GetTags()
	if err != nil {
		slog.Warn("dashboard: failed to get tags", "error", err)
	}
	if len(data.Tags) > 10 {
		data.Tags = data.Tags[:10]
	}

	// Canvas: canvas files from database
	canvasFiles, err := ix.GetCanvasFiles()
	if err == nil {
		data.Canvas = canvasFiles
	}

	return data, nil
}

// QueryResultRow represents a single row from a vault query.
type QueryResultRow struct {
	Path   string            `json:"path"`
	Title  string            `json:"title"`
	Fields map[string]string `json:"fields"`
}

// ExecuteVaultQuery executes a parsed vault query and returns results.
func (ix *Indexer) ExecuteVaultQuery(q *parser.VaultQuery) ([]QueryResultRow, error) {
	// Build SQL query
	args := []interface{}{}
	whereClauses := []string{}

	if q.From != "" {
		whereClauses = append(whereClauses, "(f.path LIKE ? || '/%' OR f.path = ?)")
		args = append(args, escapeLike(q.From), q.From)
	}

	// Process where filters via properties
	for key, value := range q.Where {
		whereClauses = append(whereClauses, `f.path IN (SELECT file_path FROM properties WHERE key = ? AND value = ?)`)
		args = append(args, key, value)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	orderBy := "f.mtime DESC"
	if q.Sort != "" {
		dir := "DESC"
		if q.Order == "asc" {
			dir = "ASC"
		}
		// Sort by property or mtime
		if q.Sort == "updated" || q.Sort == "mtime" {
			orderBy = "f.mtime " + dir
		} else if q.Sort == "title" {
			orderBy = "f.title " + dir
		} else {
			orderBy = "f.mtime " + dir
		}
	}

	limit := 20
	if q.Limit > 0 {
		limit = q.Limit
	}

	query := `SELECT f.path, f.title FROM files f ` + whereSQL + ` ORDER BY ` + orderBy + ` LIMIT ?`
	args = append(args, limit)

	rows, err := ix.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("vault query: %w", err)
	}
	defer rows.Close()

	var results []QueryResultRow
	// Collect paths for batch property lookup
	var paths []string
	for rows.Next() {
		var path, title string
		if err := rows.Scan(&path, &title); err != nil {
			return nil, err
		}
		results = append(results, QueryResultRow{
			Path:   path,
			Title:  title,
			Fields: make(map[string]string),
		})
		paths = append(paths, path)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Batch-fetch properties for all result paths
	if len(q.Fields) > 0 && len(paths) > 0 {
		placeholders := strings.Repeat("?,", len(paths))
		placeholders = placeholders[:len(placeholders)-1]
		propArgs := make([]interface{}, len(paths))
		for i, p := range paths {
			propArgs[i] = p
		}
		propRows, err := ix.db.Query(
			`SELECT file_path, key, value FROM properties WHERE file_path IN (`+placeholders+`)`,
			propArgs...)
		if err == nil {
			// Build path->properties map
			propMap := make(map[string]map[string]string)
			for propRows.Next() {
				var fp, k, v string
				if err := propRows.Scan(&fp, &k, &v); err != nil {
					continue
				}
				if _, ok := propMap[fp]; !ok {
					propMap[fp] = make(map[string]string)
				}
				propMap[fp][k] = v
			}
			propRows.Close()
			// Apply to results
			for i, r := range results {
				if props, ok := propMap[r.Path]; ok {
					results[i].Fields = props
				}
			}
		}
	}

	return results, nil
}

// expandGraphNeighbors adds neighbors of a node to the node map, up to depth levels.
func (ix *Indexer) expandGraphNeighbors(nodeMap map[string]GraphNode, centerPath string, depth, maxNodes int) {
	if depth <= 0 || len(nodeMap) >= maxNodes {
		return
	}

	// Find all links to/from centerPath
	neighbors := make(map[string]bool)
	rows, err := ix.db.Query(`
		SELECT target_path FROM links WHERE from_path = ? AND resolved = 1 AND target_path != ''
		UNION
		SELECT from_path FROM links WHERE target_path = ? AND resolved = 1 AND from_path != ''`,
		centerPath, centerPath)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			continue
		}
		if nodeMap[p].ID == "" && !neighbors[p] {
			neighbors[p] = true
		}
	}

	// Add neighbors
	for p := range neighbors {
		if len(nodeMap) >= maxNodes {
			break
		}
		var title string
		if err := ix.db.QueryRow(`SELECT title FROM files WHERE path = ?`, p).Scan(&title); err != nil {
			continue
		}
		parts := strings.Split(p, "/")
		group := ""
		if len(parts) > 1 {
			group = parts[0]
		}
		nodeMap[p] = GraphNode{ID: p, Title: title, Path: p, Group: group}
		// Recurse
		ix.expandGraphNeighbors(nodeMap, p, depth-1, maxNodes)
	}
}

// escapeLike escapes SQL LIKE wildcard characters (% and _).
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// ftsKeywordRe matches standalone FTS5 keywords (word-boundary aware).
var ftsKeywordRe = regexp.MustCompile(`\b(AND|OR|NOT)\b`)

// cleanFTSQuery sanitizes a search query for FTS5 MATCH.
func cleanFTSQuery(query string) string {
	// Remove characters that have special meaning in FTS5
	replacer := strings.NewReplacer(
		`"`, ``, `{`, ``, `}`, ``, `(`, ``, `)`, ``,
		`:`, ``, `^`, ``, `+`, ``, `*`, ``,
		`~`, ``,
	)
	cleaned := replacer.Replace(query)
	// Remove standalone AND/OR/NOT keywords (not substrings like "SANDWICH")
	cleaned = ftsKeywordRe.ReplaceAllString(cleaned, "")
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
