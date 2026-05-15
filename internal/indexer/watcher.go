package indexer

import (
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches the vault directory for changes and triggers reindex.
type Watcher struct {
	indexer   *Indexer
	vaultDir  string
	watcher   *fsnotify.Watcher
	debounce  time.Duration
	done      chan struct{}
	mu        sync.Mutex
	timer     *time.Timer
}

// NewWatcher creates a new file watcher.
func NewWatcher(ix *Indexer, vaultDir string) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		indexer:  ix,
		vaultDir: vaultDir,
		watcher:  fw,
		debounce: 2 * time.Second,
		done:     make(chan struct{}),
	}

	// Add vault directory and subdirectories
	if err := w.addWatchDirs(); err != nil {
		fw.Close()
		return nil, err
	}

	go w.loop()

	return w, nil
}

// Close stops the watcher.
func (w *Watcher) Close() error {
	close(w.done)
	return w.watcher.Close()
}

func (w *Watcher) addWatchDirs() error {
	return filepath.WalkDir(w.vaultDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip ignored directories
		rel, _ := filepath.Rel(w.vaultDir, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		for _, p := range parts {
			if p == ".obsidian" || p == ".git" || p == "node_modules" ||
				p == ".trash" || p == ".vault-reader-data" || p == ".DS_Store" {
				return filepath.SkipDir
			}
		}

		if d.IsDir() {
			if err := w.watcher.Add(path); err != nil {
				slog.Warn("failed to watch directory", "path", path, "error", err)
			}
		}
		return nil
	})
}

func (w *Watcher) loop() {
	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("watcher error", "error", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// We only care about markdown and asset files
	name := event.Name
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".md" && ext != ".markdown" && !isWatchedAsset(ext) {
		// Also handle directory creation (need to watch new dirs)
		if event.Has(fsnotify.Create) {
			if de, err := filepath.Abs(name); err == nil {
				_ = de
				// Check if it's a directory and add to watcher
				w.tryAddDir(name)
			}
		}
		return
	}

	slog.Debug("vault file changed", "event", event.Op.String(), "name", name)

	// Debounce: reset timer
	w.mu.Lock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.debounce, func() {
		w.reindex()
		w.mu.Lock()
		w.timer = nil
		w.mu.Unlock()
	})
	w.mu.Unlock()
}

func (w *Watcher) tryAddDir(path string) {
	// Simple check: try to add as watch target
	// fsnotify will only succeed for directories
	if err := w.watcher.Add(path); err == nil {
		slog.Debug("watching new directory", "path", path)
	}
}

func (w *Watcher) reindex() {
	w.mu.Lock()
	defer w.mu.Unlock()

	slog.Info("reindexing due to vault changes")
	if err := w.indexer.FullIndex(); err != nil {
		slog.Error("reindex failed", "error", err)
	} else {
		slog.Info("reindex complete")
	}
}

func isWatchedAsset(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".pdf":
		return true
	}
	return false
}
