# snipt

A terminal-based code snippet manager with fuzzy search, syntax highlighting, and clipboard integration.

<!-- Demo GIFs -->
<!-- ![snipt manage demo](./docs/demo-manage.gif) -->
<!-- ![snipt find demo](./docs/demo-find.gif) -->

## Features

- **Two TUI modes** — full-screen manager (`snipt manage`) and floating search palette (`snipt find`)
- **Fuzzy search** with scoring that rewards consecutive and word-boundary matches
- **Tag filtering** — prefix searches with `#` to match tags only (e.g., `#api`)
- **Syntax highlighting** — built-in keyword highlighting for Go, Python, Rust, TypeScript, JavaScript, Bash, SQL, Nix, Lua, and more
- **Create/edit in your editor** — opens `$EDITOR` with a YAML frontmatter template
- **Clipboard integration** — OSC52 in TUI, `pbcopy`/`xclip`/`wl-copy` for CLI
- **Pin snippets** to keep them at the top of the list
- **Use tracking** — tracks how often each snippet is copied
- **Full CLI** — add, get, list, edit, tag, pin, export, import, and more
- **Export/import** — JSON, Markdown, or tar.gz formats
- **Pure Go SQLite** with FTS5 full-text search — no CGo required
- **Catppuccin Mocha** color theme

## Install

### From source

```bash
go install github.com/infktd/snipt/src/cmd/snipt@latest
```

Or clone and build:

```bash
git clone https://github.com/infktd/snipt.git
cd snipt
go build -o snipt ./src/cmd/snipt
```

### Dependencies

- Go 1.25+
- `$EDITOR` or `$VISUAL` set for create/edit (falls back to `vi`)

## Usage

### `snipt manage` (default)

Full-screen snippet manager with a sidebar and syntax-highlighted preview pane. This is the default when you run `snipt` with no arguments.

| Key | Action |
|-----|--------|
| `j` / `k` / `↑` / `↓` | Navigate snippets |
| `/` | Search (fuzzy title match, `#tag` for tags) |
| `enter` | Copy snippet to clipboard |
| `n` | Create new snippet in `$EDITOR` |
| `e` | Edit selected snippet in `$EDITOR` |
| `d` | Delete selected snippet (confirms with `y`) |
| `p` | Toggle pin |
| `q` / `ctrl+c` | Quit |

While searching:

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate results |
| `enter` | Confirm search |
| `esc` | Cancel and clear search |

### `snipt find`

Floating search palette for quick fuzzy search and copy. Accepts an optional initial query.

```bash
snipt find            # open empty
snipt find "http"     # open with pre-filled query
```

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate results |
| `enter` | Copy to clipboard and close |
| `esc` / `ctrl+c` | Close |

### Creating a snippet

**In the TUI:** Press `n` in manage mode. Your editor opens with a YAML frontmatter template:

```yaml
---
title:
language: text
tags: []
pinned: false
---

```

Fill in the metadata, write your code below the closing `---`, save and quit. The snippet appears in your list.

**From the CLI:**

```bash
snipt new                              # open blank file in $EDITOR, then fill metadata
snipt add script.sh                    # import from file (auto-detects language)
echo "curl -s ..." | snipt add -       # pipe from stdin
snipt add --from-clipboard             # grab from system clipboard
```

### Snippet format

Snippets edited via the TUI use YAML frontmatter:

```yaml
---
title: HTTP GET with headers
language: go
tags: [http, networking]
pinned: false
---

req, _ := http.NewRequest("GET", url, nil)
req.Header.Set("Authorization", "Bearer "+token)
resp, err := http.DefaultClient.Do(req)
```

### Full CLI reference

| Command | Description |
|---------|-------------|
| `snipt` / `snipt manage` | Full-screen TUI manager |
| `snipt find [query]` | Floating search palette |
| `snipt new` | Create snippet in `$EDITOR` |
| `snipt add [file]` | Create from file, stdin, or clipboard |
| `snipt get <ref>` | Output snippet content (`-c` for clipboard) |
| `snipt list` | List snippets (`--lang`, `--tag`, `--pinned`, `--sort`, `--json`) |
| `snipt edit <ref>` | Edit content in `$EDITOR` |
| `snipt set <ref>` | Modify metadata (`--title`, `--lang`, `--desc`, `--source`) |
| `snipt tag <ref> <tags...>` | Add tags |
| `snipt untag <ref> <tags...>` | Remove tags |
| `snipt pin <ref>` / `unpin` | Toggle pinned status |
| `snipt rm <ref>` | Delete snippet (`--force` to skip confirmation) |
| `snipt stats` | Collection statistics |
| `snipt export` | Export as JSON, Markdown, or tar.gz |
| `snipt import <file>` | Import from JSON (`--overwrite`, `--dry-run`) |
| `snipt config` | Open config in `$EDITOR` |

The `<ref>` argument resolves by exact ID, exact title (case-insensitive), or full-text search. When ambiguous, an interactive picker appears.

**Global flags:** `--db <path>`, `--no-color`, `--version`

## Configuration

Config file: `~/.config/snipt/config.toml`

```toml
editor = ""               # editor command (overrides $VISUAL/$EDITOR)
default_language = "text"  # default language for new snippets
theme = "catppuccin-mocha" # color theme
```

**Editor resolution:** `config.editor` > `$VISUAL` > `$EDITOR` > `vi`

GUI editors (`code`, `zed`, `subl`, etc.) automatically get `--wait` injected so the TUI suspends properly.

**Database:** `~/.local/share/snipt/snipt.db` (override with `--db`)

## Tech Stack

- [Go](https://go.dev)
- [Bubbletea v2](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss v2](https://github.com/charmbracelet/lipgloss) — styling
- [Bubbles v2](https://github.com/charmbracelet/bubbles) — TUI components
- [SQLite](https://modernc.org/sqlite) via modernc.org (pure Go, no CGo)
- FTS5 — full-text search
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Catppuccin Mocha](https://catppuccin.com) — color theme

## Roadmap

- [ ] Gist sync (backup/share snippets via GitHub Gists)
- [ ] Sort options in TUI (by name, date, use count)
- [ ] Bulk operations
- [ ] Import/export from TUI
- [ ] Configurable keybindings

## License

[MIT](LICENSE)
