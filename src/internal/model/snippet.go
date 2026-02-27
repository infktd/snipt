package model

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Snippet represents a saved code snippet.
type Snippet struct {
	ID          string
	Title       string
	Content     string
	Language    string
	Description string
	Source      string
	Pinned      bool
	UseCount    int
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// snippetJSON is the JSON wire format — times are strings so empty values don't
// blow up json.Unmarshal (time.Time rejects "").
type snippetJSON struct {
	ID          string   `json:"ID"`
	Title       string   `json:"Title"`
	Content     string   `json:"Content"`
	Language    string   `json:"Language"`
	Description string   `json:"Description"`
	Source      string   `json:"Source"`
	Pinned      bool     `json:"Pinned"`
	UseCount    int      `json:"UseCount"`
	Tags        []string `json:"Tags"`
	CreatedAt   string   `json:"CreatedAt"`
	UpdatedAt   string   `json:"UpdatedAt"`
}

func parseTimeLoose(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

// UnmarshalJSON accepts empty or missing time strings without error.
func (s *Snippet) UnmarshalJSON(data []byte) error {
	var raw snippetJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.ID = raw.ID
	s.Title = raw.Title
	s.Content = raw.Content
	s.Language = raw.Language
	s.Description = raw.Description
	s.Source = raw.Source
	s.Pinned = raw.Pinned
	s.UseCount = raw.UseCount
	s.Tags = raw.Tags
	s.CreatedAt = parseTimeLoose(raw.CreatedAt)
	s.UpdatedAt = parseTimeLoose(raw.UpdatedAt)
	return nil
}

// NewID generates an 8-character snippet ID from a UUIDv4.
func NewID() string {
	id := uuid.New()
	return strings.ReplaceAll(id.String(), "-", "")[:8]
}

// Stats holds collection overview data.
type Stats struct {
	TotalSnippets int
	TotalTags     int
	Languages     map[string]int // language -> count
	MostUsed      *Snippet
	RecentlyAdded []Snippet
}

// SearchResult pairs a snippet with its relevance score.
type SearchResult struct {
	Snippet      Snippet
	Score        float64
	TitleIndices []int // matched character positions for highlighting
}
