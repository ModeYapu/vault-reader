package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"vault-reader/internal/config"
	"vault-reader/internal/indexer"
	"vault-reader/internal/server"
)

// version is set via -ldflags "-X main.version=..." at build time.
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	// Verify vault directory exists and is a directory
	info, err := os.Stat(cfg.VaultDir)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Error("vault directory does not exist", "path", cfg.VaultDir)
		} else {
			slog.Error("cannot access vault directory", "path", cfg.VaultDir, "error", err)
		}
		return err
	}
	if !info.IsDir() {
		slog.Error("vault path is not a directory", "path", cfg.VaultDir)
		return os.ErrNotExist
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		slog.Error("cannot create data directory", "path", cfg.DataDir, "error", err)
		return err
	}

	slog.Info("starting vault-reader",
		"version", version,
		"vault", cfg.VaultDir,
		"data", cfg.DataDir,
		"addr", cfg.Addr,
	)

	// Initialize indexer
	dbPath := filepath.Join(cfg.DataDir, "vault-reader.db")
	ix, err := indexer.New(dbPath, cfg.VaultDir)
	if err != nil {
		slog.Error("failed to initialize indexer", "error", err)
		return err
	}
	defer ix.Close()

	// Run full index — fatal on failure to prevent serving empty/corrupt data
	slog.Info("building index...")
	if err := ix.FullIndex(); err != nil {
		slog.Error("full index failed, cannot start server", "error", err)
		return err
	}

	// Start file watcher
	watcher, err := indexer.NewWatcher(ix, cfg.VaultDir)
	if err != nil {
		slog.Warn("file watcher not available — vault changes will not be detected automatically", "error", err)
	} else {
		defer watcher.Close()
		slog.Info("file watcher started")
	}

	// Create server with indexer
	handler := server.New(cfg.VaultDir, server.WithIndexer(ix), server.WithPrefix(cfg.Prefix))

	httpSrv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Use a channel to propagate server errors to main goroutine,
	// so deferred cleanups (DB close, watcher close) actually run.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for either a shutdown signal or a server error
	select {
	case <-ctx.Done():
		slog.Info("shutting down...")
	case err := <-serverErr:
		slog.Error("server exited unexpectedly", "error", err)
		stop() // release signal resources
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
	return nil
}
