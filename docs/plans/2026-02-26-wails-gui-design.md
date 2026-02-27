# snipt GUI: Wails Desktop App Design

**Date**: 2026-02-26
**Branch**: `gui`
**Status**: Approved

## Overview

Desktop GUI for snipt using Wails v2 + React 18 + TypeScript. Shares the existing SQLite data layer (`internal/db`) with the CLI. Dark Catppuccin Mocha aesthetic matching the TUI and getsnipt.dev landing page.

Scope: browse, search, create, edit, delete, copy, pin. Nothing else.

---

## 1. Backend Layer

**File**: `src/internal/gui/app.go`

Thin passthrough to `*db.Store`. No business logic.

```go
package gui

type App struct {
    ctx   context.Context
    store *db.Store
}

func NewApp(store *db.Store) *App
func (a *App) Startup(ctx context.Context)

// Exposed to frontend via Wails auto-generated bindings:
func (a *App) ListSnippets(opts db.ListOpts) ([]model.Snippet, error)
func (a *App) SearchSnippets(query string) ([]model.SearchResult, error)
func (a *App) GetSnippet(id string) (*model.Snippet, error)
func (a *App) CreateSnippet(s model.Snippet) error
func (a *App) UpdateSnippet(s model.Snippet) error
func (a *App) UpdateSnippetTags(id string, tags []string) error
func (a *App) DeleteSnippet(id string) error
func (a *App) SetPinned(id string, pinned bool) error
func (a *App) IncrementUseCount(id string) error
func (a *App) GetStats() (*model.Stats, error)
```

### Key decisions

- **`SearchSnippets` returns `[]model.SearchResult`** (not `[]model.Snippet`). Preserves `Score` and `TitleIndices` for match highlighting in the sidebar. Frontend ignores what it doesn't need, but data isn't thrown away.
- **`ListSnippets` takes `db.ListOpts`** (Language, Tag, Pinned, Sort filters). Frontend can use them for filtered views later (language dropdown, sort toggle) without rewiring the backend.
- **`UpdateSnippetTags`** is separate from `UpdateSnippet` because `Store.Update()` doesn't touch tags (junction table). Backend diffs current vs desired tags, calls `AddTags`/`RemoveTags`. Frontend just sends the final tag set.
- **`CreateSnippet`** generates ID server-side via `model.NewID()` and sets timestamps before `store.Create()`.
- **`SetPinned(id, bool)`** instead of `TogglePin` -- explicit state beats implicit toggling with async frontend/backend communication.

---

## 2. Wails Entry Point

**File**: `src/cmd/snipt-gui/main.go`

- DB path: `config.DBPath("")` -- same resolution as CLI (XDG_DATA_HOME or ~/.local/share/snipt/snipt.db)
- macOS title bar: transparent, hidden title, full-size content, traffic lights in native position
- Window background: `#0d0d14` (`bg` color) at native level -- no white flash on launch
- Window size: 1100x700 default, 800x500 minimum
- Frontend assets: embedded via `//go:embed all:frontend/dist`

---

## 3. Frontend Architecture

**Toolchain**: Vite + React 18 + TypeScript

### Component tree

```
App
├── Sidebar (300px fixed)
│   ├── SearchBar          debounced 300ms, FTS5 via SearchSnippets
│   ├── SnippetList        scrollable list
│   │   └── SnippetRow     title, pin indicator, lang badge, tags, match highlights
│   └── NewSnippetButton
├── DetailPane (flex remaining)
│   ├── DetailHeader       inline-editable title, language badge
│   ├── CodeEditor         CodeMirror 6, read-only default
│   ├── MetadataFooter     tag editor, pin toggle, use count, dates
│   └── ActionBar          Copy, Edit/Save, Delete, Pin/Unpin
└── StatusBar              snippet count, search indicator, shortcut hints
```

### State management

React Context + `useReducer`. State shape:

- `snippets: Snippet[]` -- current list
- `searchResults: SearchResult[] | null` -- non-null when searching
- `selectedId: string | null`
- `editMode: boolean`
- `searchQuery: string`

One context provider at App level.

### Data flow

1. App mounts -> `ListSnippets({})` -> populate sidebar
2. User types in search -> debounce 300ms -> `SearchSnippets(query)` -> replace sidebar with results (preserving `TitleIndices` for highlights)
3. User clears search -> `ListSnippets({})` again
4. CRUD operations -> call backend -> refetch list -> maintain selection if still exists

### Match highlighting

`SnippetRow` receives `TitleIndices` from `SearchResult`, wraps matched characters in `<mark>` spans styled with `pink` color. CSS-native.

### No client-side routing

Single-view app. Sidebar/detail split is component state, not URL-driven.

---

## 4. CodeMirror 6 Integration

**Theme**: Custom Catppuccin Mocha theme in `src/frontend/src/editor/catppuccin-theme.ts`:

- Background: `#1e1e2e` (bgTerminal)
- Gutters: `#242435` (bgSurface), line numbers `#45475a` (textMuted)
- Selection: `#45475a`
- Caret: `#cba6f7` (mauve)
- Keywords: `#cba6f7` (mauve, bold 600)
- Strings: `#a6e3a1` (green)
- Comments: `#45475a` (textMuted, italic)
- Functions: `#89b4fa` (blue)
- Numbers: `#fab387` (peach)
- Types: `#f9e2af` (yellow)
- Operators: `#89dceb` (sky)
- Punctuation: `#6c7086` (textDim)

**Language loading**: Lazy dynamic imports from a map in `src/frontend/src/editor/languages.ts`. Supported out of the box: go, javascript, typescript, python, sql, html, css, json, markdown. Bash/yaml via community packages if available, otherwise plain text fallback.

**Read-only / edit toggle**: Managed via CodeMirror compartment reconfiguration (not remounting). No flicker, no state loss. "Edit" button or double-click enters edit mode, "Save" reads doc content, calls `UpdateSnippet()`, flips back to read-only.

---

## 5. Keyboard Shortcuts

Global `keydown` listener on `window`, registered once at App mount.

| Shortcut | Action | Context |
|----------|--------|---------|
| `Cmd+N` | New snippet | Always |
| `Cmd+F` | Focus search | Always |
| `Cmd+C` | Copy snippet content | Detail pane focused, no text selection |
| `Cmd+S` | Save edit | Editing |
| `Cmd+Backspace` | Delete (with confirm dialog) | Not editing |
| `Up/Down` | Navigate snippet list | Search/list focused |
| `Escape` | Cancel edit / clear search | Contextual |
| `Cmd+P` | Toggle pin | Snippet selected |

**Note**: `Cmd+P` overrides macOS Print. If this causes issues, fallback to `Cmd+Shift+P`.

**Cmd+C interception**: Only when detail pane has focus AND no text selection in CodeMirror. Native copy-paste in the editor always works.

**Clipboard**: Uses `ClipboardSetText()` from `@wailsapp/runtime` -- no Go backend method needed.

---

## 6. Layout & Styling

### Color palette (CSS variables)

```css
:root {
  --bg: #0d0d14;
  --bg-card: #151521;
  --bg-terminal: #1e1e2e;
  --bg-surface: #242435;
  --border: #2a2a3c;
  --border-subtle: #1f1f30;
  --text: #cdd6f4;
  --text-sub: #a6adc8;
  --text-dim: #6c7086;
  --text-muted: #45475a;
  --pink: #f5c2e7;
  --mauve: #cba6f7;
  --peach: #fab387;
  --green: #a6e3a1;
  --teal: #94e2d5;
  --blue: #89b4fa;
  --yellow: #f9e2af;
  --red: #f38ba8;
  --lavender: #b4befe;
  --sky: #89dceb;
}
```

### Typography

- Monospace: `"Berkeley Mono", "JetBrains Mono", "Fira Code", monospace`
- Body: `"DM Sans", "Helvetica Neue", sans-serif` (imported from Google Fonts)

### macOS title bar

- 40px top padding on sidebar and detail pane
- `--wails-draggable: drag` on top bar div
- Traffic lights in native position

### Sidebar

- 300px fixed width
- SNIPT badge pill: gradient pink-to-mauve background, dark text, bold
- Search input: `bgSurface` background, `border` bottom separator
- Snippet rows: two lines (pin + title / lang badge + tags)
- Selected: `bgSurface` background, left border accent in `pink`
- Hover: subtle background shift
- Sorted: pinned first, then by most recently updated
- `+ New snippet` button at bottom

### Detail pane

- Title: large, bold, inline-editable on click
- Language badge: right-aligned colored pill
- Code: CodeMirror 6 with custom theme (see section 4)
- Metadata: tag pills (editable), pin toggle, use count, dates in `textMuted`
- Actions: Copy (brief "Copied!" feedback), Edit/Save toggle, Delete (confirmation dialog), Pin/Unpin

### Status bar

- Left: snippet count, search indicator
- Right: keyboard shortcut hints (dim text)
- `bgCard` background, thin `border` top

---

## 7. File Structure

### New files

```
src/
├── cmd/snipt-gui/
│   └── main.go
├── internal/gui/
│   └── app.go
└── frontend/
    ├── index.html
    ├── package.json
    ├── tsconfig.json
    ├── vite.config.ts
    ├── wailsjs/                         # auto-generated (gitignored)
    └── src/
        ├── main.tsx
        ├── App.tsx
        ├── state/
        │   ├── context.tsx
        │   └── types.ts
        ├── components/
        │   ├── Sidebar.tsx
        │   ├── SearchBar.tsx
        │   ├── SnippetList.tsx
        │   ├── SnippetRow.tsx
        │   ├── NewSnippetButton.tsx
        │   ├── DetailPane.tsx
        │   ├── DetailHeader.tsx
        │   ├── CodeEditor.tsx
        │   ├── MetadataFooter.tsx
        │   ├── ActionBar.tsx
        │   ├── StatusBar.tsx
        │   └── ConfirmDialog.tsx
        ├── editor/
        │   ├── catppuccin-theme.ts
        │   └── languages.ts
        ├── hooks/
        │   ├── useKeyboardShortcuts.ts
        │   └── useDebounce.ts
        └── styles/
            ├── global.css
            └── colors.ts
```

### Untouched

`cmd/snipt/`, `internal/cli/`, `internal/tui/`, `internal/db/`, `internal/model/`, `internal/config/` -- all existing code stays as-is.

### Dependencies

- **Go**: adds `github.com/wailsapp/wails/v2`
- **Frontend**: react 18, react-dom, @codemirror/* (view, state, language, lang-go, lang-javascript, lang-python, lang-sql, lang-html, lang-css, lang-json, lang-markdown), @lezer/highlight

---

## 8. Verification Checklist

- [ ] `wails build` succeeds
- [ ] App launches with correct window size and dark background
- [ ] Snippets load from same SQLite database as CLI
- [ ] Search filters snippets in real-time (FTS5) with match highlighting
- [ ] Clicking a snippet shows detail with syntax highlighting
- [ ] Create new snippet works, appears in sidebar
- [ ] Edit snippet works, changes persist
- [ ] Delete snippet works with confirmation, removed from sidebar and database
- [ ] Copy button copies content to clipboard with "Copied!" feedback
- [ ] Pin toggle works, pinned snippets sort to top
- [ ] Keyboard shortcuts work (Cmd+N, Cmd+F, arrows, Escape, etc.)
- [ ] CLI still works independently (`snipt`, `snipt find`, `snipt manage`)
- [ ] Both GUI and CLI read/write to the same database
- [ ] CodeMirror theme matches Catppuccin Mocha
- [ ] macOS title bar: transparent, draggable, traffic lights positioned correctly
