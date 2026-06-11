# API Compatibility Documentation

This document describes the API endpoints for vault-reader-go and ensures compatibility with the original vault-reader implementation.

## Base URL

```
http://localhost:3000/api
```

## Authentication

The API supports optional Basic Authentication:

```go
server.WithBasicAuth(username, password)
```

When enabled, all requests require:
```
Authorization: Basic base64(username:password)
```

## Rate Limiting

Optional rate limiting can be enabled:

```go
server.WithRateLimiting(requests, window)
```

Default: 100 requests per minute per IP address.

## Response Format

All endpoints return JSON with the following structure:

```json
{
  "items": [...]
}
```

Error responses:
```json
{
  "error": "Error message"
}
```

## Endpoints

### 1. GET /tree

Get the complete directory tree of the vault.

**Response:**
```json
{
  "name": "root",
  "type": "folder",
  "path": "",
  "children": [
    {
      "name": "notes",
      "type": "dir",
      "path": "notes",
      "children": [
        {
          "name": "example.md",
          "type": "file",
          "path": "notes/example.md",
          "isMarkdown": true,
          "isCanvas": false
        }
      ]
    }
  ]
}
```

### 2. GET /note?path={path}

Get rendered markdown note with metadata.

**Parameters:**
- `path` (required): Path to the note file

**Response:**
```json
{
  "path": "notes/example.md",
  "title": "Example Note",
  "html": "<p>Rendered content...</p>",
  "backlinks": [
    {
      "fromPath": "notes/other.md",
      "title": "Other Note",
      "raw": "[[Example Note]]"
    }
  ],
  "tags": ["golang", "tutorial"],
  "headings": [
    {
      "level": 1,
      "text": "Introduction",
      "slug": "introduction"
    }
  ],
  "properties": {
    "status": "active",
    "created": "2024-01-15"
  }
}
```

### 3. GET /search?q={query}

Full-text search across all notes using SQLite FTS5.

**Parameters:**
- `q` (required): Search query string

**Response:**
```json
{
  "items": [
    {
      "path": "notes/golang.md",
      "title": "Golang Tutorial",
      "snippet": "... <mark>golang</mark> ...",
      "rank": 0.95
    }
  ]
}
```

### 4. GET /backlinks?path={path}

Get all notes that link to the specified note.

**Parameters:**
- `path` (required): Path to the note file

**Response:**
```json
{
  "items": [
    {
      "fromPath": "notes/other.md",
      "title": "Other Note",
      "raw": "[[Example Note]]"
    }
  ]
}
```

### 5. GET /tags

Get all tags with their file counts.

**Response:**
```json
{
  "items": [
    {
      "name": "golang",
      "count": 5
    },
    {
      "name": "tutorial",
      "count": 3
    }
  ]
}
```

### 6. GET /tag?name={tag}

Get all files that have the specified tag.

**Parameters:**
- `name` (required): Tag name

**Response:**
```json
{
  "items": [
    {
      "path": "notes/golang.md",
      "title": "Golang Tutorial"
    }
  ]
}
```

### 7. GET /tag-tree

Get tags organized in a hierarchical structure.

**Response:**
```json
{
  "items": [
    {
      "name": "dev",
      "fullName": "dev",
      "count": 10,
      "children": [
        {
          "name": "golang",
          "fullName": "dev/golang",
          "count": 5,
          "children": []
        }
      ]
    }
  ]
}
```

### 8. GET /graph?folder={folder}&tag={tag}&depth={depth}&max={max}

Get nodes and edges for graph visualization.

**Parameters:**
- `folder` (optional): Filter by folder path
- `tag` (optional): Filter by tag
- `depth` (optional): Graph depth (default: 1)
- `max` (optional): Maximum nodes (default: 500)

**Response:**
```json
{
  "nodes": [
    {
      "id": "notes/a.md",
      "label": "Note A",
      "type": "file",
      "tags": ["tag1"]
    }
  ],
  "edges": [
    {
      "source": "notes/a.md",
      "target": "notes/b.md",
      "type": "wikilink"
    }
  ]
}
```

### 9. GET /dashboard

Get aggregated dashboard information.

**Response:**
```json
{
  "recent": [
    {
      "path": "notes/recent.md",
      "title": "Recent Note",
      "updated": "2024-01-15T10:30:00Z"
    }
  ],
  "inbox": [...],
  "active": [...],
  "debug": [...],
  "tags": [
    {
      "name": "golang",
      "count": 5
    }
  ],
  "canvas": [...]
}
```

### 10. POST /vault-query

Execute a YAML-based vault query (table/list/cards).

**Request Body (YAML):**
```yaml
type: table
from: notes/
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

**Response:**
```json
{
  "type": "table",
  "fields": ["title", "status", "updated"],
  "results": [
    {
      "path": "notes/active.md",
      "title": "Active Note",
      "fields": {
        "status": "active",
        "updated": "2024-01-15"
      }
    }
  ]
}
```

### 11. GET /properties?path={path}

Get YAML frontmatter properties for a note.

**Parameters:**
- `path` (required): Path to the note file

**Response:**
```json
{
  "items": [
    {
      "key": "status",
      "value": "active",
      "valueType": "string"
    },
    {
      "key": "tags",
      "value": ["golang", "tutorial"],
      "valueType": "array"
    }
  ]
}
```

### 12. GET /filter?key={key}&value={value}

Filter notes by property key-value pair.

**Parameters:**
- `key` (required): Property key
- `value` (required): Property value

**Response:**
```json
{
  "items": [
    {
      "path": "notes/active.md",
      "title": "Active Note",
      "updated": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### 13. GET /canvas?path={path}

Get JSON canvas data for visualization.

**Parameters:**
- `path` (required): Path to the canvas file

**Response:**
```json
{
  "nodes": [
    {
      "id": "n1",
      "type": "text",
      "text": "Hello",
      "x": 0,
      "y": 0,
      "width": 300,
      "height": 200
    }
  ],
  "edges": []
}
```

### 14. GET /assets?path={path}

Serve vault assets (images, PDFs, etc.).

**Parameters:**
- `path` (required): Path to the asset file

**Response:** Binary file with appropriate Content-Type header

**Supported Formats:**
- Images: PNG, JPG, GIF, SVG, WebP
- Documents: PDF, TXT

## Status Codes

- `200 OK`: Successful request
- `400 Bad Request`: Invalid parameters
- `403 Forbidden`: Path traversal attempt
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Indexer not initialized

## CORS

CORS is enabled by default for all origins. Configure with:

```go
server.WithCORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: false,
})
```

## Request ID

All requests include a unique `X-Request-Id` header for tracing.

## Performance Features

- **Zero-copy responses**: JSON encoding uses pooled buffers
- **HTTP connection reuse**: Built-in Go HTTP/2 support
- **sync.Pool**: Buffer pooling for reduced allocations
- **Streaming**: Large responses use efficient streaming
