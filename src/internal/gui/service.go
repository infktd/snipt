package gui

import (
	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
)

// SnippetService exposes snippet operations to the Wails v3 frontend.
// Registered as a v3 Service — no context, no startup hook.
type SnippetService struct {
	store   *db.Store
	version string
}

// NewSnippetService creates a new SnippetService backed by the given store.
func NewSnippetService(store *db.Store, version string) *SnippetService {
	return &SnippetService{store: store, version: version}
}

func (s *SnippetService) ListSnippets(opts db.ListOpts) ([]model.Snippet, error) {
	return s.store.List(opts)
}

func (s *SnippetService) SearchSnippets(query string) ([]model.SearchResult, error) {
	return s.store.Search(query)
}

func (s *SnippetService) GetSnippet(id string) (*model.Snippet, error) {
	return s.store.Get(id)
}

func (s *SnippetService) CreateSnippet(snippet model.Snippet) error {
	snippet.ID = model.NewID()
	return s.store.Create(&snippet)
}

func (s *SnippetService) UpdateSnippet(snippet model.Snippet) error {
	return s.store.Update(&snippet)
}

func (s *SnippetService) UpdateSnippetTags(id string, tags []string) error {
	current, err := s.store.Get(id)
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
		if err := s.store.RemoveTags(id, toRemove); err != nil {
			return err
		}
	}
	if len(toAdd) > 0 {
		if err := s.store.AddTags(id, toAdd); err != nil {
			return err
		}
	}

	return nil
}

func (s *SnippetService) DeleteSnippet(id string) error {
	return s.store.Delete(id)
}

func (s *SnippetService) SetPinned(id string, pinned bool) error {
	return s.store.SetPinned(id, pinned)
}

func (s *SnippetService) IncrementUseCount(id string) error {
	return s.store.IncrementUseCount(id)
}

func (s *SnippetService) GetStats() (*model.Stats, error) {
	return s.store.Stats()
}

func (s *SnippetService) GetConfig() (*config.Config, error) {
	return config.Load()
}

func (s *SnippetService) UpdateConfig(cfg config.Config) error {
	return cfg.Save()
}

func (s *SnippetService) GetDBPath() string {
	return config.DBPath("")
}

func (s *SnippetService) GetVersion() string {
	return s.version
}
