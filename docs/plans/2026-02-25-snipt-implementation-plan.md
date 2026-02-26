# snipt Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a CLI snippet manager with SQLite storage and a Bubbletea fuzzy-find command palette.

**Architecture:** Go module at `github.com/infktd/snipt` with all source under `src/`. Standalone packages (`model`, `lang`, `clipboard`, `config`) have no internal imports. `db` imports only `model`. `cli` and `tui` wire everything together.

**Tech Stack:** Go 1.25, modernc.org/sqlite, cobra, bubbletea/lipgloss/bubbles, BurntSushi/toml, google/uuid

**Design doc:** `docs/plans/2026-02-25-snipt-phase1-phase2-design.md`

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `src/cmd/snipt/main.go`
- Create: `src/internal/model/snippet.go`
- Create: `src/internal/model/exits.go`
- Create: `.goreleaser.yaml`
- Create: `LICENSE`

**Step 1: Initialize Go module**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt
go mod init github.com/infktd/snipt
```

**Step 2: Create main.go entry point**

Create `src/cmd/snipt/main.go`:
```go
package main

import (
	"fmt"
	"os"
)

// Version is set by goreleaser at build time.
var version = "dev"

func main() {
	fmt.Fprintln(os.Stderr, "snipt", version)
	os.Exit(0)
}
```

**Step 3: Create model package — Snippet struct**

Create `src/internal/model/snippet.go`:
```go
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
```

**Step 4: Create model package — exit codes**

Create `src/internal/model/exits.go`:
```go
package model

const (
	ExitOK          = 0
	ExitError       = 1
	ExitNotFound    = 2
	ExitDBError     = 3
	ExitInterrupted = 130
)
```

**Step 5: Create model package — stats type**

Add to `src/internal/model/snippet.go` (append):
```go
// Stats holds collection overview data.
type Stats struct {
	TotalSnippets int
	TotalTags     int
	Languages     map[string]int // language → count
	MostUsed      *Snippet
	RecentlyAdded []Snippet
}

// SearchResult pairs a snippet with its relevance score.
type SearchResult struct {
	Snippet      Snippet
	Score        float64
	TitleIndices []int // matched character positions for highlighting
}
```

**Step 6: Create .goreleaser.yaml**

Create `.goreleaser.yaml`:
```yaml
version: 2
project_name: snipt

builds:
  - main: ./src/cmd/snipt
    binary: snipt
    ldflags:
      - -s -w -X main.version={{.Version}}
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - formats: [tar.gz]
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

changelog:
  sort: asc
```

**Step 7: Create LICENSE**

Create `LICENSE` with MIT license, copyright `infktd`.

**Step 8: Add uuid dependency and verify build**

Run:
```bash
go get github.com/google/uuid
go build ./src/cmd/snipt
./snipt
```
Expected: prints `snipt dev`

**Step 9: Write test for NewID**

Create `src/internal/model/snippet_test.go`:
```go
package model

import "testing"

func TestNewID(t *testing.T) {
	id := NewID()
	if len(id) != 8 {
		t.Errorf("expected 8-char ID, got %d chars: %q", len(id), id)
	}

	// IDs should be unique
	id2 := NewID()
	if id == id2 {
		t.Errorf("expected unique IDs, got same: %q", id)
	}
}
```

**Step 10: Run test**

Run: `go test ./src/internal/model/ -v`
Expected: PASS

**Step 11: Commit**

```bash
git add go.mod go.sum src/ .goreleaser.yaml LICENSE
git commit -m "feat: scaffold project with model package and entry point"
```

---

## Task 2: Language Detection Package

**Files:**
- Create: `src/internal/lang/lang.go`
- Create: `src/internal/lang/lang_test.go`

**Step 1: Write failing test**

Create `src/internal/lang/lang_test.go`:
```go
package lang

import "testing"

func TestFromExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"script.py", "python"},
		{"app.ts", "typescript"},
		{"index.js", "javascript"},
		{"lib.rs", "rust"},
		{"init.lua", "lua"},
		{"deploy.sh", "bash"},
		{"setup.bash", "bash"},
		{"query.sql", "sql"},
		{"flake.nix", "nix"},
		{"Gemfile.rb", "ruby"},
		{"Main.java", "java"},
		{"util.c", "c"},
		{"util.cpp", "cpp"},
		{"README.md", "markdown"},
		{"config.toml", "toml"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"data.json", "json"},
		{"noext", ""},
		{"weird.xyz", ""},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := FromExtension(tt.filename)
			if got != tt.want {
				t.Errorf("FromExtension(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/internal/lang/ -v`
Expected: FAIL — `FromExtension` not defined

**Step 3: Implement lang.go**

Create `src/internal/lang/lang.go`:
```go
package lang

import (
	"path/filepath"
	"strings"
)

var extMap = map[string]string{
	".go":   "go",
	".py":   "python",
	".ts":   "typescript",
	".js":   "javascript",
	".rs":   "rust",
	".lua":  "lua",
	".sh":   "bash",
	".bash": "bash",
	".sql":  "sql",
	".nix":  "nix",
	".rb":   "ruby",
	".java": "java",
	".c":    "c",
	".cpp":  "cpp",
	".md":   "markdown",
	".toml": "toml",
	".yaml": "yaml",
	".yml":  "yaml",
	".json": "json",
}

// FromExtension returns the language for a filename based on its extension.
// Returns empty string if the extension is not recognized.
func FromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return extMap[ext]
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/internal/lang/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/internal/lang/
git commit -m "feat: add language detection from file extension"
```

---

## Task 3: Clipboard Package

**Files:**
- Create: `src/internal/clipboard/clipboard.go`
- Create: `src/internal/clipboard/clipboard_test.go`

**Step 1: Write the test**

Create `src/internal/clipboard/clipboard_test.go`:
```go
package clipboard

import (
	"runtime"
	"testing"
)

func TestAvailable(t *testing.T) {
	// On macOS, pbcopy should always be available
	if runtime.GOOS == "darwin" {
		if !Available() {
			t.Error("expected clipboard to be available on macOS")
		}
	}
}

func TestRoundTrip(t *testing.T) {
	if !Available() {
		t.Skip("no clipboard tool available")
	}

	text := "snipt-test-clipboard-roundtrip"
	if err := Write(text); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got, err := Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if got != text {
		t.Errorf("clipboard round-trip failed: got %q, want %q", got, text)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/internal/clipboard/ -v`
Expected: FAIL — functions not defined

**Step 3: Implement clipboard.go**

Create `src/internal/clipboard/clipboard.go`:
```go
package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type tool struct {
	copyCmd  string
	copyArgs []string
	pasteCmd string
	pasteArgs []string
}

func detect() *tool {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("pbcopy"); err == nil {
			return &tool{
				copyCmd: "pbcopy", copyArgs: nil,
				pasteCmd: "pbpaste", pasteArgs: nil,
			}
		}
	default: // linux, etc.
		if _, err := exec.LookPath("wl-copy"); err == nil {
			return &tool{
				copyCmd: "wl-copy", copyArgs: nil,
				pasteCmd: "wl-paste", pasteArgs: []string{"--no-newline"},
			}
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return &tool{
				copyCmd: "xclip", copyArgs: []string{"-selection", "clipboard"},
				pasteCmd: "xclip", pasteArgs: []string{"-selection", "clipboard", "-o"},
			}
		}
	}
	return nil
}

// Available returns true if a clipboard tool is detected.
func Available() bool {
	return detect() != nil
}

// Write copies text to the system clipboard.
func Write(text string) error {
	t := detect()
	if t == nil {
		return fmt.Errorf("no clipboard tool found (need pbcopy, xclip, or wl-copy)")
	}
	cmd := exec.Command(t.copyCmd, t.copyArgs...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// Read returns the current clipboard contents.
func Read() (string, error) {
	t := detect()
	if t == nil {
		return "", fmt.Errorf("no clipboard tool found (need pbcopy, xclip, or wl-copy)")
	}
	cmd := exec.Command(t.pasteCmd, t.pasteArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("clipboard read failed: %w", err)
	}
	return string(out), nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/internal/clipboard/ -v`
Expected: PASS (on macOS)

**Step 5: Commit**

```bash
git add src/internal/clipboard/
git commit -m "feat: add platform clipboard read/write support"
```

---

## Task 4: Config Package

**Files:**
- Create: `src/internal/config/config.go`
- Create: `src/internal/config/config_test.go`

**Step 1: Write failing test**

Create `src/internal/config/config_test.go`:
```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Use a temp dir so we don't touch real config
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Editor != "" {
		t.Errorf("expected empty editor default, got %q", cfg.Editor)
	}
	if cfg.DefaultLanguage != "text" {
		t.Errorf("expected default_language=text, got %q", cfg.DefaultLanguage)
	}
	if cfg.Theme != "catppuccin-mocha" {
		t.Errorf("expected theme=catppuccin-mocha, got %q", cfg.Theme)
	}

	// Config file should have been created
	cfgFile := filepath.Join(dir, "snipt", "config.toml")
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "snipt")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(`
editor = "code"
default_language = "go"
theme = "dracula"
`), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Editor != "code" {
		t.Errorf("expected editor=code, got %q", cfg.Editor)
	}
	if cfg.DefaultLanguage != "go" {
		t.Errorf("expected default_language=go, got %q", cfg.DefaultLanguage)
	}
}

func TestResolveEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	cfg := &Config{Editor: "nvim"}
	if got := cfg.ResolveEditor(); got != "nvim" {
		t.Errorf("expected nvim from config, got %q", got)
	}

	cfg.Editor = ""
	t.Setenv("VISUAL", "code")
	if got := cfg.ResolveEditor(); got != "code" {
		t.Errorf("expected code from $VISUAL, got %q", got)
	}

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nano")
	if got := cfg.ResolveEditor(); got != "nano" {
		t.Errorf("expected nano from $EDITOR, got %q", got)
	}

	t.Setenv("EDITOR", "")
	if got := cfg.ResolveEditor(); got != "vi" {
		t.Errorf("expected vi fallback, got %q", got)
	}
}

func TestDBPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	path := DBPath("")
	expected := filepath.Join(dir, "snipt", "snipt.db")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestDBPath_Override(t *testing.T) {
	override := "/tmp/custom.db"
	path := DBPath(override)
	if path != override {
		t.Errorf("expected override %q, got %q", override, path)
	}
}

func TestConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path := ConfigPath()
	expected := filepath.Join(dir, "snipt", "config.toml")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/internal/config/ -v`
Expected: FAIL

**Step 3: Implement config.go**

Create `src/internal/config/config.go`:
```go
package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds user preferences loaded from config.toml.
type Config struct {
	Editor          string `toml:"editor"`
	DefaultLanguage string `toml:"default_language"`
	Theme           string `toml:"theme"`
}

const defaultConfig = `editor = ""
default_language = "text"
theme = "catppuccin-mocha"
`

// Load reads the config file, creating it with defaults if it doesn't exist.
func Load() (*Config, error) {
	path := ConfigPath()

	cfg := &Config{
		DefaultLanguage: "text",
		Theme:           "catppuccin-mocha",
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Create default config
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(defaultConfig), 0o644); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ResolveEditor returns the editor to use, following the chain:
// config editor → $VISUAL → $EDITOR → vi
func (c *Config) ResolveEditor() string {
	if c.Editor != "" {
		return c.Editor
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vi"
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	cfgHome := os.Getenv("XDG_CONFIG_HOME")
	if cfgHome == "" {
		home, _ := os.UserHomeDir()
		cfgHome = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgHome, "snipt", "config.toml")
}

// DBPath returns the path to the SQLite database.
// If override is non-empty, it is returned directly.
func DBPath(override string) string {
	if override != "" {
		return override
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "snipt", "snipt.db")
}
```

**Step 4: Add dependency and run tests**

Run:
```bash
go get github.com/BurntSushi/toml
go test ./src/internal/config/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add src/internal/config/ go.mod go.sum
git commit -m "feat: add config package with XDG paths and editor resolution"
```

---

## Task 5: SQLite Database Layer — Schema & Open

**Files:**
- Create: `src/internal/db/db.go`
- Create: `src/internal/db/migrate.go`
- Create: `src/internal/db/db_test.go`

**Step 1: Write failing test for Open + migration**

Create `src/internal/db/db_test.go`:
```go
package db

import (
	"path/filepath"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestOpen(t *testing.T) {
	store := testStore(t)

	// Verify schema version
	var version string
	err := store.db.QueryRow("SELECT value FROM meta WHERE key = 'schema_version'").Scan(&version)
	if err != nil {
		t.Fatalf("failed to read schema_version: %v", err)
	}
	if version != "1" {
		t.Errorf("expected schema_version=1, got %q", version)
	}
}

func TestOpen_Idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")

	// Open twice — second open should not fail
	s1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open failed: %v", err)
	}
	s1.Close()

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open failed: %v", err)
	}
	s2.Close()
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/internal/db/ -v`
Expected: FAIL

**Step 3: Implement db.go and migrate.go**

Create `src/internal/db/db.go`:
```go
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store wraps a SQLite database connection for snippet operations.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at path and runs migrations.
func Open(path string) (*Store, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode and foreign keys
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return store, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
```

Create `src/internal/db/migrate.go`:
```go
package db

import "database/sql"

const schemaV1 = `
CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT
);

CREATE TABLE IF NOT EXISTS snippets (
    id          TEXT PRIMARY KEY,
    title       TEXT,
    content     TEXT NOT NULL,
    language    TEXT,
    description TEXT,
    source      TEXT,
    pinned      INTEGER DEFAULT 0,
    use_count   INTEGER DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tags (
    snippet_id TEXT REFERENCES snippets(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (snippet_id, tag)
);

CREATE VIRTUAL TABLE IF NOT EXISTS snippets_fts USING fts5(
    title, content, description,
    content='snippets',
    content_rowid='rowid'
);

CREATE TRIGGER IF NOT EXISTS snippets_ai AFTER INSERT ON snippets BEGIN
    INSERT INTO snippets_fts(rowid, title, content, description)
    VALUES (new.rowid, new.title, new.content, new.description);
END;

CREATE TRIGGER IF NOT EXISTS snippets_ad AFTER DELETE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, title, content, description)
    VALUES ('delete', old.rowid, old.title, old.content, old.description);
END;

CREATE TRIGGER IF NOT EXISTS snippets_au AFTER UPDATE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, title, content, description)
    VALUES ('delete', old.rowid, old.title, old.content, old.description);
    INSERT INTO snippets_fts(rowid, title, content, description)
    VALUES (new.rowid, new.title, new.content, new.description);
END;
`

func (s *Store) migrate() error {
	// Check current version
	var version string
	err := s.db.QueryRow("SELECT value FROM meta WHERE key = 'schema_version'").Scan(&version)
	if err == sql.ErrNoRows || err != nil {
		// Fresh database — run initial migration
		if _, err := s.db.Exec(schemaV1); err != nil {
			return err
		}
		_, err := s.db.Exec("INSERT OR REPLACE INTO meta (key, value) VALUES ('schema_version', '1')")
		return err
	}

	// Future migrations would go here:
	// if version == "1" { migrate to v2 }

	return nil
}
```

**Step 4: Add dependency and run tests**

Run:
```bash
go get modernc.org/sqlite
go test ./src/internal/db/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add src/internal/db/db.go src/internal/db/migrate.go src/internal/db/db_test.go go.mod go.sum
git commit -m "feat: add SQLite layer with schema migration and FTS5"
```

---

## Task 6: DB CRUD — Create & Get

**Files:**
- Create: `src/internal/db/crud.go`
- Modify: `src/internal/db/db_test.go`

**Step 1: Write failing tests**

Append to `src/internal/db/db_test.go`:
```go
import (
	"testing"
	"time"
	"path/filepath"

	"github.com/infktd/snipt/src/internal/model"
)

func TestCreate_and_Get(t *testing.T) {
	store := testStore(t)
	now := time.Now()

	s := &model.Snippet{
		ID:        "abcd1234",
		Title:     "Test snippet",
		Content:   "fmt.Println(\"hello\")",
		Language:  "go",
		Tags:      []string{"test", "hello"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.Create(s); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := store.Get("abcd1234")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.Title != "Test snippet" {
		t.Errorf("title = %q, want %q", got.Title, "Test snippet")
	}
	if got.Content != "fmt.Println(\"hello\")" {
		t.Errorf("content mismatch")
	}
	if got.Language != "go" {
		t.Errorf("language = %q, want %q", got.Language, "go")
	}
	if len(got.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(got.Tags))
	}
}

func TestGet_NotFound(t *testing.T) {
	store := testStore(t)

	_, err := store.Get("nonexist")
	if err == nil {
		t.Error("expected error for nonexistent snippet")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/internal/db/ -v -run TestCreate`
Expected: FAIL

**Step 3: Implement crud.go**

Create `src/internal/db/crud.go`:
```go
package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/infktd/snipt/src/internal/model"
)

// Create inserts a new snippet and its tags.
func (s *Store) Create(snip *model.Snippet) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO snippets (id, title, content, language, description, source, pinned, use_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		snip.ID, snip.Title, snip.Content, snip.Language, snip.Description,
		snip.Source, boolToInt(snip.Pinned), snip.UseCount,
		snip.CreatedAt.Format(time.RFC3339), snip.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert snippet: %w", err)
	}

	for _, tag := range snip.Tags {
		if _, err := tx.Exec("INSERT INTO tags (snippet_id, tag) VALUES (?, ?)", snip.ID, tag); err != nil {
			return fmt.Errorf("insert tag %q: %w", tag, err)
		}
	}

	return tx.Commit()
}

// Get retrieves a snippet by exact ID.
func (s *Store) Get(id string) (*model.Snippet, error) {
	snip, err := s.scanSnippet(s.db.QueryRow(`
		SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at
		FROM snippets WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("snippet %q not found", id)
	}
	if err != nil {
		return nil, err
	}

	tags, err := s.getTags(id)
	if err != nil {
		return nil, err
	}
	snip.Tags = tags

	return snip, nil
}

// Update modifies an existing snippet.
func (s *Store) Update(snip *model.Snippet) error {
	snip.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE snippets SET title=?, content=?, language=?, description=?, source=?, pinned=?, updated_at=?
		WHERE id=?`,
		snip.Title, snip.Content, snip.Language, snip.Description,
		snip.Source, boolToInt(snip.Pinned), snip.UpdatedAt.Format(time.RFC3339),
		snip.ID,
	)
	return err
}

// Delete removes a snippet by ID. Tags are cascade-deleted.
func (s *Store) Delete(id string) error {
	result, err := s.db.Exec("DELETE FROM snippets WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("snippet %q not found", id)
	}
	return nil
}

// AddTags adds tags to a snippet. Idempotent.
func (s *Store) AddTags(id string, tags []string) error {
	for _, tag := range tags {
		_, err := s.db.Exec("INSERT OR IGNORE INTO tags (snippet_id, tag) VALUES (?, ?)", id, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveTags removes tags from a snippet. Idempotent.
func (s *Store) RemoveTags(id string, tags []string) error {
	for _, tag := range tags {
		if _, err := s.db.Exec("DELETE FROM tags WHERE snippet_id = ? AND tag = ?", id, tag); err != nil {
			return err
		}
	}
	return nil
}

// SetPinned sets the pinned status of a snippet.
func (s *Store) SetPinned(id string, pinned bool) error {
	_, err := s.db.Exec("UPDATE snippets SET pinned = ?, updated_at = ? WHERE id = ?",
		boolToInt(pinned), time.Now().Format(time.RFC3339), id)
	return err
}

// IncrementUseCount bumps the use_count for a snippet.
func (s *Store) IncrementUseCount(id string) error {
	_, err := s.db.Exec("UPDATE snippets SET use_count = use_count + 1 WHERE id = ?", id)
	return err
}

// ListOpts configures filtering and sorting for List.
type ListOpts struct {
	Language string
	Tag      string
	Pinned   *bool // nil = no filter
	Sort     string // "created", "updated", "usage", "title"
}

// List returns snippets matching the given filters.
func (s *Store) List(opts ListOpts) ([]model.Snippet, error) {
	query := "SELECT DISTINCT s.id, s.title, s.content, s.language, s.description, s.source, s.pinned, s.use_count, s.created_at, s.updated_at FROM snippets s"
	var conditions []string
	var args []any

	if opts.Tag != "" {
		query += " JOIN tags t ON s.id = t.snippet_id"
		conditions = append(conditions, "t.tag = ?")
		args = append(args, opts.Tag)
	}

	if opts.Language != "" {
		conditions = append(conditions, "s.language = ?")
		args = append(args, opts.Language)
	}

	if opts.Pinned != nil {
		conditions = append(conditions, "s.pinned = ?")
		args = append(args, boolToInt(*opts.Pinned))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch opts.Sort {
	case "updated":
		query += " ORDER BY s.updated_at DESC"
	case "usage":
		query += " ORDER BY s.use_count DESC"
	case "title":
		query += " ORDER BY s.title ASC"
	default: // "created" or empty
		query += " ORDER BY s.created_at DESC"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []model.Snippet
	for rows.Next() {
		snip, err := s.scanSnippetRows(rows)
		if err != nil {
			return nil, err
		}
		tags, err := s.getTags(snip.ID)
		if err != nil {
			return nil, err
		}
		snip.Tags = tags
		snippets = append(snippets, *snip)
	}
	return snippets, rows.Err()
}

// Stats returns collection overview data.
func (s *Store) Stats() (*model.Stats, error) {
	stats := &model.Stats{Languages: make(map[string]int)}

	// Total snippets
	s.db.QueryRow("SELECT COUNT(*) FROM snippets").Scan(&stats.TotalSnippets)

	// Total unique tags
	s.db.QueryRow("SELECT COUNT(DISTINCT tag) FROM tags").Scan(&stats.TotalTags)

	// Language counts
	rows, err := s.db.Query("SELECT language, COUNT(*) FROM snippets WHERE language != '' GROUP BY language ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var lang string
		var count int
		if err := rows.Scan(&lang, &count); err != nil {
			return nil, err
		}
		stats.Languages[lang] = count
	}

	// Most used snippet
	row := s.db.QueryRow("SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at FROM snippets ORDER BY use_count DESC LIMIT 1")
	most, err := s.scanSnippet(row)
	if err == nil {
		tags, _ := s.getTags(most.ID)
		most.Tags = tags
		stats.MostUsed = most
	}

	// Recently added (5)
	recentRows, err := s.db.Query("SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at FROM snippets ORDER BY created_at DESC LIMIT 5")
	if err != nil {
		return nil, err
	}
	defer recentRows.Close()
	for recentRows.Next() {
		snip, err := s.scanSnippetRows(recentRows)
		if err != nil {
			return nil, err
		}
		tags, _ := s.getTags(snip.ID)
		snip.Tags = tags
		stats.RecentlyAdded = append(stats.RecentlyAdded, *snip)
	}

	return stats, nil
}

// --- helpers ---

func (s *Store) getTags(snippetID string) ([]string, error) {
	rows, err := s.db.Query("SELECT tag FROM tags WHERE snippet_id = ? ORDER BY tag", snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, rows.Err()
}

func (s *Store) scanSnippet(row *sql.Row) (*model.Snippet, error) {
	var snip model.Snippet
	var pinned int
	var createdAt, updatedAt string
	err := row.Scan(&snip.ID, &snip.Title, &snip.Content, &snip.Language,
		&snip.Description, &snip.Source, &pinned, &snip.UseCount,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	snip.Pinned = pinned != 0
	snip.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	snip.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &snip, nil
}

func (s *Store) scanSnippetRows(rows *sql.Rows) (*model.Snippet, error) {
	var snip model.Snippet
	var pinned int
	var createdAt, updatedAt string
	err := rows.Scan(&snip.ID, &snip.Title, &snip.Content, &snip.Language,
		&snip.Description, &snip.Source, &pinned, &snip.UseCount,
		&createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	snip.Pinned = pinned != 0
	snip.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	snip.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &snip, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
```

**Step 4: Run tests**

Run: `go test ./src/internal/db/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/internal/db/
git commit -m "feat: add CRUD operations, list, tags, stats to db layer"
```

---

## Task 7: DB Search & Resolve

**Files:**
- Create: `src/internal/db/search.go`
- Modify: `src/internal/db/db_test.go`

**Step 1: Write failing tests**

Append to `src/internal/db/db_test.go`:
```go
func seedSnippets(t *testing.T, store *Store) {
	t.Helper()
	now := time.Now()
	snippets := []model.Snippet{
		{ID: "aaaa1111", Title: "HTTP server with middleware", Content: "func main() { http.ListenAndServe() }", Language: "go", Tags: []string{"server", "http"}, Pinned: true, CreatedAt: now, UpdatedAt: now},
		{ID: "bbbb2222", Title: "Retry with exponential backoff", Content: "func retry(attempts int) {}", Language: "go", Tags: []string{"resilience"}, CreatedAt: now, UpdatedAt: now},
		{ID: "cccc3333", Title: "Nix flake dev shell", Content: "{ inputs, outputs }", Language: "nix", Tags: []string{"devshell"}, CreatedAt: now, UpdatedAt: now},
	}
	for _, s := range snippets {
		s := s
		if err := store.Create(&s); err != nil {
			t.Fatalf("seed Create failed: %v", err)
		}
	}
}

func TestSearch(t *testing.T) {
	store := testStore(t)
	seedSnippets(t, store)

	results, err := store.Search("http server")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}
	if results[0].Snippet.ID != "aaaa1111" {
		t.Errorf("expected top result to be aaaa1111, got %s", results[0].Snippet.ID)
	}
}

func TestResolveRef_ExactID(t *testing.T) {
	store := testStore(t)
	seedSnippets(t, store)

	results, err := store.ResolveRef("aaaa1111")
	if err != nil {
		t.Fatalf("ResolveRef failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Snippet.ID != "aaaa1111" {
		t.Errorf("expected aaaa1111, got %s", results[0].Snippet.ID)
	}
}

func TestResolveRef_ExactTitle(t *testing.T) {
	store := testStore(t)
	seedSnippets(t, store)

	results, err := store.ResolveRef("Nix flake dev shell")
	if err != nil {
		t.Fatalf("ResolveRef failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Snippet.ID != "cccc3333" {
		t.Errorf("expected cccc3333, got %s", results[0].Snippet.ID)
	}
}

func TestResolveRef_NoMatch(t *testing.T) {
	store := testStore(t)
	seedSnippets(t, store)

	results, err := store.ResolveRef("zzzznonexistent")
	if err != nil {
		t.Fatalf("ResolveRef failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestGetAndTrack(t *testing.T) {
	store := testStore(t)
	seedSnippets(t, store)

	snip, err := store.GetAndTrack("aaaa1111")
	if err != nil {
		t.Fatalf("GetAndTrack failed: %v", err)
	}
	if snip.ID != "aaaa1111" {
		t.Errorf("expected aaaa1111, got %s", snip.ID)
	}

	// Verify use_count was bumped
	got, _ := store.Get("aaaa1111")
	if got.UseCount != 1 {
		t.Errorf("expected use_count=1 after GetAndTrack, got %d", got.UseCount)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/internal/db/ -v -run "TestSearch|TestResolve|TestGetAndTrack"`
Expected: FAIL

**Step 3: Implement search.go**

Create `src/internal/db/search.go`:
```go
package db

import (
	"database/sql"
	"strings"

	"github.com/infktd/snipt/src/internal/model"
)

// Search performs an FTS5 search and returns scored results.
func (s *Store) Search(query string) ([]model.SearchResult, error) {
	// FTS5 query: wrap each word in quotes for phrase matching
	words := strings.Fields(query)
	ftsQuery := strings.Join(words, " AND ")

	rows, err := s.db.Query(`
		SELECT s.id, s.title, s.content, s.language, s.description, s.source,
		       s.pinned, s.use_count, s.created_at, s.updated_at,
		       rank
		FROM snippets_fts fts
		JOIN snippets s ON s.rowid = fts.rowid
		WHERE snippets_fts MATCH ?
		ORDER BY rank`, ftsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.SearchResult
	for rows.Next() {
		var snip model.Snippet
		var pinned int
		var createdAt, updatedAt string
		var rank float64

		err := rows.Scan(&snip.ID, &snip.Title, &snip.Content, &snip.Language,
			&snip.Description, &snip.Source, &pinned, &snip.UseCount,
			&createdAt, &updatedAt, &rank)
		if err != nil {
			return nil, err
		}
		snip.Pinned = pinned != 0
		snip.CreatedAt, _ = parseTime(createdAt)
		snip.UpdatedAt, _ = parseTime(updatedAt)

		tags, _ := s.getTags(snip.ID)
		snip.Tags = tags

		// FTS5 rank is negative (lower = better), invert for our scoring
		results = append(results, model.SearchResult{
			Snippet: snip,
			Score:   -rank,
		})
	}
	return results, rows.Err()
}

// ResolveRef resolves a snippet reference by:
// 1. Exact ID match
// 2. Exact title match (case-insensitive)
// 3. FTS5 fuzzy search
func (s *Store) ResolveRef(ref string) ([]model.SearchResult, error) {
	// 1. Exact ID match
	snip, err := s.Get(ref)
	if err == nil {
		return []model.SearchResult{{Snippet: *snip, Score: 1000}}, nil
	}

	// 2. Exact title match (case-insensitive)
	row := s.db.QueryRow(`
		SELECT id, title, content, language, description, source, pinned, use_count, created_at, updated_at
		FROM snippets WHERE LOWER(title) = LOWER(?)`, ref)
	snip, err = s.scanSnippet(row)
	if err == nil {
		tags, _ := s.getTags(snip.ID)
		snip.Tags = tags
		return []model.SearchResult{{Snippet: *snip, Score: 999}}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// 3. FTS5 search
	return s.Search(ref)
}

// GetAndTrack resolves a ref and bumps use_count on the top result.
// Returns the top-scoring snippet.
func (s *Store) GetAndTrack(ref string) (*model.Snippet, error) {
	results, err := s.ResolveRef(ref)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, &NotFoundError{Ref: ref}
	}

	top := &results[0].Snippet
	s.IncrementUseCount(top.ID)
	return top, nil
}

// NotFoundError is returned when no snippet matches a reference.
type NotFoundError struct {
	Ref string
}

func (e *NotFoundError) Error() string {
	return "no snippet matching \"" + e.Ref + "\""
}
```

Add to `src/internal/db/db.go` (helper):
```go
import "time"

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
```

Actually, since `parseTime` is a small helper, put it at the bottom of `crud.go` instead of modifying `db.go`. Append to `crud.go`:
```go
func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
```

**Step 4: Run tests**

Run: `go test ./src/internal/db/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/internal/db/
git commit -m "feat: add FTS5 search, ref resolution, and use-count tracking"
```

---

## Task 8: Cobra CLI Root + Version + Config Commands

**Files:**
- Create: `src/internal/cli/root.go`
- Create: `src/internal/cli/config.go`
- Modify: `src/cmd/snipt/main.go`

**Step 1: Implement root.go**

Create `src/internal/cli/root.go`:
```go
package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

var (
	dbPath  string
	noColor bool
	store   *db.Store
	cfg     *config.Config
)

// NewRootCmd creates the root cobra command.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:     "snipt",
		Short:   "Cut once, paste forever",
		Long:    "snipt is a CLI snippet manager. Save, search, and reuse code snippets from your terminal.",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip DB init for commands that don't need it
			if cmd.Name() == "path" || (cmd.Name() == "config" && len(args) == 0) {
				return nil
			}

			var err error
			cfg, err = config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
				cfg = &config.Config{DefaultLanguage: "text", Theme: "catppuccin-mocha"}
			}

			path := config.DBPath(dbPath)
			store, err = db.Open(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: could not open database: %v\n", err)
				os.Exit(model.ExitDBError)
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if store != nil {
				store.Close()
			}
		},
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "", "path to database file")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")

	root.AddCommand(newConfigCmd())

	return root
}
```

**Step 2: Implement config.go**

Create `src/internal/cli/config.go`:
```go
package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Open config file in editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := config.ConfigPath()

			// Ensure config exists
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			editor := cfg.ResolveEditor()
			c := exec.Command(editor, path)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.ConfigPath())
		},
	})

	return cmd
}
```

**Step 3: Update main.go**

Replace `src/cmd/snipt/main.go`:
```go
package main

import (
	"os"

	"github.com/infktd/snipt/src/internal/cli"
	"github.com/infktd/snipt/src/internal/model"
)

var version = "dev"

func main() {
	root := cli.NewRootCmd(version)
	if err := root.Execute(); err != nil {
		os.Exit(model.ExitError)
	}
}
```

**Step 4: Add cobra dependency and build**

Run:
```bash
go get github.com/spf13/cobra
go build ./src/cmd/snipt
./snipt --version
./snipt config path
```
Expected: prints version, then config path

**Step 5: Commit**

```bash
git add src/ go.mod go.sum
git commit -m "feat: add Cobra CLI root with config commands"
```

---

## Task 9: `snipt add` Command

**Files:**
- Create: `src/internal/cli/add.go`
- Modify: `src/internal/cli/root.go` (register command)

**Step 1: Implement add.go**

Create `src/internal/cli/add.go`:
```go
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/infktd/snipt/src/internal/clipboard"
	"github.com/infktd/snipt/src/internal/lang"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var (
		title         string
		language      string
		tags          string
		desc          string
		source        string
		fromClipboard bool
	)

	cmd := &cobra.Command{
		Use:   "add [file]",
		Short: "Add a snippet from a file, stdin, or clipboard",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string
			var detectedLang string

			switch {
			case fromClipboard:
				text, err := clipboard.Read()
				if err != nil {
					return err
				}
				content = text

			case len(args) == 1:
				// Read from file
				filename := args[0]
				data, err := os.ReadFile(filename)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				content = string(data)
				detectedLang = lang.FromExtension(filename)

				if title == "" {
					// Use filename as default title
					title = filename
				}

			default:
				// Read from stdin
				stat, _ := os.Stdin.Stat()
				if (stat.Mode() & os.ModeCharDevice) != 0 {
					return fmt.Errorf("no input: provide a file argument, pipe stdin, or use --from-clipboard")
				}
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				content = string(data)
			}

			content = strings.TrimRight(content, "\n")
			if content == "" {
				return fmt.Errorf("empty content")
			}

			// Language: flag > detected > config default
			if language == "" {
				language = detectedLang
			}
			if language == "" {
				language = cfg.DefaultLanguage
			}

			now := time.Now()
			snip := &model.Snippet{
				ID:          model.NewID(),
				Title:       title,
				Content:     content,
				Language:    language,
				Description: desc,
				Source:      source,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			if tags != "" {
				for _, t := range strings.Split(tags, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						snip.Tags = append(snip.Tags, t)
					}
				}
			}

			if err := store.Create(snip); err != nil {
				return fmt.Errorf("save snippet: %w", err)
			}

			fmt.Printf("saved %s", snip.ID)
			if snip.Title != "" {
				fmt.Printf(" (%s)", snip.Title)
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "snippet title")
	cmd.Flags().StringVar(&language, "lang", "", "language (auto-detected from file extension)")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	cmd.Flags().StringVar(&source, "source", "", "source URL or reference")
	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "read content from clipboard")

	return cmd
}
```

**Step 2: Register in root.go**

Add `root.AddCommand(newAddCmd())` in `NewRootCmd`, after the config command registration.

**Step 3: Build and test manually**

Run:
```bash
go build ./src/cmd/snipt
echo 'fmt.Println("hello")' | ./snipt add --title "hello world" --lang go --tags "test,demo"
```
Expected: prints `saved <id> (hello world)`

**Step 4: Commit**

```bash
git add src/internal/cli/add.go src/internal/cli/root.go
git commit -m "feat: add 'snipt add' command (file, stdin, clipboard)"
```

---

## Task 10: `snipt get` Command

**Files:**
- Create: `src/internal/cli/get.go`
- Modify: `src/internal/cli/root.go`

**Step 1: Implement get.go**

Create `src/internal/cli/get.go`:
```go
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/clipboard"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	var (
		toClipboard bool
		outputID    bool
	)

	cmd := &cobra.Command{
		Use:   "get <id|title>",
		Short: "Get a snippet's content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]

			// Try GetAndTrack for single-match cases
			snip, err := store.GetAndTrack(ref)
			if err != nil {
				var nfe *db.NotFoundError
				if errors.As(err, &nfe) {
					fmt.Fprintf(os.Stderr, "error: %s\n", err)
					os.Exit(model.ExitNotFound)
				}
				return err
			}

			if outputID {
				fmt.Println(snip.ID)
				return nil
			}

			if toClipboard {
				if err := clipboard.Write(snip.Content); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "copied %s to clipboard\n", snip.ID)
				return nil
			}

			fmt.Print(snip.Content)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&toClipboard, "clipboard", "c", false, "copy to clipboard")
	cmd.Flags().BoolVarP(&outputID, "id", "i", false, "output snippet ID instead of content")

	return cmd
}
```

**Step 2: Register in root.go**

Add `root.AddCommand(newGetCmd())` in `NewRootCmd`.

**Step 3: Build and test**

Run:
```bash
go build ./src/cmd/snipt
echo 'print("hi")' | ./snipt add --title "python hello" --lang python
./snipt get "python hello"
```
Expected: prints `print("hi")`

**Step 4: Commit**

```bash
git add src/internal/cli/get.go src/internal/cli/root.go
git commit -m "feat: add 'snipt get' command with clipboard and ID output"
```

---

## Task 11: `snipt list` Command

**Files:**
- Create: `src/internal/cli/list.go`
- Modify: `src/internal/cli/root.go`

**Step 1: Implement list.go**

Create `src/internal/cli/list.go`:
```go
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		language string
		tag      string
		pinned   bool
		sort     string
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := db.ListOpts{
				Language: language,
				Tag:      tag,
				Sort:     sort,
			}
			if cmd.Flags().Changed("pinned") {
				opts.Pinned = &pinned
			}

			snippets, err := store.List(opts)
			if err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(snippets)
			}

			if len(snippets) == 0 {
				fmt.Println("no snippets found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tLANG\tTAGS\tPIN\tUSES")
			for _, s := range snippets {
				pin := ""
				if s.Pinned {
					pin = "*"
				}
				tags := strings.Join(s.Tags, ", ")
				title := s.Title
				if len(title) > 40 {
					title = title[:37] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
					s.ID, title, s.Language, tags, pin, s.UseCount)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&language, "lang", "", "filter by language")
	cmd.Flags().StringVar(&tag, "tag", "", "filter by tag")
	cmd.Flags().BoolVar(&pinned, "pinned", false, "show only pinned snippets")
	cmd.Flags().StringVar(&sort, "sort", "created", "sort by: created, updated, usage, title")
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")

	return cmd
}
```

**Step 2: Register in root.go, build, test**

Add `root.AddCommand(newListCmd())` in `NewRootCmd`.

Run:
```bash
go build ./src/cmd/snipt && ./snipt list
```

**Step 3: Commit**

```bash
git add src/internal/cli/list.go src/internal/cli/root.go
git commit -m "feat: add 'snipt list' command with filters and JSON output"
```

---

## Task 12: `snipt edit`, `snipt set`, `snipt new` Commands

**Files:**
- Create: `src/internal/cli/edit.go`
- Create: `src/internal/cli/set.go`
- Create: `src/internal/cli/new.go`
- Modify: `src/internal/cli/root.go`

**Step 1: Implement edit.go**

Create `src/internal/cli/edit.go`:
```go
package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id|title>",
		Short: "Edit a snippet's content in your editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "error: no snippet matching %q\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			snip := &results[0].Snippet

			// Write content to temp file
			tmpFile, err := os.CreateTemp("", "snipt-*.txt")
			if err != nil {
				return err
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(snip.Content); err != nil {
				return err
			}
			tmpFile.Close()

			// Open in editor
			editor := cfg.ResolveEditor()
			c := exec.Command(editor, tmpFile.Name())
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("editor failed: %w", err)
			}

			// Read back edited content
			data, err := os.ReadFile(tmpFile.Name())
			if err != nil {
				return err
			}

			snip.Content = string(data)
			if err := store.Update(snip); err != nil {
				var nfe *db.NotFoundError
				if errors.As(err, &nfe) {
					fmt.Fprintf(os.Stderr, "error: %s\n", err)
					os.Exit(model.ExitNotFound)
				}
				return err
			}

			fmt.Printf("updated %s\n", snip.ID)
			return nil
		},
	}
}
```

**Step 2: Implement set.go**

Create `src/internal/cli/set.go`:
```go
package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	var (
		title    string
		language string
		desc     string
		source   string
	)

	cmd := &cobra.Command{
		Use:   "set <id|title>",
		Short: "Set snippet metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("title") && !cmd.Flags().Changed("lang") &&
				!cmd.Flags().Changed("desc") && !cmd.Flags().Changed("source") {
				return fmt.Errorf("at least one of --title, --lang, --desc, --source is required")
			}

			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "error: no snippet matching %q\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			snip := &results[0].Snippet

			if cmd.Flags().Changed("title") {
				snip.Title = title
			}
			if cmd.Flags().Changed("lang") {
				snip.Language = language
			}
			if cmd.Flags().Changed("desc") {
				snip.Description = desc
			}
			if cmd.Flags().Changed("source") {
				snip.Source = source
			}

			if err := store.Update(snip); err != nil {
				return err
			}

			fmt.Printf("updated %s\n", snip.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "set title")
	cmd.Flags().StringVar(&language, "lang", "", "set language")
	cmd.Flags().StringVar(&desc, "desc", "", "set description")
	cmd.Flags().StringVar(&source, "source", "", "set source URL")

	return cmd
}
```

**Step 3: Implement new.go**

Create `src/internal/cli/new.go`:
```go
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Create a new snippet from scratch",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create temp file for editor
			tmpFile, err := os.CreateTemp("", "snipt-new-*.txt")
			if err != nil {
				return err
			}
			defer os.Remove(tmpFile.Name())
			tmpFile.Close()

			// Open editor
			editor := cfg.ResolveEditor()
			c := exec.Command(editor, tmpFile.Name())
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("editor failed: %w", err)
			}

			// Read content
			data, err := os.ReadFile(tmpFile.Name())
			if err != nil {
				return err
			}
			content := strings.TrimRight(string(data), "\n")
			if content == "" {
				fmt.Println("empty content, nothing saved")
				return nil
			}

			// For now, use simple prompts. Will be replaced with Bubbletea form in a later task.
			now := time.Now()
			snip := &model.Snippet{
				ID:        model.NewID(),
				Content:   content,
				Language:  cfg.DefaultLanguage,
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := store.Create(snip); err != nil {
				return err
			}

			fmt.Printf("saved %s\n", snip.ID)
			return nil
		},
	}
}
```

**Step 4: Register all in root.go, build**

Add to `NewRootCmd`:
```go
root.AddCommand(newAddCmd())
root.AddCommand(newGetCmd())
root.AddCommand(newListCmd())
root.AddCommand(newEditCmd())
root.AddCommand(newSetCmd())
root.AddCommand(newNewCmd())
```

Run: `go build ./src/cmd/snipt`

**Step 5: Commit**

```bash
git add src/internal/cli/
git commit -m "feat: add edit, set, and new commands"
```

---

## Task 13: `snipt tag`, `untag`, `pin`, `unpin`, `rm` Commands

**Files:**
- Create: `src/internal/cli/tag.go`
- Create: `src/internal/cli/pin.go`
- Create: `src/internal/cli/rm.go`
- Modify: `src/internal/cli/root.go`

**Step 1: Implement tag.go**

Create `src/internal/cli/tag.go`:
```go
package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <id|title> <tags...>",
		Short: "Add tags to a snippet",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "error: no snippet matching %q\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			id := results[0].Snippet.ID
			tags := args[1:]
			if err := store.AddTags(id, tags); err != nil {
				return err
			}
			fmt.Printf("tagged %s with %v\n", id, tags)
			return nil
		},
	}
}

func newUntagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "untag <id|title> <tags...>",
		Short: "Remove tags from a snippet",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "error: no snippet matching %q\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			id := results[0].Snippet.ID
			tags := args[1:]
			if err := store.RemoveTags(id, tags); err != nil {
				return err
			}
			fmt.Printf("untagged %s from %v\n", id, tags)
			return nil
		},
	}
}
```

**Step 2: Implement pin.go**

Create `src/internal/cli/pin.go`:
```go
package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newPinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pin <id|title>",
		Short: "Pin a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "error: no snippet matching %q\n", args[0])
				os.Exit(model.ExitNotFound)
			}
			id := results[0].Snippet.ID
			if err := store.SetPinned(id, true); err != nil {
				return err
			}
			fmt.Printf("pinned %s\n", id)
			return nil
		},
	}
}

func newUnpinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpin <id|title>",
		Short: "Unpin a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "error: no snippet matching %q\n", args[0])
				os.Exit(model.ExitNotFound)
			}
			id := results[0].Snippet.ID
			if err := store.SetPinned(id, false); err != nil {
				return err
			}
			fmt.Printf("unpinned %s\n", id)
			return nil
		},
	}
}
```

**Step 3: Implement rm.go**

Create `src/internal/cli/rm.go`:
```go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "rm <id|title>",
		Short: "Delete a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "error: no snippet matching %q\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			snip := &results[0].Snippet

			if !force {
				// Check if stdout is a TTY
				stat, _ := os.Stdout.Stat()
				isTTY := (stat.Mode() & os.ModeCharDevice) != 0

				if !isTTY {
					return fmt.Errorf("refusing to delete without --force in non-interactive mode")
				}

				fmt.Fprintf(os.Stderr, "Delete %q (%s)? [y/N] ", snip.Title, snip.ID)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("cancelled")
					return nil
				}
			}

			if err := store.Delete(snip.ID); err != nil {
				return err
			}
			fmt.Printf("deleted %s\n", snip.ID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation")

	return cmd
}
```

**Step 4: Register all, build, test**

Add to `NewRootCmd`:
```go
root.AddCommand(newTagCmd())
root.AddCommand(newUntagCmd())
root.AddCommand(newPinCmd())
root.AddCommand(newUnpinCmd())
root.AddCommand(newRmCmd())
```

Run: `go build ./src/cmd/snipt`

**Step 5: Commit**

```bash
git add src/internal/cli/
git commit -m "feat: add tag, untag, pin, unpin, rm commands"
```

---

## Task 14: `snipt stats`, `snipt export`, `snipt import` Commands

**Files:**
- Create: `src/internal/cli/stats.go`
- Create: `src/internal/cli/export.go`
- Create: `src/internal/cli/import.go`
- Modify: `src/internal/cli/root.go`

**Step 1: Implement stats.go**

Create `src/internal/cli/stats.go`:
```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show collection statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := store.Stats()
			if err != nil {
				return err
			}

			fmt.Printf("Snippets:  %d\n", stats.TotalSnippets)
			fmt.Printf("Tags:      %d\n", stats.TotalTags)

			if len(stats.Languages) > 0 {
				fmt.Println("\nLanguages:")
				for lang, count := range stats.Languages {
					fmt.Printf("  %-15s %d\n", lang, count)
				}
			}

			if stats.MostUsed != nil && stats.MostUsed.UseCount > 0 {
				fmt.Printf("\nMost used: %s (%s) — %d uses\n",
					stats.MostUsed.Title, stats.MostUsed.ID, stats.MostUsed.UseCount)
			}

			if len(stats.RecentlyAdded) > 0 {
				fmt.Println("\nRecent:")
				for _, s := range stats.RecentlyAdded {
					fmt.Printf("  %s  %s\n", s.ID, s.Title)
				}
			}

			return nil
		},
	}
}
```

**Step 2: Implement export.go**

Create `src/internal/cli/export.go`:
```go
package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

type exportEnvelope struct {
	Version    int              `json:"version"`
	ExportedAt string           `json:"exported_at"`
	Count      int              `json:"count"`
	Snippets   []model.Snippet  `json:"snippets"`
}

func newExportCmd() *cobra.Command {
	var (
		format string
		output string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export all snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			snippets, err := store.List(db.ListOpts{})
			if err != nil {
				return err
			}

			var w io.Writer = os.Stdout
			if output != "" {
				f, err := os.Create(output)
				if err != nil {
					return err
				}
				defer f.Close()
				w = f
			}

			switch format {
			case "json":
				return exportJSON(w, snippets)
			case "markdown", "md":
				return exportMarkdown(w, snippets)
			case "tar":
				return exportTar(w, snippets)
			default:
				return fmt.Errorf("unknown format %q (use json, markdown, or tar)", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "export format: json, markdown, tar")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (default: stdout)")

	return cmd
}

func exportJSON(w io.Writer, snippets []model.Snippet) error {
	env := exportEnvelope{
		Version:    1,
		ExportedAt: time.Now().Format(time.RFC3339),
		Count:      len(snippets),
		Snippets:   snippets,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}

func exportMarkdown(w io.Writer, snippets []model.Snippet) error {
	for i, s := range snippets {
		if i > 0 {
			fmt.Fprintln(w, "---")
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "# %s\n\n", s.Title)
		fmt.Fprintf(w, "- **ID:** %s\n", s.ID)
		fmt.Fprintf(w, "- **Language:** %s\n", s.Language)
		if len(s.Tags) > 0 {
			fmt.Fprintf(w, "- **Tags:** %s\n", strings.Join(s.Tags, ", "))
		}
		if s.Pinned {
			fmt.Fprintln(w, "- **Pinned:** yes")
		}
		fmt.Fprintf(w, "- **Created:** %s\n", s.CreatedAt.Format(time.RFC3339))
		fmt.Fprintln(w)
		fmt.Fprintf(w, "```%s\n%s\n```\n\n", s.Language, s.Content)
	}
	return nil
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	slug := strings.ToLower(s)
	slug = nonAlphaNum.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func exportTar(w io.Writer, snippets []model.Snippet) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, s := range snippets {
		name := fmt.Sprintf("%s-%s.md", s.ID, slugify(s.Title))
		var buf strings.Builder
		fmt.Fprintf(&buf, "---\n")
		fmt.Fprintf(&buf, "id: %s\n", s.ID)
		fmt.Fprintf(&buf, "title: %s\n", s.Title)
		fmt.Fprintf(&buf, "language: %s\n", s.Language)
		if len(s.Tags) > 0 {
			fmt.Fprintf(&buf, "tags: [%s]\n", strings.Join(s.Tags, ", "))
		}
		if s.Pinned {
			fmt.Fprintf(&buf, "pinned: true\n")
		}
		fmt.Fprintf(&buf, "created_at: %s\n", s.CreatedAt.Format(time.RFC3339))
		fmt.Fprintf(&buf, "---\n\n")
		fmt.Fprintf(&buf, "```%s\n%s\n```\n", s.Language, s.Content)

		content := buf.String()
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return err
		}
	}
	return nil
}
```

**Step 3: Implement import.go**

Create `src/internal/cli/import.go`:
```go
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var (
		overwrite bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import snippets from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			snippets, err := parseImport(data)
			if err != nil {
				return fmt.Errorf("parse import: %w", err)
			}

			var created, skipped, updated int
			for _, s := range snippets {
				// Generate new ID if missing
				if s.ID == "" {
					s.ID = model.NewID()
				}

				// Check if exists
				existing, _ := store.Get(s.ID)

				if existing != nil {
					if overwrite {
						if dryRun {
							fmt.Printf("[dry-run] would overwrite %s (%s)\n", s.ID, s.Title)
						} else {
							if err := store.Update(&s); err != nil {
								return err
							}
						}
						updated++
					} else {
						if dryRun {
							fmt.Printf("[dry-run] would skip %s (%s) — already exists\n", s.ID, s.Title)
						}
						skipped++
					}
					continue
				}

				if s.CreatedAt.IsZero() {
					s.CreatedAt = time.Now()
				}
				if s.UpdatedAt.IsZero() {
					s.UpdatedAt = time.Now()
				}

				if dryRun {
					fmt.Printf("[dry-run] would import %s (%s)\n", s.ID, s.Title)
				} else {
					if err := store.Create(&s); err != nil {
						return fmt.Errorf("create %s: %w", s.ID, err)
					}
				}
				created++
			}

			action := "imported"
			if dryRun {
				action = "would import"
			}
			fmt.Printf("%s %d, skipped %d, updated %d (total in file: %d)\n",
				action, created, skipped, updated, len(snippets))
			return nil
		},
	}

	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing snippets with matching IDs")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would happen without writing")

	return cmd
}

func parseImport(data []byte) ([]model.Snippet, error) {
	// Try envelope format first
	var env exportEnvelope
	if err := json.Unmarshal(data, &env); err == nil && env.Snippets != nil {
		return env.Snippets, nil
	}

	// Try flat array
	var snippets []model.Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return nil, fmt.Errorf("unrecognized format: expected JSON envelope or array")
	}
	return snippets, nil
}
```

**Step 4: Register all, build, test**

Add to `NewRootCmd`:
```go
root.AddCommand(newStatsCmd())
root.AddCommand(newExportCmd())
root.AddCommand(newImportCmd())
```

Run:
```bash
go build ./src/cmd/snipt
./snipt stats
./snipt export --format json
```

**Step 5: Commit**

```bash
git add src/internal/cli/
git commit -m "feat: add stats, export, and import commands"
```

---

## Task 15: Integration Test — Full CLI Round Trip

**Files:**
- Create: `src/internal/cli/cli_test.go`

**Step 1: Write integration test**

Create `src/internal/cli/cli_test.go`:
```go
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_RoundTrip(t *testing.T) {
	// Use temp DB
	dbFile := filepath.Join(t.TempDir(), "test.db")

	// Helper to run CLI commands
	run := func(args ...string) (string, error) {
		root := NewRootCmd("test")
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetErr(buf)
		root.SetArgs(append([]string{"--db", dbFile}, args...))
		err := root.Execute()
		return buf.String(), err
	}

	// Add a snippet via stdin
	// (We'll test file-based add since stdin is harder in tests)
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	os.WriteFile(tmpFile, []byte("package main\n\nfunc main() {}"), 0o644)

	out, err := run("add", tmpFile, "--title", "test snippet", "--tags", "test,go")
	if err != nil {
		t.Fatalf("add failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "saved") {
		t.Errorf("expected 'saved' in output, got: %s", out)
	}

	// List
	out, err = run("list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(out, "test snippet") {
		t.Errorf("expected 'test snippet' in list output, got: %s", out)
	}

	// Stats
	out, err = run("stats")
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if !strings.Contains(out, "Snippets:  1") {
		t.Errorf("expected 1 snippet in stats, got: %s", out)
	}
}
```

**Step 2: Run test**

Run: `go test ./src/internal/cli/ -v -run TestCLI_RoundTrip`
Expected: PASS

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add src/internal/cli/cli_test.go
git commit -m "test: add CLI integration round-trip test"
```

---

## Task 16: TUI — Bubbletea New-Snippet Form

**Files:**
- Create: `src/internal/tui/form.go`
- Modify: `src/internal/cli/new.go` (wire up the form)

**Step 1: Implement form.go**

Create `src/internal/tui/form.go` — a Bubbletea model with 4 text inputs (title, language, tags, description) and Enter to submit. Style with Lip Gloss using the Catppuccin palette. The form launches after saving the editor content.

Key behaviors:
- Tab/Shift+Tab to navigate between fields
- Language field pre-filled with auto-detected value if available
- Enter on last field submits
- Esc cancels (returns empty result)
- Returns a struct with the filled values

**Step 2: Wire into `snipt new`**

Update `new.go` to launch the Bubbletea form after editor closes, then use the form output to populate snippet metadata.

**Step 3: Build and test manually**

Run: `go build ./src/cmd/snipt && ./snipt new`

**Step 4: Commit**

```bash
git add src/internal/tui/ src/internal/cli/new.go
git commit -m "feat: add Bubbletea form for snipt new metadata input"
```

---

## Task 17: TUI — Fuzzy Matching Engine

**Files:**
- Create: `src/internal/tui/fuzzy.go`
- Create: `src/internal/tui/fuzzy_test.go`

**Step 1: Write failing test**

Create `src/internal/tui/fuzzy_test.go`:
```go
package tui

import "testing"

func TestFuzzyMatch_Exact(t *testing.T) {
	result := FuzzyMatch("HTTP server", "HTTP server")
	if !result.Match {
		t.Error("expected exact match")
	}
}

func TestFuzzyMatch_Partial(t *testing.T) {
	result := FuzzyMatch("HTTP server with middleware", "http serv")
	if !result.Match {
		t.Error("expected partial match")
	}
	if result.Score <= 0 {
		t.Error("expected positive score")
	}
	if len(result.Indices) != 9 { // "http serv" = 9 chars (excluding space in query matching)
		t.Errorf("expected 9 matched indices, got %d", len(result.Indices))
	}
}

func TestFuzzyMatch_NoMatch(t *testing.T) {
	result := FuzzyMatch("HTTP server", "zzz")
	if result.Match {
		t.Error("expected no match")
	}
}

func TestFuzzyMatch_ConsecutiveBonus(t *testing.T) {
	r1 := FuzzyMatch("abcdef", "abc")
	r2 := FuzzyMatch("aXbXcX", "abc")
	if r1.Score <= r2.Score {
		t.Errorf("consecutive match should score higher: %d vs %d", r1.Score, r2.Score)
	}
}

func TestFuzzyMatch_WordBoundaryBonus(t *testing.T) {
	r1 := FuzzyMatch("http_server", "hs")
	r2 := FuzzyMatch("ahahs", "hs")
	if r1.Score <= r2.Score {
		t.Errorf("word boundary match should score higher: %d vs %d", r1.Score, r2.Score)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/internal/tui/ -v -run TestFuzzy`
Expected: FAIL

**Step 3: Implement fuzzy.go**

Create `src/internal/tui/fuzzy.go`:
```go
package tui

import "strings"

// FuzzyResult holds the result of a fuzzy match.
type FuzzyResult struct {
	Match   bool
	Score   int
	Indices []int // character positions that matched in the text
}

// FuzzyMatch performs fuzzy matching of query against text.
// Scoring: +1 per match, +3 consecutive, +2 word boundary.
func FuzzyMatch(text, query string) FuzzyResult {
	if query == "" {
		return FuzzyResult{Match: true, Score: 0}
	}

	lower := strings.ToLower(text)
	q := strings.ToLower(query)
	qi := 0
	score := 0
	indices := make([]int, 0, len(q))
	lastMatchIdx := -1

	for i := 0; i < len(lower) && qi < len(q); i++ {
		if lower[i] == q[qi] {
			indices = append(indices, i)
			// Consecutive match bonus
			if lastMatchIdx == i-1 {
				score += 3
			}
			// Word boundary bonus
			if i == 0 || text[i-1] == ' ' || text[i-1] == '_' || text[i-1] == '-' {
				score += 2
			}
			score += 1
			lastMatchIdx = i
			qi++
		}
	}

	return FuzzyResult{
		Match:   qi == len(q),
		Score:   score,
		Indices: indices,
	}
}
```

**Step 4: Run test**

Run: `go test ./src/internal/tui/ -v -run TestFuzzy`
Expected: PASS

**Step 5: Commit**

```bash
git add src/internal/tui/fuzzy.go src/internal/tui/fuzzy_test.go
git commit -m "feat: add fuzzy matching engine with consecutive and boundary bonuses"
```

---

## Task 18: TUI — Theme / Color Palette

**Files:**
- Create: `src/internal/tui/theme.go`

**Step 1: Implement theme.go**

Create `src/internal/tui/theme.go` with all Catppuccin Mocha colors as Lip Gloss color constants, the language color map, and reusable Lip Gloss styles (badge, border, selected row, etc.).

**Step 2: Commit**

```bash
git add src/internal/tui/theme.go
git commit -m "feat: add Catppuccin Mocha theme and language color palette"
```

---

## Task 19: TUI — Find Palette (Core)

**Files:**
- Create: `src/internal/tui/find.go`
- Create: `src/internal/tui/resultlist.go`
- Create: `src/internal/cli/find.go`
- Modify: `src/internal/cli/root.go`

**Step 1: Implement resultlist.go**

The reusable list component. Renders rows with:
- Pinned indicator (yellow dot)
- Title with fuzzy-match highlighted characters (pink)
- Language badge (colored text on tinted bg)
- Tags (dimmed)
- Inline code preview on selected row (right side)

This component is reused by both `snipt find` and the mini-picker.

**Step 2: Implement find.go (Bubbletea model)**

The main `snipt find` model:
- `textinput.Model` for search bar with SNIPT badge
- `resultlist` for filtered results
- Bottom bar with hint pills
- Keybindings: Up/Down navigate, Enter select, Tab toggle preview scroll, Esc quit
- Result count display (N/total)
- "Copied" feedback on Enter

**Step 3: Implement cli/find.go**

Wire the `find` command:
- Flags: `-c`, `--lang`, `--tag`, `--pinned`, `-i`
- Load all snippets, pass to TUI model
- On selection: output content (stdout or clipboard), bump use count via `GetAndTrack`
- Exit code 130 on cancel

**Step 4: Register, build, test**

Add `root.AddCommand(newFindCmd())`, build, launch `./snipt find`.

**Step 5: Commit**

```bash
git add src/internal/tui/ src/internal/cli/find.go src/internal/cli/root.go
git commit -m "feat: add snipt find command palette with fuzzy search"
```

---

## Task 20: TUI — Mini-Picker for Ambiguous Resolution

**Files:**
- Create: `src/internal/tui/picker.go`
- Modify: `src/internal/cli/get.go` (and other commands using ResolveRef)

**Step 1: Implement picker.go**

A stripped-down Bubbletea model reusing `resultlist`:
- No search bar
- Pre-filled with the ambiguous matches
- Arrow keys + Enter to select
- Esc to cancel (exit code 130)

**Step 2: Create a resolve helper in CLI**

Create a shared helper in `src/internal/cli/resolve.go` that wraps `store.ResolveRef` with the resolution strategy:
- Single result → use it
- Multiple results + TTY → launch mini-picker
- Multiple results + non-TTY → take top match
- Zero results → exit code 2

Update `get.go`, `edit.go`, `set.go`, `tag.go`, `pin.go`, `rm.go` to use this helper.

**Step 3: Build and test**

**Step 4: Commit**

```bash
git add src/internal/tui/picker.go src/internal/cli/
git commit -m "feat: add mini-picker for ambiguous snippet resolution"
```

---

## Task 21: Final Integration — Run All Tests, Build, Verify

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 2: Build release binary**

Run: `go build -o snipt ./src/cmd/snipt`

**Step 3: End-to-end manual test**

```bash
./snipt add testfile.go --title "Test" --tags "demo"
./snipt list
./snipt get "Test"
./snipt find
./snipt stats
./snipt export --format json
./snipt rm "Test" --force
```

**Step 4: Run vet and format**

Run:
```bash
go vet ./...
gofmt -l ./src/
```
Expected: no issues

**Step 5: Commit any fixes**

```bash
git add -A
git commit -m "chore: final cleanup and verification"
```
