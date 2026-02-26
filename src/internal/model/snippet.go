package model

import (
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
