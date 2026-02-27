import { useState, useEffect, useRef, useCallback } from "react";
import { C } from "../styles/colors";
import { langColors, highlightTitle } from "../utils/snippetDisplay";
import {
  GetConfig,
  ListSnippets,
  SearchSnippets,
  IncrementUseCount,
} from "../bindings/snippetservice";
import { Clipboard, Events } from "@wailsio/runtime";
import { useDebounce } from "../hooks/useDebounce";
import type { Snippet, SearchResult } from "../state/types";

export function FindPalette() {
  const [query, setQuery] = useState("");
  const [snippets, setSnippets] = useState<Snippet[]>([]);
  const [searchResults, setSearchResults] = useState<SearchResult[] | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const debouncedQuery = useDebounce(query, 150);

  // Load snippets
  const loadSnippets = useCallback(() => {
    GetConfig()
      .then((cfg: any) => {
        const sort = cfg.Find?.Sort || "recent";
        return ListSnippets({ Sort: sort });
      })
      .then((s: any) => setSnippets(s ?? []))
      .catch((err: any) => console.error("Failed to load snippets:", err));
  }, []);

  // Load on mount
  // Transparent background so CSS border-radius shows through to desktop
  useEffect(() => {
    document.body.style.background = "transparent";
    document.documentElement.style.background = "transparent";
  }, []);

  useEffect(() => {
    loadSnippets();
    inputRef.current?.focus();
  }, [loadSnippets]);

  // Reset state when palette is shown from the tray
  useEffect(() => {
    const cancel = Events.On("find-opened", () => {
      setQuery("");
      setActiveIndex(0);
      setSearchResults(null);
      setCopiedId(null);
      loadSnippets();
      inputRef.current?.focus();
    });
    return cancel;
  }, [loadSnippets]);

  // Search when debounced query changes
  useEffect(() => {
    if (!debouncedQuery.trim()) {
      setSearchResults(null);
      setActiveIndex(0);
      return;
    }
    SearchSnippets(debouncedQuery).then((results: any) => {
      setSearchResults(results ?? []);
      setActiveIndex(0);
    });
  }, [debouncedQuery]);

  // Derive display list
  const displayList: Snippet[] = searchResults
    ? searchResults.map((r) => r.Snippet)
    : snippets;

  const getTitleIndices = (id: string): number[] | null => {
    if (!searchResults) return null;
    const result = searchResults.find((r) => r.Snippet.ID === id);
    return result?.TitleIndices ?? null;
  };

  // Copy snippet content, flash green, then hide palette
  const handleSelect = useCallback(async (snippet: Snippet) => {
    setCopiedId(snippet.ID);
    try {
      await Clipboard.SetText(snippet.Content);
    } catch {
      await navigator.clipboard.writeText(snippet.Content);
    }
    IncrementUseCount(snippet.ID).catch(() => {});
    setTimeout(() => Events.Emit("find-done"), 180);
  }, []);

  // Keyboard navigation
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setActiveIndex((i) => Math.min(i + 1, displayList.length - 1));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setActiveIndex((i) => Math.max(i - 1, 0));
      } else if (e.key === "Enter") {
        e.preventDefault();
        if (displayList[activeIndex]) {
          handleSelect(displayList[activeIndex]);
        }
      } else if (e.key === "Escape") {
        e.preventDefault();
        Events.Emit("find-done");
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [displayList, activeIndex, handleSelect]);

  // Scroll active item into view
  useEffect(() => {
    const container = listRef.current;
    if (!container) return;
    const activeEl = container.children[activeIndex] as HTMLElement;
    activeEl?.scrollIntoView({ block: "nearest" });
  }, [activeIndex]);

  return (
    <div className="find-palette">
      <div className="find-palette-inner">
        {/* Search bar — draggable region */}
        <div
          className="find-search"
          style={{ WebkitAppRegion: "drag" } as React.CSSProperties}
        >
          <span
            className="snipt-badge"
            style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          >
            SNIPT
          </span>
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search snippets..."
            style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
          />
          <span className="result-count">
            {searchResults
              ? `${displayList.length}/${snippets.length}`
              : `${snippets.length}`}
          </span>
        </div>

        {/* Results list */}
        <div ref={listRef} className="find-results">
          {displayList.map((snippet, i) => {
            const isActive = i === activeIndex;
            const isCopied = copiedId === snippet.ID;
            const badgeColor =
              langColors[snippet.Language.toLowerCase()] ?? C.textDim;
            return (
              <div
                key={snippet.ID}
                className={`find-result-row${isActive ? " selected" : ""}${isCopied ? " copied" : ""}`}
                onClick={() => handleSelect(snippet)}
                onMouseEnter={() => setActiveIndex(i)}
              >
                {snippet.Pinned && <span className="pin">●</span>}
                <span className="title">
                  {highlightTitle(snippet.Title, getTitleIndices(snippet.ID))}
                </span>
                {snippet.Language && (
                  <span
                    className="lang"
                    style={{
                      color: badgeColor,
                      background: `${badgeColor}14`,
                    }}
                  >
                    {snippet.Language}
                  </span>
                )}
                {snippet.Tags?.map((tag) => (
                  <span key={tag} className="tag">
                    #{tag}
                  </span>
                ))}
              </div>
            );
          })}
          {displayList.length === 0 && (
            <div className="find-empty">No snippets found</div>
          )}
        </div>

        {/* Footer hints */}
        <div className="find-footer">
          <span className="key-hint">
            <kbd>↑↓</kbd>
            <span className="key-label">navigate</span>
          </span>
          <span className="key-hint">
            <kbd>enter</kbd>
            <span className="key-label">copy</span>
          </span>
          <span className="key-hint">
            <kbd>esc</kbd>
            <span className="key-label">close</span>
          </span>
        </div>
      </div>
    </div>
  );
}
