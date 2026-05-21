package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"vault-reader/internal/indexer"
	"vault-reader/internal/middleware"
	"vault-reader/internal/parser"
	"vault-reader/internal/resolver"
	"vault-reader/internal/scanner"
	"vault-reader/internal/security"
)

// Server is the HTTP server for vault-reader.
type Server struct {
	vaultDir      string
	mux           *http.ServeMux
	resolver      *resolver.Resolver
	indexer       *indexer.Indexer
	middleware    []func(http.Handler) http.Handler
	corsConfig    middleware.CORSConfig
	rateLimiter   *middleware.RateLimiter
	authConfig    *middleware.BasicAuthConfig
	metrics       *middleware.Metrics
	configReload  func() error
}

// Option configures a Server.
type Option func(*Server)

// WithIndexer sets the indexer for the server.
func WithIndexer(ix *indexer.Indexer) Option {
	return func(s *Server) {
		s.indexer = ix
	}
}

// WithCORS sets the CORS configuration.
func WithCORS(cfg middleware.CORSConfig) Option {
	return func(s *Server) {
		s.corsConfig = cfg
	}
}

// WithRateLimiting enables rate limiting.
func WithRateLimiting(requests int, window time.Duration) Option {
	return func(s *Server) {
		s.rateLimiter = middleware.NewRateLimiter(requests, window)
	}
}

// WithBasicAuth enables basic authentication.
func WithBasicAuth(username, password string) Option {
	return func(s *Server) {
		s.authConfig = &middleware.BasicAuthConfig{
			Username: username,
			Password: password,
		}
	}
}

// WithMiddleware adds custom middleware to the chain.
func WithMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *Server) {
		s.middleware = append(s.middleware, mw)
	}
}

// WithMetrics enables Prometheus metrics collection.
func WithMetrics(m *middleware.Metrics) Option {
	return func(s *Server) {
		s.metrics = m
	}
}

// WithConfigReload sets the config reload function.
func WithConfigReload(reload func() error) Option {
	return func(s *Server) {
		s.configReload = reload
	}
}

// New creates a new Server serving the given vault directory.
func New(vaultDir string, opts ...Option) *Server {
	s := &Server{
		vaultDir:   vaultDir,
		mux:        http.NewServeMux(),
		corsConfig: middleware.DefaultCORSConfig(),
		middleware: []func(http.Handler) http.Handler{},
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
	// Build middleware chain in correct order:
	// Recovery → RequestID → Logging → CORS → Metrics → RateLimiter → BasicAuth
	middlewares := []func(http.Handler) http.Handler{
		middleware.Recovery,
		middleware.RequestID,
		middleware.Logging,
		middleware.CORS(s.corsConfig),
	}
	if s.metrics != nil {
		middlewares = append(middlewares, s.metrics.Handler)
	}
	if s.rateLimiter != nil {
		middlewares = append(middlewares, s.rateLimiter.Handler)
	}
	if s.authConfig != nil {
		middlewares = append(middlewares, middleware.BasicAuth(*s.authConfig))
	}
	// Add custom middleware last
	middlewares = append(middlewares, s.middleware...)

	chain := middleware.Chain(middlewares...)

	// Apply security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")

	chain(s.mux).ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	// Health checks
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ready", s.handleReady)
	// Metrics endpoint
	s.mux.HandleFunc("/metrics", s.handleMetrics)
	// API documentation
	s.mux.HandleFunc("/api/docs", s.handleAPIDocs)
	s.mux.HandleFunc("/api/openapi.yaml", s.handleOpenAPISpec)
	// API endpoints
	s.mux.HandleFunc("/api/tree", s.handleTree)
	s.mux.HandleFunc("/api/note", s.handleNote)
	s.mux.HandleFunc("/api/search", s.handleSearch)
	s.mux.HandleFunc("/api/search/advanced", s.handleAdvancedSearch)
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

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}

	// Check database connection
	if s.indexer != nil {
		db := s.indexer.DB()
		if err := db.Ping(); err != nil {
			health["status"] = "unhealthy"
			health["database"] = "disconnected"
			slog.Warn("health check: database ping failed", "error", err)
		} else {
			health["database"] = "connected"
		}
	}

	// Check disk space
	var stat runtime.MemStats
	runtime.ReadMemStats(&stat)
	health["memory_alloc"] = stat.Alloc
	health["memory_sys"] = stat.Sys
	health["goroutines"] = runtime.NumGoroutine()

	// Check vault directory accessibility
	if _, err := os.Stat(s.vaultDir); err != nil {
		health["status"] = "unhealthy"
		health["vault"] = "inaccessible"
	} else {
		health["vault"] = "accessible"
	}

	status := http.StatusOK
	if health["status"] != "healthy" {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, health)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Write HTTP metrics
	if s.metrics != nil {
		s.metrics.WritePrometheus(w)
	}

	// Write runtime metrics
	middleware.WriteRuntimeMetrics(w)
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.indexer == nil {
		http.Error(w, "indexer not ready", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ready"}`))
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

func (s *Server) handleAdvancedSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.indexer == nil {
		http.Error(w, "index not available", http.StatusServiceUnavailable)
		return
	}

	var query string
	var tags []string
	var pathPrefix string
	var limit int

	if r.Method == http.MethodPost {
		var req struct {
			Query      string   `json:"query"`
			Tags       []string `json:"tags"`
			PathPrefix string   `json:"pathPrefix"`
			Limit      int      `json:"limit"`
		}
		if err := readJSON(r, &req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		query = req.Query
		tags = req.Tags
		pathPrefix = req.PathPrefix
		limit = req.Limit
	} else {
		query = r.URL.Query().Get("q")
		tags = r.URL.Query()["tag"]
		pathPrefix = r.URL.Query().Get("path")
		if l := r.URL.Query().Get("limit"); l != "" {
			limit, _ = strconv.Atoi(l)
		}
	}

	if limit <= 0 {
		limit = 50
	}

	results, err := s.indexer.AdvancedSearch(query, tags, pathPrefix, limit)
	if err != nil {
		slog.Error("advanced search failed", "query", query, "error", err)
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
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
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
		w.Header().Set("Content-Disposition", "attachment; filename="+info.Name())
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

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(apiDocsHTML))
}

func (s *Server) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Write([]byte(openapiSpec))
}

func readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

const apiDocsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vault Reader API Documentation</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; background: #f8f9fa; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .header h1 { color: #1a73e8; margin-bottom: 10px; }
        .header p { color: #5f6368; }
        .endpoint { background: white; padding: 20px; margin-bottom: 15px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .method { display: inline-block; padding: 4px 12px; border-radius: 4px; font-weight: 600; font-size: 12px; margin-right: 10px; }
        .get { background: #e8f0fe; color: #1967d2; }
        .post { background: #e6f4ea; color: #137333; }
        .path { font-family: monospace; font-size: 14px; color: #202124; }
        .description { margin: 10px 0; color: #5f6368; }
        .params { background: #f8f9fa; padding: 15px; border-radius: 6px; margin-top: 10px; }
        .params h4 { margin-bottom: 8px; color: #202124; }
        .param { display: flex; padding: 8px 0; border-bottom: 1px solid #e8eaed; }
        .param:last-child { border-bottom: none; }
        .param-name { font-family: monospace; font-weight: 600; width: 150px; color: #1967d2; }
        .param-type { width: 100px; color: #5f6368; font-size: 12px; }
        .param-desc { flex: 1; color: #5f6368; }
        .response { background: #f8f9fa; padding: 15px; border-radius: 6px; margin-top: 10px; }
        .response h4 { margin-bottom: 8px; }
        code { background: #e8eaed; padding: 2px 6px; border-radius: 4px; font-family: monospace; font-size: 13px; }
        .footer { text-align: center; margin-top: 40px; color: #5f6368; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📚 Vault Reader API</h1>
            <p>A lightweight, read-only Obsidian Vault web reader API v1.0</p>
            <p style="margin-top: 15px;">
                <a href="/" style="color: #1a73e8; text-decoration: none;">← Back to App</a> |
                <a href="/api/openapi.yaml" style="color: #1a73e8; text-decoration: none;" download>Download OpenAPI Spec</a>
            </p>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/tree</span></h2>
            <p class="description">Get the complete directory tree of the vault</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "name": "root",
  "type": "folder",
  "children": [
    { "name": "notes", "type": "dir", "children": [...] }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/note?path={path}</span></h2>
            <p class="description">Get rendered markdown note with metadata</p>
            <div class="params">
                <h4>Parameters</h4>
                <div class="param">
                    <span class="param-name">path</span>
                    <span class="param-type">string</span>
                    <span class="param-desc">Path to the note file (e.g., "notes/example.md")</span>
                </div>
            </div>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "path": "notes/example.md",
  "title": "Example Note",
  "html": "<p>Rendered content...</p>",
  "backlinks": [...],
  "tags": ["golang", "tutorial"],
  "properties": { "status": "active" }
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/search?q={query}</span></h2>
            <p class="description">Full-text search across all notes using SQLite FTS5</p>
            <div class="params">
                <h4>Parameters</h4>
                <div class="param">
                    <span class="param-name">q</span>
                    <span class="param-type">string</span>
                    <span class="param-desc">Search query string</span>
                </div>
            </div>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "items": [
    { "path": "notes/golang.md", "title": "Golang Tutorial", "snippet": "...", "rank": 0.95 }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/backlinks?path={path}</span></h2>
            <p class="description">Get all notes that link to the specified note</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "items": [
    { "fromPath": "notes/other.md", "title": "Other Note", "raw": "[[link text]]" }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/tags</span></h2>
            <p class="description">Get all tags with their file counts</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "items": [
    { "name": "golang", "count": 5 },
    { "name": "tutorial", "count": 3 }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/tag?name={tag}</span></h2>
            <p class="description">Get all files that have the specified tag</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "items": [
    { "path": "notes/golang.md", "title": "Golang Tutorial" }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/tag-tree</span></h2>
            <p class="description">Get tags organized in a hierarchical structure</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "items": [
    { "name": "dev", "count": 10, "children": [...] }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/graph?folder={folder}&tag={tag}&depth={depth}&max={max}</span></h2>
            <p class="description">Get nodes and edges for graph visualization</p>
            <div class="params">
                <h4>Parameters</h4>
                <div class="param">
                    <span class="param-name">folder</span>
                    <span class="param-type">string</span>
                    <span class="param-desc">Filter by folder path (optional)</span>
                </div>
                <div class="param">
                    <span class="param-name">tag</span>
                    <span class="param-type">string</span>
                    <span class="param-desc">Filter by tag (optional)</span>
                </div>
                <div class="param">
                    <span class="param-name">depth</span>
                    <span class="param-type">integer</span>
                    <span class="param-desc">Graph depth (default: 1)</span>
                </div>
                <div class="param">
                    <span class="param-name">max</span>
                    <span class="param-type">integer</span>
                    <span class="param-desc">Maximum nodes (default: 500)</span>
                </div>
            </div>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "nodes": [{ "id": "notes/a.md", "label": "Note A", "type": "file" }],
  "edges": [{ "source": "notes/a.md", "target": "notes/b.md" }]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/dashboard</span></h2>
            <p class="description">Get aggregated dashboard information</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "recent": [...],
  "inbox": [...],
  "active": [...],
  "debug": [...],
  "tags": [...],
  "canvas": [...]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method post">POST</span><span class="path">/api/vault-query</span></h2>
            <p class="description">Execute a YAML-based vault query (table/list/cards)</p>
            <div class="params">
                <h4>Request Body (YAML)</h4>
                <pre><code>type: table
from: notes/
where:
  status: active
sort: updated
order: desc
limit: 20
fields:
  - title
  - status
  - updated</code></pre>
            </div>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "type": "table",
  "fields": ["title", "status", "updated"],
  "results": [...]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/properties?path={path}</span></h2>
            <p class="description">Get YAML frontmatter properties for a note</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "items": [
    { "key": "status", "value": "active", "valueType": "string" },
    { "key": "tags", "value": ["golang", "tutorial"], "valueType": "array" }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/filter?key={key}&value={value}</span></h2>
            <p class="description">Filter notes by property key-value pair</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "items": [
    { "path": "notes/active.md", "title": "Active Note", "updated": "2024-01-15T10:30:00Z" }
  ]
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/api/canvas?path={path}</span></h2>
            <p class="description">Get JSON canvas data for visualization</p>
            <div class="response">
                <h4>Response (200)</h4>
                <pre><code>{
  "nodes": [
    { "id": "n1", "type": "text", "text": "Hello", "x": 0, "y": 0, "width": 300, "height": 200 }
  ],
  "edges": []
}</code></pre>
            </div>
        </div>

        <div class="endpoint">
            <h2><span class="method get">GET</span><span class="path">/assets?path={path}</span></h2>
            <p class="description">Serve vault assets (images, PDFs, etc.)</p>
            <div class="params">
                <h4>Supported Formats</h4>
                <p>PNG, JPG, GIF, SVG, WebP, PDF, TXT</p>
            </div>
            <div class="response">
                <h4>Response (200)</h4>
                <p>Binary file with appropriate Content-Type header</p>
            </div>
        </div>

        <div class="footer">
            <p>Vault Reader API v1.0 • MIT License • <a href="/api/openapi.yaml" style="color: #1a73e8;">Download OpenAPI Spec</a></p>
        </div>
    </div>
</body>
</html>
`

const openapiSpec = `openapi: 3.0.3
info:
  title: Vault Reader API
  description: A lightweight, read-only Obsidian Vault web reader API
  version: 1.0.0
  contact:
    name: API Support
    url: https://github.com/openclaw-vault
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT

servers:
  - url: http://localhost:3000
    description: Local development server

tags:
  - name: health
    description: Health check operations
  - name: notes
    description: Note operations
  - name: search
    description: Search operations
  - name: tags
    description: Tag operations
  - name: graph
    description: Graph visualization

paths:
  /health:
    get:
      tags: [health]
      summary: Health check
      description: Basic liveness probe
      operationId: health
      responses:
        '200':
          description: Service is healthy
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string

  /ready:
    get:
      tags: [health]
      summary: Readiness probe
      description: Check if service is ready to handle requests
      operationId: ready
      responses:
        '200':
          description: Service is ready
        '503':
          description: Service not ready
  /tree:
    get:
      tags: [notes]
      summary: Get file tree
      description: Returns the complete directory tree of the vault
      operationId: getTree
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TreeNode'

  /note:
    get:
      tags: [notes]
      summary: Get note content
      description: Returns rendered markdown note with metadata
      operationId: getNote
      parameters:
        - name: path
          in: query
          description: Path to the note file
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Note'

  /search:
    get:
      tags: [search]
      summary: Search notes
      description: Full-text search across all notes
      operationId: searchNotes
      parameters:
        - name: q
          in: query
          description: Search query string
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SearchResults'

  /backlinks:
    get:
      tags: [notes]
      summary: Get backlinks
      description: Returns all notes that link to the specified note
      operationId: getBacklinks
      parameters:
        - name: path
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response

  /tags:
    get:
      tags: [tags]
      summary: Get all tags
      description: Returns all tags with counts
      operationId: getTags
      responses:
        '200':
          description: Successful response

  /tag:
    get:
      tags: [tags]
      summary: Get files by tag
      operationId: getFilesByTag
      parameters:
        - name: name
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response

  /graph:
    get:
      tags: [graph]
      summary: Get graph data
      description: Returns nodes and edges for graph visualization
      operationId: getGraph
      parameters:
        - name: folder
          in: query
          schema:
            type: string
        - name: tag
          in: query
          schema:
            type: string
        - name: depth
          in: query
          schema:
            type: integer
        - name: max
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GraphData'

  /tag-tree:
    get:
      tags: [tags]
      summary: Get tag tree
      description: Returns tags organized in a hierarchical structure
      operationId: getTagTree
      responses:
        '200':
          description: Successful response

  /canvas:
    get:
      tags: [notes]
      summary: Get canvas data
      description: Returns JSON canvas data for visualization
      operationId: getCanvas
      parameters:
        - name: path
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response

  /dashboard:
    get:
      tags: [notes]
      summary: Get dashboard data
      description: Returns aggregated dashboard information
      operationId: getDashboard
      responses:
        '200':
          description: Successful response

  /vault-query:
    post:
      tags: [search]
      summary: Execute vault query
      description: Execute a YAML-based vault query (table/list/cards)
      operationId: vaultQuery
      requestBody:
        required: true
        content:
          application/x-yaml:
            schema:
              type: string
      responses:
        '200':
          description: Successful response

  /properties:
    get:
      tags: [notes]
      summary: Get note properties
      description: Get YAML frontmatter properties for a note
      operationId: getProperties
      parameters:
        - name: path
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response

  /filter:
    get:
      tags: [notes]
      summary: Filter notes by property
      description: Filter notes by property key-value pair
      operationId: filterByProperty
      parameters:
        - name: key
          in: query
          required: true
          schema:
            type: string
        - name: value
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response

components:
  schemas:
    TreeNode:
      type: object
      properties:
        name:
          type: string
        path:
          type: string
        type:
          type: string
          enum: [file, folder]
        children:
          type: array
          items:
            $ref: '#/components/schemas/TreeNode'

    Note:
      type: object
      properties:
        path:
          type: string
        title:
          type: string
        html:
          type: string
        backlinks:
          type: array
          items:
            $ref: '#/components/schemas/Backlink'
        tags:
          type: array
          items:
            type: string

    SearchResults:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/SearchResult'

    SearchResult:
      type: object
      properties:
        path:
          type: string
        title:
          type: string
        snippet:
          type: string

    GraphData:
      type: object
      properties:
        nodes:
          type: array
          items:
            $ref: '#/components/schemas/GraphNode'
        edges:
          type: array
          items:
            $ref: '#/components/schemas/GraphEdge'

    GraphNode:
      type: object
      properties:
        id:
          type: string
        label:
          type: string

    GraphEdge:
      type: object
      properties:
        source:
          type: string
        target:
          type: string
`

