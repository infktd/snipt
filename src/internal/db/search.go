package db

import (
	"fmt"
	"strings"

	"github.com/infktd/snipt/src/internal/model"
)

// NotFoundError indicates that no snippet matched the given reference.
type NotFoundError struct {
	Ref string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("snippet %q not found", e.Ref)
}

// Search performs a full-text search using FTS5 MATCH and returns results
// ordered by relevance score.
func (s *Store) Search(query string) ([]model.SearchResult, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.title, s.content, s.language, s.description, s.source,
		       s.pinned, s.use_count, s.created_at, s.updated_at,
		       rank
		FROM snippets_fts fts
		JOIN snippets s ON s.rowid = fts.rowid
		WHERE snippets_fts MATCH ?
		ORDER BY rank
	`, query)
	if err != nil {
		return nil, fmt.Errorf("FTS5 search: %w", err)
	}
	defer rows.Close()

	var results []model.SearchResult
	for rows.Next() {
		var sn model.Snippet
		var pinned int
		var createdAt, updatedAt string
		var rank float64

		err := rows.Scan(
			&sn.ID, &sn.Title, &sn.Content, &sn.Language,
			&sn.Description, &sn.Source, &pinned, &sn.UseCount,
			&createdAt, &updatedAt, &rank,
		)
		if err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}

		sn.Pinned = pinned != 0
		sn.CreatedAt = parseTime(createdAt)
		sn.UpdatedAt = parseTime(updatedAt)

		tags, err := s.getTags(sn.ID)
		if err != nil {
			return nil, err
		}
		sn.Tags = tags

		// FTS5 rank is negative (lower = more relevant), negate for a positive score.
		results = append(results, model.SearchResult{
			Snippet: sn,
			Score:   -rank,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}

	return results, nil
}

// ResolveRef resolves a reference to snippets using a resolution chain:
//  1. Exact ID match
//  2. Exact title match (case-insensitive)
//  3. FTS5 search
func (s *Store) ResolveRef(ref string) ([]model.SearchResult, error) {
	// 1. Exact ID match.
	row := s.db.QueryRow(`
		SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at
		FROM snippets WHERE id = ?
	`, ref)
	sn, err := scanSnippet(row)
	if err == nil {
		tags, err := s.getTags(sn.ID)
		if err != nil {
			return nil, err
		}
		sn.Tags = tags
		return []model.SearchResult{{Snippet: *sn, Score: 1.0}}, nil
	}

	// 2. Exact title match (case-insensitive).
	titleRows, err := s.db.Query(`
		SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at
		FROM snippets WHERE LOWER(title) = LOWER(?)
	`, ref)
	if err != nil {
		return nil, fmt.Errorf("title lookup: %w", err)
	}
	defer titleRows.Close()

	var titleResults []model.SearchResult
	snippets, err := scanSnippetRows(titleRows)
	if err != nil {
		return nil, err
	}
	for _, sn := range snippets {
		tags, err := s.getTags(sn.ID)
		if err != nil {
			return nil, err
		}
		sn.Tags = tags
		titleResults = append(titleResults, model.SearchResult{Snippet: sn, Score: 0.9})
	}
	if len(titleResults) > 0 {
		return titleResults, nil
	}

	// 3. FTS5 search.
	// Escape special FTS5 characters by quoting each term.
	terms := strings.Fields(ref)
	for i, t := range terms {
		terms[i] = `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
	}
	ftsQuery := strings.Join(terms, " ")

	results, err := s.Search(ftsQuery)
	if err != nil {
		// If the FTS query fails (e.g. syntax error), return empty results.
		return nil, nil
	}

	return results, nil
}

// GetAndTrack resolves a reference, returns the top matching snippet,
// and increments its use_count. Returns NotFoundError if no match.
func (s *Store) GetAndTrack(ref string) (*model.Snippet, error) {
	results, err := s.ResolveRef(ref)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, &NotFoundError{Ref: ref}
	}

	top := &results[0].Snippet

	if err := s.IncrementUseCount(top.ID); err != nil {
		return nil, fmt.Errorf("track usage: %w", err)
	}

	// Re-fetch to get the updated use_count.
	return s.Get(top.ID)
}

// IsNotFound returns true if the error is a NotFoundError.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*NotFoundError)
	return ok
}
