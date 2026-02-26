# snipt — Phase 1 & 2 Design

## Overview

CLI snippet manager in Go with Bubbletea TUI. Phase 1 builds the core CLI + SQLite foundation. Phase 2 adds the `snipt find` command palette.

## Architecture

**Module:** `github.com/infktd/snipt`

```
snipt/
├── src/
│   ├── cmd/snipt/            # main.go entry point
│   └── internal/
│       ├── model/            # Snippet struct, ID generation, exit codes
│       ├── db/               # SQLite: migrations, CRUD, FTS5, fuzzy resolve
│       ├── config/           # XDG paths, TOML parsing, editor resolution
│       ├── lang/             # Extension→language map
│       ├── clipboard/        # Platform read/write (pbcopy/pbpaste, xclip, wl-copy/wl-paste)
│       ├── cli/              # Cobra root + subcommands
│       └── tui/              # Bubbletea components (new-snippet form, find palette, mini-picker)
├── go.mod
├── go.sum
├── .goreleaser.yaml
└── LICENSE
```

**Dependency graph** (arrows = imports):
```
cmd/snipt → cli → db, config, model, clipboard, lang, tui
                   db → model
                   tui → model, db
                   config (standalone)
                   lang (standalone)
                   clipboard (standalone)
```

Standalone packages never import other internal packages. `db` only imports `model`. Acyclic and testable.

**Core dependencies:**
- `modernc.org/sqlite` — pure Go SQLite, no CGO
- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` v2 — styling
- `github.com/charmbracelet/bubbles` — text input, viewport components
- `github.com/BurntSushi/toml` — config parsing
- `github.com/google/uuid` — snippet ID generation (first 8 chars)

## Data Model

```go
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
```

**ID generation:** First 8 hex chars of a UUIDv4.

**Exit codes:**
- 0 = success
- 1 = general error
- 2 = not found
- 3 = db error
- 130 = interrupted (Esc/Ctrl+C)

## ID Resolution Strategy

When a command takes `<id|title>`:

1. Exact ID match → use it
2. Exact title match (case-insensitive) → use it
3. FTS5 fuzzy search → score results
4. Single high-confidence match (top score ≥ 2x second-best) → use it
5. Multiple close matches + stdout is TTY → launch mini-picker
6. Multiple close matches + non-TTY (piped) → take top match (deterministic)
7. Zero matches → exit code 2, message: `no snippet matching "whatever"`

**Use count tracking:** `ResolveRef` does pure resolution with no side effects. `GetAndTrack` wraps it and bumps `use_count` atomically. Only `snipt get` and `snipt find` use `GetAndTrack`. Commands like `edit`, `set`, `tag`, `pin`, `rm` use `ResolveRef` without inflating the count.

## SQLite Schema

```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;

CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT);
INSERT INTO meta VALUES ('schema_version', '1');

CREATE TABLE snippets (
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

CREATE TABLE tags (
    snippet_id TEXT REFERENCES snippets(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (snippet_id, tag)
);

CREATE VIRTUAL TABLE snippets_fts USING fts5(
    title, content, description,
    content='snippets', content_rowid='rowid'
);

CREATE TRIGGER snippets_ai AFTER INSERT ON snippets BEGIN
    INSERT INTO snippets_fts(rowid, title, content, description)
    VALUES (new.rowid, new.title, new.content, new.description);
END;

CREATE TRIGGER snippets_ad AFTER DELETE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, title, content, description)
    VALUES ('delete', old.rowid, old.title, old.content, old.description);
END;

CREATE TRIGGER snippets_au AFTER UPDATE ON snippets BEGIN
    INSERT INTO snippets_fts(snippets_fts, rowid, title, content, description)
    VALUES ('delete', old.rowid, old.title, old.content, old.description);
    INSERT INTO snippets_fts(rowid, title, content, description)
    VALUES (new.rowid, new.title, new.content, new.description);
END;
```

Timestamps stored as RFC3339 strings (sort lexicographically in SQLite).

## DB Interface

`Store` struct wrapping `*sql.DB`:

- `Open(path string) (*Store, error)` — open DB, run migrations
- `Close() error`
- `Create(snippet *model.Snippet) error`
- `Get(id string) (*model.Snippet, error)` — exact ID lookup
- `Update(snippet *model.Snippet) error`
- `Delete(id string) error`
- `List(opts ListOpts) ([]model.Snippet, error)` — filtered listing
- `Search(query string) ([]SearchResult, error)` — FTS5 search with scores
- `ResolveRef(ref string) ([]SearchResult, error)` — ID resolution strategy
- `GetAndTrack(ref string) (*model.Snippet, error)` — resolve + bump use_count
- `AddTags(id string, tags []string) error`
- `RemoveTags(id string, tags []string) error`
- `SetPinned(id string, pinned bool) error`
- `IncrementUseCount(id string) error` — internal, called by GetAndTrack
- `Stats() (*model.Stats, error)`

Migrations: embedded SQL with version check against `meta` table.

## Config

**XDG paths:**
- Config: `$XDG_CONFIG_HOME/snipt/config.toml` (default `~/.config/snipt/config.toml`)
- Data: `$XDG_DATA_HOME/snipt/snipt.db` (default `~/.local/share/snipt/snipt.db`)

```toml
editor = "nvim"
default_language = "text"
theme = "catppuccin-mocha"
```

**Editor resolution:** config `editor` → `$VISUAL` → `$EDITOR` → `vi`

## Clipboard

Detection order:
- macOS: `pbcopy` / `pbpaste`
- Wayland: `wl-copy` / `wl-paste`
- X11: `xclip`

Error if none found and clipboard requested: `"no clipboard tool found (need pbcopy, xclip, or wl-copy)"`

## Language Detection

Extension map: `.go`→go, `.py`→python, `.ts`→typescript, `.js`→javascript, `.rs`→rust, `.lua`→lua, `.sh/.bash`→bash, `.sql`→sql, `.nix`→nix, `.rb`→ruby, `.java`→java, `.c`→c, `.cpp`→cpp, `.md`→markdown, `.toml`→toml, `.yaml/.yml`→yaml, `.json`→json

Fallback: config `default_language` (defaults to `"text"`).

## Phase 1 Commands

- **`snipt new`** — open $EDITOR, then Bubbletea mini-form for title/lang/tags/description
- **`snipt add <file>`** — create from file, auto-detect language from extension
- **`snipt add` (stdin)** — create from piped input, requires `--lang`
- **`snipt add --from-clipboard`** — create from clipboard contents
- **`snipt get <ref>`** — output content via `GetAndTrack`, `-c` for clipboard, `-i` for ID only
- **`snipt list`** — formatted table with `--lang`, `--tag`, `--pinned`, `--sort`, `--json`
- **`snipt edit <ref>`** — open content in $EDITOR, update on save
- **`snipt set <ref>`** — modify metadata via `--title`, `--lang`, `--desc`, `--source`
- **`snipt tag <ref> <tags...>`** — add tags (idempotent)
- **`snipt untag <ref> <tags...>`** — remove tags (idempotent)
- **`snipt pin <ref>` / `snipt unpin <ref>`** — toggle pinned (idempotent)
- **`snipt rm <ref>`** — delete with confirmation, `--force` to skip, non-TTY requires `--force`
- **`snipt export`** — `--format` json/markdown/tar, `-o` output path
- **`snipt import <file>`** — `--format` auto-detect, `--overwrite`, `--dry-run`
- **`snipt stats`** — collection overview
- **`snipt config`** — open config in $EDITOR
- **`snipt config path`** — print config path

**Global flags:** `--db <path>`, `--no-color`, `--help`, `--version`

## Phase 2: `snipt find`

### Layout

Centered floating panel with inline preview on selected row:

```
╭─────────────────────────────────────────────╮
│  SNIPT   🔍 http serv█              5/8    │
├─────────────────────────────────────────────┤
│  ● HTTP server w/ middleware  go  #server   │  func main() {         │
│    Retry with backoff          go           │      mux := http...    │
│    Parse JSON request body     go  #api     │                        │
╰─────────────────────────────────────────────╯
  ↑↓ navigate  enter copy  tab preview  esc close
```

Selected row expands to show inline code preview to the right. Non-selected rows are compact.

### Bubbletea Components

- **Search bar:** `textinput.Model`, 50ms debounce
- **Results list:** Custom model with match highlighting, inline preview on selected row
- **Mini-picker:** Reuses the same list renderer for ambiguous ID resolution

### Interaction

- Up/Down: navigate results
- Enter: output selected snippet (stdout or clipboard with `-c`)
- Tab: toggle focus for scrolling preview
- Esc/Ctrl+C: exit code 130
- Typing: real-time fuzzy filtering

### Fuzzy Scoring

- +1 per matched character
- +3 for consecutive matches
- +2 for word-boundary matches (after space, `_`, `-`, or at start)
- +3 pinned bonus
- +5 if any tag matches
- +2 for content substring match
- Matched characters highlighted in pink
- Results sorted by score descending

### Color Palette (Catppuccin Mocha)

```
bg=#1e1e2e  bgSurface=#242435  bgOverlay=#2a2a3c  bgHighlight=#313147
bgSelected=#3e3e5e  border=#45456a  borderDim=#363654  borderFocus=#cba6f7
text=#cdd6f4  textSub=#a6adc8  textDim=#6c7086  textMuted=#45475a
pink=#f5c2e7  mauve=#cba6f7  peach=#fab387  green=#a6e3a1  teal=#94e2d5
blue=#89b4fa  yellow=#f9e2af  red=#f38ba8  lavender=#b4befe  sky=#89dceb
flamingo=#f2cdcd  rosewater=#f5e0dc
```

**Language badge colors:** go=sky, nix=mauve, sql=yellow, bash=green, python=blue, typescript=blue, rust=peach, lua=lavender

**SNIPT badge:** gradient from pink→mauve

**Language badges:** colored text on subtle tinted background (color at 15% opacity)

### Flags

- `-c`: clipboard output
- `--lang`, `--tag`, `--pinned`: pre-filter
- `-i`: output ID instead of content

### Copied Feedback

On Enter, bottom bar shows "✓ copied to clipboard" in green for ~2 seconds, replacing the hint pills.

## Export/Import Formats

### JSON (envelope)

```json
{
  "version": 1,
  "exported_at": "2026-02-25T10:30:00Z",
  "count": 42,
  "snippets": [...]
}
```

Import auto-detects envelope (has `snippets` key) vs flat array.

### Markdown

One file per snippet with YAML frontmatter containing id, title, language, tags, pinned, created_at. Content in fenced code block.

### Tar

`.tar.gz` of markdown files named `{id}-{slugified-title}.md`.

### Import Behavior

- `--overwrite`: replace existing snippets with matching IDs
- Without `--overwrite`: skip duplicates, report them
- `--dry-run`: print what would happen
- New IDs generated for snippets without IDs
