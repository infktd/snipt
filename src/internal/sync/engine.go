package sync

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
)

// SyncMeta is the structure of .snipt-meta.json in the Gist.
type SyncMeta struct {
	Version  int               `json:"version"`
	LastSync string            `json:"last_sync,omitempty"`
	Hashes   map[string]string `json:"snippet_hashes"`
}

// SyncResult reports what happened during a sync operation.
type SyncResult struct {
	Pushed    int      `json:"pushed"`
	Pulled    int      `json:"pulled"`
	Deleted   int      `json:"deleted"`
	Conflicts int      `json:"conflicts"`
	Errors    []string `json:"errors,omitempty"`
}

// SyncStatus reports the current sync configuration state.
type SyncStatus struct {
	Configured bool   `json:"configured"`
	GistID     string `json:"gist_id"`
	GistURL    string `json:"gist_url"`
	LastSync   string `json:"last_sync"`
	Username   string `json:"username"`
}

// SyncEngine coordinates syncing between the local DB and a GitHub Gist.
type SyncEngine struct {
	store  *db.Store
	client *GistClient
	config *config.SyncConfig
}

// NewSyncEngine creates a new SyncEngine.
func NewSyncEngine(store *db.Store, client *GistClient, cfg *config.SyncConfig) *SyncEngine {
	return &SyncEngine{store: store, client: client, config: cfg}
}

// Setup validates a token, creates a new private Gist, and returns the updated config.
func (e *SyncEngine) Setup(token string) (*config.SyncConfig, error) {
	username, err := e.client.ValidateToken()
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	meta := SyncMeta{Version: 1, Hashes: map[string]string{}}
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")

	gist, err := e.client.CreateGist("snipt-sync", map[string]GistFile{
		".snipt-meta.json": {Content: string(metaJSON)},
	}, false)
	if err != nil {
		return nil, fmt.Errorf("create gist: %w", err)
	}

	syncCfg := &config.SyncConfig{
		GistID:   gist.ID,
		Token:    token,
		Username: username,
		LastSync: time.Now().UTC().Format(time.RFC3339),
	}
	e.config = syncCfg

	if _, err := e.Push(); err != nil {
		return syncCfg, fmt.Errorf("initial push: %w", err)
	}

	return syncCfg, nil
}

// Sync performs a bidirectional sync: pull first, then push.
func (e *SyncEngine) Sync() (*SyncResult, error) {
	pullResult, err := e.Pull()
	if err != nil {
		return nil, fmt.Errorf("pull: %w", err)
	}

	pushResult, err := e.Push()
	if err != nil {
		return nil, fmt.Errorf("push: %w", err)
	}

	return &SyncResult{
		Pushed:    pushResult.Pushed,
		Pulled:    pullResult.Pulled,
		Deleted:   pullResult.Deleted + pushResult.Deleted,
		Conflicts: pullResult.Conflicts + pushResult.Conflicts,
		Errors:    append(pullResult.Errors, pushResult.Errors...),
	}, nil
}

// Push sends local snippets to the Gist.
func (e *SyncEngine) Push() (*SyncResult, error) {
	result := &SyncResult{}

	snippets, err := e.store.List(db.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}

	gist, err := e.client.GetGist(e.config.GistID)
	if err != nil {
		return nil, fmt.Errorf("get gist: %w", err)
	}

	meta := e.parseMeta(gist)

	localFiles := make(map[string]model.Snippet)
	for _, sn := range snippets {
		slug := Slugify(sn.Title)
		localFiles[slug] = sn
	}

	update := GistUpdate{Files: make(map[string]*GistFile)}
	newHashes := make(map[string]string)

	for slug, sn := range localFiles {
		hash := ComputeHash(sn)
		newHashes[slug] = hash

		// Skip if meta hash matches (content unchanged since last sync).
		if oldHash, ok := meta.Hashes[slug]; ok && oldHash == hash {
			continue
		}

		// Also skip if the remote file already has identical content.
		if remoteFile, ok := gist.Files[slug]; ok {
			content := ToFrontmatter(sn)
			if remoteFile.Content == content {
				continue
			}
		}

		content := ToFrontmatter(sn)
		update.Files[slug] = &GistFile{Content: content}
		result.Pushed++
	}

	// Delete remote files that no longer have a local counterpart.
	for filename := range gist.Files {
		if filename == ".snipt-meta.json" {
			continue
		}
		if _, ok := localFiles[filename]; !ok {
			update.Files[filename] = nil
			result.Deleted++
		}
	}

	if result.Pushed == 0 && result.Deleted == 0 {
		return result, nil
	}

	meta.Hashes = newHashes
	meta.LastSync = time.Now().UTC().Format(time.RFC3339)
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")
	update.Files[".snipt-meta.json"] = &GistFile{Content: string(metaJSON)}

	if _, err := e.client.UpdateGist(e.config.GistID, update); err != nil {
		return nil, fmt.Errorf("update gist: %w", err)
	}

	return result, nil
}

// Pull fetches snippets from the Gist into the local DB.
func (e *SyncEngine) Pull() (*SyncResult, error) {
	result := &SyncResult{}

	gist, err := e.client.GetGist(e.config.GistID)
	if err != nil {
		return nil, fmt.Errorf("get gist: %w", err)
	}

	localSnippets, err := e.store.List(db.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}
	localBySlug := make(map[string]model.Snippet)
	for _, sn := range localSnippets {
		slug := Slugify(sn.Title)
		localBySlug[slug] = sn
	}

	for filename, file := range gist.Files {
		if filename == ".snipt-meta.json" {
			continue
		}
		if !strings.HasSuffix(filename, ".md") {
			continue
		}

		remoteSn, err := FromFrontmatter(filename, file.Content)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("parse %s: %v", filename, err))
			continue
		}

		if localSn, exists := localBySlug[filename]; exists {
			remoteHash := ComputeHash(remoteSn)
			localHash := ComputeHash(localSn)
			if remoteHash != localHash {
				localSn.Title = remoteSn.Title
				localSn.Content = remoteSn.Content
				localSn.Language = remoteSn.Language
				localSn.Description = remoteSn.Description
				localSn.Pinned = remoteSn.Pinned
				if err := e.store.Update(&localSn); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("update %s: %v", filename, err))
					continue
				}
				e.syncTags(localSn.ID, remoteSn.Tags)
				result.Pulled++
			}
		} else {
			remoteSn.ID = model.NewID()
			if err := e.store.Create(&remoteSn); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("create %s: %v", filename, err))
				continue
			}
			result.Pulled++
		}
	}

	return result, nil
}

func (e *SyncEngine) syncTags(id string, tags []string) {
	current, err := e.store.Get(id)
	if err != nil {
		return
	}
	if len(current.Tags) > 0 {
		e.store.RemoveTags(id, current.Tags)
	}
	if len(tags) > 0 {
		e.store.AddTags(id, tags)
	}
}

func (e *SyncEngine) parseMeta(gist *Gist) SyncMeta {
	meta := SyncMeta{Version: 1, Hashes: map[string]string{}}
	if file, ok := gist.Files[".snipt-meta.json"]; ok {
		json.Unmarshal([]byte(file.Content), &meta)
		if meta.Hashes == nil {
			meta.Hashes = map[string]string{}
		}
	}
	return meta
}
