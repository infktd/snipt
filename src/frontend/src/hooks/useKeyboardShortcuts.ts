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

      if (meta && e.key === "n") {
        e.preventDefault();
        handlers.onNewSnippet();
        return;
      }

      if (meta && e.key === "f") {
        e.preventDefault();
        handlers.onFocusSearch();
        return;
      }

      if (meta && e.key === "s") {
        e.preventDefault();
        if (state.editMode) {
          handlers.onSave();
        }
        return;
      }

      if (meta && e.key === "Backspace") {
        e.preventDefault();
        if (!state.editMode && state.selectedId) {
          handlers.onDelete();
        }
        return;
      }

      if (meta && e.key === "p") {
        e.preventDefault();
        if (state.selectedId) {
          handlers.onTogglePin();
        }
        return;
      }

      if (meta && e.key === "c" && !state.editMode) {
        const selection = window.getSelection();
        if (!selection || selection.isCollapsed) {
          e.preventDefault();
          handlers.onCopyContent();
          return;
        }
      }

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
