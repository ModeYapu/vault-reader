package indexer

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches the vault directory for changes and triggers reindex.
type Watcher struct {
	indexer  *Indexer
	vaultDir string
	watcher  *fsnotify.Watcher
	debounce time.Duration
	done     chan struct{}
	closeOnce sync.Once
	mu       sync.Mutex
	timer    *time.Timer
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

// Close stops the watcher. Safe to call multiple times.
func (w *Watcher) Close() error {
	w.closeOnce.Do(func() {
		close(w.done)
	})
	return w.watcher.Close()
}

func (w *Watcher) addWatchDirs() error {
	return filepath.WalkDir(w.vaultDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Warn("watcher: skipping directory", "path", path, "error", err)
			return nil
		}

		// Skip ignored directories
		rel, relErr := filepath.Rel(w.vaultDir, path)
		if relErr != nil {
			slog.Warn("watcher: cannot compute relative path", "path", path, "error", relErr)
			return nil
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		for _, p := range parts {
			if p == ".obsidian" || p == ".git" || p == "node_modules" ||
				p == ".trash" || p == ".vault-reader-data" {
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
			w.tryAddDir(name)
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
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return
	}
	if err := w.watcher.Add(path); err == nil {
		slog.Debug("watching new directory", "path", path)
	}
}

func (w *Watcher) reindex() {
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
