# Wails GUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a Wails v2 desktop GUI to snipt that shares the existing SQLite data layer with the CLI.

**Architecture:** Thin Go backend (`internal/gui/app.go`) wrapping `*db.Store` methods, exposed to a React 18 + TypeScript frontend via Wails auto-generated bindings. CodeMirror 6 for syntax highlighting/editing. Catppuccin Mocha theming throughout.

**Tech Stack:** Go 1.25, Wails v2, React 18, TypeScript, Vite, CodeMirror 6, SQLite (existing)

**Design doc:** `docs/plans/2026-02-26-wails-gui-design.md`

**Module path:** `github.com/infktd/snipt`

**Existing packages used:**
- `github.com/infktd/snipt/src/internal/db` — `Store`, `ListOpts`, `Open(path) (*Store, error)`
- `github.com/infktd/snipt/src/internal/model` — `Snippet`, `SearchResult`, `Stats`, `NewID()`
- `github.com/infktd/snipt/src/internal/config` — `DBPath(override string) string`

---

## Task 1: Install Wails CLI and Verify

**Step 1: Install Wails CLI**

Run:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**Step 2: Verify installation**

Run:
```bash
wails doctor
```

Expected: Shows Go version, platform info, dependencies. May warn about missing optional deps (npm/node) — that's fine as long as core requirements pass.

**Step 3: Verify node/npm available**

Run:
```bash
node --version && npm --version
```

Expected: Node 18+ and npm available. If not, install Node first.

---

## Task 2: Scaffold Wails Project Structure

This task creates the directory structure and config files. We use `wails init` to generate a React-TS template, then move/adapt the files to fit our project layout.

**Step 1: Generate a temporary Wails React-TS project**

Run from a temp directory:
```bash
cd /tmp && wails init -n snipt-gui -t react-ts
```

This gives us reference files for `wails.json`, `vite.config.ts`, `tsconfig.json`, and the Wails React boilerplate.

**Step 2: Create the project directories**

Run:
```bash
mkdir -p src/cmd/snipt-gui
mkdir -p src/internal/gui
mkdir -p src/frontend/src/{components,state,editor,hooks,styles}
```

**Step 3: Create `wails.json` in project root**

Create file: `src/cmd/snipt-gui/wails.json`

This tells Wails where the frontend lives and how to build it. Adapt from the template — key fields:

```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "snipt-gui",
  "outputfilename": "snipt-gui",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "frontend:dir": "../../frontend",
  "wailsjsdir": "../../frontend/src",
  "author": {
    "name": "infktd"
  }
}
```

Note: `frontend:dir` is relative to `wails.json` location (`src/cmd/snipt-gui/`), so `../../frontend` points to `src/frontend/`.

**Step 4: Create `src/frontend/package.json`**

```json
{
  "name": "snipt-gui-frontend",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@types/react": "^18.3.12",
    "@types/react-dom": "^18.3.1",
    "@vitejs/plugin-react": "^4.3.4",
    "typescript": "^5.6.3",
    "vite": "^6.0.0"
  }
}
```

CodeMirror deps are added in Task 6. Keep this minimal for now.

**Step 5: Create `src/frontend/tsconfig.json`**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"]
}
```

**Step 6: Create `src/frontend/vite.config.ts`**

```typescript
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "dist",
  },
});
```

**Step 7: Create `src/frontend/index.html`**

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link
      href="https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&display=swap"
      rel="stylesheet"
    />
    <title>snipt</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

**Step 8: Install frontend dependencies**

Run:
```bash
cd src/frontend && npm install
```

**Step 9: Clean up temp project**

Run:
```bash
rm -rf /tmp/snipt-gui
```

**Step 10: Commit**

```bash
git add src/cmd/snipt-gui/wails.json src/frontend/
git commit -m "feat(gui): scaffold Wails project structure with React-TS frontend"
```

---

## Task 3: Go Backend — `internal/gui/app.go`

**Files:**
- Create: `src/internal/gui/app.go`

**Step 1: Write `app.go`**

```go
package gui

import (
	"context"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
)

// App exposes snippet operations to the Wails frontend.
type App struct {
	ctx   context.Context
	store *db.Store
}

// NewApp creates a new App backed by the given store.
func NewApp(store *db.Store) *App {
	return &App{store: store}
}

// Startup is called by Wails at application startup.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ListSnippets returns snippets filtered and sorted by opts.
func (a *App) ListSnippets(opts db.ListOpts) ([]model.Snippet, error) {
	return a.store.List(opts)
}

// SearchSnippets performs FTS5 search and returns results with scores.
func (a *App) SearchSnippets(query string) ([]model.SearchResult, error) {
	return a.store.Search(query)
}

// GetSnippet retrieves a snippet by exact ID.
func (a *App) GetSnippet(id string) (*model.Snippet, error) {
	return a.store.Get(id)
}

// CreateSnippet generates an ID and inserts a new snippet.
func (a *App) CreateSnippet(s model.Snippet) error {
	s.ID = model.NewID()
	return a.store.Create(&s)
}

// UpdateSnippet modifies an existing snippet's fields (not tags).
func (a *App) UpdateSnippet(s model.Snippet) error {
	return a.store.Update(&s)
}

// UpdateSnippetTags replaces a snippet's tags with the given set.
// Diffs current vs desired and calls AddTags/RemoveTags.
func (a *App) UpdateSnippetTags(id string, tags []string) error {
	current, err := a.store.Get(id)
	if err != nil {
		return err
	}

	// Build sets for diffing.
	oldSet := make(map[string]bool, len(current.Tags))
	for _, t := range current.Tags {
		oldSet[t] = true
	}
	newSet := make(map[string]bool, len(tags))
	for _, t := range tags {
		newSet[t] = true
	}

	// Tags to add: in new but not old.
	var toAdd []string
	for _, t := range tags {
		if !oldSet[t] {
			toAdd = append(toAdd, t)
		}
	}

	// Tags to remove: in old but not new.
	var toRemove []string
	for _, t := range current.Tags {
		if !newSet[t] {
			toRemove = append(toRemove, t)
		}
	}

	if len(toRemove) > 0 {
		if err := a.store.RemoveTags(id, toRemove); err != nil {
			return err
		}
	}
	if len(toAdd) > 0 {
		if err := a.store.AddTags(id, toAdd); err != nil {
			return err
		}
	}

	return nil
}

// DeleteSnippet removes a snippet by ID.
func (a *App) DeleteSnippet(id string) error {
	return a.store.Delete(id)
}

// SetPinned sets the pinned state of a snippet.
func (a *App) SetPinned(id string, pinned bool) error {
	return a.store.SetPinned(id, pinned)
}

// IncrementUseCount bumps the use count of a snippet by 1.
func (a *App) IncrementUseCount(id string) error {
	return a.store.IncrementUseCount(id)
}

// GetStats returns collection overview data.
func (a *App) GetStats() (*model.Stats, error) {
	return a.store.Stats()
}
```

**Step 2: Verify it compiles**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt && go build ./src/internal/gui/
```

Expected: Clean build, no errors. (Wails dependency not needed yet — this file only imports db and model.)

**Step 3: Commit**

```bash
git add src/internal/gui/app.go
git commit -m "feat(gui): add backend bindings for Wails frontend"
```

---

## Task 4: Wails Entry Point — `cmd/snipt-gui/main.go`

**Files:**
- Create: `src/cmd/snipt-gui/main.go`

**Step 1: Add Wails v2 dependency**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt && go get github.com/wailsapp/wails/v2
```

**Step 2: Write `main.go`**

```go
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/gui"
)

//go:embed all:../../frontend/dist
var assets embed.FS

func main() {
	dbPath := config.DBPath("")
	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer store.Close()

	app := gui.NewApp(store)

	err = wails.Run(&options.App{
		Title:     "snipt",
		Width:     1100,
		Height:    700,
		MinWidth:  800,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 13, G: 13, B: 20, A: 255},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
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
	})
	if err != nil {
		log.Fatalf("wails: %v", err)
	}
}
```

Note: The `embed` path `all:../../frontend/dist` is relative to this file's location (`src/cmd/snipt-gui/`). The `all:` prefix includes dotfiles. Verify this path resolves correctly at build time — `wails build` handles the frontend build first, then the Go embed.

**Step 3: Run `go mod tidy`**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt && go mod tidy
```

**Step 4: Verify Go compilation**

Run:
```bash
go vet ./src/cmd/snipt-gui/
```

Expected: This may fail because `frontend/dist` doesn't exist yet (embed will complain). That's expected — full build comes after frontend setup. The Go code itself should be syntactically valid.

**Step 5: Commit**

```bash
git add src/cmd/snipt-gui/main.go go.mod go.sum
git commit -m "feat(gui): add Wails entry point with macOS titlebar config"
```

---

## Task 5: Frontend Foundation — React Entry, Global Styles, Colors

**Files:**
- Create: `src/frontend/src/main.tsx`
- Create: `src/frontend/src/styles/global.css`
- Create: `src/frontend/src/styles/colors.ts`
- Create: `src/frontend/src/App.tsx` (placeholder)

**Step 1: Create `src/frontend/src/styles/colors.ts`**

```typescript
export const C = {
  bg: "#0d0d14",
  bgCard: "#151521",
  bgTerminal: "#1e1e2e",
  bgSurface: "#242435",
  border: "#2a2a3c",
  borderSubtle: "#1f1f30",
  text: "#cdd6f4",
  textSub: "#a6adc8",
  textDim: "#6c7086",
  textMuted: "#45475a",
  pink: "#f5c2e7",
  mauve: "#cba6f7",
  peach: "#fab387",
  green: "#a6e3a1",
  teal: "#94e2d5",
  blue: "#89b4fa",
  yellow: "#f9e2af",
  red: "#f38ba8",
  lavender: "#b4befe",
  sky: "#89dceb",
} as const;

export const MONO = '"Berkeley Mono", "JetBrains Mono", "Fira Code", monospace';
export const BODY = '"DM Sans", "Helvetica Neue", sans-serif';
```

**Step 2: Create `src/frontend/src/styles/global.css`**

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

*,
*::before,
*::after {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html,
body,
#root {
  height: 100%;
  overflow: hidden;
  background: var(--bg);
  color: var(--text);
  font-family: "DM Sans", "Helvetica Neue", sans-serif;
  font-size: 14px;
  -webkit-font-smoothing: antialiased;
}

/* Disable text selection on non-content areas */
.no-select {
  user-select: none;
  -webkit-user-select: none;
}

/* Scrollbar styling */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--text-muted);
}

/* Match highlighting */
mark {
  background: none;
  color: var(--pink);
  font-weight: 600;
}
```

**Step 3: Create `src/frontend/src/main.tsx`**

```tsx
import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./styles/global.css";

const root = createRoot(document.getElementById("root")!);
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

**Step 4: Create placeholder `src/frontend/src/App.tsx`**

```tsx
export default function App() {
  return (
    <div style={{ padding: "60px 24px 24px", fontFamily: "DM Sans, sans-serif" }}>
      <h1 style={{ color: "var(--text)", fontSize: 24 }}>snipt</h1>
      <p style={{ color: "var(--text-dim)", marginTop: 8 }}>GUI loading...</p>
    </div>
  );
}
```

**Step 5: First Wails build test**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt-gui && wails build
```

Expected: Wails installs frontend deps, runs Vite build, compiles Go binary. If successful, produces `build/bin/snipt-gui` (or `build/bin/snipt-gui.app` on macOS).

**Step 6: Launch and verify**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt-gui && wails dev
```

Expected: Dark window (#0d0d14 background) with "snipt" heading and "GUI loading..." text. macOS traffic lights visible. Window is 1100x700.

**Step 7: Commit**

```bash
git add src/frontend/src/
git commit -m "feat(gui): add React entry point with Catppuccin global styles"
```

---

## Task 6: State Management — Context + Reducer

**Files:**
- Create: `src/frontend/src/state/types.ts`
- Create: `src/frontend/src/state/context.tsx`

**Step 1: Create `src/frontend/src/state/types.ts`**

Type definitions mirroring the Wails-generated bindings. These may need adjustment once Wails generates its bindings, but the shape matches the Go structs exactly.

```typescript
export interface Snippet {
  ID: string;
  Title: string;
  Content: string;
  Language: string;
  Description: string;
  Source: string;
  Pinned: boolean;
  UseCount: number;
  Tags: string[];
  CreatedAt: string; // RFC3339
  UpdatedAt: string; // RFC3339
}

export interface SearchResult {
  Snippet: Snippet;
  Score: number;
  TitleIndices: number[] | null;
}

export interface ListOpts {
  Language?: string;
  Tag?: string;
  Pinned?: boolean;
  Sort?: string;
}

export interface Stats {
  TotalSnippets: number;
  TotalTags: number;
  Languages: Record<string, number>;
  MostUsed: Snippet | null;
  RecentlyAdded: Snippet[];
}

export interface AppState {
  snippets: Snippet[];
  searchResults: SearchResult[] | null;
  selectedId: string | null;
  editMode: boolean;
  searchQuery: string;
  createMode: boolean;
}

export type AppAction =
  | { type: "SET_SNIPPETS"; snippets: Snippet[] }
  | { type: "SET_SEARCH_RESULTS"; results: SearchResult[] | null }
  | { type: "SET_SELECTED"; id: string | null }
  | { type: "SET_EDIT_MODE"; editing: boolean }
  | { type: "SET_SEARCH_QUERY"; query: string }
  | { type: "SET_CREATE_MODE"; creating: boolean }
  | { type: "CLEAR_SEARCH" };
```

**Step 2: Create `src/frontend/src/state/context.tsx`**

```tsx
import { createContext, useContext, useReducer, type Dispatch, type ReactNode } from "react";
import type { AppState, AppAction } from "./types";

const initialState: AppState = {
  snippets: [],
  searchResults: null,
  selectedId: null,
  editMode: false,
  searchQuery: "",
  createMode: false,
};

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "SET_SNIPPETS":
      return { ...state, snippets: action.snippets };
    case "SET_SEARCH_RESULTS":
      return { ...state, searchResults: action.results };
    case "SET_SELECTED":
      return { ...state, selectedId: action.id, editMode: false, createMode: false };
    case "SET_EDIT_MODE":
      return { ...state, editMode: action.editing };
    case "SET_SEARCH_QUERY":
      return { ...state, searchQuery: action.query };
    case "SET_CREATE_MODE":
      return { ...state, createMode: action.creating, editMode: true, selectedId: null };
    case "CLEAR_SEARCH":
      return { ...state, searchQuery: "", searchResults: null };
    default:
      return state;
  }
}

const AppStateContext = createContext<AppState>(initialState);
const AppDispatchContext = createContext<Dispatch<AppAction>>(() => {});

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(appReducer, initialState);
  return (
    <AppStateContext.Provider value={state}>
      <AppDispatchContext.Provider value={dispatch}>
        {children}
      </AppDispatchContext.Provider>
    </AppStateContext.Provider>
  );
}

export function useAppState() {
  return useContext(AppStateContext);
}

export function useAppDispatch() {
  return useContext(AppDispatchContext);
}
```

**Step 3: Wire provider into App.tsx**

Update `src/frontend/src/App.tsx`:

```tsx
import { AppProvider } from "./state/context";

export default function App() {
  return (
    <AppProvider>
      <div style={{ padding: "60px 24px 24px", fontFamily: "DM Sans, sans-serif" }}>
        <h1 style={{ color: "var(--text)", fontSize: 24 }}>snipt</h1>
        <p style={{ color: "var(--text-dim)", marginTop: 8 }}>GUI loading...</p>
      </div>
    </AppProvider>
  );
}
```

**Step 4: Verify build still works**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt-gui && wails build
```

Expected: Clean build.

**Step 5: Commit**

```bash
git add src/frontend/src/state/ src/frontend/src/App.tsx
git commit -m "feat(gui): add state management with Context + useReducer"
```

---

## Task 7: Hooks — useDebounce and useKeyboardShortcuts

**Files:**
- Create: `src/frontend/src/hooks/useDebounce.ts`
- Create: `src/frontend/src/hooks/useKeyboardShortcuts.ts`

**Step 1: Create `src/frontend/src/hooks/useDebounce.ts`**

```typescript
import { useEffect, useState } from "react";

export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}
```

**Step 2: Create `src/frontend/src/hooks/useKeyboardShortcuts.ts`**

```typescript
import { useEffect } from "react";
import type { AppState } from "../state/types";

interface ShortcutHandlers {
  onNewSnippet: () => void;
  onFocusSearch: () => void;
  onCopyContent: () => void;
  onSave: () => void;
  onDelete: () => void;
  onNavigateUp: () => void;
  onNavigateDown: () => void;
  onEscape: () => void;
  onTogglePin: () => void;
}

export function useKeyboardShortcuts(state: AppState, handlers: ShortcutHandlers) {
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const meta = e.metaKey || e.ctrlKey;

      // Cmd+N: New snippet
      if (meta && e.key === "n") {
        e.preventDefault();
        handlers.onNewSnippet();
        return;
      }

      // Cmd+F: Focus search
      if (meta && e.key === "f") {
        e.preventDefault();
        handlers.onFocusSearch();
        return;
      }

      // Cmd+S: Save (when editing)
      if (meta && e.key === "s") {
        e.preventDefault();
        if (state.editMode) {
          handlers.onSave();
        }
        return;
      }

      // Cmd+Backspace: Delete (when not editing)
      if (meta && e.key === "Backspace") {
        e.preventDefault();
        if (!state.editMode && state.selectedId) {
          handlers.onDelete();
        }
        return;
      }

      // Cmd+P: Toggle pin
      if (meta && e.key === "p") {
        e.preventDefault();
        if (state.selectedId) {
          handlers.onTogglePin();
        }
        return;
      }

      // Cmd+C: Copy content (only when not in editor with selection)
      if (meta && e.key === "c" && !state.editMode) {
        const selection = window.getSelection();
        if (!selection || selection.isCollapsed) {
          e.preventDefault();
          handlers.onCopyContent();
          return;
        }
      }

      // Arrow keys: Navigate list (when not editing)
      if (!state.editMode) {
        if (e.key === "ArrowUp") {
          e.preventDefault();
          handlers.onNavigateUp();
          return;
        }
        if (e.key === "ArrowDown") {
          e.preventDefault();
          handlers.onNavigateDown();
          return;
        }
      }

      // Escape: Cancel edit / clear search
      if (e.key === "Escape") {
        e.preventDefault();
        handlers.onEscape();
        return;
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [state, handlers]);
}
```

**Step 3: Commit**

```bash
git add src/frontend/src/hooks/
git commit -m "feat(gui): add useDebounce and useKeyboardShortcuts hooks"
```

---

## Task 8: CodeMirror 6 — Theme and Language Loading

**Files:**
- Create: `src/frontend/src/editor/catppuccin-theme.ts`
- Create: `src/frontend/src/editor/languages.ts`

**Step 1: Install CodeMirror dependencies**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt/src/frontend && npm install @codemirror/view @codemirror/state @codemirror/language @codemirror/lang-go @codemirror/lang-javascript @codemirror/lang-python @codemirror/lang-sql @codemirror/lang-html @codemirror/lang-css @codemirror/lang-json @codemirror/lang-markdown @lezer/highlight
```

**Step 2: Create `src/frontend/src/editor/catppuccin-theme.ts`**

```typescript
import { EditorView } from "@codemirror/view";
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags } from "@lezer/highlight";

const catppuccinMochaTheme = EditorView.theme(
  {
    "&": {
      backgroundColor: "#1e1e2e",
      color: "#cdd6f4",
    },
    ".cm-content": {
      fontFamily: '"Berkeley Mono", "JetBrains Mono", "Fira Code", monospace',
      fontSize: "13px",
      caretColor: "#cba6f7",
      padding: "12px 0",
    },
    ".cm-gutters": {
      backgroundColor: "#242435",
      color: "#45475a",
      border: "none",
      paddingLeft: "8px",
    },
    ".cm-activeLineGutter": {
      backgroundColor: "#2a2a3c",
    },
    ".cm-activeLine": {
      backgroundColor: "#2a2a3c40",
    },
    "&.cm-focused .cm-selectionBackground, .cm-selectionBackground": {
      backgroundColor: "#45475a !important",
    },
    ".cm-cursor, .cm-dropCursor": {
      borderLeftColor: "#cba6f7",
    },
    ".cm-lineNumbers .cm-gutterElement": {
      padding: "0 8px 0 0",
      minWidth: "32px",
    },
    "&.cm-focused": {
      outline: "none",
    },
  },
  { dark: true }
);

const catppuccinMochaHighlight = HighlightStyle.define([
  { tag: tags.keyword, color: "#cba6f7", fontWeight: "600" },
  { tag: tags.string, color: "#a6e3a1" },
  { tag: tags.comment, color: "#45475a", fontStyle: "italic" },
  { tag: tags.function(tags.variableName), color: "#89b4fa" },
  { tag: tags.number, color: "#fab387" },
  { tag: tags.typeName, color: "#f9e2af" },
  { tag: tags.bool, color: "#fab387" },
  { tag: tags.operator, color: "#89dceb" },
  { tag: tags.propertyName, color: "#89b4fa" },
  { tag: tags.punctuation, color: "#6c7086" },
  { tag: tags.className, color: "#f9e2af" },
  { tag: tags.definition(tags.variableName), color: "#cdd6f4" },
  { tag: tags.variableName, color: "#cdd6f4" },
  { tag: tags.tagName, color: "#f38ba8" },
  { tag: tags.attributeName, color: "#f9e2af" },
  { tag: tags.attributeValue, color: "#a6e3a1" },
]);

export const catppuccinMocha = [
  catppuccinMochaTheme,
  syntaxHighlighting(catppuccinMochaHighlight),
];
```

**Step 3: Create `src/frontend/src/editor/languages.ts`**

```typescript
import type { LanguageSupport } from "@codemirror/language";

type LanguageLoader = () => Promise<LanguageSupport>;

const languageMap: Record<string, LanguageLoader> = {
  go: () => import("@codemirror/lang-go").then((m) => m.go()),
  javascript: () => import("@codemirror/lang-javascript").then((m) => m.javascript()),
  js: () => import("@codemirror/lang-javascript").then((m) => m.javascript()),
  typescript: () =>
    import("@codemirror/lang-javascript").then((m) => m.javascript({ typescript: true })),
  ts: () =>
    import("@codemirror/lang-javascript").then((m) => m.javascript({ typescript: true })),
  jsx: () =>
    import("@codemirror/lang-javascript").then((m) => m.javascript({ jsx: true })),
  tsx: () =>
    import("@codemirror/lang-javascript").then((m) =>
      m.javascript({ jsx: true, typescript: true })
    ),
  python: () => import("@codemirror/lang-python").then((m) => m.python()),
  py: () => import("@codemirror/lang-python").then((m) => m.python()),
  sql: () => import("@codemirror/lang-sql").then((m) => m.sql()),
  html: () => import("@codemirror/lang-html").then((m) => m.html()),
  css: () => import("@codemirror/lang-css").then((m) => m.css()),
  json: () => import("@codemirror/lang-json").then((m) => m.json()),
  markdown: () => import("@codemirror/lang-markdown").then((m) => m.markdown()),
  md: () => import("@codemirror/lang-markdown").then((m) => m.markdown()),
};

export async function loadLanguage(lang: string): Promise<LanguageSupport | null> {
  const loader = languageMap[lang.toLowerCase()];
  if (!loader) return null;
  return loader();
}

export function isLanguageSupported(lang: string): boolean {
  return lang.toLowerCase() in languageMap;
}
```

**Step 4: Verify build**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt-gui && wails build
```

Expected: Clean build with CodeMirror packages resolved.

**Step 5: Commit**

```bash
git add src/frontend/src/editor/ src/frontend/package.json src/frontend/package-lock.json
git commit -m "feat(gui): add CodeMirror 6 Catppuccin theme and language loaders"
```

---

## Task 9: Sidebar Components — SearchBar, SnippetRow, SnippetList

**Files:**
- Create: `src/frontend/src/components/SearchBar.tsx`
- Create: `src/frontend/src/components/SnippetRow.tsx`
- Create: `src/frontend/src/components/SnippetList.tsx`
- Create: `src/frontend/src/components/NewSnippetButton.tsx`
- Create: `src/frontend/src/components/Sidebar.tsx`

**Step 1: Create `SearchBar.tsx`**

```tsx
import { useRef, useEffect, forwardRef, useImperativeHandle } from "react";
import { C, BODY } from "../styles/colors";

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
}

export interface SearchBarHandle {
  focus: () => void;
}

export const SearchBar = forwardRef<SearchBarHandle, SearchBarProps>(
  function SearchBar({ value, onChange }, ref) {
    const inputRef = useRef<HTMLInputElement>(null);

    useImperativeHandle(ref, () => ({
      focus: () => inputRef.current?.focus(),
    }));

    return (
      <div
        className="no-select"
        style={{
          padding: "12px 16px",
          borderBottom: `1px solid ${C.border}`,
          display: "flex",
          alignItems: "center",
          gap: 12,
        }}
      >
        <span
          style={{
            background: `linear-gradient(135deg, ${C.pink}, ${C.mauve})`,
            color: C.bg,
            fontFamily: BODY,
            fontWeight: 700,
            fontSize: 11,
            padding: "3px 10px",
            borderRadius: 12,
            letterSpacing: "0.5px",
            flexShrink: 0,
          }}
        >
          SNIPT
        </span>
        <input
          ref={inputRef}
          type="text"
          placeholder="Search snippets..."
          value={value}
          onChange={(e) => onChange(e.target.value)}
          style={{
            flex: 1,
            background: "transparent",
            border: "none",
            outline: "none",
            color: C.text,
            fontFamily: BODY,
            fontSize: 13,
          }}
        />
      </div>
    );
  }
);
```

**Step 2: Create `SnippetRow.tsx`**

```tsx
import { C, BODY, MONO } from "../styles/colors";
import type { Snippet } from "../state/types";

interface SnippetRowProps {
  snippet: Snippet;
  selected: boolean;
  titleIndices?: number[] | null;
  onClick: () => void;
}

// Map common languages to badge colors
const langColors: Record<string, string> = {
  go: C.blue,
  javascript: C.yellow,
  js: C.yellow,
  typescript: C.blue,
  ts: C.blue,
  python: C.green,
  py: C.green,
  bash: C.peach,
  sh: C.peach,
  sql: C.mauve,
  html: C.red,
  css: C.sky,
  json: C.peach,
  markdown: C.teal,
  md: C.teal,
  rust: C.peach,
  ruby: C.red,
  yaml: C.lavender,
  toml: C.lavender,
};

function highlightTitle(title: string, indices: number[] | null | undefined) {
  if (!indices || indices.length === 0) return title;

  const indexSet = new Set(indices);
  const parts: JSX.Element[] = [];
  let current = "";
  let inMatch = false;

  for (let i = 0; i < title.length; i++) {
    const isMatch = indexSet.has(i);
    if (isMatch !== inMatch) {
      if (current) {
        parts.push(
          inMatch ? <mark key={i}>{current}</mark> : <span key={i}>{current}</span>
        );
      }
      current = "";
      inMatch = isMatch;
    }
    current += title[i];
  }
  if (current) {
    parts.push(
      inMatch ? <mark key="last">{current}</mark> : <span key="last">{current}</span>
    );
  }

  return <>{parts}</>;
}

export function SnippetRow({ snippet, selected, titleIndices, onClick }: SnippetRowProps) {
  const badgeColor = langColors[snippet.Language.toLowerCase()] ?? C.textDim;

  return (
    <div
      onClick={onClick}
      style={{
        padding: "10px 16px",
        cursor: "pointer",
        borderLeft: selected ? `3px solid ${C.pink}` : "3px solid transparent",
        background: selected ? C.bgSurface : "transparent",
        transition: "background 0.1s",
      }}
      onMouseEnter={(e) => {
        if (!selected) e.currentTarget.style.background = C.bgCard;
      }}
      onMouseLeave={(e) => {
        if (!selected) e.currentTarget.style.background = "transparent";
      }}
    >
      {/* Line 1: pin + title */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          fontFamily: BODY,
          fontSize: 13,
          fontWeight: 500,
          color: C.text,
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
        }}
      >
        {snippet.Pinned && (
          <span style={{ color: C.yellow, fontSize: 10, flexShrink: 0 }}>●</span>
        )}
        <span style={{ overflow: "hidden", textOverflow: "ellipsis" }}>
          {highlightTitle(snippet.Title, titleIndices)}
        </span>
      </div>

      {/* Line 2: lang badge + tags */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          marginTop: 4,
          fontSize: 11,
        }}
      >
        {snippet.Language && (
          <span
            style={{
              color: badgeColor,
              fontFamily: MONO,
              fontSize: 10,
              padding: "1px 6px",
              borderRadius: 4,
              background: `${badgeColor}15`,
              flexShrink: 0,
            }}
          >
            {snippet.Language}
          </span>
        )}
        {snippet.Tags?.map((tag) => (
          <span
            key={tag}
            style={{
              color: C.textDim,
              fontFamily: BODY,
              fontSize: 10,
            }}
          >
            #{tag}
          </span>
        ))}
      </div>
    </div>
  );
}
```

**Step 3: Create `SnippetList.tsx`**

```tsx
import { SnippetRow } from "./SnippetRow";
import type { Snippet, SearchResult } from "../state/types";

interface SnippetListProps {
  snippets: Snippet[];
  searchResults: SearchResult[] | null;
  selectedId: string | null;
  onSelect: (id: string) => void;
}

export function SnippetList({ snippets, searchResults, selectedId, onSelect }: SnippetListProps) {
  // When searching, use search results (which have TitleIndices). Otherwise use snippets.
  if (searchResults) {
    return (
      <div style={{ flex: 1, overflowY: "auto" }}>
        {searchResults.map((result) => (
          <SnippetRow
            key={result.Snippet.ID}
            snippet={result.Snippet}
            selected={result.Snippet.ID === selectedId}
            titleIndices={result.TitleIndices}
            onClick={() => onSelect(result.Snippet.ID)}
          />
        ))}
        {searchResults.length === 0 && (
          <div
            style={{
              padding: "24px 16px",
              color: "var(--text-dim)",
              fontSize: 13,
              textAlign: "center",
            }}
          >
            No results found
          </div>
        )}
      </div>
    );
  }

  return (
    <div style={{ flex: 1, overflowY: "auto" }}>
      {snippets.map((snippet) => (
        <SnippetRow
          key={snippet.ID}
          snippet={snippet}
          selected={snippet.ID === selectedId}
          onClick={() => onSelect(snippet.ID)}
        />
      ))}
      {snippets.length === 0 && (
        <div
          style={{
            padding: "24px 16px",
            color: "var(--text-dim)",
            fontSize: 13,
            textAlign: "center",
          }}
        >
          No snippets yet
        </div>
      )}
    </div>
  );
}
```

**Step 4: Create `NewSnippetButton.tsx`**

```tsx
import { C, BODY } from "../styles/colors";

interface NewSnippetButtonProps {
  onClick: () => void;
}

export function NewSnippetButton({ onClick }: NewSnippetButtonProps) {
  return (
    <button
      onClick={onClick}
      style={{
        width: "100%",
        padding: "12px 16px",
        background: C.bgCard,
        border: "none",
        borderTop: `1px solid ${C.border}`,
        color: C.textDim,
        fontFamily: BODY,
        fontSize: 13,
        cursor: "pointer",
        textAlign: "left",
        transition: "color 0.1s",
      }}
      onMouseEnter={(e) => (e.currentTarget.style.color = C.text)}
      onMouseLeave={(e) => (e.currentTarget.style.color = C.textDim)}
    >
      + New snippet
    </button>
  );
}
```

**Step 5: Create `Sidebar.tsx`**

```tsx
import { useRef } from "react";
import { SearchBar, type SearchBarHandle } from "./SearchBar";
import { SnippetList } from "./SnippetList";
import { NewSnippetButton } from "./NewSnippetButton";
import { C } from "../styles/colors";
import type { Snippet, SearchResult } from "../state/types";

interface SidebarProps {
  snippets: Snippet[];
  searchResults: SearchResult[] | null;
  searchQuery: string;
  selectedId: string | null;
  onSearchChange: (query: string) => void;
  onSelect: (id: string) => void;
  onNewSnippet: () => void;
  searchBarRef?: React.Ref<SearchBarHandle>;
}

export function Sidebar({
  snippets,
  searchResults,
  searchQuery,
  selectedId,
  onSearchChange,
  onSelect,
  onNewSnippet,
  searchBarRef,
}: SidebarProps) {
  return (
    <div
      style={{
        width: 300,
        minWidth: 300,
        height: "100%",
        display: "flex",
        flexDirection: "column",
        borderRight: `1px solid ${C.border}`,
        background: C.bg,
        paddingTop: 40, // macOS title bar space
      }}
    >
      <SearchBar ref={searchBarRef} value={searchQuery} onChange={onSearchChange} />
      <SnippetList
        snippets={snippets}
        searchResults={searchResults}
        selectedId={selectedId}
        onSelect={onSelect}
      />
      <NewSnippetButton onClick={onNewSnippet} />
    </div>
  );
}
```

**Step 6: Commit**

```bash
git add src/frontend/src/components/SearchBar.tsx src/frontend/src/components/SnippetRow.tsx src/frontend/src/components/SnippetList.tsx src/frontend/src/components/NewSnippetButton.tsx src/frontend/src/components/Sidebar.tsx
git commit -m "feat(gui): add Sidebar with search, snippet list, and new button"
```

---

## Task 10: Detail Pane Components — Header, CodeEditor, Metadata, Actions

**Files:**
- Create: `src/frontend/src/components/CodeEditor.tsx`
- Create: `src/frontend/src/components/DetailHeader.tsx`
- Create: `src/frontend/src/components/MetadataFooter.tsx`
- Create: `src/frontend/src/components/ActionBar.tsx`
- Create: `src/frontend/src/components/ConfirmDialog.tsx`
- Create: `src/frontend/src/components/DetailPane.tsx`

**Step 1: Create `CodeEditor.tsx`**

CodeMirror 6 wrapper with compartment-based read-only toggle.

```tsx
import { useEffect, useRef, useCallback } from "react";
import { EditorState, Compartment } from "@codemirror/state";
import { EditorView, lineNumbers } from "@codemirror/view";
import { catppuccinMocha } from "../editor/catppuccin-theme";
import { loadLanguage } from "../editor/languages";

const readOnlyCompartment = new Compartment();
const languageCompartment = new Compartment();

interface CodeEditorProps {
  content: string;
  language: string;
  readOnly: boolean;
  onContentChange?: (content: string) => void;
  onDoubleClick?: () => void;
}

export interface CodeEditorHandle {
  getContent: () => string;
}

export function CodeEditor({
  content,
  language,
  readOnly,
  onContentChange,
  onDoubleClick,
}: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const contentRef = useRef(content);

  // Initialize editor
  useEffect(() => {
    if (!containerRef.current) return;

    const state = EditorState.create({
      doc: content,
      extensions: [
        lineNumbers(),
        ...catppuccinMocha,
        readOnlyCompartment.of(EditorState.readOnly.of(readOnly)),
        languageCompartment.of([]),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            contentRef.current = update.state.doc.toString();
            onContentChange?.(contentRef.current);
          }
        }),
      ],
    });

    const view = new EditorView({
      state,
      parent: containerRef.current,
    });

    viewRef.current = view;

    // Load language async
    loadLanguage(language).then((lang) => {
      if (lang && viewRef.current) {
        viewRef.current.dispatch({
          effects: languageCompartment.reconfigure(lang),
        });
      }
    });

    return () => {
      view.destroy();
      viewRef.current = null;
    };
  }, [content, language]); // Remount when content or language changes

  // Toggle read-only via compartment reconfiguration
  useEffect(() => {
    if (viewRef.current) {
      viewRef.current.dispatch({
        effects: readOnlyCompartment.reconfigure(EditorState.readOnly.of(readOnly)),
      });
    }
  }, [readOnly]);

  return (
    <div
      ref={containerRef}
      onDoubleClick={onDoubleClick}
      style={{
        flex: 1,
        overflow: "auto",
        borderRadius: 8,
        border: `1px solid var(--border)`,
      }}
    />
  );
}
```

Note: The `content` and `language` deps in the first useEffect cause a full remount when switching snippets, which is correct — you want a fresh editor state for each snippet. The read-only toggle within the same snippet uses compartment reconfiguration (no remount).

**Step 2: Create `DetailHeader.tsx`**

```tsx
import { useState, useRef, useEffect } from "react";
import { C, BODY, MONO } from "../styles/colors";

interface DetailHeaderProps {
  title: string;
  language: string;
  editMode: boolean;
  onTitleChange: (title: string) => void;
}

export function DetailHeader({ title, language, editMode, onTitleChange }: DetailHeaderProps) {
  const [editing, setEditing] = useState(false);
  const [localTitle, setLocalTitle] = useState(title);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setLocalTitle(title);
    setEditing(false);
  }, [title]);

  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editing]);

  function commitTitle() {
    setEditing(false);
    if (localTitle.trim() && localTitle !== title) {
      onTitleChange(localTitle.trim());
    } else {
      setLocalTitle(title);
    }
  }

  const langColor = C.mauve; // default badge color

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "0 0 16px",
        borderBottom: `1px solid ${C.border}`,
        marginBottom: 16,
      }}
    >
      {editing || editMode ? (
        <input
          ref={inputRef}
          value={localTitle}
          onChange={(e) => setLocalTitle(e.target.value)}
          onBlur={commitTitle}
          onKeyDown={(e) => {
            if (e.key === "Enter") commitTitle();
            if (e.key === "Escape") {
              setLocalTitle(title);
              setEditing(false);
            }
          }}
          style={{
            flex: 1,
            background: "transparent",
            border: "none",
            outline: "none",
            color: C.text,
            fontFamily: BODY,
            fontSize: 20,
            fontWeight: 700,
          }}
        />
      ) : (
        <h2
          onClick={() => setEditing(true)}
          style={{
            flex: 1,
            color: C.text,
            fontFamily: BODY,
            fontSize: 20,
            fontWeight: 700,
            cursor: "pointer",
          }}
        >
          {title}
        </h2>
      )}

      {language && (
        <span
          style={{
            color: langColor,
            fontFamily: MONO,
            fontSize: 12,
            padding: "3px 10px",
            borderRadius: 8,
            background: `${langColor}15`,
            marginLeft: 12,
            flexShrink: 0,
          }}
        >
          {language}
        </span>
      )}
    </div>
  );
}
```

**Step 3: Create `MetadataFooter.tsx`**

```tsx
import { useState } from "react";
import { C, BODY, MONO } from "../styles/colors";

interface MetadataFooterProps {
  tags: string[];
  pinned: boolean;
  useCount: number;
  createdAt: string;
  updatedAt: string;
  editMode: boolean;
  onTagsChange: (tags: string[]) => void;
  onTogglePin: () => void;
}

export function MetadataFooter({
  tags,
  pinned,
  useCount,
  createdAt,
  updatedAt,
  editMode,
  onTagsChange,
  onTogglePin,
}: MetadataFooterProps) {
  const [tagInput, setTagInput] = useState("");

  function addTag() {
    const tag = tagInput.trim().toLowerCase().replace(/^#/, "");
    if (tag && !tags.includes(tag)) {
      onTagsChange([...tags, tag]);
    }
    setTagInput("");
  }

  function removeTag(tag: string) {
    onTagsChange(tags.filter((t) => t !== tag));
  }

  function formatDate(dateStr: string): string {
    try {
      return new Date(dateStr).toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    } catch {
      return dateStr;
    }
  }

  return (
    <div
      style={{
        padding: "16px 0 0",
        borderTop: `1px solid ${C.border}`,
        marginTop: 16,
        display: "flex",
        flexDirection: "column",
        gap: 12,
      }}
    >
      {/* Tags */}
      <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap" }}>
        <span style={{ color: C.textDim, fontFamily: BODY, fontSize: 12 }}>Tags:</span>
        {tags.map((tag) => (
          <span
            key={tag}
            style={{
              color: C.textSub,
              fontFamily: MONO,
              fontSize: 11,
              padding: "2px 8px",
              borderRadius: 4,
              background: C.bgSurface,
              display: "flex",
              alignItems: "center",
              gap: 4,
            }}
          >
            #{tag}
            {editMode && (
              <span
                onClick={() => removeTag(tag)}
                style={{ cursor: "pointer", color: C.textMuted, marginLeft: 2 }}
              >
                x
              </span>
            )}
          </span>
        ))}
        {editMode && (
          <input
            value={tagInput}
            onChange={(e) => setTagInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") addTag();
            }}
            placeholder="add tag..."
            style={{
              background: "transparent",
              border: `1px solid ${C.border}`,
              borderRadius: 4,
              padding: "2px 8px",
              color: C.text,
              fontFamily: MONO,
              fontSize: 11,
              outline: "none",
              width: 100,
            }}
          />
        )}
      </div>

      {/* Meta row */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 16,
          fontFamily: BODY,
          fontSize: 11,
          color: C.textMuted,
        }}
      >
        <span
          onClick={onTogglePin}
          style={{
            cursor: "pointer",
            color: pinned ? C.yellow : C.textMuted,
          }}
        >
          {pinned ? "● Pinned" : "○ Not pinned"}
        </span>
        <span>Used {useCount}x</span>
        <span>Created {formatDate(createdAt)}</span>
        <span>Updated {formatDate(updatedAt)}</span>
      </div>
    </div>
  );
}
```

**Step 4: Create `ActionBar.tsx`**

```tsx
import { useState } from "react";
import { C, BODY } from "../styles/colors";
import { ClipboardSetText } from "../../wailsjs/runtime/runtime";

interface ActionBarProps {
  content: string;
  editMode: boolean;
  onEdit: () => void;
  onSave: () => void;
  onDelete: () => void;
  onCancelEdit: () => void;
}

export function ActionBar({
  content,
  editMode,
  onEdit,
  onSave,
  onDelete,
  onCancelEdit,
}: ActionBarProps) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await ClipboardSetText(content);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Fallback: navigator.clipboard
      await navigator.clipboard.writeText(content);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    }
  }

  const buttonStyle = (color: string): React.CSSProperties => ({
    padding: "6px 14px",
    borderRadius: 6,
    border: `1px solid ${C.border}`,
    background: C.bgCard,
    color,
    fontFamily: BODY,
    fontSize: 12,
    cursor: "pointer",
    transition: "background 0.1s, border-color 0.1s",
  });

  return (
    <div style={{ display: "flex", gap: 8, paddingTop: 12 }}>
      <button onClick={handleCopy} style={buttonStyle(copied ? C.green : C.textSub)}>
        {copied ? "Copied!" : "Copy"}
      </button>
      {editMode ? (
        <>
          <button onClick={onSave} style={buttonStyle(C.green)}>
            Save
          </button>
          <button onClick={onCancelEdit} style={buttonStyle(C.textDim)}>
            Cancel
          </button>
        </>
      ) : (
        <button onClick={onEdit} style={buttonStyle(C.textSub)}>
          Edit
        </button>
      )}
      <button onClick={onDelete} style={buttonStyle(C.red)}>
        Delete
      </button>
    </div>
  );
}
```

Note: `ClipboardSetText` import path depends on Wails binding generation. The exact path (`../../wailsjs/runtime/runtime`) may vary — check what Wails generates. The `navigator.clipboard` fallback handles dev mode.

**Step 5: Create `ConfirmDialog.tsx`**

```tsx
import { C, BODY } from "../styles/colors";

interface ConfirmDialogProps {
  title: string;
  message: string;
  confirmLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog({
  title,
  message,
  confirmLabel = "Delete",
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "rgba(0, 0, 0, 0.6)",
        zIndex: 100,
      }}
      onClick={onCancel}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: C.bgCard,
          border: `1px solid ${C.border}`,
          borderRadius: 12,
          padding: 24,
          minWidth: 340,
          fontFamily: BODY,
        }}
      >
        <h3 style={{ color: C.text, fontSize: 16, marginBottom: 8 }}>{title}</h3>
        <p style={{ color: C.textDim, fontSize: 13, marginBottom: 20 }}>{message}</p>
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <button
            onClick={onCancel}
            style={{
              padding: "8px 16px",
              borderRadius: 6,
              border: `1px solid ${C.border}`,
              background: C.bgSurface,
              color: C.textSub,
              fontFamily: BODY,
              fontSize: 13,
              cursor: "pointer",
            }}
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            style={{
              padding: "8px 16px",
              borderRadius: 6,
              border: "none",
              background: C.red,
              color: C.bg,
              fontFamily: BODY,
              fontSize: 13,
              fontWeight: 600,
              cursor: "pointer",
            }}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
```

**Step 6: Create `DetailPane.tsx`**

This is the main detail/editor area. It shows either the selected snippet's details, a new snippet form, or an empty state.

```tsx
import { useState, useRef, useEffect } from "react";
import { DetailHeader } from "./DetailHeader";
import { CodeEditor } from "./CodeEditor";
import { MetadataFooter } from "./MetadataFooter";
import { ActionBar } from "./ActionBar";
import { C, BODY } from "../styles/colors";
import type { Snippet } from "../state/types";

interface DetailPaneProps {
  snippet: Snippet | null;
  editMode: boolean;
  createMode: boolean;
  onUpdate: (snippet: Snippet) => void;
  onUpdateTags: (id: string, tags: string[]) => void;
  onDelete: (id: string) => void;
  onTogglePin: (id: string, pinned: boolean) => void;
  onCreate: (snippet: Partial<Snippet>) => void;
  onSetEditMode: (editing: boolean) => void;
}

export function DetailPane({
  snippet,
  editMode,
  createMode,
  onUpdate,
  onUpdateTags,
  onDelete,
  onTogglePin,
  onCreate,
  onSetEditMode,
}: DetailPaneProps) {
  const [editedContent, setEditedContent] = useState("");
  const [editedTitle, setEditedTitle] = useState("");
  const [editedLanguage, setEditedLanguage] = useState("");
  const [editedTags, setEditedTags] = useState<string[]>([]);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  // Sync local state when snippet changes
  useEffect(() => {
    if (snippet) {
      setEditedContent(snippet.Content);
      setEditedTitle(snippet.Title);
      setEditedLanguage(snippet.Language);
      setEditedTags([...snippet.Tags]);
    }
  }, [snippet]);

  // Reset for create mode
  useEffect(() => {
    if (createMode) {
      setEditedContent("");
      setEditedTitle("");
      setEditedLanguage("");
      setEditedTags([]);
    }
  }, [createMode]);

  if (!snippet && !createMode) {
    return (
      <div
        style={{
          flex: 1,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          paddingTop: 40,
          color: C.textMuted,
          fontFamily: BODY,
          fontSize: 14,
        }}
      >
        Select a snippet or create a new one
      </div>
    );
  }

  function handleSave() {
    if (createMode) {
      onCreate({
        Title: editedTitle || "Untitled",
        Content: editedContent,
        Language: editedLanguage,
        Tags: editedTags,
      });
    } else if (snippet) {
      onUpdate({
        ...snippet,
        Title: editedTitle,
        Content: editedContent,
        Language: editedLanguage,
      });
      if (JSON.stringify(editedTags) !== JSON.stringify(snippet.Tags)) {
        onUpdateTags(snippet.ID, editedTags);
      }
    }
    onSetEditMode(false);
  }

  return (
    <div
      style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        paddingTop: 40, // macOS title bar space
        padding: "40px 24px 16px 24px",
        overflow: "hidden",
      }}
    >
      <DetailHeader
        title={createMode ? editedTitle : (snippet?.Title ?? "")}
        language={createMode ? editedLanguage : (snippet?.Language ?? "")}
        editMode={editMode}
        onTitleChange={setEditedTitle}
      />

      {/* Language input in create/edit mode */}
      {editMode && (
        <div style={{ marginBottom: 12 }}>
          <input
            value={editedLanguage}
            onChange={(e) => setEditedLanguage(e.target.value)}
            placeholder="Language (go, python, bash...)"
            style={{
              background: "transparent",
              border: `1px solid ${C.border}`,
              borderRadius: 6,
              padding: "6px 12px",
              color: C.text,
              fontFamily: BODY,
              fontSize: 12,
              outline: "none",
              width: 240,
            }}
          />
        </div>
      )}

      <CodeEditor
        content={createMode ? "" : (snippet?.Content ?? "")}
        language={editMode ? editedLanguage : (snippet?.Language ?? "")}
        readOnly={!editMode}
        onContentChange={setEditedContent}
        onDoubleClick={() => !editMode && onSetEditMode(true)}
      />

      {!createMode && snippet && (
        <MetadataFooter
          tags={editMode ? editedTags : snippet.Tags}
          pinned={snippet.Pinned}
          useCount={snippet.UseCount}
          createdAt={snippet.CreatedAt}
          updatedAt={snippet.UpdatedAt}
          editMode={editMode}
          onTagsChange={setEditedTags}
          onTogglePin={() => onTogglePin(snippet.ID, !snippet.Pinned)}
        />
      )}

      {createMode && editMode && (
        <div style={{ paddingTop: 16 }}>
          <MetadataFooter
            tags={editedTags}
            pinned={false}
            useCount={0}
            createdAt=""
            updatedAt=""
            editMode={true}
            onTagsChange={setEditedTags}
            onTogglePin={() => {}}
          />
        </div>
      )}

      <ActionBar
        content={editMode ? editedContent : (snippet?.Content ?? "")}
        editMode={editMode}
        onEdit={() => onSetEditMode(true)}
        onSave={handleSave}
        onDelete={() => setShowDeleteConfirm(true)}
        onCancelEdit={() => onSetEditMode(false)}
      />

      {showDeleteConfirm && snippet && (
        <ConfirmDialog
          title={`Delete "${snippet.Title}"?`}
          message="This is permanent. The snippet will be removed from your database."
          onConfirm={() => {
            onDelete(snippet.ID);
            setShowDeleteConfirm(false);
          }}
          onCancel={() => setShowDeleteConfirm(false)}
        />
      )}
    </div>
  );
}
```

Note: You'll need to add the ConfirmDialog import at the top. The component is already created in Step 5.

**Step 7: Commit**

```bash
git add src/frontend/src/components/
git commit -m "feat(gui): add DetailPane with CodeEditor, header, metadata, and actions"
```

---

## Task 11: StatusBar Component

**Files:**
- Create: `src/frontend/src/components/StatusBar.tsx`

**Step 1: Create `StatusBar.tsx`**

```tsx
import { C, BODY, MONO } from "../styles/colors";

interface StatusBarProps {
  snippetCount: number;
  searching: boolean;
  editMode: boolean;
}

export function StatusBar({ snippetCount, searching, editMode }: StatusBarProps) {
  return (
    <div
      className="no-select"
      style={{
        height: 32,
        minHeight: 32,
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "0 16px",
        background: C.bgCard,
        borderTop: `1px solid ${C.border}`,
        fontFamily: BODY,
        fontSize: 11,
        color: C.textMuted,
      }}
    >
      <div style={{ display: "flex", gap: 12 }}>
        <span>
          {snippetCount} snippet{snippetCount !== 1 ? "s" : ""}
        </span>
        {searching && <span style={{ color: C.pink }}>searching...</span>}
      </div>
      <div style={{ display: "flex", gap: 16, fontFamily: MONO, fontSize: 10 }}>
        {editMode ? (
          <>
            <span>Cmd+S save</span>
            <span>Esc cancel</span>
          </>
        ) : (
          <>
            <span>Cmd+N new</span>
            <span>Cmd+F search</span>
            <span>Up/Down navigate</span>
            <span>Cmd+P pin</span>
          </>
        )}
      </div>
    </div>
  );
}
```

**Step 2: Commit**

```bash
git add src/frontend/src/components/StatusBar.tsx
git commit -m "feat(gui): add StatusBar with snippet count and shortcut hints"
```

---

## Task 12: Wire Everything Together in App.tsx

**Files:**
- Modify: `src/frontend/src/App.tsx`

This is the main orchestration — connects state, backend calls, keyboard shortcuts, and all components.

**Step 1: Rewrite `App.tsx`**

```tsx
import { useEffect, useRef, useCallback } from "react";
import { AppProvider, useAppState, useAppDispatch } from "./state/context";
import { Sidebar } from "./components/Sidebar";
import { DetailPane } from "./components/DetailPane";
import { StatusBar } from "./components/StatusBar";
import { useDebounce } from "./hooks/useDebounce";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import type { SearchBarHandle } from "./components/SearchBar";
import type { Snippet } from "./state/types";

// These imports come from Wails-generated bindings.
// The exact paths depend on your wails.json wailsjsdir setting.
// Adjust if Wails generates them elsewhere.
import { ListSnippets, SearchSnippets, GetSnippet, CreateSnippet, UpdateSnippet, UpdateSnippetTags, DeleteSnippet, SetPinned, IncrementUseCount } from "../wailsjs/go/gui/App";

function AppContent() {
  const state = useAppState();
  const dispatch = useAppDispatch();
  const searchBarRef = useRef<SearchBarHandle>(null);
  const debouncedQuery = useDebounce(state.searchQuery, 300);

  // Load snippets on mount
  const loadSnippets = useCallback(async () => {
    try {
      const snippets = await ListSnippets({});
      dispatch({ type: "SET_SNIPPETS", snippets: snippets ?? [] });
    } catch (err) {
      console.error("Failed to load snippets:", err);
    }
  }, [dispatch]);

  useEffect(() => {
    loadSnippets();
  }, [loadSnippets]);

  // Debounced search
  useEffect(() => {
    if (!debouncedQuery.trim()) {
      dispatch({ type: "SET_SEARCH_RESULTS", results: null });
      return;
    }
    SearchSnippets(debouncedQuery)
      .then((results) => dispatch({ type: "SET_SEARCH_RESULTS", results: results ?? [] }))
      .catch((err) => console.error("Search failed:", err));
  }, [debouncedQuery, dispatch]);

  // Get selected snippet
  const selectedSnippet =
    state.snippets.find((s) => s.ID === state.selectedId) ?? null;

  // Handlers
  async function handleSelect(id: string) {
    dispatch({ type: "SET_SELECTED", id });
  }

  async function handleUpdate(snippet: Snippet) {
    try {
      await UpdateSnippet(snippet);
      await loadSnippets();
    } catch (err) {
      console.error("Update failed:", err);
    }
  }

  async function handleUpdateTags(id: string, tags: string[]) {
    try {
      await UpdateSnippetTags(id, tags);
      await loadSnippets();
    } catch (err) {
      console.error("Update tags failed:", err);
    }
  }

  async function handleDelete(id: string) {
    try {
      await DeleteSnippet(id);
      dispatch({ type: "SET_SELECTED", id: null });
      await loadSnippets();
    } catch (err) {
      console.error("Delete failed:", err);
    }
  }

  async function handleTogglePin(id: string, pinned: boolean) {
    try {
      await SetPinned(id, pinned);
      await loadSnippets();
    } catch (err) {
      console.error("Pin toggle failed:", err);
    }
  }

  async function handleCreate(partial: Partial<Snippet>) {
    try {
      await CreateSnippet({
        ID: "",
        Title: partial.Title ?? "Untitled",
        Content: partial.Content ?? "",
        Language: partial.Language ?? "",
        Description: partial.Description ?? "",
        Source: "",
        Pinned: false,
        UseCount: 0,
        Tags: partial.Tags ?? [],
        CreatedAt: "",
        UpdatedAt: "",
      });
      dispatch({ type: "SET_CREATE_MODE", creating: false });
      await loadSnippets();
    } catch (err) {
      console.error("Create failed:", err);
    }
  }

  async function handleCopyContent() {
    if (selectedSnippet) {
      try {
        const { ClipboardSetText } = await import("../wailsjs/runtime/runtime");
        await ClipboardSetText(selectedSnippet.Content);
      } catch {
        await navigator.clipboard.writeText(selectedSnippet.Content);
      }
      await IncrementUseCount(selectedSnippet.ID).catch(() => {});
    }
  }

  // Keyboard shortcut handlers
  const shortcutHandlers = {
    onNewSnippet: () => dispatch({ type: "SET_CREATE_MODE", creating: true }),
    onFocusSearch: () => searchBarRef.current?.focus(),
    onCopyContent: handleCopyContent,
    onSave: () => {}, // Handled by DetailPane internally via ActionBar
    onDelete: () => {
      if (selectedSnippet) handleDelete(selectedSnippet.ID);
    },
    onNavigateUp: () => {
      const list = state.searchResults?.map((r) => r.Snippet) ?? state.snippets;
      const idx = list.findIndex((s) => s.ID === state.selectedId);
      if (idx > 0) dispatch({ type: "SET_SELECTED", id: list[idx - 1].ID });
    },
    onNavigateDown: () => {
      const list = state.searchResults?.map((r) => r.Snippet) ?? state.snippets;
      const idx = list.findIndex((s) => s.ID === state.selectedId);
      if (idx < list.length - 1) dispatch({ type: "SET_SELECTED", id: list[idx + 1].ID });
    },
    onEscape: () => {
      if (state.editMode) {
        dispatch({ type: "SET_EDIT_MODE", editing: false });
      } else if (state.searchQuery) {
        dispatch({ type: "CLEAR_SEARCH" });
      } else if (state.createMode) {
        dispatch({ type: "SET_CREATE_MODE", creating: false });
      }
    },
    onTogglePin: () => {
      if (selectedSnippet) {
        handleTogglePin(selectedSnippet.ID, !selectedSnippet.Pinned);
      }
    },
  };

  useKeyboardShortcuts(state, shortcutHandlers);

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      {/* Title bar drag region */}
      <div
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          right: 0,
          height: 40,
          // @ts-ignore — Wails-specific CSS property
          "--wails-draggable": "drag",
          WebkitAppRegion: "drag",
          zIndex: 50,
        } as React.CSSProperties}
      />

      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
        <Sidebar
          snippets={state.snippets}
          searchResults={state.searchResults}
          searchQuery={state.searchQuery}
          selectedId={state.selectedId}
          onSearchChange={(q) => dispatch({ type: "SET_SEARCH_QUERY", query: q })}
          onSelect={handleSelect}
          onNewSnippet={() => dispatch({ type: "SET_CREATE_MODE", creating: true })}
          searchBarRef={searchBarRef}
        />
        <DetailPane
          snippet={selectedSnippet}
          editMode={state.editMode}
          createMode={state.createMode}
          onUpdate={handleUpdate}
          onUpdateTags={handleUpdateTags}
          onDelete={handleDelete}
          onTogglePin={handleTogglePin}
          onCreate={handleCreate}
          onSetEditMode={(editing) => dispatch({ type: "SET_EDIT_MODE", editing })}
        />
      </div>

      <StatusBar
        snippetCount={state.snippets.length}
        searching={state.searchQuery.length > 0}
        editMode={state.editMode}
      />
    </div>
  );
}

export default function App() {
  return (
    <AppProvider>
      <AppContent />
    </AppProvider>
  );
}
```

Note: The Wails binding imports (`../wailsjs/go/gui/App`) are auto-generated by `wails dev` or `wails build`. On the first build, these files won't exist until Wails generates them. The `wailsjsdir` in `wails.json` controls where they land. If the import paths don't match, check the generated `wailsjs/` directory structure and adjust.

**Step 2: Build and test**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt-gui && wails dev
```

Expected: Full app renders with sidebar showing snippets from the database, search works, clicking a snippet shows detail pane with CodeMirror syntax highlighting, and keyboard shortcuts are functional.

**Step 3: Commit**

```bash
git add src/frontend/src/App.tsx
git commit -m "feat(gui): wire all components together in App with backend integration"
```

---

## Task 13: Build Verification and Polish

**Step 1: Full production build**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt/src/cmd/snipt-gui && wails build
```

Expected: Produces a macOS `.app` bundle in `build/bin/`.

**Step 2: Verify CLI still works**

Run:
```bash
cd /Users/jayne/Desktop/codingProjects/snipt && go build -o /tmp/snipt-test ./src/cmd/snipt/ && /tmp/snipt-test list
```

Expected: CLI builds and runs independently. Lists the same snippets visible in the GUI.

**Step 3: Update `.gitignore`**

Add these entries if not already present:

```
# Wails
src/frontend/node_modules/
src/frontend/dist/
src/frontend/src/wailsjs/
src/cmd/snipt-gui/build/
```

**Step 4: Commit**

```bash
git add .gitignore
git commit -m "chore: update gitignore for Wails build artifacts"
```

---

## Task 14: Verification Checklist

Run through each item manually:

- [ ] `wails build` succeeds
- [ ] App launches with correct window size (1100x700) and dark background (#0d0d14)
- [ ] macOS title bar: transparent, draggable, traffic lights positioned correctly
- [ ] Snippets load from same SQLite database as CLI
- [ ] Search filters snippets in real-time (FTS5) with match highlighting via `<mark>`
- [ ] Clicking a snippet shows detail with CodeMirror syntax highlighting
- [ ] CodeMirror theme matches Catppuccin Mocha (dark bg, mauve keywords, green strings)
- [ ] Create new snippet works, appears in sidebar
- [ ] Edit snippet works (double-click or Edit button), changes persist after save
- [ ] Delete snippet works with confirmation dialog, removed from sidebar
- [ ] Copy button copies content to clipboard, shows "Copied!" feedback
- [ ] Pin toggle works, pinned snippets have yellow indicator
- [ ] Keyboard shortcuts: Cmd+N, Cmd+F, Cmd+S, Cmd+Backspace, Up/Down, Escape, Cmd+P
- [ ] CLI still works independently (`snipt list`, `snipt find`)
- [ ] Both GUI and CLI read/write to the same database

Fix any issues found during verification and commit fixes individually with descriptive messages.

---

## Summary

| Task | Description | Key files |
|------|------------|-----------|
| 1 | Install Wails CLI | (system) |
| 2 | Scaffold project structure | `wails.json`, `package.json`, `vite.config.ts`, `tsconfig.json`, `index.html` |
| 3 | Go backend bindings | `src/internal/gui/app.go` |
| 4 | Wails entry point | `src/cmd/snipt-gui/main.go` |
| 5 | React entry + global styles | `main.tsx`, `global.css`, `colors.ts`, `App.tsx` placeholder |
| 6 | State management | `state/types.ts`, `state/context.tsx` |
| 7 | Hooks | `useDebounce.ts`, `useKeyboardShortcuts.ts` |
| 8 | CodeMirror theme + languages | `catppuccin-theme.ts`, `languages.ts` |
| 9 | Sidebar components | `SearchBar`, `SnippetRow`, `SnippetList`, `NewSnippetButton`, `Sidebar` |
| 10 | Detail pane components | `CodeEditor`, `DetailHeader`, `MetadataFooter`, `ActionBar`, `ConfirmDialog`, `DetailPane` |
| 11 | Status bar | `StatusBar.tsx` |
| 12 | App orchestration | `App.tsx` (full rewrite) |
| 13 | Build verification | `.gitignore`, production build test |
| 14 | Manual verification checklist | (manual testing) |
