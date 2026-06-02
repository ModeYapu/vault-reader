# Changelog

All notable changes to vault-reader are documented here.

## [0.2.0] - 2026-05-15

### Added

**Tag Tree (Milestone 3)**
- `GET /api/tag-tree` — hierarchical tag tree API with nested tag support (e.g. `debug/proxy`)
- Frontend tag tree panel in right sidebar with expand/collapse
- Clickable tag pills that show associated files

**JSON Canvas Viewer (Milestone 6)**
- Scanner recognizes `.canvas` files with `IsCanvas` flag
- `GET /api/canvas?path=xxx` — serves parsed canvas JSON
- Frontend canvas viewer with:
  - Absolute-positioned nodes (text, file, link, group)
  - SVG edge rendering
  - Pan and zoom (mouse drag + scroll wheel)
  - File nodes navigate to notes, link nodes open external URLs
  - Canvas icon in sidebar tree

**Graph View (Milestone 7)**
- `GET /api/graph` — graph data API with filters: `?folder=`, `?tag=`, `?path=`, `?depth=`, `?max=`
- Frontend force-directed graph with SVG rendering
- Node colors by folder group
- Click nodes to open notes
- Graph toggle button in header

**Dashboard (Milestone 8)**
- `GET /api/dashboard` — aggregated dashboard data (recent, inbox, active, debug, tags, canvas)
- Frontend dashboard homepage with card layout
- Loads automatically on app start

**Vault Query / Dataview Lite (Milestone 9)**
- `POST /api/vault-query` — execute YAML-based queries against vault index
- Supports `type` (table/list/cards), `from`, `where`, `sort`, `order`, `limit`, `fields`
- Frontend renders vault-query code blocks as tables, lists, or card grids
- Graceful error display on query failures

### Tests

- `internal/indexer/indexer_test.go` — TagTree (5 tests), Graph (4 tests)
- `internal/server/server_test.go` — TagTree (2), Canvas (3), Graph (3), Dashboard (2)
- `internal/scanner/scanner_test.go` — Canvas scan test
- `internal/parser/vault_query_test.go` — 5 tests
- `internal/parser/callout_test.go` — 7 tests

---

## [0.1.0] - 2026-05-14

### Added

- Vault directory scanning with `.obsidian`/`.git`/`node_modules` exclusion
- Markdown rendering with goldmark
- File tree sidebar with Explorer
- `[[Wikilink]]` resolution (name, path, alias, heading)
- `![[Asset embed]]` (images, PDFs, notes)
- YAML frontmatter parsing
- SQLite FTS5 full-text search
- Backlinks
- Tags (frontmatter + inline `#tag`)
- Properties display in right sidebar
- Aliases participate in wikilink resolution
- Block references (`^block-id`) with scroll-to and highlight
- Obsidian Callouts (`> [!type]`) with foldable support
- Mermaid diagram rendering
- Dark/light theme toggle
- Path traversal protection
- Docker deployment
- Chinese filename and space-in-path support
