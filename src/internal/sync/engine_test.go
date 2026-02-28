package sync

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
)

func openTestStore(t *testing.T) *db.Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func makeSnippet(id, title, content, lang string, tags []string) *model.Snippet {
	return &model.Snippet{
		ID:       id,
		Title:    title,
		Content:  content,
		Language: lang,
		Tags:     tags,
	}
}

func TestPush_NewSnippets(t *testing.T) {
	store := openTestStore(t)

	sn1 := makeSnippet("push0001", "HTTP Server", "package main", "go", []string{"web"})
	sn2 := makeSnippet("push0002", "Flask App", "from flask import Flask", "python", []string{"web"})
	store.Create(sn1)
	store.Create(sn2)

	var patchPayload GistUpdate
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/gists/aabbccddee0011223344":
			json.NewEncoder(w).Encode(Gist{
				ID: "aabbccddee0011223344",
				Files: map[string]GistFile{
					".snipt-meta.json": {Content: `{"version":1,"snippet_hashes":{}}`},
				},
			})
		case r.Method == "PATCH" && r.URL.Path == "/gists/aabbccddee0011223344":
			json.NewDecoder(r.Body).Decode(&patchPayload)
			json.NewEncoder(w).Encode(Gist{ID: "aabbccddee0011223344"})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	cfg := &config.SyncConfig{GistID: "aabbccddee0011223344", Token: "token"}
	engine := NewSyncEngine(store, client, cfg)

	result, err := engine.Push()
	if err != nil {
		t.Fatalf("Push() error: %v", err)
	}

	if result.Pushed != 2 {
		t.Errorf("Pushed = %d, want 2", result.Pushed)
	}

	if _, ok := patchPayload.Files["http-server.md"]; !ok {
		t.Error("expected http-server.md in patch")
	}
	if _, ok := patchPayload.Files["flask-app.md"]; !ok {
		t.Error("expected flask-app.md in patch")
	}
	if _, ok := patchPayload.Files[".snipt-meta.json"]; !ok {
		t.Error("expected .snipt-meta.json in patch")
	}
}

func TestPull_NewRemoteSnippets(t *testing.T) {
	store := openTestStore(t)

	remoteMD := ToFrontmatter(model.Snippet{
		Title:    "Remote Snippet",
		Language: "rust",
		Content:  "fn main() {}",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Gist{
			ID: "aabbccddee0011223344",
			Files: map[string]GistFile{
				"remote-snippet.md":  {Content: remoteMD},
				".snipt-meta.json": {Content: `{"version":1,"snippet_hashes":{}}`},
			},
		})
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	cfg := &config.SyncConfig{GistID: "aabbccddee0011223344", Token: "token"}
	engine := NewSyncEngine(store, client, cfg)

	result, err := engine.Pull()
	if err != nil {
		t.Fatalf("Pull() error: %v", err)
	}

	if result.Pulled != 1 {
		t.Errorf("Pulled = %d, want 1", result.Pulled)
	}

	all, _ := store.List(db.ListOpts{})
	if len(all) != 1 {
		t.Fatalf("DB snippet count = %d, want 1", len(all))
	}
	if all[0].Title != "Remote Snippet" {
		t.Errorf("Title = %q, want %q", all[0].Title, "Remote Snippet")
	}
	if all[0].Language != "rust" {
		t.Errorf("Language = %q, want %q", all[0].Language, "rust")
	}
}

func TestPush_UnchangedSnippetsSkipped(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("skip0001", "Unchanged", "content", "go", nil)
	store.Create(sn)

	hash := ComputeHash(*sn)
	metaJSON, _ := json.Marshal(SyncMeta{
		Version: 1,
		Hashes:  map[string]string{"unchanged.md": hash},
	})

	var patched bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET":
			json.NewEncoder(w).Encode(Gist{
				ID: "aabbccddee0011223344",
				Files: map[string]GistFile{
					"unchanged.md":     {Content: ToFrontmatter(*sn)},
					".snipt-meta.json": {Content: string(metaJSON)},
				},
			})
		case r.Method == "PATCH":
			patched = true
			json.NewEncoder(w).Encode(Gist{ID: "aabbccddee0011223344"})
		}
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	cfg := &config.SyncConfig{GistID: "aabbccddee0011223344", Token: "token"}
	engine := NewSyncEngine(store, client, cfg)

	result, err := engine.Push()
	if err != nil {
		t.Fatalf("Push() error: %v", err)
	}

	if result.Pushed != 0 {
		t.Errorf("Pushed = %d, want 0 (nothing changed)", result.Pushed)
	}
	if patched {
		t.Error("should not have sent PATCH when nothing changed")
	}
}

func TestSetup(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("set00001", "Setup Test", "hello", "text", nil)
	store.Create(sn)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/user":
			json.NewEncoder(w).Encode(map[string]string{"login": "testuser"})
		case r.Method == "POST" && r.URL.Path == "/gists":
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(Gist{
				ID:      "bb00cc11dd22ee33ff44",
				HTMLURL: "https://gist.github.com/bb00cc11dd22ee33ff44",
			})
		case r.Method == "GET" && r.URL.Path == "/gists/bb00cc11dd22ee33ff44":
			json.NewEncoder(w).Encode(Gist{
				ID: "bb00cc11dd22ee33ff44",
				Files: map[string]GistFile{
					".snipt-meta.json": {Content: `{"version":1,"snippet_hashes":{}}`},
				},
			})
		case r.Method == "PATCH" && r.URL.Path == "/gists/bb00cc11dd22ee33ff44":
			json.NewEncoder(w).Encode(Gist{ID: "bb00cc11dd22ee33ff44"})
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	client := NewGistClient("test-token")
	client.baseURL = srv.URL

	cfg := &config.SyncConfig{}
	engine := NewSyncEngine(store, client, cfg)

	syncCfg, err := engine.Setup("test-token")
	if err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	if syncCfg.GistID != "bb00cc11dd22ee33ff44" {
		t.Errorf("GistID = %q, want %q", syncCfg.GistID, "bb00cc11dd22ee33ff44")
	}
	if syncCfg.Username != "testuser" {
		t.Errorf("Username = %q, want %q", syncCfg.Username, "testuser")
	}
	if syncCfg.Token != "test-token" {
		t.Errorf("Token = %q, want %q", syncCfg.Token, "test-token")
	}
}

func TestPull_SkipsLocallyDeletedSnippets(t *testing.T) {
	store := openTestStore(t)

	// Simulate: a snippet was previously synced (exists in meta hashes and in Gist)
	// but has been deleted locally. Pull should NOT re-import it.
	deletedSlug := "deleted-locally.md"
	deletedMD := ToFrontmatter(model.Snippet{
		Title:    "Deleted Locally",
		Language: "go",
		Content:  "package deleted",
	})
	deletedHash := ComputeHash(model.Snippet{
		Title:    "Deleted Locally",
		Language: "go",
		Content:  "package deleted",
	})

	// Also include a genuinely new remote snippet (not in meta hashes).
	newRemoteMD := ToFrontmatter(model.Snippet{
		Title:    "Genuinely New",
		Language: "python",
		Content:  "print('hello')",
	})

	metaJSON, _ := json.Marshal(SyncMeta{
		Version: 1,
		Hashes:  map[string]string{deletedSlug: deletedHash},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Gist{
			ID: "aabbccddee0011223344",
			Files: map[string]GistFile{
				deletedSlug:        {Content: deletedMD},
				"genuinely-new.md": {Content: newRemoteMD},
				".snipt-meta.json": {Content: string(metaJSON)},
			},
		})
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	cfg := &config.SyncConfig{GistID: "aabbccddee0011223344", Token: "token"}
	engine := NewSyncEngine(store, client, cfg)

	result, err := engine.Pull()
	if err != nil {
		t.Fatalf("Pull() error: %v", err)
	}

	// Only the genuinely new snippet should be pulled — not the locally deleted one.
	if result.Pulled != 1 {
		t.Errorf("Pulled = %d, want 1", result.Pulled)
	}

	all, _ := store.List(db.ListOpts{})
	if len(all) != 1 {
		t.Fatalf("DB snippet count = %d, want 1", len(all))
	}
	if all[0].Title != "Genuinely New" {
		t.Errorf("Title = %q, want %q", all[0].Title, "Genuinely New")
	}
}

func TestSync_BidirectionalMerge(t *testing.T) {
	store := openTestStore(t)

	localOnly := makeSnippet("loc00001", "Local Only", "local content", "go", nil)
	localOnly.CreatedAt = time.Now().UTC()
	localOnly.UpdatedAt = localOnly.CreatedAt
	store.Create(localOnly)

	remoteOnlyMD := ToFrontmatter(model.Snippet{
		Title:    "Remote Only",
		Language: "python",
		Content:  "remote content",
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET":
			json.NewEncoder(w).Encode(Gist{
				ID: "aabbccddee0011223344",
				Files: map[string]GistFile{
					"remote-only.md":   {Content: remoteOnlyMD},
					".snipt-meta.json": {Content: `{"version":1,"snippet_hashes":{}}`},
				},
			})
		case r.Method == "PATCH":
			json.NewEncoder(w).Encode(Gist{ID: "aabbccddee0011223344"})
		}
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	cfg := &config.SyncConfig{GistID: "aabbccddee0011223344", Token: "token"}
	engine := NewSyncEngine(store, client, cfg)

	result, err := engine.Sync()
	if err != nil {
		t.Fatalf("Sync() error: %v", err)
	}

	if result.Pulled != 1 {
		t.Errorf("Pulled = %d, want 1", result.Pulled)
	}

	if result.Pushed != 1 {
		t.Errorf("Pushed = %d, want 1", result.Pushed)
	}

	all, _ := store.List(db.ListOpts{Sort: "title"})
	if len(all) != 2 {
		t.Fatalf("DB snippet count = %d, want 2", len(all))
	}
}
