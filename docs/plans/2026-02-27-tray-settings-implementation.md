# System Tray + Settings Page Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a macOS system tray icon that keeps snipt alive in the background, and a settings page in the manage window's detail pane.

**Architecture:** System tray runs in a goroutine alongside the Wails main thread, sharing the `App` struct. Tray "Find" shells out to `snipt find` subprocess. Settings is a frontend view that swaps into the detail pane via a `detailView` discriminated union in React state. Config is extended with new sections persisted to `~/.config/snipt/config.toml`.

**Tech Stack:** Go + Wails v2, `fyne.io/systray`, `BurntSushi/toml` (already in go.mod), React/TypeScript frontend

---

### Task 1: Extend Config Package

**Files:**
- Modify: `src/internal/config/config.go`
- Modify: `src/internal/config/config_test.go`

**Step 1: Write failing tests for new config fields and Save()**

Add to `src/internal/config/config_test.go`:

```go
func TestLoad_NewSections(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "snipt")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(`
editor = "nvim"
default_language = "go"
theme = "catppuccin-mocha"

[general]
hotkey = "cmd+shift+s"

[find]
preview = true
sort = "alpha"
copy_to_clipboard = false
`), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.General.Hotkey != "cmd+shift+s" {
		t.Errorf("expected hotkey=cmd+shift+s, got %q", cfg.General.Hotkey)
	}
	if cfg.Find.Preview != true {
		t.Error("expected find.preview=true")
	}
	if cfg.Find.Sort != "alpha" {
		t.Errorf("expected find.sort=alpha, got %q", cfg.Find.Sort)
	}
	if cfg.Find.CopyToClipboard != false {
		t.Error("expected find.copy_to_clipboard=false")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.General.Hotkey != "cmd+shift+s" {
		t.Errorf("expected default hotkey=cmd+shift+s, got %q", cfg.General.Hotkey)
	}
	if cfg.Find.Sort != "recent" {
		t.Errorf("expected default sort=recent, got %q", cfg.Find.Sort)
	}
	if cfg.Find.CopyToClipboard != true {
		t.Error("expected default copy_to_clipboard=true")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := DefaultConfig()
	cfg.Editor = "code"
	cfg.Find.Sort = "alpha"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.Editor != "code" {
		t.Errorf("expected editor=code after save, got %q", loaded.Editor)
	}
	if loaded.Find.Sort != "alpha" {
		t.Errorf("expected find.sort=alpha after save, got %q", loaded.Find.Sort)
	}
	if loaded.Find.CopyToClipboard != true {
		t.Error("expected copy_to_clipboard preserved after save")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt && go test ./src/internal/config/ -v -run "TestLoad_NewSections|TestDefaultConfig|TestSave"`
Expected: FAIL — `DefaultConfig` and `Save` not defined, `General`/`Find` fields don't exist

**Step 3: Implement config extensions**

Replace the contents of `src/internal/config/config.go`. Keep all existing functions. Add:

- `GeneralConfig` struct with `Hotkey string`
- `FindConfig` struct with `Preview bool`, `Sort string`, `CopyToClipboard bool`
- Add `General GeneralConfig` and `Find FindConfig` fields to `Config`
- `DefaultConfig() *Config` — returns sensible defaults
- `Save() error` — marshals config to TOML and writes to `ConfigPath()`
- Update `Load()` to apply defaults for new fields when they're zero-valued (so existing config files without `[general]`/`[find]` sections still work)

```go
package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds user preferences loaded from config.toml.
type Config struct {
	Editor          string `toml:"editor"`
	DefaultLanguage string `toml:"default_language"`
	Theme           string `toml:"theme"`

	General GeneralConfig `toml:"general"`
	Find    FindConfig    `toml:"find"`
}

// GeneralConfig holds general app settings.
type GeneralConfig struct {
	Hotkey string `toml:"hotkey"`
}

// FindConfig holds find palette preferences.
type FindConfig struct {
	Preview         bool   `toml:"preview"`
	Sort            string `toml:"sort"`
	CopyToClipboard bool   `toml:"copy_to_clipboard"`
}

const defaultConfig = `editor = ""
default_language = "text"
theme = "catppuccin-mocha"

[general]
hotkey = "cmd+shift+s"

[find]
preview = false
sort = "recent"
copy_to_clipboard = true
`

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultLanguage: "text",
		Theme:           "catppuccin-mocha",
		General: GeneralConfig{
			Hotkey: "cmd+shift+s",
		},
		Find: FindConfig{
			Sort:            "recent",
			CopyToClipboard: true,
		},
	}
}

// Load reads the config file, creating it with defaults if it doesn't exist.
func Load() (*Config, error) {
	path := ConfigPath()
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
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

// Save writes the config to disk at ConfigPath().
func (c *Config) Save() error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(c); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// ResolveEditor returns the editor to use, following the chain:
// config editor -> $VISUAL -> $EDITOR -> vi
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

**Step 4: Run tests to verify they pass**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt && go test ./src/internal/config/ -v`
Expected: ALL PASS (existing tests + new tests)

**Step 5: Commit**

```bash
git add src/internal/config/config.go src/internal/config/config_test.go
git commit -m "feat(config): add General/Find sections, Save(), DefaultConfig()"
```

---

### Task 2: Add Config + Version Backend Methods to App

**Files:**
- Modify: `src/internal/gui/app.go`
- Modify: `src/internal/gui/launch.go`
- Modify: `src/internal/cli/root.go`
- Modify: `src/internal/cli/manage.go`
- Modify: `src/internal/cli/find.go`
- Modify: `src/cmd/snipt/main.go`

**Step 1: Thread version string through to App**

The `version` variable is set via ldflags in `main.go` and passed to `cli.NewRootCmd(version)`. We need it to reach `gui.LaunchGUI` and the `App` struct.

In `src/internal/gui/app.go`, add a `version` field to `App`:

```go
type App struct {
	ctx     context.Context
	store   *db.Store
	mode    string
	version string
}

func NewApp(store *db.Store, mode, version string) *App {
	return &App{store: store, mode: mode, version: version}
}
```

In `src/internal/gui/launch.go`, update signature:

```go
func LaunchGUI(store *db.Store, mode, version string) error {
	app := NewApp(store, mode, version)
	// ... rest unchanged
}
```

In `src/internal/cli/root.go`, store version in a package-level var and pass it through:

```go
var appVersion string

func NewRootCmd(version string) *cobra.Command {
	appVersion = version
	// ...
	root.RunE = func(cmd *cobra.Command, args []string) error {
		return gui.LaunchGUI(store, "manage", appVersion)
	}
	// ...
}
```

In `src/internal/cli/manage.go`, update the `gui.LaunchGUI` call:

```go
return gui.LaunchGUI(store, "manage", appVersion)
```

In `src/internal/cli/find.go`, update the `gui.LaunchGUI` call:

```go
return gui.LaunchGUI(store, "find", appVersion)
```

**Step 2: Add config and version methods to App**

Add to `src/internal/gui/app.go`:

```go
import (
	"github.com/infktd/snipt/src/internal/config"
)

// GetConfig returns the current user config.
func (a *App) GetConfig() (*config.Config, error) {
	return config.Load()
}

// UpdateConfig saves the given config to disk.
func (a *App) UpdateConfig(cfg config.Config) error {
	return cfg.Save()
}

// GetDBPath returns the path to the SQLite database.
func (a *App) GetDBPath() string {
	return config.DBPath("")
}

// GetVersion returns the app version string.
func (a *App) GetVersion() string {
	return a.version
}
```

**Step 3: Build to verify**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt && go build ./src/cmd/snipt/`
Expected: Compiles without errors

**Step 4: Commit**

```bash
git add src/internal/gui/app.go src/internal/gui/launch.go src/internal/cli/root.go src/internal/cli/manage.go src/internal/cli/find.go src/cmd/snipt/main.go
git commit -m "feat(gui): add GetConfig, UpdateConfig, GetDBPath, GetVersion to App; thread version string"
```

---

### Task 3: Add System Tray

**Files:**
- Create: `src/internal/gui/tray.go`
- Create: `src/internal/gui/tray_icon.png` (embedded)
- Modify: `src/internal/gui/launch.go`

**Step 1: Add `fyne.io/systray` dependency**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt && go get fyne.io/systray`

**Step 2: Create tray icon PNG**

The SVG at `src/cmd/snipt/build/tray-icon.svg` is a 22x22 macOS template image (black strokes on transparent). Convert it to a PNG file for embedding. The icon needs to be a proper macOS template image: black content on transparent background, 22x22 at 1x and 44x44 at 2x.

Use a tool or script to convert the SVG to PNG. If `rsvg-convert` or `cairosvg` is available:

```bash
# Generate 22x22 (1x) PNG
rsvg-convert -w 22 -h 22 src/cmd/snipt/build/tray-icon.svg > src/internal/gui/tray_icon.png
```

If no SVG converter is available, create a minimal Go program to generate the PNG, or use ImageMagick:

```bash
convert -background none -resize 22x22 src/cmd/snipt/build/tray-icon.svg src/internal/gui/tray_icon.png
```

**Step 3: Create `src/internal/gui/tray.go`**

```go
package gui

import (
	_ "embed"
	"os"
	"os/exec"
	"time"

	"fyne.io/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed tray_icon.png
var trayIconBytes []byte

// setupTray starts the system tray icon and menu.
// It blocks until systray.Quit() is called.
func setupTray(app *App) {
	systray.Run(func() {
		onTrayReady(app)
	}, func() {
		// onExit — nothing to clean up
	})
}

func onTrayReady(app *App) {
	systray.SetIcon(trayIconBytes)
	systray.SetTooltip("snipt")

	// Title item (disabled, just a label)
	mTitle := systray.AddMenuItem("✂ snipt", "")
	mTitle.Disable()

	systray.AddSeparator()

	mFind := systray.AddMenuItem("Find", "Open find palette")
	mManage := systray.AddMenuItem("Manage", "Show manage window")

	systray.AddSeparator()

	mSettings := systray.AddMenuItem("Settings", "Open settings")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit snipt", "Quit the application")

	go func() {
		for {
			select {
			case <-mFind.ClickedCh:
				launchFindPalette()
			case <-mManage.ClickedCh:
				showManageWindow(app)
			case <-mSettings.ClickedCh:
				showSettings(app)
			case <-mQuit.ClickedCh:
				systray.Quit()
				wailsRuntime.Quit(app.ctx)
				return
			}
		}
	}()
}

// launchFindPalette spawns "snipt find" as a separate process.
func launchFindPalette() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe, "find")
	cmd.Start()
}

// showManageWindow brings the manage window to front.
func showManageWindow(app *App) {
	// Wait for Wails context to be available
	for app.ctx == nil {
		time.Sleep(50 * time.Millisecond)
	}
	wailsRuntime.Show(app.ctx)
	wailsRuntime.WindowShow(app.ctx)
}

// showSettings shows the manage window and emits an event to open settings.
func showSettings(app *App) {
	showManageWindow(app)
	wailsRuntime.EventsEmit(app.ctx, "open-settings")
}
```

**Step 4: Update `src/internal/gui/launch.go` — start tray + hide on close**

```go
package gui

import (
	"fyne.io/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/infktd/snipt/src/frontend"
	"github.com/infktd/snipt/src/internal/db"
)

// LaunchGUI starts a Wails window in the given mode ("manage" or "find").
func LaunchGUI(store *db.Store, mode, version string) error {
	app := NewApp(store, mode, version)

	if mode == "manage" {
		go setupTray(app)
	}

	opts := &options.App{
		Title: "snipt",
		AssetServer: &assetserver.Options{
			Assets: frontend.Assets,
		},
		BackgroundColour: &options.RGBA{R: 13, G: 13, B: 20, A: 255},
		OnStartup:        app.Startup,
		Bind:             []interface{}{app},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                 true,
				HideTitleBar:              false,
				FullSizeContent:           true,
				UseToolbar:                false,
			},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
		},
	}

	switch mode {
	case "find":
		opts.Title = "snipt find"
		opts.Width = 680
		opts.Height = 420
		opts.MaxHeight = 500
		opts.MinWidth = 500
		opts.Frameless = true
		opts.AlwaysOnTop = true
		opts.BackgroundColour = &options.RGBA{R: 36, G: 36, B: 53, A: 255}
		opts.Mac.TitleBar.HideTitleBar = true
	default: // "manage"
		opts.Width = 1100
		opts.Height = 700
		opts.MinWidth = 800
		opts.MinHeight = 500
		opts.Mac.OnClose = func() {
			wailsRuntime.Hide(app.ctx)
		}
	}

	err := wails.Run(opts)

	// If Wails exits (e.g. from Quit), also quit the tray
	if mode == "manage" {
		systray.Quit()
	}

	return err
}
```

**Step 5: Build to verify**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt && go build ./src/cmd/snipt/`
Expected: Compiles. (Tray icon PNG must exist for the embed to work.)

**Step 6: Manual test**

Run the built binary. Verify:
- Tray icon appears in macOS menu bar
- Tray menu shows: ✂ snipt (disabled), Find, Manage, Settings, Quit snipt
- Closing the manage window hides it (doesn't quit)
- Tray > Manage brings the window back
- Tray > Find opens a find palette window
- Tray > Quit exits the app completely

**Step 7: Commit**

```bash
git add src/internal/gui/tray.go src/internal/gui/tray_icon.png src/internal/gui/launch.go go.mod go.sum
git commit -m "feat(gui): add system tray with Find, Manage, Settings, Quit"
```

---

### Task 4: Frontend State — Add detailView to Reducer

**Files:**
- Modify: `src/frontend/src/state/types.ts`
- Modify: `src/frontend/src/state/context.tsx`

**Step 1: Add DetailView type and OPEN_SETTINGS action to types.ts**

Add to `src/frontend/src/state/types.ts`:

```typescript
export type DetailView =
  | { kind: "snippet" }
  | { kind: "settings" }
  | { kind: "empty" };
```

Add `detailView: DetailView;` to the `AppState` interface.

Add `| { type: "OPEN_SETTINGS" }` to the `AppAction` union.

**Step 2: Update reducer in context.tsx**

In `src/frontend/src/state/context.tsx`:

Add `detailView: { kind: "snippet" }` to `initialState`.

Add reducer case:

```typescript
case "OPEN_SETTINGS":
  return {
    ...state,
    detailView: { kind: "settings" },
    editMode: false,
    createMode: false,
  };
```

Modify `SELECT_SINGLE` case to also set `detailView: { kind: "snippet" }`:

```typescript
case "SELECT_SINGLE":
  return {
    ...state,
    selectedIds: new Set([action.id]),
    anchorId: action.id,
    focusId: action.id,
    editMode: false,
    createMode: false,
    detailView: { kind: "snippet" },
  };
```

**Step 3: Commit**

```bash
git add src/frontend/src/state/types.ts src/frontend/src/state/context.tsx
git commit -m "feat(state): add detailView and OPEN_SETTINGS action"
```

---

### Task 5: Settings Component

**Files:**
- Create: `src/frontend/src/components/Settings.tsx`

**Step 1: Create the Settings component**

This component:
- Calls `GetConfig()`, `GetStats()`, `GetDBPath()`, `GetVersion()` on mount
- Renders four sections: General, Find, Data, About
- Toggle switches for `preview` and `copy_to_clipboard`
- Dropdown for `sort`
- Read-only display for hotkey, editor, theme, db path, stats, version
- On change → `UpdateConfig()` → re-fetch

```typescript
import { useState, useEffect, useCallback } from "react";
import { C, MONO, BODY } from "../styles/colors";
import {
  GetConfig,
  UpdateConfig,
  GetStats,
  GetDBPath,
  GetVersion,
} from "../wailsjs/go/gui/App";
import { BrowserOpenURL } from "../wailsjs/runtime/runtime";

interface Config {
  Editor: string;
  DefaultLanguage: string;
  Theme: string;
  General: {
    Hotkey: string;
  };
  Find: {
    Preview: boolean;
    Sort: string;
    CopyToClipboard: boolean;
  };
}

interface Stats {
  TotalSnippets: number;
  TotalTags: number;
  Languages: Record<string, number>;
}

export function Settings() {
  const [config, setConfig] = useState<Config | null>(null);
  const [stats, setStats] = useState<Stats | null>(null);
  const [dbPath, setDbPath] = useState("");
  const [version, setVersion] = useState("");

  const loadConfig = useCallback(async () => {
    try {
      const cfg = await GetConfig();
      setConfig(cfg);
    } catch (err) {
      console.error("Failed to load config:", err);
    }
  }, []);

  useEffect(() => {
    loadConfig();
    GetStats()
      .then((s) => setStats(s))
      .catch(console.error);
    GetDBPath()
      .then(setDbPath)
      .catch(console.error);
    GetVersion()
      .then(setVersion)
      .catch(console.error);
  }, [loadConfig]);

  async function updateField(updater: (cfg: Config) => Config) {
    if (!config) return;
    const updated = updater({ ...config });
    try {
      await UpdateConfig(updated as never);
      await loadConfig();
    } catch (err) {
      console.error("Failed to update config:", err);
    }
  }

  if (!config) return null;

  const langCount = stats ? Object.keys(stats.Languages ?? {}).length : 0;

  return (
    <div
      style={{
        flex: 1,
        paddingTop: 40,
        overflow: "auto",
        background: C.bg,
      }}
    >
      <div style={{ padding: 24, maxWidth: 600 }}>
        <h2
          style={{
            fontFamily: MONO,
            fontSize: 16,
            fontWeight: 600,
            color: C.text,
            margin: "0 0 24px",
          }}
        >
          Settings
        </h2>

        {/* GENERAL */}
        <SectionLabel>General</SectionLabel>
        <Divider />
        <Row label="Global Hotkey" value={config.General?.Hotkey || "cmd+shift+s"} muted />
        <Row label="Editor" value={config.Editor || "system default"} muted />
        <Row label="Theme" value="Catppuccin Mocha" muted />

        {/* FIND */}
        <SectionLabel>Find</SectionLabel>
        <Divider />
        <ToggleRow
          label="Show preview"
          value={config.Find?.Preview ?? false}
          onChange={(v) =>
            updateField((c) => ({
              ...c,
              Find: { ...c.Find, Preview: v },
            }))
          }
        />
        <SelectRow
          label="Default sort"
          value={config.Find?.Sort || "recent"}
          options={[
            { value: "recent", label: "Most recent" },
            { value: "usage", label: "Most used" },
            { value: "alpha", label: "Alphabetical" },
          ]}
          onChange={(v) =>
            updateField((c) => ({
              ...c,
              Find: { ...c.Find, Sort: v },
            }))
          }
        />
        <ToggleRow
          label="Copy to clipboard"
          value={config.Find?.CopyToClipboard ?? true}
          onChange={(v) =>
            updateField((c) => ({
              ...c,
              Find: { ...c.Find, CopyToClipboard: v },
            }))
          }
        />

        {/* DATA */}
        <SectionLabel>Data</SectionLabel>
        <Divider />
        <Row label="Database path" value={dbPath || "~/.local/share/snipt/"} />
        <Row label="Snippets" value={String(stats?.TotalSnippets ?? 0)} />
        <Row label="Languages" value={String(langCount)} />
        <Row label="Tags" value={String(stats?.TotalTags ?? 0)} />
        <div style={{ padding: "10px 0" }}>
          <Row label="Gist Sync" value="Not configured" muted />
          <button
            style={{
              background: "transparent",
              color: C.mauve,
              fontFamily: MONO,
              fontSize: 12,
              border: `1px solid ${C.border}`,
              borderRadius: 6,
              padding: "6px 14px",
              cursor: "not-allowed",
              opacity: 0.5,
              marginTop: 4,
            }}
            disabled
          >
            Set up sync
          </button>
        </div>

        {/* ABOUT */}
        <SectionLabel>About</SectionLabel>
        <Divider />
        <Row label="" value={`snipt ${version || "dev"}`} />
        <Row label="" value="Built with Go, Bubbletea, Wails" muted />
        <div style={{ padding: "10px 0" }}>
          <span
            style={{
              fontFamily: MONO,
              fontSize: 13,
              color: C.mauve,
              cursor: "pointer",
            }}
            onClick={() => BrowserOpenURL("https://github.com/infktd/snipt")}
          >
            github.com/infktd/snipt
          </span>
        </div>
      </div>
    </div>
  );
}

// --- Sub-components ---

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div
      style={{
        fontFamily: MONO,
        fontSize: 11,
        fontWeight: 600,
        color: C.mauve,
        textTransform: "uppercase",
        letterSpacing: 1.5,
        margin: "24px 0 12px",
      }}
    >
      {children}
    </div>
  );
}

function Divider() {
  return (
    <hr
      style={{
        border: "none",
        borderTop: `1px solid ${C.borderSubtle}`,
        margin: "0 0 16px",
      }}
    />
  );
}

function Row({
  label,
  value,
  muted,
}: {
  label: string;
  value: string;
  muted?: boolean;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "10px 0",
        fontFamily: MONO,
        fontSize: 13,
      }}
    >
      {label && <span style={{ color: C.textSub }}>{label}</span>}
      <span style={{ color: muted ? C.textDim : C.text }}>{value}</span>
    </div>
  );
}

function ToggleRow({
  label,
  value,
  onChange,
}: {
  label: string;
  value: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "10px 0",
        fontFamily: MONO,
        fontSize: 13,
      }}
    >
      <span style={{ color: C.textSub }}>{label}</span>
      <div
        onClick={() => onChange(!value)}
        style={{
          width: 36,
          height: 20,
          borderRadius: 10,
          background: value ? C.mauve : C.bgSurface,
          border: `1px solid ${value ? C.mauve : C.border}`,
          cursor: "pointer",
          position: "relative",
          transition: "background 0.15s ease",
        }}
      >
        <div
          style={{
            width: 14,
            height: 14,
            borderRadius: "50%",
            background: value ? C.bg : C.text,
            position: "absolute",
            top: 2,
            left: value ? 20 : 2,
            transition: "left 0.15s ease",
          }}
        />
      </div>
    </div>
  );
}

function SelectRow({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (v: string) => void;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "10px 0",
        fontFamily: MONO,
        fontSize: 13,
      }}
    >
      <span style={{ color: C.textSub }}>{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        style={{
          background: C.bgSurface,
          border: `1px solid ${C.border}`,
          borderRadius: 6,
          color: C.text,
          fontFamily: MONO,
          fontSize: 12,
          padding: "4px 10px",
          cursor: "pointer",
          outline: "none",
        }}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add src/frontend/src/components/Settings.tsx
git commit -m "feat(gui): add Settings component"
```

---

### Task 6: Wire Settings into App.tsx + Sidebar

**Files:**
- Modify: `src/frontend/src/App.tsx`
- Modify: `src/frontend/src/components/Sidebar.tsx`

**Step 1: Update Sidebar to include gear button**

In `src/frontend/src/components/Sidebar.tsx`:

- Add `onOpenSettings` and `settingsActive` to `SidebarProps`
- Add a settings button in the sidebar footer alongside `NewSnippetButton`

The footer area needs to wrap both buttons. Currently it's just `<NewSnippetButton />` at the bottom. Wrap in a div:

```tsx
interface SidebarProps {
  // ... existing props ...
  onOpenSettings: () => void;
  settingsActive: boolean;
}

// In the return, replace the bare <NewSnippetButton /> with:
<div style={{ display: "flex", borderTop: `1px solid ${C.borderSubtle}` }}>
  <button
    onClick={onOpenSettings}
    style={{
      padding: "10px 14px",
      background: settingsActive ? C.bgSurface : "transparent",
      border: "none",
      borderRight: `1px solid ${C.borderSubtle}`,
      color: settingsActive ? C.mauve : C.textDim,
      fontFamily: BODY,
      fontSize: 13,
      cursor: "pointer",
      transition: "color 0.12s ease",
    }}
    onMouseEnter={(e) => { if (!settingsActive) e.currentTarget.style.color = C.textSub; }}
    onMouseLeave={(e) => { if (!settingsActive) e.currentTarget.style.color = C.textDim; }}
    title="Settings"
  >
    ⚙
  </button>
  <NewSnippetButton onClick={onNewSnippet} />
</div>
```

Remove `borderTop` from `NewSnippetButton` since the wrapper div now has it. Actually — `NewSnippetButton` has its own `borderTop` in its inline styles. We have two options: modify `NewSnippetButton` to not render `borderTop`, or override it. Simplest: remove `borderTop` from `NewSnippetButton`'s style and let the parent div handle it.

**Step 2: Update App.tsx**

In `src/frontend/src/App.tsx`:

- Import `Settings` component
- Import `EventsOn` from Wails runtime
- Add `useEffect` to listen for `"open-settings"` event
- Pass `onOpenSettings` and `settingsActive` to `Sidebar`
- Switch detail pane rendering based on `state.detailView`

Key changes in `AppContent`:

```tsx
import { Settings } from "./components/Settings";
import { EventsOn } from "./wailsjs/runtime/runtime";

// Inside AppContent:
useEffect(() => {
  const cancel = EventsOn("open-settings", () => {
    dispatch({ type: "OPEN_SETTINGS" });
  });
  return cancel;
}, [dispatch]);

// In the JSX, replace the bare <DetailPane ... /> with:
{state.detailView.kind === "settings" ? (
  <Settings />
) : (
  <DetailPane ... />
)}
```

Pass new props to `Sidebar`:

```tsx
<Sidebar
  // ... existing props ...
  onOpenSettings={() => dispatch({ type: "OPEN_SETTINGS" })}
  settingsActive={state.detailView.kind === "settings"}
/>
```

**Step 3: Update NewSnippetButton to remove borderTop**

In `src/frontend/src/components/NewSnippetButton.tsx`, remove the `borderTop` from the button's inline style (the parent sidebar footer div now provides it).

**Step 4: Regenerate Wails bindings**

After adding `GetConfig`, `UpdateConfig`, `GetDBPath`, `GetVersion` to `App`, the Wails TypeScript bindings need to be regenerated:

Run: `cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt && wails generate module`

This updates `src/frontend/src/wailsjs/go/gui/App.js`, `App.d.ts`, and `models.ts` to include the new methods and the `config.Config` type.

**Step 5: Build and test**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt && wails build`

Manual test:
- Click gear icon in sidebar → settings page appears in detail pane
- Click a snippet in sidebar → switches back to snippet detail
- Toggle switches work and persist (check `~/.config/snipt/config.toml`)
- Sort dropdown works
- Stats display correct numbers
- Version displays

**Step 6: Commit**

```bash
git add src/frontend/src/App.tsx src/frontend/src/components/Sidebar.tsx src/frontend/src/components/NewSnippetButton.tsx src/frontend/src/wailsjs/
git commit -m "feat(gui): wire Settings into App.tsx and Sidebar"
```

---

### Task 7: Full Integration Test

**Files:** None (testing only)

**Step 1: Full build**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt && wails build`

**Step 2: Run all Go tests**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt && go test ./...`
Expected: ALL PASS

**Step 3: Manual verification checklist**

```
[ ] Tray icon appears in macOS menu bar when snipt launches
[ ] Tray menu shows: ✂ snipt (disabled), Find, Manage, Settings, Quit snipt
[ ] Tray > Find opens the floating palette (separate process)
[ ] Tray > Manage opens/shows the manage window
[ ] Tray > Settings opens settings in the manage window
[ ] Tray > Quit fully exits the app
[ ] Closing the manage window hides it (doesn't quit the app)
[ ] Tray icon stays visible after closing the manage window
[ ] Settings page renders in the detail pane
[ ] Settings: find preview toggle works and persists to config.toml
[ ] Settings: sort dropdown works and persists
[ ] Settings: copy to clipboard toggle works and persists
[ ] Settings: shows correct snippet count, language count, tag count
[ ] Settings: shows database path
[ ] Settings: shows version
[ ] Settings: Gist sync shows "Not configured" with disabled button
[ ] Settings: About section shows version + repo link
[ ] Gear icon in sidebar opens settings
[ ] Gear icon highlights when settings is active
[ ] Clicking a snippet while in settings switches back to detail view
[ ] Config file created at ~/.config/snipt/config.toml on first launch
[ ] Config file preserves existing values when updating
[ ] Existing config files without [general]/[find] sections still load fine
```

**Step 4: Commit (if any fixes needed)**

```bash
git add -A
git commit -m "fix: integration fixes for tray + settings"
```
