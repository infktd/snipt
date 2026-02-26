# Manage TUI View() Rewrite Design

## Goal

Rewrite the rendering (`View()` and all `render*` methods) in `manage.go` to match the layout spec in `MANAGE-PROMPT.md`. No behavioral changes -- keybindings, navigation, CRUD, filtering all stay the same.

## Scope

1. **Rewrite `View()` and all `render*` helpers** in `manage.go`
2. **Refactor syntax highlighter** to accept a background color parameter
3. **Add minimal accessors** to `ResultList` so manage owns its sidebar rendering
4. **Keep all existing color values** as-is (no Catppuccin standardization)

## Changes

### 1. Syntax Highlighter (`resultlist.go`)

Add `bg color.Color` parameter to:
- `SyntaxHighlightLine(line, language string, bg color.Color) string`
- `HighlightTokens(code, language string, bg color.Color) string`

Update callers:
- `resultlist.go` internal calls: pass `common.ColorBgSurface` (unchanged behavior)
- `manage.go` preview: pass `common.ColorBg`

### 2. ResultList Accessors (`resultlist.go` / `accessors.go`)

Add to `ResultList`:
```go
func (r *ResultList) Cursor() int       // returns r.cursor
func (r *ResultList) Items() []ResultItem  // returns r.items
func (r *ResultList) Len() int          // returns len(r.items)
```

Manage computes its own scroll window (8 lines of math using `contentHeight / 2`). No `VisibleRange()` accessor needed.

### 3. Manage View() Rewrite (`manage.go`)

Replace all rendering functions. New structure:

```
View()
  ├── renderHeader()          -- badge + search + count, ColorBgSurface bg
  ├── renderHorizontalRule()  -- ─ full width, ColorBorderDim fg
  ├── renderContent()         -- per-line: sidebar + │ + preview
  │   ├── sidebar lines       -- 2 lines per snippet, ColorBg / ColorBgSelected
  │   └── preview lines       -- title, code, footer on ColorBg
  ├── renderHorizontalRule()  -- reused
  └── renderStatusBar()       -- ColorMauve bg, ColorBg text
```

**Layout math:**
```
sidebarWidth  = clamp(termWidth * 30 / 100, 25, 40)
sepWidth      = 1
previewWidth  = termWidth - sidebarWidth - sepWidth
contentHeight = termHeight - 4   (header + 2 separators + status bar)
```

Note: the spec says `termHeight - 3` but counts 4 rows (header, sep, sep, status). Using the correct math: `termHeight - 4`.

**Key rendering changes from current code:**
- Sidebar bg: `ColorBg` (was `ColorBgSurface`)
- Sidebar rows: rendered in manage.go, not delegated to `ResultList.View()`
- Preview footer: metadata line with `N lines  used M×  created DATE  #tags`
- Status bar: `ColorMauve` bg with `ColorBg` text (was surface bg with colored text)
- Two horizontal separators (header-content, content-status) added

### 4. What Stays Unchanged

- `Update()`, all keybinding logic, modes, CRUD operations
- `NewManageModel()` initialization
- `applyFilter()`, `reloadSnippets()`, frontmatter handling
- Find palette rendering in `resultlist.go` (its own `View()` untouched)
- All color constant values in `colors.go`

## Rendering Rules

Per hard-won lessons from the find palette:
1. Every styled element gets explicit `.Background(bg)`
2. Every row gets `.Width(targetWidth)` on its container
3. Two-stage padding: pad segments to allocated width, then verify total
4. Per-line rendering for variable-height elements (no border wrappers)
5. Split-and-zip for horizontal composition of sidebar + separator + preview
6. `View()` builds entire screen as one string, returns via `tea.NewView()` with `AltScreen = true`
