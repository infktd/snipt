import { useState, useEffect, useRef, useCallback } from "react";
import { C, BODY, MONO } from "../styles/colors";
import { langColors, highlightTitle } from "../utils/snippetDisplay";
import {
  GetConfig,
  ListSnippets,
  SearchSnippets,
  IncrementUseCount,
} from "../wailsjs/go/gui/App";
import { ClipboardSetText, Quit } from "../wailsjs/runtime/runtime";
import { useDebounce } from "../hooks/useDebounce";
import type { Snippet, SearchResult } from "../state/types";

export function FindPalette() {
  const [query, setQuery] = useState("");
  const [snippets, setSnippets] = useState<Snippet[]>([]);
  const [searchResults, setSearchResults] = useState<SearchResult[] | null>(
    null
  );
  const [activeIndex, setActiveIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const debouncedQuery = useDebounce(query, 150);

  // Load all snippets on mount with sort from config
  useEffect(() => {
    GetConfig()
      .then((cfg) => {
        const sort = cfg.Find?.Sort || "recent";
        return ListSnippets({ Sort: sort } as never);
      })
      .then((s) => setSnippets(s ?? []))
      .catch((err) => console.error("Failed to load snippets:", err));
    inputRef.current?.focus();
  }, []);

  // Search when debounced query changes
  useEffect(() => {
    if (!debouncedQuery.trim()) {
      setSearchResults(null);
      setActiveIndex(0);
      return;
    }
    SearchSnippets(debouncedQuery).then((results) => {
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

  // Copy snippet content and close
  const handleSelect = useCallback(async (snippet: Snippet) => {
    try {
      await ClipboardSetText(snippet.Content);
    } catch {
      await navigator.clipboard.writeText(snippet.Content);
    }
    IncrementUseCount(snippet.ID).catch(() => {});
    Quit();
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
        Quit();
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
    <div
      style={{
        height: "100vh",
        display: "flex",
        flexDirection: "column",
        background: C.bgSurface,
        overflow: "hidden",
      }}
    >
      {/* Search header — draggable region */}
      <div
        style={
          {
            padding: "14px 16px",
            display: "flex",
            alignItems: "center",
            gap: 10,
            borderBottom: `1px solid ${C.borderSubtle}`,
            "--wails-draggable": "drag",
            WebkitAppRegion: "drag",
          } as React.CSSProperties
        }
      >
        <span
          style={{
            background: `linear-gradient(135deg, ${C.pink}, ${C.mauve})`,
            color: C.bg,
            fontFamily: MONO,
            fontWeight: 700,
            fontSize: 10,
            padding: "3px 8px",
            borderRadius: 5,
            letterSpacing: "0.5px",
            flexShrink: 0,
            // @ts-expect-error Wails drag property
            WebkitAppRegion: "no-drag",
          }}
        >
          SNIPT
        </span>
        <input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search snippets..."
          style={
            {
              flex: 1,
              background: "transparent",
              border: "none",
              outline: "none",
              color: C.text,
              fontFamily: BODY,
              fontSize: 14,
              WebkitAppRegion: "no-drag",
            } as React.CSSProperties
          }
        />
        <span
          style={{
            color: C.textDim,
            fontFamily: MONO,
            fontSize: 11,
            flexShrink: 0,
          }}
        >
          {searchResults
            ? `${displayList.length}/${snippets.length}`
            : `${snippets.length}`}
        </span>
      </div>

      {/* Results list */}
      <div
        ref={listRef}
        style={{
          flex: 1,
          overflowY: "auto",
          scrollbarWidth: "thin" as const,
          scrollbarColor: `${C.border} transparent`,
        }}
      >
        {displayList.map((snippet, i) => {
          const isActive = i === activeIndex;
          const badgeColor =
            langColors[snippet.Language.toLowerCase()] ?? C.textDim;
          return (
            <div
              key={snippet.ID}
              onClick={() => handleSelect(snippet)}
              onMouseEnter={() => setActiveIndex(i)}
              style={{
                padding: "8px 16px",
                cursor: "pointer",
                background: isActive ? "#3e3e5e" : "transparent",
                borderLeft: `3px solid ${isActive ? C.pink : "transparent"}`,
                transition: "background 0.08s ease",
              }}
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                  fontFamily: MONO,
                  fontSize: 13,
                  fontWeight: isActive ? 500 : 400,
                  color: C.text,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {snippet.Pinned && (
                  <span
                    style={{ color: C.yellow, fontSize: 8, flexShrink: 0 }}
                  >
                    ●
                  </span>
                )}
                <span style={{ overflow: "hidden", textOverflow: "ellipsis" }}>
                  {highlightTitle(snippet.Title, getTitleIndices(snippet.ID))}
                </span>
              </div>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                  marginTop: 3,
                  paddingLeft: 3,
                }}
              >
                {snippet.Language && (
                  <span
                    style={{
                      color: badgeColor,
                      fontFamily: MONO,
                      fontSize: 10,
                      padding: "1px 6px",
                      borderRadius: 10,
                      background: `${badgeColor}1A`,
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
                      fontFamily: MONO,
                      fontSize: 10,
                    }}
                  >
                    #{tag}
                  </span>
                ))}
              </div>
            </div>
          );
        })}
        {displayList.length === 0 && (
          <div
            style={{
              padding: "24px 16px",
              color: C.textDim,
              fontSize: 13,
              textAlign: "center",
              fontFamily: BODY,
            }}
          >
            No snippets found
          </div>
        )}
      </div>

      {/* Footer hints */}
      <div
        style={{
          padding: "8px 16px",
          borderTop: `1px solid ${C.borderSubtle}`,
          display: "flex",
          gap: 14,
          alignItems: "center",
        }}
      >
        <span style={hintStyle}>
          <kbd style={kbdStyle}>↑↓</kbd>
          <span style={labelStyle}>navigate</span>
        </span>
        <span style={hintStyle}>
          <kbd style={kbdStyle}>enter</kbd>
          <span style={labelStyle}>copy</span>
        </span>
        <span style={hintStyle}>
          <kbd style={kbdStyle}>esc</kbd>
          <span style={labelStyle}>close</span>
        </span>
      </div>
    </div>
  );
}

const hintStyle: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  gap: 4,
};

const kbdStyle: React.CSSProperties = {
  fontFamily: MONO,
  background: "#313147",
  color: C.textSub,
  padding: "2px 6px",
  borderRadius: 3,
  fontSize: 10,
  border: `1px solid ${C.borderSubtle}`,
};

const labelStyle: React.CSSProperties = {
  fontFamily: MONO,
  color: C.textDim,
  fontSize: 10,
};
