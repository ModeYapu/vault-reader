package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"vault-reader/internal/config"
	"vault-reader/internal/indexer"
	"vault-reader/internal/middleware"
	"vault-reader/internal/server"
)

func main() {
	cfg, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		slog.Error("invalid arguments", "error", err)
		os.Exit(1)
	}

	// Verify vault directory exists and is a directory
	info, err := os.Stat(cfg.VaultDir)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Error("vault directory does not exist", "path", cfg.VaultDir)
		} else {
			slog.Error("cannot access vault directory", "path", cfg.VaultDir, "error", err)
		}
		os.Exit(1)
	}
	if !info.IsDir() {
		slog.Error("vault path is not a directory", "path", cfg.VaultDir)
		os.Exit(1)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		slog.Error("cannot create data directory", "path", cfg.DataDir, "error", err)
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
		slog.Error("failed to initialize indexer", "error", err)
		os.Exit(1)
	}
	defer ix.Close()

	// Run full index — fatal on failure to prevent serving empty/corrupt data
	slog.Info("building index...")
	if err := ix.FullIndex(); err != nil {
		slog.Error("full index failed, cannot start server", "error", err)
		os.Exit(1)
	}

	// Start file watcher
	watcher, err := indexer.NewWatcher(ix, cfg.VaultDir)
	if err != nil {
		slog.Warn("file watcher not available — vault changes will not be detected automatically", "error", err)
	} else {
		defer watcher.Close()
		slog.Info("file watcher started")
	}

	// Initialize metrics
	metrics := middleware.NewMetrics()

	// Build server options from config
	opts := []server.Option{
		server.WithIndexer(ix),
		server.WithMetrics(metrics),
		server.WithConfigReload(func() error {
			reloaded, err := cfg.Reload()
			if err != nil {
				return err
			}
			if reloaded {
				slog.Info("configuration reloaded", "config", cfg.GetConfigPath())
			}
			return nil
		}),
	}

	// Configure CORS
	corsOrigins := strings.Split(cfg.CORSOrigins, ",")
	for i, origin := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(origin)
	}
	opts = append(opts, server.WithCORS(middleware.CORSConfig{
		AllowedOrigins: corsOrigins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		ExposedHeaders: []string{"X-Request-Id"},
	}))

	// Configure rate limiting if enabled
	if cfg.RateLimit > 0 {
		window, err := time.ParseDuration(cfg.RateWindow)
		if err != nil {
			slog.Warn("invalid rate limit window, using default", "error", err)
			window = time.Minute
		}
		opts = append(opts, server.WithRateLimiting(cfg.RateLimit, window))
		slog.Info("rate limiting enabled", "requests", cfg.RateLimit, "window", cfg.RateWindow)
	}

	// Configure basic authentication if both username and password are provided
	if cfg.AuthUsername != "" && cfg.AuthPassword != "" {
		opts = append(opts, server.WithBasicAuth(cfg.AuthUsername, cfg.AuthPassword))
		slog.Info("basic authentication enabled")
	}

	// Configure base URL for reverse proxy
	if cfg.BaseURL != "" {
		opts = append(opts, server.WithBaseURL(cfg.BaseURL))
		slog.Info("base URL configured", "baseURL", cfg.BaseURL)
	}

	// Create server with all options
	handler := server.New(cfg.VaultDir, opts...)

	httpSrv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Set up signal handling for graceful shutdown and config reload
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Use a channel to propagate server errors to main goroutine,
	// so deferred cleanups (DB close, watcher close) actually run.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for signals
	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				slog.Info("shutting down...")
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := httpSrv.Shutdown(shutdownCtx); err != nil {
					slog.Error("shutdown error", "error", err)
				}
				slog.Info("server stopped")
				return
			case syscall.SIGHUP:
				slog.Info("received SIGHUP, reloading configuration")
				if _, err := cfg.Reload(); err != nil {
					slog.Error("config reload failed", "error", err)
				} else {
					slog.Info("configuration reloaded successfully",
						"vault", cfg.VaultDir,
						"addr", cfg.Addr,
					)
				}
			}
		case err := <-serverErr:
			slog.Error("server exited unexpectedly", "error", err)
			os.Exit(1)
		}
	}
}
