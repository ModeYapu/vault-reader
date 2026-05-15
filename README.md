# Vault Reader

A lightweight, read-only Obsidian Vault web reader built with Go.

## Features

### Core
- Directory tree with Explorer sidebar
- Markdown rendering with full syntax support
- Full-text search (SQLite FTS5, CJK support)
- Backlinks
- Chinese filename and space-in-path support
- Dark/light theme

### Obsidian Compatibility
- `[[Wikilinks]]` — name, path, alias, heading resolution
- `![[Embeds]]` — images, PDFs, note embeds
- YAML Properties & Aliases
- Callouts (`> [!note]`, `> [!tip]-`, etc.) with foldable support
- Block references (`^block-id`) with scroll-to-highlight
- Inline tags (`#tag`, `#nested/tag`)
- Mermaid diagram rendering
- JSON Canvas read-only viewer (pan, zoom, SVG edges)

### Enhanced
- **Tag Tree** — hierarchical tag navigation
- **Graph View** — force-directed graph with folder/tag filters
- **Dashboard** — homepage with recent, inbox, active, debug notes
- **Vault Query** — YAML-based query blocks (table/list/cards)
- Path traversal protection

### What's NOT Supported
- Online editing
- Plugin system
- Full Dataview language
- Canvas editor
- Multi-user collaboration
- Obsidian Sync replacement

## Quick Start

```bash
# Build
go build -o vault-reader ./cmd/vault-reader

# Run
./vault-reader --vault /path/to/your/vault

# Open
open http://localhost:3000
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--vault` | required | Path to Obsidian Vault |
| `--addr` | `:3000` | Listen address |
| `--data` | `<vault>/.vault-reader-data` | Data directory for index |

Environment variables: `VAULT_DIR`, `DATA_DIR`, `ADDR`

## Docker

```bash
docker build -t vault-reader .

docker run -d \
  -p 3000:3000 \
  -v /path/to/vault:/vault:ro \
  -v vault-data:/data \
  vault-reader
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GET /api/tree` | GET | File tree |
| `GET /api/note?path=` | GET | Rendered note |
| `GET /api/search?q=` | GET | Full-text search |
| `GET /api/backlinks?path=` | GET | Backlinks for note |
| `GET /api/tags` | GET | All tags with counts |
| `GET /api/tag?name=` | GET | Files by tag |
| `GET /api/tag-tree` | GET | Hierarchical tag tree |
| `GET /api/properties?path=` | GET | Note properties |
| `GET /api/filter?key=&value=` | GET | Filter by property |
| `GET /api/canvas?path=` | GET | Canvas JSON data |
| `GET /api/graph` | GET | Graph data |
| `GET /api/dashboard` | GET | Dashboard data |
| `GET /api/block?id=` | GET | Block reference data |
| `POST /api/vault-query` | POST | Execute vault query |
| `GET /assets?path=` | GET | Serve vault asset |
| `GET /health` | GET | Health check |

## Vault Query Syntax

Use `vault-query` code blocks in notes:

````markdown
```vault-query
type: table
from: 20_Debug
where:
  status: active
sort: updated
order: desc
limit: 20
fields:
  - title
  - status
  - updated
```
````

Supported types: `table`, `list`, `cards`

## Architecture

```
cmd/vault-reader/       — Entrypoint
internal/
  config/               — CLI configuration
  scanner/              — Vault directory scanner
  parser/               — Markdown, callout, canvas, vault-query parsers
  resolver/             — Wikilink resolver
  indexer/              — SQLite indexer (FTS5, tags, properties, blocks, graph)
  security/             — Path traversal prevention
  server/               — HTTP server + embedded frontend
```

- Single Go binary, no Node.js build chain
- Frontend: HTML + CSS + Vanilla JS (embedded)
- Database: SQLite (pure-Go, no CGO required)

## License

MIT
