package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/infktd/snipt/src/internal/model"
)

// ListOpts controls filtering and sorting for List().
type ListOpts struct {
	Language string
	Tag      string
	Pinned   *bool
	Sort     string // "created", "updated", "usage", "title"
}

// Create inserts a new snippet and its tags in a single transaction.
func (s *Store) Create(snippet *model.Snippet) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	if snippet.CreatedAt.IsZero() {
		snippet.CreatedAt, _ = time.Parse(time.RFC3339, now)
	}
	if snippet.UpdatedAt.IsZero() {
		snippet.UpdatedAt = snippet.CreatedAt
	}

	_, err = tx.Exec(`
		INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		snippet.ID,
		snippet.Title,
		snippet.Content,
		snippet.Language,
		snippet.Description,
		snippet.Source,
		boolToInt(snippet.Pinned),
		snippet.UseCount,
		snippet.CreatedAt.UTC().Format(time.RFC3339),
		snippet.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert snippet: %w", err)
	}

	for _, tag := range snippet.Tags {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO tags (snippet_id, tag) VALUES (?, ?)`, snippet.ID, tag); err != nil {
			return fmt.Errorf("insert tag %q: %w", tag, err)
		}
	}

	return tx.Commit()
}

// Get retrieves a snippet by exact ID. Returns an error if not found.
func (s *Store) Get(id string) (*model.Snippet, error) {
	row := s.db.QueryRow(`
		SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at
		FROM snippets WHERE id = ?
	`, id)

	snippet, err := scanSnippet(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("snippet %q not found", id)
		}
		return nil, fmt.Errorf("scan snippet: %w", err)
	}

	tags, err := s.getTags(id)
	if err != nil {
		return nil, err
	}
	snippet.Tags = tags

	return snippet, nil
}

// Update modifies an existing snippet's fields and sets updated_at.
func (s *Store) Update(snippet *model.Snippet) error {
	snippet.UpdatedAt = time.Now().UTC()

	result, err := s.db.Exec(`
		UPDATE snippets
		SET title = ?, content = ?, language = ?, description = ?, source = ?, pinned = ?, updated_at = ?
		WHERE id = ?
	`,
		snippet.Title,
		snippet.Content,
		snippet.Language,
		snippet.Description,
		snippet.Source,
		boolToInt(snippet.Pinned),
		snippet.UpdatedAt.Format(time.RFC3339),
		snippet.ID,
	)
	if err != nil {
		return fmt.Errorf("update snippet: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("snippet %q not found", snippet.ID)
	}

	return nil
}

// Delete removes a snippet by ID. Tags are deleted via CASCADE.
func (s *Store) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM snippets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete snippet: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("snippet %q not found", id)
	}

	return nil
}

// AddTags adds tags to a snippet (idempotent via INSERT OR IGNORE).
func (s *Store) AddTags(id string, tags []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, tag := range tags {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO tags (snippet_id, tag) VALUES (?, ?)`, id, tag); err != nil {
			return fmt.Errorf("add tag %q: %w", tag, err)
		}
	}

	return tx.Commit()
}

// RemoveTags removes tags from a snippet (idempotent).
func (s *Store) RemoveTags(id string, tags []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, tag := range tags {
		if _, err := tx.Exec(`DELETE FROM tags WHERE snippet_id = ? AND tag = ?`, id, tag); err != nil {
			return fmt.Errorf("remove tag %q: %w", tag, err)
		}
	}

	return tx.Commit()
}

// SetPinned sets the pinned state of a snippet.
func (s *Store) SetPinned(id string, pinned bool) error {
	result, err := s.db.Exec(`UPDATE snippets SET pinned = ?, updated_at = ? WHERE id = ?`,
		boolToInt(pinned), time.Now().UTC().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("set pinned: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("snippet %q not found", id)
	}

	return nil
}

// IncrementUseCount bumps the use_count of a snippet by 1.
func (s *Store) IncrementUseCount(id string) error {
	result, err := s.db.Exec(`UPDATE snippets SET use_count = use_count + 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("increment use_count: %w", err)
	}

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("snippet %q not found", id)
	}

	return nil
}

// List returns snippets filtered and sorted by the given options.
func (s *Store) List(opts ListOpts) ([]model.Snippet, error) {
	var clauses []string
	var args []interface{}

	if opts.Language != "" {
		clauses = append(clauses, "s.language = ?")
		args = append(args, opts.Language)
	}
	if opts.Tag != "" {
		clauses = append(clauses, "s.id IN (SELECT snippet_id FROM tags WHERE tag = ?)")
		args = append(args, opts.Tag)
	}
	if opts.Pinned != nil {
		clauses = append(clauses, "s.pinned = ?")
		args = append(args, boolToInt(*opts.Pinned))
	}

	query := `SELECT s.id, s.title, s.content, s.language, s.description, s.source, s.pinned, s.use_count, s.created_at, s.updated_at FROM snippets s`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	switch opts.Sort {
	case "usage":
		query += " ORDER BY s.use_count DESC, s.updated_at DESC"
	case "alpha", "title":
		query += " ORDER BY s.title ASC"
	case "created":
		query += " ORDER BY s.created_at DESC"
	default: // "recent", "updated", or empty
		query += " ORDER BY s.updated_at DESC"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}
	defer rows.Close()

	snippets, err := scanSnippetRows(rows)
	if err != nil {
		return nil, err
	}

	// Load tags for each snippet.
	for i := range snippets {
		tags, err := s.getTags(snippets[i].ID)
		if err != nil {
			return nil, err
		}
		snippets[i].Tags = tags
	}

	return snippets, nil
}

// Stats returns collection overview data.
func (s *Store) Stats() (*model.Stats, error) {
	stats := &model.Stats{
		Languages: make(map[string]int),
	}

	// Total snippets.
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM snippets`).Scan(&stats.TotalSnippets); err != nil {
		return nil, fmt.Errorf("count snippets: %w", err)
	}

	// Total distinct tags.
	if err := s.db.QueryRow(`SELECT COUNT(DISTINCT tag) FROM tags`).Scan(&stats.TotalTags); err != nil {
		return nil, fmt.Errorf("count tags: %w", err)
	}

	// Languages breakdown.
	langRows, err := s.db.Query(`SELECT language, COUNT(*) FROM snippets WHERE language != '' GROUP BY language`)
	if err != nil {
		return nil, fmt.Errorf("query languages: %w", err)
	}
	defer langRows.Close()
	for langRows.Next() {
		var lang string
		var count int
		if err := langRows.Scan(&lang, &count); err != nil {
			return nil, fmt.Errorf("scan language: %w", err)
		}
		stats.Languages[lang] = count
	}
	if err := langRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate languages: %w", err)
	}

	// Most used snippet.
	mostUsedRow := s.db.QueryRow(`
		SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at
		FROM snippets ORDER BY use_count DESC LIMIT 1
	`)
	mostUsed, err := scanSnippet(mostUsedRow)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("most used snippet: %w", err)
	}
	if err == nil {
		tags, err := s.getTags(mostUsed.ID)
		if err != nil {
			return nil, err
		}
		mostUsed.Tags = tags
		stats.MostUsed = mostUsed
	}

	// Recently added (up to 5).
	recentRows, err := s.db.Query(`
		SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at
		FROM snippets ORDER BY created_at DESC LIMIT 5
	`)
	if err != nil {
		return nil, fmt.Errorf("query recent snippets: %w", err)
	}
	defer recentRows.Close()
	recent, err := scanSnippetRows(recentRows)
	if err != nil {
		return nil, err
	}
	for i := range recent {
		tags, err := s.getTags(recent[i].ID)
		if err != nil {
			return nil, err
		}
		recent[i].Tags = tags
	}
	stats.RecentlyAdded = recent

	return stats, nil
}

// --- helpers ---

// getTags returns all tags for a snippet.
func (s *Store) getTags(snippetID string) ([]string, error) {
	rows, err := s.db.Query(`SELECT tag FROM tags WHERE snippet_id = ? ORDER BY tag`, snippetID)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// scanSnippet scans a single row into a Snippet.
func scanSnippet(row *sql.Row) (*model.Snippet, error) {
	var s model.Snippet
	var pinned int
	var createdAt, updatedAt string

	err := row.Scan(
		&s.ID, &s.Title, &s.Content, &s.Language,
		&s.Description, &s.Source, &pinned, &s.UseCount,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	s.Pinned = pinned != 0
	s.CreatedAt = parseTime(createdAt)
	s.UpdatedAt = parseTime(updatedAt)

	return &s, nil
}

// scanSnippetRows scans multiple rows into a slice of Snippets.
func scanSnippetRows(rows *sql.Rows) ([]model.Snippet, error) {
	var snippets []model.Snippet
	for rows.Next() {
		var s model.Snippet
		var pinned int
		var createdAt, updatedAt string

		err := rows.Scan(
			&s.ID, &s.Title, &s.Content, &s.Language,
			&s.Description, &s.Source, &pinned, &s.UseCount,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan snippet row: %w", err)
		}

		s.Pinned = pinned != 0
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		snippets = append(snippets, s)
	}
	return snippets, rows.Err()
}

// boolToInt converts a bool to 0 or 1 for SQLite storage.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// parseTime tries RFC3339 first, then falls back to SQLite's datetime format.
func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}
