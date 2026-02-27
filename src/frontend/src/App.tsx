import { useEffect, useRef, useCallback } from "react";
import { AppProvider, useAppState, useAppDispatch } from "./state/context";
import { Sidebar } from "./components/Sidebar";
import { DetailPane } from "./components/DetailPane";
import { StatusBar } from "./components/StatusBar";
import { useDebounce } from "./hooks/useDebounce";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import type { SearchBarHandle } from "./components/SearchBar";
import type { Snippet } from "./state/types";

import {
  ListSnippets,
  SearchSnippets,
  CreateSnippet,
  UpdateSnippet,
  UpdateSnippetTags,
  DeleteSnippet,
  SetPinned,
  IncrementUseCount,
} from "./wailsjs/go/gui/App";
import { ClipboardSetText } from "./wailsjs/runtime/runtime";

function AppContent() {
  const state = useAppState();
  const dispatch = useAppDispatch();
  const searchBarRef = useRef<SearchBarHandle>(null);
  const debouncedQuery = useDebounce(state.searchQuery, 300);

  // Load snippets on mount
  const loadSnippets = useCallback(async () => {
    try {
      const snippets = await ListSnippets({} as never);
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
      .then((results) =>
        dispatch({ type: "SET_SEARCH_RESULTS", results: results ?? [] })
      )
      .catch((err) => console.error("Search failed:", err));
  }, [debouncedQuery, dispatch]);

  // Get selected snippet
  const selectedSnippet =
    state.snippets.find((s) => s.ID === state.selectedId) ?? null;

  // Handlers
  function handleSelect(id: string) {
    dispatch({ type: "SET_SELECTED", id });
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
      } as never);
      dispatch({ type: "SET_CREATE_MODE", creating: false });
      await loadSnippets();
    } catch (err) {
      console.error("Create failed:", err);
    }
  }

  async function handleCopyContent() {
    if (selectedSnippet) {
      try {
        await ClipboardSetText(selectedSnippet.Content);
      } catch {
        await navigator.clipboard.writeText(selectedSnippet.Content);
      }
      IncrementUseCount(selectedSnippet.ID).catch(() => {});
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
      const list =
        state.searchResults?.map((r) => r.Snippet) ?? state.snippets;
      const idx = list.findIndex((s) => s.ID === state.selectedId);
      if (idx > 0) dispatch({ type: "SET_SELECTED", id: list[idx - 1].ID });
    },
    onNavigateDown: () => {
      const list =
        state.searchResults?.map((r) => r.Snippet) ?? state.snippets;
      const idx = list.findIndex((s) => s.ID === state.selectedId);
      if (idx < list.length - 1)
        dispatch({ type: "SET_SELECTED", id: list[idx + 1].ID });
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
        style={
          {
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            height: 40,
            "--wails-draggable": "drag",
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
          selectedId={state.selectedId}
          onSearchChange={(q) =>
            dispatch({ type: "SET_SEARCH_QUERY", query: q })
          }
          onSelect={handleSelect}
          onNewSnippet={() =>
            dispatch({ type: "SET_CREATE_MODE", creating: true })
          }
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
          onSetEditMode={(editing) =>
            dispatch({ type: "SET_EDIT_MODE", editing })
          }
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
