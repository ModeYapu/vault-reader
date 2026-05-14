package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"vault-reader/internal/config"
	"vault-reader/internal/indexer"
	"vault-reader/internal/server"
)

func main() {
	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Verify vault directory exists
	if info, err := os.Stat(cfg.VaultDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: vault directory %q does not exist or is not a directory\n", cfg.VaultDir)
		os.Exit(1)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot create data directory %q: %v\n", cfg.DataDir, err)
		os.Exit(1)
	}

	slog.Info("starting vault-reader",
		"vault", cfg.VaultDir,
		"data", cfg.DataDir,
		"addr", cfg.Addr,
	)

	// Initialize indexer
	dbPath := filepath.Join(cfg.DataDir, "vault-reader.db")
	ix, err := indexer.New(dbPath, cfg.VaultDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize indexer: %v\n", err)
		os.Exit(1)
	}
	defer ix.Close()

	// Run full index
	slog.Info("building index...")
	if err := ix.FullIndex(); err != nil {
		slog.Error("full index failed", "error", err)
	}

	// Start file watcher
	watcher, err := indexer.NewWatcher(ix, cfg.VaultDir)
	if err != nil {
		slog.Warn("file watcher not available", "error", err)
	} else {
		defer watcher.Close()
		slog.Info("file watcher started")
	}

	// Create server with indexer
	srv := server.New(cfg.VaultDir, server.WithIndexer(ix))

	slog.Info("listening", "addr", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, srv); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
