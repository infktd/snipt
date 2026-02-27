package gui

import (
	"fmt"
	"log"
	gosync "sync"
	"time"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/sync"
)

// SnippetService exposes snippet operations to the Wails v3 frontend.
// Registered as a v3 Service — no context, no startup hook.
type SnippetService struct {
	store   *db.Store
	version string

	syncMu    gosync.Mutex
	syncTimer *time.Timer
}

const autoSyncDelay = 2 * time.Second

// NewSnippetService creates a new SnippetService backed by the given store.
func NewSnippetService(store *db.Store, version string) *SnippetService {
	return &SnippetService{store: store, version: version}
}

// triggerAutoSync schedules a background sync after a short debounce delay.
// Rapid successive calls reset the timer so only one sync fires.
func (s *SnippetService) triggerAutoSync() {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	if s.syncTimer != nil {
		s.syncTimer.Stop()
	}
	s.syncTimer = time.AfterFunc(autoSyncDelay, func() {
		if _, err := s.SyncNow(); err != nil {
			log.Printf("[auto-sync] %v", err)
		}
	})
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
	if err := s.store.Create(&snippet); err != nil {
		return err
	}
	s.triggerAutoSync()
	return nil
}

func (s *SnippetService) UpdateSnippet(snippet model.Snippet) error {
	if err := s.store.Update(&snippet); err != nil {
		return err
	}
	s.triggerAutoSync()
	return nil
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

	s.triggerAutoSync()
	return nil
}

func (s *SnippetService) DeleteSnippet(id string) error {
	if err := s.store.Delete(id); err != nil {
		return err
	}
	s.triggerAutoSync()
	return nil
}

func (s *SnippetService) SetPinned(id string, pinned bool) error {
	if err := s.store.SetPinned(id, pinned); err != nil {
		return err
	}
	s.triggerAutoSync()
	return nil
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

// SyncSetup validates a token, creates a Gist, does initial push, and saves config.
func (s *SnippetService) SyncSetup(token string) (*sync.SyncResult, error) {
	client := sync.NewGistClient(token)
	engine := sync.NewSyncEngine(s.store, client, &config.SyncConfig{})

	syncCfg, err := engine.Setup(token)
	if err != nil {
		return nil, err
	}

	appCfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	appCfg.Sync = *syncCfg
	if err := appCfg.Save(); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}

	return &sync.SyncResult{Pushed: 0}, nil
}

// SyncNow performs a bidirectional sync and updates last_sync.
// Returns nil, nil when sync is not configured (no-op).
func (s *SnippetService) SyncNow() (*sync.SyncResult, error) {
	appCfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if appCfg.Sync.GistID == "" {
		return nil, nil
	}

	client := sync.NewGistClient(appCfg.Sync.Token)
	engine := sync.NewSyncEngine(s.store, client, &appCfg.Sync)

	result, err := engine.Sync()
	if err != nil {
		return nil, err
	}

	appCfg.Sync.LastSync = time.Now().UTC().Format(time.RFC3339)
	appCfg.Save()

	return result, nil
}

// SyncStatus returns the current sync configuration state.
func (s *SnippetService) SyncStatus() (*sync.SyncStatus, error) {
	appCfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &sync.SyncStatus{
		Configured: appCfg.Sync.GistID != "",
		GistID:     appCfg.Sync.GistID,
		GistURL:    fmt.Sprintf("https://gist.github.com/%s", appCfg.Sync.GistID),
		LastSync:   appCfg.Sync.LastSync,
		Username:   appCfg.Sync.Username,
	}, nil
}

// SyncDisconnect removes the sync configuration.
func (s *SnippetService) SyncDisconnect() error {
	appCfg, err := config.Load()
	if err != nil {
		return err
	}
	appCfg.Sync = config.SyncConfig{}
	return appCfg.Save()
}

// IsSyncConfigured returns whether sync is set up.
func (s *SnippetService) IsSyncConfigured() (bool, error) {
	appCfg, err := config.Load()
	if err != nil {
		return false, err
	}
	return appCfg.Sync.GistID != "", nil
}
