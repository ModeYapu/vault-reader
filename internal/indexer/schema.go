package indexer

const schemaSQL = `
CREATE TABLE IF NOT EXISTS files (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT UNIQUE NOT NULL,
  title TEXT NOT NULL,
  ext TEXT NOT NULL,
  size INTEGER NOT NULL,
  mtime INTEGER NOT NULL,
  content TEXT,
  html TEXT,
  frontmatter_json TEXT,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  from_path TEXT NOT NULL,
  raw TEXT NOT NULL,
  target TEXT NOT NULL,
  target_path TEXT,
  alias TEXT,
  heading TEXT,
  is_embed INTEGER NOT NULL DEFAULT 0,
  is_asset INTEGER NOT NULL DEFAULT 0,
  resolved INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  tag TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS headings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  level INTEGER NOT NULL,
  text TEXT NOT NULL,
  slug TEXT NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS file_fts USING fts5(
  title,
  path,
  content,
  tokenize="unicode61 categories 'L* N* Co'"
);

CREATE INDEX IF NOT EXISTS idx_links_from_resolved ON links(from_path, resolved, target_path);
CREATE INDEX IF NOT EXISTS idx_links_from ON links(from_path);
CREATE INDEX IF NOT EXISTS idx_links_target ON links(target_path);
CREATE INDEX IF NOT EXISTS idx_tags_file ON tags(file_path);
CREATE INDEX IF NOT EXISTS idx_tags_tag ON tags(tag);
CREATE TABLE IF NOT EXISTS properties (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  key TEXT NOT NULL,
  value TEXT,
  value_type TEXT NOT NULL DEFAULT 'string'
);

CREATE INDEX IF NOT EXISTS idx_properties_key ON properties(key);
CREATE INDEX IF NOT EXISTS idx_properties_file_path ON properties(file_path);
CREATE INDEX IF NOT EXISTS idx_properties_key_value ON properties(key, value);

CREATE INDEX IF NOT EXISTS idx_headings_file ON headings(file_path);

CREATE TABLE IF NOT EXISTS blocks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  block_id TEXT NOT NULL,
  text TEXT,
  line INTEGER
);

CREATE INDEX IF NOT EXISTS idx_blocks_file ON blocks(file_path);
CREATE INDEX IF NOT EXISTS idx_blocks_id ON blocks(block_id);
`
