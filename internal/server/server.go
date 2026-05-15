package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"vault-reader/internal/indexer"
	"vault-reader/internal/parser"
	"vault-reader/internal/resolver"
	"vault-reader/internal/scanner"
	"vault-reader/internal/security"
)

// Server is the HTTP server for vault-reader.
type Server struct {
	vaultDir string
	mux      *http.ServeMux
	resolver *resolver.Resolver
	indexer  *indexer.Indexer
}

// Option configures a Server.
type Option func(*Server)

// WithIndexer sets the indexer for the server.
func WithIndexer(ix *indexer.Indexer) Option {
	return func(s *Server) {
		s.indexer = ix
	}
}

// New creates a new Server serving the given vault directory.
func New(vaultDir string, opts ...Option) *Server {
	s := &Server{
		vaultDir: vaultDir,
		mux:      http.NewServeMux(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.buildResolver()
	s.routes()
	return s
}

// buildResolver scans the vault and builds the link resolver.
func (s *Server) buildResolver() {
	if s.indexer != nil {
		return
	}

	files, err := scanner.Scan(s.vaultDir)
	if err != nil {
		slog.Error("failed to scan vault for resolver", "error", err)
		s.resolver = resolver.New(nil)
		return
	}

	metas := make([]resolver.FileMeta, 0, len(files))
	for _, f := range files {
		if !f.IsMarkdown {
			continue
		}
		title := s.extractTitle(f)
		metas = append(metas, resolver.BuildFileMeta(f.Path, title))
	}

	s.resolver = resolver.New(metas)
	slog.Info("resolver built", "files", len(metas))
}

func (s *Server) extractTitle(f scanner.VaultFile) string {
	fullPath := filepath.Join(s.vaultDir, filepath.FromSlash(f.Path))
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}
	doc, err := parser.ParseDocument(string(content), f.Path)
	if err != nil {
		return ""
	}
	return doc.Title
}

func (s *Server) resolveFunc() parser.ResolveFunc {
	return func(target string) (string, bool) {
		if s.indexer != nil {
			result := s.indexer.Resolve(target)
			return result.TargetPath, result.Found
		}
		if s.resolver != nil {
			result := s.resolver.Resolve(target)
			return result.TargetPath, result.Found
		}
		return "", false
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/tree", s.handleTree)
	s.mux.HandleFunc("/api/note", s.handleNote)
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/backlinks", s.handleBacklinks)
	s.mux.HandleFunc("/api/tags", s.handleTags)
	s.mux.HandleFunc("/api/tag", s.handleTag)
	s.mux.HandleFunc("/api/tag-tree", s.handleTagTree)
	s.mux.HandleFunc("/api/canvas", s.handleCanvas)
	s.mux.HandleFunc("/api/graph", s.handleGraph)
	s.mux.HandleFunc("/api/dashboard", s.handleDashboard)
	s.mux.HandleFunc("/api/vault-query", s.handleVaultQuery)
	s.mux.HandleFunc("/api/properties", s.handleProperties)
	s.mux.HandleFunc("/api/filter", s.handleFilter)
	s.mux.HandleFunc("/api/block", s.handleBlock)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/assets", s.handleAssets)
	s.mux.HandleFunc("/vendor/", vendorHandler())
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.indexer != nil {
		// Use indexed data instead of filesystem scan
		paths, err := s.indexer.GetFileList()
		if err != nil {
			slog.Error("file list query failed", "error", err)
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		vfiles := make([]scanner.VaultFile, len(paths))
		for i, p := range paths {
			vfiles[i] = scanner.VaultFile{Path: p, Name: filepath.Base(p)}
		}
		tree := scanner.BuildTree(vfiles)
		writeJSON(w, http.StatusOK, tree)
		return
	}

	// Fallback: scan filesystem when no indexer
	files, err := scanner.Scan(s.vaultDir)
	if err != nil {
		slog.Error("scan failed", "error", err)
		http.Error(w, "scan failed", http.StatusInternalServerError)
		return
	}
	tree := scanner.BuildTree(files)
	writeJSON(w, http.StatusOK, tree)
}

func (s *Server) handleNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	if err := security.ValidatePath(s.vaultDir, path); err != nil {
		slog.Warn("path validation failed", "path", path, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(s.vaultDir, filepath.FromSlash(path))

	// Limit file size to prevent OOM
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	if info.Size() > 10*1024*1024 {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		slog.Error("read file failed", "path", path, "error", err)
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	doc, err := parser.ParseDocument(string(content), path)
	if err != nil {
		slog.Error("parse failed", "path", path, "error", err)
		http.Error(w, "parse error", http.StatusInternalServerError)
		return
	}

	doc.HTML = parser.RenderWikiLinksInHTML(doc.HTML, s.resolveFunc())

	if s.indexer != nil {
		backlinks, err := s.indexer.GetBacklinks(path)
		if err != nil {
			slog.Error("backlinks query failed", "path", path, "error", err)
		} else {
			doc.Backlinks = backlinks
		}
	}

	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{"items": []interface{}{}})
		return
	}

	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	results, err := s.indexer.Search(q, 50)
	if err != nil {
		slog.Error("search failed", "query", q, "error", err)
		http.Error(w, "search error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": results})
}

func (s *Server) handleBacklinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	if err := security.ValidatePath(s.vaultDir, path); err != nil {
		slog.Warn("backlinks path validation failed", "path", path, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	backlinks, err := s.indexer.GetBacklinks(path)
	if err != nil {
		slog.Error("backlinks query failed", "path", path, "error", err)
		http.Error(w, "backlinks error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": backlinks})
}

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	tags, err := s.indexer.GetTags()
	if err != nil {
		slog.Error("tags query failed", "error", err)
		http.Error(w, "tags error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": tags})
}

func (s *Server) handleTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name parameter required", http.StatusBadRequest)
		return
	}

	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	files, err := s.indexer.GetFilesByTag(name)
	if err != nil {
		slog.Error("tag files query failed", "tag", name, "error", err)
		http.Error(w, "tag error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": files})
}

func (s *Server) handleTagTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	tree, err := s.indexer.GetTagTree()
	if err != nil {
		slog.Error("tag tree query failed", "error", err)
		http.Error(w, "tag tree error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": tree})
}

func (s *Server) handleCanvas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	if err := security.ValidatePath(s.vaultDir, path); err != nil {
		slog.Warn("canvas path validation failed", "path", path, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(s.vaultDir, filepath.FromSlash(path))

	// Limit file size to prevent OOM
	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	if stat.Size() > 10*1024*1024 {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		slog.Error("read canvas failed", "path", path, "error", err)
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	doc, err := parser.ParseCanvas(string(content), path)
	if err != nil {
		slog.Error("parse canvas failed", "path", path, "error", err)
		http.Error(w, "invalid canvas JSON", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	folder := r.URL.Query().Get("folder")
	tag := r.URL.Query().Get("tag")
	path := r.URL.Query().Get("path")
	depth := 1
	if d := r.URL.Query().Get("depth"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			depth = parsed
		}
	}
	maxNodes := 500
	if m := r.URL.Query().Get("max"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 {
			maxNodes = parsed
		}
	}

	nodes, edges, err := s.indexer.GetGraph(folder, tag, path, depth, maxNodes)
	if err != nil {
		slog.Error("graph query failed", "error", err)
		http.Error(w, "graph error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	data, err := s.indexer.GetDashboard()
	if err != nil {
		slog.Error("dashboard query failed", "error", err)
		http.Error(w, "dashboard error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleVaultQuery(w http.ResponseWriter, r *http.Request) {
	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	q, err := parser.ParseVaultQuery(string(body))
	if err != nil {
		http.Error(w, "invalid query YAML", http.StatusBadRequest)
		return
	}

	results, err := s.indexer.ExecuteVaultQuery(q)
	if err != nil {
		slog.Error("vault query failed", "error", err)
		http.Error(w, "query error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"type":    q.Type,
		"fields":  q.Fields,
		"results": results,
	})
}

func (s *Server) handleProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	if err := security.ValidatePath(s.vaultDir, path); err != nil {
		slog.Warn("properties path validation failed", "path", path, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	props, err := s.indexer.GetProperties(path)
	if err != nil {
		slog.Error("properties query failed", "path", path, "error", err)
		http.Error(w, "properties error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": props})
}

func (s *Server) handleFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	if key == "" || value == "" {
		http.Error(w, "key and value parameters required", http.StatusBadRequest)
		return
	}

	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	files, err := s.indexer.FilterByProperty(key, value)
	if err != nil {
		slog.Error("filter query failed", "key", key, "value", value, "error", err)
		http.Error(w, "filter error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"items": files})
}

func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	if err := security.ValidatePath(s.vaultDir, path); err != nil {
		slog.Warn("asset path validation failed", "path", path, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(s.vaultDir, filepath.FromSlash(path))

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
		if info.IsDir() {
			http.Error(w, "not a file", http.StatusBadRequest)
			return
		}

	contentType := contentTypeFromExt(filepath.Ext(info.Name()))
	w.Header().Set("Content-Type", contentType)
	// SVG can contain scripts; serve as attachment to prevent XSS
	if contentType == "image/svg+xml" {
				w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(info.Name()))
	}
	http.ServeFile(w, r, fullPath)
}

func contentTypeFromExt(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("json encode error", "error", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}
	block, err := s.indexer.GetBlock(id)
	if err != nil {
		slog.Error("block query failed", "id", id, "error", err)
		http.Error(w, "block error", http.StatusInternalServerError)
		return
	}
	if block == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, block)
}
