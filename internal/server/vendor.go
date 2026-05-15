package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed vendor/*
var vendorFS embed.FS

// vendorHandler serves embedded vendor files (fonts, mermaid.js, etc.)
func vendorHandler() http.HandlerFunc {
	sub, err := fs.Sub(vendorFS, "vendor")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return func(w http.ResponseWriter, r *http.Request) {
		// Set content type based on extension
		path := strings.TrimPrefix(r.URL.Path, "/vendor/")
		switch {
		case strings.HasSuffix(path, ".js"):
			w.Header().Set("Content-Type", "application/javascript")
		case strings.HasSuffix(path, ".woff2"):
			w.Header().Set("Content-Type", "font/woff2")
		case strings.HasSuffix(path, ".ttf"):
			w.Header().Set("Content-Type", "font/ttf")
		case strings.HasSuffix(path, ".css"):
			w.Header().Set("Content-Type", "text/css")
		}
		// Cache for 7 days
		w.Header().Set("Cache-Control", "public, max-age=604800")
		r.URL.Path = "/" + path
		fileServer.ServeHTTP(w, r)
	}
}
