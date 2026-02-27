package gui

import (
	"context"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
)

// App exposes snippet operations to the Wails frontend.
type App struct {
	ctx   context.Context
	store *db.Store
	mode  string
}

// NewApp creates a new App backed by the given store.
// mode is "manage" or "find".
func NewApp(store *db.Store, mode string) *App {
	return &App{store: store, mode: mode}
}

// GetMode returns the GUI mode ("manage" or "find").
func (a *App) GetMode() string {
	return a.mode
}

// Startup is called by Wails at application startup.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ListSnippets returns snippets filtered and sorted by opts.
func (a *App) ListSnippets(opts db.ListOpts) ([]model.Snippet, error) {
	return a.store.List(opts)
}

// SearchSnippets performs FTS5 search and returns results with scores.
func (a *App) SearchSnippets(query string) ([]model.SearchResult, error) {
	return a.store.Search(query)
}

// GetSnippet retrieves a snippet by exact ID.
func (a *App) GetSnippet(id string) (*model.Snippet, error) {
	return a.store.Get(id)
}

// CreateSnippet generates an ID and inserts a new snippet.
func (a *App) CreateSnippet(s model.Snippet) error {
	s.ID = model.NewID()
	return a.store.Create(&s)
}

// UpdateSnippet modifies an existing snippet's fields (not tags).
func (a *App) UpdateSnippet(s model.Snippet) error {
	return a.store.Update(&s)
}

// UpdateSnippetTags replaces a snippet's tags with the given set.
// Diffs current vs desired and calls AddTags/RemoveTags.
func (a *App) UpdateSnippetTags(id string, tags []string) error {
	current, err := a.store.Get(id)
	if err != nil {
		return err
	}

	oldSet := make(map[string]bool, len(current.Tags))
	for _, t := range current.Tags {
		oldSet[t] = true
	}
	newSet := make(map[string]bool, len(tags))
	for _, t := range tags {
		newSet[t] = true
	}

	var toAdd []string
	for _, t := range tags {
		if !oldSet[t] {
			toAdd = append(toAdd, t)
		}
	}

	var toRemove []string
	for _, t := range current.Tags {
		if !newSet[t] {
			toRemove = append(toRemove, t)
		}
	}

	if len(toRemove) > 0 {
		if err := a.store.RemoveTags(id, toRemove); err != nil {
			return err
		}
	}
	if len(toAdd) > 0 {
		if err := a.store.AddTags(id, toAdd); err != nil {
			return err
		}
	}

	return nil
}

// DeleteSnippet removes a snippet by ID.
func (a *App) DeleteSnippet(id string) error {
	return a.store.Delete(id)
}

// SetPinned sets the pinned state of a snippet.
func (a *App) SetPinned(id string, pinned bool) error {
	return a.store.SetPinned(id, pinned)
}

// IncrementUseCount bumps the use count of a snippet by 1.
func (a *App) IncrementUseCount(id string) error {
	return a.store.IncrementUseCount(id)
}

// GetStats returns collection overview data.
func (a *App) GetStats() (*model.Stats, error) {
	return a.store.Stats()
}
