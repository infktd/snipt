package db

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/infktd/snipt/src/internal/model"
)

// --- helpers ---

// openTestStore creates a Store backed by a temp database.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

// makeSnippet creates a test snippet with sensible defaults.
func makeSnippet(id, title, content, lang string, tags []string) *model.Snippet {
	return &model.Snippet{
		ID:          id,
		Title:       title,
		Content:     content,
		Language:    lang,
		Description: "test description",
		Source:      "test",
		Tags:        tags,
	}
}

// --- Task 5 tests ---

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	// Verify schema_version is 1.
	var version string
	err = store.db.QueryRow(`SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&version)
	if err != nil {
		t.Fatalf("query schema_version: %v", err)
	}
	if version != "1" {
		t.Errorf("schema_version = %q, want %q", version, "1")
	}
}

func TestOpen_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	// First open -- creates schema.
	store1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open() error: %v", err)
	}
	store1.Close()

	// Second open -- should succeed without error.
	store2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open() error: %v", err)
	}
	defer store2.Close()

	var version string
	err = store2.db.QueryRow(`SELECT value FROM meta WHERE key = 'schema_version'`).Scan(&version)
	if err != nil {
		t.Fatalf("query schema_version: %v", err)
	}
	if version != "1" {
		t.Errorf("schema_version = %q, want %q", version, "1")
	}
}

// --- Task 6 tests ---

func TestCreate_and_Get(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("abc12345", "HTTP Server", "package main\nfunc main() {}", "go", []string{"web", "http"})
	sn.Pinned = true
	sn.UseCount = 3

	if err := store.Create(sn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	got, err := store.Get("abc12345")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	// Verify all fields.
	if got.ID != sn.ID {
		t.Errorf("ID = %q, want %q", got.ID, sn.ID)
	}
	if got.Title != sn.Title {
		t.Errorf("Title = %q, want %q", got.Title, sn.Title)
	}
	if got.Content != sn.Content {
		t.Errorf("Content = %q, want %q", got.Content, sn.Content)
	}
	if got.Language != sn.Language {
		t.Errorf("Language = %q, want %q", got.Language, sn.Language)
	}
	if got.Description != sn.Description {
		t.Errorf("Description = %q, want %q", got.Description, sn.Description)
	}
	if got.Source != sn.Source {
		t.Errorf("Source = %q, want %q", got.Source, sn.Source)
	}
	if got.Pinned != sn.Pinned {
		t.Errorf("Pinned = %v, want %v", got.Pinned, sn.Pinned)
	}
	if got.UseCount != sn.UseCount {
		t.Errorf("UseCount = %d, want %d", got.UseCount, sn.UseCount)
	}
	if len(got.Tags) != 2 {
		t.Fatalf("Tags count = %d, want 2", len(got.Tags))
	}
	// Tags are sorted alphabetically.
	if got.Tags[0] != "http" || got.Tags[1] != "web" {
		t.Errorf("Tags = %v, want [http web]", got.Tags)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero")
	}
}

func TestGet_NotFound(t *testing.T) {
	store := openTestStore(t)

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("Get() expected error for nonexistent ID, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to contain 'not found'", err.Error())
	}
}

func TestUpdate(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("upd12345", "Original", "original content", "go", nil)
	// Set CreatedAt in the past so UpdatedAt will be strictly after it.
	sn.CreatedAt = time.Now().UTC().Add(-10 * time.Second)
	sn.UpdatedAt = sn.CreatedAt
	if err := store.Create(sn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Modify fields.
	sn.Title = "Updated Title"
	sn.Content = "updated content"
	sn.Language = "python"

	if err := store.Update(sn); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, err := store.Get("upd12345")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.Title != "Updated Title" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated Title")
	}
	if got.Content != "updated content" {
		t.Errorf("Content = %q, want %q", got.Content, "updated content")
	}
	if got.Language != "python" {
		t.Errorf("Language = %q, want %q", got.Language, "python")
	}
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Errorf("UpdatedAt (%v) should be after CreatedAt (%v)", got.UpdatedAt, got.CreatedAt)
	}
}

func TestDelete(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("del12345", "To Delete", "delete me", "go", []string{"temp"})
	if err := store.Create(sn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := store.Delete("del12345"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err := store.Get("del12345")
	if err == nil {
		t.Fatal("Get() expected error after delete, got nil")
	}
}

func TestAddTags(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("tag12345", "Tagged", "content", "go", []string{"original"})
	if err := store.Create(sn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := store.AddTags("tag12345", []string{"new", "original"}); err != nil {
		t.Fatalf("AddTags() error: %v", err)
	}

	got, err := store.Get("tag12345")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if len(got.Tags) != 2 {
		t.Fatalf("Tags count = %d, want 2", len(got.Tags))
	}
	// Should have both "new" and "original" (sorted).
	if got.Tags[0] != "new" || got.Tags[1] != "original" {
		t.Errorf("Tags = %v, want [new original]", got.Tags)
	}
}

func TestRemoveTags(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("rtg12345", "Remove Tag", "content", "go", []string{"keep", "remove"})
	if err := store.Create(sn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := store.RemoveTags("rtg12345", []string{"remove", "nonexistent"}); err != nil {
		t.Fatalf("RemoveTags() error: %v", err)
	}

	got, err := store.Get("rtg12345")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "keep" {
		t.Errorf("Tags = %v, want [keep]", got.Tags)
	}
}

func TestSetPinned(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("pin12345", "Pinnable", "content", "go", nil)
	if err := store.Create(sn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := store.SetPinned("pin12345", true); err != nil {
		t.Fatalf("SetPinned(true) error: %v", err)
	}
	got, _ := store.Get("pin12345")
	if !got.Pinned {
		t.Error("Pinned should be true")
	}

	if err := store.SetPinned("pin12345", false); err != nil {
		t.Fatalf("SetPinned(false) error: %v", err)
	}
	got, _ = store.Get("pin12345")
	if got.Pinned {
		t.Error("Pinned should be false")
	}
}

func TestIncrementUseCount(t *testing.T) {
	store := openTestStore(t)

	sn := makeSnippet("use12345", "Usable", "content", "go", nil)
	if err := store.Create(sn); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := store.IncrementUseCount("use12345"); err != nil {
			t.Fatalf("IncrementUseCount() error: %v", err)
		}
	}

	got, _ := store.Get("use12345")
	if got.UseCount != 3 {
		t.Errorf("UseCount = %d, want 3", got.UseCount)
	}
}

func TestList(t *testing.T) {
	store := openTestStore(t)

	// Create snippets with small time gaps to ensure ordering.
	snippets := []*model.Snippet{
		makeSnippet("lst00001", "Alpha", "content a", "go", []string{"web"}),
		makeSnippet("lst00002", "Beta", "content b", "python", []string{"web", "api"}),
		makeSnippet("lst00003", "Gamma", "content c", "go", []string{"cli"}),
	}
	snippets[0].Pinned = true
	snippets[1].UseCount = 5

	for i, sn := range snippets {
		sn.CreatedAt = time.Now().UTC().Add(time.Duration(i) * time.Second)
		sn.UpdatedAt = sn.CreatedAt
		if err := store.Create(sn); err != nil {
			t.Fatalf("Create(%q) error: %v", sn.ID, err)
		}
	}

	t.Run("no filter", func(t *testing.T) {
		results, err := store.List(ListOpts{})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("count = %d, want 3", len(results))
		}
	})

	t.Run("filter by language", func(t *testing.T) {
		results, err := store.List(ListOpts{Language: "go"})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("count = %d, want 2", len(results))
		}
	})

	t.Run("filter by tag", func(t *testing.T) {
		results, err := store.List(ListOpts{Tag: "web"})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("count = %d, want 2", len(results))
		}
	})

	t.Run("filter by pinned", func(t *testing.T) {
		pinned := true
		results, err := store.List(ListOpts{Pinned: &pinned})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("count = %d, want 1", len(results))
		}
		if results[0].ID != "lst00001" {
			t.Errorf("ID = %q, want %q", results[0].ID, "lst00001")
		}
	})

	t.Run("sort by title", func(t *testing.T) {
		results, err := store.List(ListOpts{Sort: "title"})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if results[0].Title != "Alpha" {
			t.Errorf("first title = %q, want Alpha", results[0].Title)
		}
	})

	t.Run("sort by usage", func(t *testing.T) {
		results, err := store.List(ListOpts{Sort: "usage"})
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if results[0].ID != "lst00002" {
			t.Errorf("first ID = %q, want lst00002 (highest usage)", results[0].ID)
		}
	})
}

func TestStats(t *testing.T) {
	store := openTestStore(t)

	snippets := []*model.Snippet{
		makeSnippet("st000001", "Stats One", "content", "go", []string{"web"}),
		makeSnippet("st000002", "Stats Two", "content", "go", []string{"api"}),
		makeSnippet("st000003", "Stats Three", "content", "python", []string{"web", "cli"}),
	}
	snippets[0].UseCount = 10

	for i, sn := range snippets {
		sn.CreatedAt = time.Now().UTC().Add(time.Duration(i) * time.Second)
		sn.UpdatedAt = sn.CreatedAt
		if err := store.Create(sn); err != nil {
			t.Fatalf("Create() error: %v", err)
		}
	}

	stats, err := store.Stats()
	if err != nil {
		t.Fatalf("Stats() error: %v", err)
	}

	if stats.TotalSnippets != 3 {
		t.Errorf("TotalSnippets = %d, want 3", stats.TotalSnippets)
	}
	if stats.TotalTags != 3 {
		t.Errorf("TotalTags = %d, want 3 (web, api, cli)", stats.TotalTags)
	}
	if stats.Languages["go"] != 2 {
		t.Errorf("Languages[go] = %d, want 2", stats.Languages["go"])
	}
	if stats.Languages["python"] != 1 {
		t.Errorf("Languages[python] = %d, want 1", stats.Languages["python"])
	}
	if stats.MostUsed == nil || stats.MostUsed.ID != "st000001" {
		t.Errorf("MostUsed = %v, want st000001", stats.MostUsed)
	}
	if len(stats.RecentlyAdded) != 3 {
		t.Errorf("RecentlyAdded count = %d, want 3", len(stats.RecentlyAdded))
	}
}

// --- Task 7 tests ---

// seedSnippets creates 3 test snippets for search tests and returns the store.
func seedSnippets(t *testing.T) *Store {
	t.Helper()
	store := openTestStore(t)

	snippets := []*model.Snippet{
		makeSnippet("srch0001", "HTTP Server", "package main\nimport \"net/http\"\nfunc main() { http.ListenAndServe(\":8080\", nil) }", "go", []string{"web", "server"}),
		makeSnippet("srch0002", "Python Flask App", "from flask import Flask\napp = Flask(__name__)", "python", []string{"web", "flask"}),
		makeSnippet("srch0003", "Bash Backup Script", "#!/bin/bash\ntar -czf backup.tar.gz /data", "bash", []string{"ops", "backup"}),
	}

	for i, sn := range snippets {
		sn.CreatedAt = time.Now().UTC().Add(time.Duration(i) * time.Second)
		sn.UpdatedAt = sn.CreatedAt
		if err := store.Create(sn); err != nil {
			t.Fatalf("seed Create(%q) error: %v", sn.ID, err)
		}
	}

	return store
}

func TestSearch(t *testing.T) {
	store := seedSnippets(t)

	results, err := store.Search("http server")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search() returned 0 results, want at least 1")
	}

	// The top result should be the HTTP Server snippet.
	top := results[0]
	if top.Snippet.ID != "srch0001" {
		t.Errorf("top result ID = %q, want %q", top.Snippet.ID, "srch0001")
	}
	if top.Score <= 0 {
		t.Errorf("score = %f, want > 0", top.Score)
	}
}

func TestResolveRef_ExactID(t *testing.T) {
	store := seedSnippets(t)

	results, err := store.ResolveRef("srch0002")
	if err != nil {
		t.Fatalf("ResolveRef() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	if results[0].Snippet.ID != "srch0002" {
		t.Errorf("ID = %q, want %q", results[0].Snippet.ID, "srch0002")
	}
	if results[0].Score != 1.0 {
		t.Errorf("score = %f, want 1.0 (exact ID match)", results[0].Score)
	}
}

func TestResolveRef_ExactTitle(t *testing.T) {
	store := seedSnippets(t)

	// Case-insensitive title match.
	results, err := store.ResolveRef("http server")
	if err != nil {
		t.Fatalf("ResolveRef() error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	if results[0].Snippet.ID != "srch0001" {
		t.Errorf("ID = %q, want %q", results[0].Snippet.ID, "srch0001")
	}
	if results[0].Score != 0.9 {
		t.Errorf("score = %f, want 0.9 (exact title match)", results[0].Score)
	}
}

func TestResolveRef_NoMatch(t *testing.T) {
	store := seedSnippets(t)

	results, err := store.ResolveRef("zzz_nonexistent_zzz")
	if err != nil {
		t.Fatalf("ResolveRef() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results count = %d, want 0", len(results))
	}
}

func TestGetAndTrack(t *testing.T) {
	store := seedSnippets(t)

	// Verify initial use_count is 0.
	before, err := store.Get("srch0001")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if before.UseCount != 0 {
		t.Fatalf("initial UseCount = %d, want 0", before.UseCount)
	}

	// GetAndTrack should resolve and bump use_count.
	got, err := store.GetAndTrack("HTTP Server")
	if err != nil {
		t.Fatalf("GetAndTrack() error: %v", err)
	}
	if got.ID != "srch0001" {
		t.Errorf("ID = %q, want %q", got.ID, "srch0001")
	}
	if got.UseCount != 1 {
		t.Errorf("UseCount = %d, want 1", got.UseCount)
	}

	// Call again -- use_count should increment again.
	got2, err := store.GetAndTrack("srch0001")
	if err != nil {
		t.Fatalf("GetAndTrack() second call error: %v", err)
	}
	if got2.UseCount != 2 {
		t.Errorf("UseCount = %d, want 2", got2.UseCount)
	}
}

func TestGetAndTrack_NotFound(t *testing.T) {
	store := seedSnippets(t)

	_, err := store.GetAndTrack("zzz_nonexistent_zzz")
	if err == nil {
		t.Fatal("GetAndTrack() expected error for nonexistent ref, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("error type = %T, want *NotFoundError", err)
	}
	nfe := err.(*NotFoundError)
	if nfe.Ref != "zzz_nonexistent_zzz" {
		t.Errorf("NotFoundError.Ref = %q, want %q", nfe.Ref, "zzz_nonexistent_zzz")
	}
}
