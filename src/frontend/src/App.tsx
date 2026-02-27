import { useEffect, useRef, useCallback, useMemo } from "react";
import { AppProvider, useAppState, useAppDispatch } from "./state/context";
import { Sidebar } from "./components/Sidebar";
import { DetailPane } from "./components/DetailPane";
import { StatusBar } from "./components/StatusBar";
import { FindPalette } from "./components/FindPalette";
import { Settings } from "./components/Settings";
import { useDebounce } from "./hooks/useDebounce";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import type { SearchBarHandle } from "./components/SearchBar";
import type { Snippet } from "./state/types";

import {
  GetConfig,
  ListSnippets,
  SearchSnippets,
  CreateSnippet,
  UpdateSnippet,
  UpdateSnippetTags,
  DeleteSnippet,
  SetPinned,
  IncrementUseCount,
} from "./bindings/snippetservice";
import { Clipboard, Events } from "@wailsio/runtime";

function AppContent() {
  const state = useAppState();
  const dispatch = useAppDispatch();
  const searchBarRef = useRef<SearchBarHandle>(null);
  const debouncedQuery = useDebounce(state.searchQuery, 300);

  // Load snippets with sort from config
  const loadSnippets = useCallback(async () => {
    try {
      const cfg = await GetConfig();
      const sort = cfg.Find?.Sort || "recent";
      const snippets = await ListSnippets({ Sort: sort } as never);
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
      .then((results: any) =>
        dispatch({ type: "SET_SEARCH_RESULTS", results: results ?? [] })
      )
      .catch((err: any) => console.error("Search failed:", err));
  }, [debouncedQuery, dispatch]);

  // Listen for tray "open-settings" event
  useEffect(() => {
    const cancel = Events.On("open-settings", () => {
      dispatch({ type: "OPEN_SETTINGS" });
    });
    return cancel;
  }, [dispatch]);

  // Ordered display list and IDs
  const displayList: Snippet[] = useMemo(
    () => state.searchResults?.map((r) => r.Snippet) ?? state.snippets,
    [state.searchResults, state.snippets],
  );
  const displayIds: string[] = useMemo(
    () => displayList.map((s) => s.ID),
    [displayList],
  );

  // Derive selected snippets
  const selectedSnippets = useMemo(
    () => state.snippets.filter((s) => state.selectedIds.has(s.ID)),
    [state.snippets, state.selectedIds],
  );
  const singleSelected = selectedSnippets.length === 1 ? selectedSnippets[0] : null;

  // Handlers
  function handleSelect(id: string, e: React.MouseEvent) {
    if (e.metaKey || e.ctrlKey) {
      dispatch({ type: "SELECT_TOGGLE", id });
    } else if (e.shiftKey) {
      dispatch({ type: "SELECT_RANGE", id, list: displayIds });
    } else {
      dispatch({ type: "SELECT_SINGLE", id });
    }
  }

  async function handleUpdate(snippet: Snippet) {
    try {
      await UpdateSnippet(snippet as never);
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

  async function handleDeleteSelected() {
    const ids = [...state.selectedIds];
    try {
      for (const id of ids) {
        await DeleteSnippet(id);
      }
      dispatch({ type: "SELECT_CLEAR" });
      await loadSnippets();
    } catch (err) {
      console.error("Delete failed:", err);
    }
  }

  async function handleTogglePinSelected() {
    const shouldPin = selectedSnippets.some((s) => !s.Pinned);
    try {
      for (const s of selectedSnippets) {
        await SetPinned(s.ID, shouldPin);
      }
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
        Pinned: partial.Pinned ?? false,
        UseCount: 0,
        Tags: partial.Tags ?? [],
        CreatedAt: "0001-01-01T00:00:00Z",
        UpdatedAt: "0001-01-01T00:00:00Z",
      } as never);
      dispatch({ type: "SET_CREATE_MODE", creating: false });
      await loadSnippets();
    } catch (err) {
      console.error("Create failed:", err);
    }
  }

  async function handleCopyContent() {
    if (selectedSnippets.length === 0) return;
    const combined = selectedSnippets.map((s) => s.Content).join("\n\n---\n\n");
    try {
      await Clipboard.SetText(combined);
    } catch {
      await navigator.clipboard.writeText(combined);
    }
    for (const s of selectedSnippets) {
      IncrementUseCount(s.ID).catch(() => {});
    }
  }

  // Keyboard shortcut handlers
  const shortcutHandlers = {
    onNewSnippet: () => dispatch({ type: "SET_CREATE_MODE", creating: true }),
    onFocusSearch: () => searchBarRef.current?.focus(),
    onCopyContent: handleCopyContent,
    onSave: () => {}, // Handled by DetailPane internally via ActionBar
    onDelete: () => {
      if (state.selectedIds.size > 0) handleDeleteSelected();
    },
    onNavigateUp: () => {
      const idx = displayIds.indexOf(state.focusId ?? "");
      if (idx > 0) {
        dispatch({ type: "SELECT_SINGLE", id: displayIds[idx - 1] });
      }
    },
    onNavigateDown: () => {
      const idx = displayIds.indexOf(state.focusId ?? "");
      if (idx < displayIds.length - 1) {
        dispatch({ type: "SELECT_SINGLE", id: displayIds[idx + 1] });
      }
    },
    onNavigateUpExtend: () => {
      const idx = displayIds.indexOf(state.focusId ?? "");
      if (idx > 0) {
        dispatch({ type: "SELECT_RANGE", id: displayIds[idx - 1], list: displayIds });
      }
    },
    onNavigateDownExtend: () => {
      const idx = displayIds.indexOf(state.focusId ?? "");
      if (idx < displayIds.length - 1) {
        dispatch({ type: "SELECT_RANGE", id: displayIds[idx + 1], list: displayIds });
      }
    },
    onEscape: () => {
      if (state.detailView.kind === "settings") {
        dispatch({ type: "CLOSE_SETTINGS" });
      } else if (state.createMode) {
        dispatch({ type: "SET_CREATE_MODE", creating: false });
      } else if (state.selectedIds.size > 1) {
        if (state.focusId) {
          dispatch({ type: "SELECT_SINGLE", id: state.focusId });
        } else {
          dispatch({ type: "SELECT_CLEAR" });
        }
      } else if (state.editMode) {
        dispatch({ type: "SET_EDIT_MODE", editing: false });
      } else if (state.searchQuery) {
        dispatch({ type: "CLEAR_SEARCH" });
      } else if (state.selectedIds.size === 1) {
        dispatch({ type: "SELECT_CLEAR" });
      }
    },
    onTogglePin: () => {
      if (selectedSnippets.length > 0) {
        handleTogglePinSelected();
      }
    },
  };

  useKeyboardShortcuts(state, shortcutHandlers);

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      {/* Title bar drag region */}
      <div
        style={
          {
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            height: 40,
            WebkitAppRegion: "drag",
            zIndex: 50,
          } as React.CSSProperties
        }
      />

      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
        <Sidebar
          snippets={state.snippets}
          searchResults={state.searchResults}
          searchQuery={state.searchQuery}
          selectedIds={state.selectedIds}
          onSearchChange={(q) =>
            dispatch({ type: "SET_SEARCH_QUERY", query: q })
          }
          onSelect={handleSelect}
          onNewSnippet={() =>
            dispatch({ type: "SET_CREATE_MODE", creating: true })
          }
          searchBarRef={searchBarRef}
        />
        {state.detailView.kind === "settings" ? (
          <Settings
            onSortChanged={loadSnippets}
            onClose={() => dispatch({ type: "CLOSE_SETTINGS" })}
          />
        ) : (
          <DetailPane
            snippet={singleSelected}
            selectedCount={selectedSnippets.length}
            editMode={state.editMode}
            createMode={state.createMode}
            onUpdate={handleUpdate}
            onUpdateTags={handleUpdateTags}
            onDelete={handleDeleteSelected}
            onTogglePin={handleTogglePinSelected}
            onBulkCopy={handleCopyContent}
            onCreate={handleCreate}
            onSetEditMode={(editing) =>
              dispatch({ type: "SET_EDIT_MODE", editing })
            }
          />
        )}
      </div>

      <StatusBar
        snippetCount={state.snippets.length}
        selectedCount={selectedSnippets.length}
        searching={state.searchQuery.length > 0}
        editMode={state.editMode}
      />
    </div>
  );
}

export default function App() {
  const params = new URLSearchParams(window.location.search);
  const mode = params.get("mode") === "find" ? "find" : "manage";

  if (mode === "find") return <FindPalette />;

  return (
    <AppProvider>
      <AppContent />
    </AppProvider>
  );
}
