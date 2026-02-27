# System Tray + Settings Page Design

**Date**: 2026-02-27
**Status**: Approved

## Overview

Add a macOS system tray icon that keeps snipt running in the background, and a settings page accessible from both the tray menu and the manage GUI sidebar.

## Decisions

- **Tray library**: `fyne.io/systray` — battle-tested, macOS/Windows/Linux
- **Global hotkey**: Skipped — display-only in settings, users bind via macOS System Settings / skhd / Hammerspoon
- **Find from tray**: Shells out to `snipt find` subprocess (fire and forget, clean separation)
- **Settings location**: Detail pane within the manage window, not a separate window

## Part 1: System Tray + Window Lifecycle

### Process Architecture

```
snipt (no args)
  ├─ goroutine: systray.Run(onReady, onExit)
  │    └─ Tray icon + menu in macOS menu bar
  │         ├─ Find    → exec "snipt find" subprocess
  │         ├─ Manage  → wailsRuntime.Show(ctx)
  │         ├─ Settings → wailsRuntime.Show(ctx) + emit "open-settings" event
  │         └─ Quit    → wailsRuntime.Quit(ctx)
  │
  └─ main thread: wails.Run(opts)
       └─ Manage window (hidden on close, not destroyed)
```

### Key Behaviors

- **Startup**: Tray goroutine starts first, then `wails.Run()` blocks on main thread. Both share the `App` struct (and its `context.Context` once Wails calls `Startup`).
- **Window close**: `Mac.OnClose` → `runtime.Hide(ctx)`. App stays alive in tray.
- **Tray > Find**: `exec.Command(os.Executable(), "find").Start()` — fire and forget. The find palette is its own Wails process with its own db connection.
- **Tray > Settings**: Show the manage window + emit a Wails event (`"open-settings"`) that the frontend listens for to switch the detail pane.
- **Quit**: `systray.Quit()` + `wailsRuntime.Quit(ctx)`. Clean shutdown.

### Tray Icon

Existing `build/tray-icon.svg` converted to PNG. Embedded via Go `embed` as `[]byte`. macOS template images (monochrome + alpha) auto-adapt to light/dark menu bars.

### Tray Menu

```
┌──────────────────────────┐
│  ✂ snipt                 │
│  ─────────────────────── │
│  Find          ⌘⇧S       │
│  Manage                  │
│  ─────────────────────── │
│  Settings                │
│  ─────────────────────── │
│  Quit snipt    ⌘Q        │
└──────────────────────────┘
```

### Changes to `launch.go`

- Start tray goroutine before `wails.Run()` (manage mode only)
- Set `Mac.OnClose` to hide instead of quit

### New file: `internal/gui/tray.go`

- `setupTray(app *App)` — waits for `app.ctx`, registers menu items + callbacks
- Embeds tray icon PNG

## Part 2: Config Package Extension

### Current State

`internal/config/config.go` has flat `Config` struct: `Editor`, `DefaultLanguage`, `Theme`. Uses `BurntSushi/toml`.

### Extended Struct

```go
type Config struct {
    // Existing fields (kept at top level for backward compat)
    Editor          string `toml:"editor"`
    DefaultLanguage string `toml:"default_language"`
    Theme           string `toml:"theme"`

    // New sections
    General GeneralConfig `toml:"general"`
    Find    FindConfig    `toml:"find"`
}

type GeneralConfig struct {
    Hotkey string `toml:"hotkey"`  // display-only
}

type FindConfig struct {
    Preview         bool   `toml:"preview"`
    Sort            string `toml:"sort"`             // "recent" | "usage" | "alpha"
    CopyToClipboard bool   `toml:"copy_to_clipboard"`
}
```

- `Editor` stays at top level — no breaking change
- `Hotkey` is display-only (no built-in registration)
- No `Sync` section yet — YAGNI, just a UI placeholder
- New `Save() error` method writes config back to disk
- `DefaultConfig()` returns sensible defaults

### New Backend Methods on `App`

```go
func (a *App) GetConfig() (*config.Config, error)
func (a *App) UpdateConfig(cfg config.Config) error
func (a *App) GetDBPath() string
func (a *App) GetVersion() string
```

Version string threaded through: `main.go` → `cli.NewRootCmd` → `gui.LaunchGUI` → `App`.

## Part 3: Frontend — Settings Page + Detail View State

### State Management

New field in `AppState`:

```typescript
type DetailView =
  | { kind: "snippet" }
  | { kind: "settings" }
  | { kind: "empty" };
```

New action: `OPEN_SETTINGS`. Existing `SELECT_SINGLE` implicitly resets to `"snippet"`.

### App.tsx Detail Pane Switching

```tsx
{state.detailView.kind === "settings" && <Settings />}
{state.detailView.kind === "snippet" && <DetailPane ... />}
{state.detailView.kind === "empty" && <EmptyState />}
```

Frontend listens for Wails `"open-settings"` event (emitted by tray Settings menu):

```typescript
EventsOn("open-settings", () => dispatch({ type: "OPEN_SETTINGS" }));
```

### Sidebar Changes

Gear button added to sidebar footer alongside `NewSnippetButton`. Highlighted when settings is active.

### Settings Component (`Settings.tsx`)

Four sections: **General**, **Find**, **Data**, **About**.

- Calls `GetConfig()`, `GetStats()`, `GetDBPath()`, `GetVersion()` on mount
- Toggle switches for `preview` and `copy_to_clipboard`
- Dropdown for `sort`
- Read-only display for hotkey, editor, theme, db path, stats
- "Set up sync" placeholder button (disabled)
- About: version, tech stack, repo link
- On change → `UpdateConfig()` → re-fetch to confirm

### Styling

Uses existing CSS variables from `global.css`. Monospace labels, subtle dividers, Catppuccin Mocha palette. Consistent with `DetailPane` aesthetic.

## What's NOT Built

- Gist sync (placeholder only)
- Theme selection (hardcoded Catppuccin Mocha)
- Hotkey recording UI (display-only, change via config file)
- Import/export
- Multiple profiles
