package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	// If indexer has a resolver, use it
	if s.indexer != nil {
		// The indexer builds its own resolver during FullIndex
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

// extractTitle reads a markdown file and extracts its title.
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

// resolveFunc returns a function that resolves wikilink targets.
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

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	s.mux.HandleFunc("/api/properties", s.handleProperties)
	s.mux.HandleFunc("/api/filter", s.handleFilter)
	s.mux.HandleFunc("/assets", s.handleAssets)
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
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}

	// Security check
	if err := security.ValidatePath(s.vaultDir, path); err != nil {
		slog.Warn("path validation failed", "path", path, "error", err)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Read file
	fullPath := filepath.Join(s.vaultDir, filepath.FromSlash(path))
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("read file failed", "path", path, "error", err)
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	// Parse markdown
	doc, err := parser.ParseDocument(string(content), path)
	if err != nil {
		slog.Error("parse failed", "path", path, "error", err)
		http.Error(w, "parse error", http.StatusInternalServerError)
		return
	}

	// Resolve wikilinks in HTML
	doc.HTML = parser.RenderWikiLinksInHTML(doc.HTML, s.resolveFunc())

	// Add backlinks
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
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
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

func (s *Server) handleProperties(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
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

	// Verify file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	// Determine content type
	contentType := contentTypeFromExt(filepath.Ext(info.Name()))
	w.Header().Set("Content-Type", contentType)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
