package db

import "fmt"

// migrate runs schema migrations up to the current version.
func (s *Store) migrate() error {
	// Create the meta table to track schema version.
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS meta (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create meta table: %w", err)
	}

	version := 0
	row := s.db.QueryRow(`SELECT value FROM meta WHERE key = 'schema_version'`)
	var v string
	if err := row.Scan(&v); err == nil {
		fmt.Sscanf(v, "%d", &version)
	}

	if version < 1 {
		if err := s.migrateV1(); err != nil {
			return fmt.Errorf("migrate v1: %w", err)
		}
	}

	return nil
}

// migrateV1 creates the initial schema: snippets, tags, FTS5, and triggers.
func (s *Store) migrateV1() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Snippets table.
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS snippets (
			id          TEXT PRIMARY KEY,
			title       TEXT NOT NULL DEFAULT '',
			content     TEXT NOT NULL DEFAULT '',
			language    TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			source      TEXT NOT NULL DEFAULT '',
			pinned      INTEGER NOT NULL DEFAULT 0,
			use_count   INTEGER NOT NULL DEFAULT 0,
			created_at  TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create snippets table: %w", err)
	}

	// Tags table.
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			snippet_id TEXT NOT NULL REFERENCES snippets(id) ON DELETE CASCADE,
			tag        TEXT NOT NULL,
			PRIMARY KEY (snippet_id, tag)
		)
	`); err != nil {
		return fmt.Errorf("create tags table: %w", err)
	}

	// FTS5 virtual table for full-text search.
	if _, err := tx.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts USING fts5(
			title,
			content,
			description,
			language,
			content='snippets',
			content_rowid='rowid'
		)
	`); err != nil {
		return fmt.Errorf("create FTS5 table: %w", err)
	}

	// Triggers to keep FTS in sync with snippets.
	if _, err := tx.Exec(`
		CREATE TRIGGER IF NOT EXISTS snippets_ai AFTER INSERT ON snippets BEGIN
			INSERT INTO snippets_fts(rowid, title, content, description, language)
			VALUES (new.rowid, new.title, new.content, new.description, new.language);
		END
	`); err != nil {
		return fmt.Errorf("create insert trigger: %w", err)
	}

	if _, err := tx.Exec(`
		CREATE TRIGGER IF NOT EXISTS snippets_ad AFTER DELETE ON snippets BEGIN
			INSERT INTO snippets_fts(snippets_fts, rowid, title, content, description, language)
			VALUES ('delete', old.rowid, old.title, old.content, old.description, old.language);
		END
	`); err != nil {
		return fmt.Errorf("create delete trigger: %w", err)
	}

	if _, err := tx.Exec(`
		CREATE TRIGGER IF NOT EXISTS snippets_au AFTER UPDATE ON snippets BEGIN
			INSERT INTO snippets_fts(snippets_fts, rowid, title, content, description, language)
			VALUES ('delete', old.rowid, old.title, old.content, old.description, old.language);
			INSERT INTO snippets_fts(rowid, title, content, description, language)
			VALUES (new.rowid, new.title, new.content, new.description, new.language);
		END
	`); err != nil {
		return fmt.Errorf("create update trigger: %w", err)
	}

	// Record schema version.
	if _, err := tx.Exec(`
		INSERT OR REPLACE INTO meta (key, value) VALUES ('schema_version', '1')
	`); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}

	return tx.Commit()
}
