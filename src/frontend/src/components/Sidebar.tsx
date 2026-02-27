import { SearchBar, type SearchBarHandle } from "./SearchBar";
import { SnippetList } from "./SnippetList";
import { NewSnippetButton } from "./NewSnippetButton";
import { C, BODY } from "../styles/colors";
import type { Snippet, SearchResult } from "../state/types";

interface SidebarProps {
  snippets: Snippet[];
  searchResults: SearchResult[] | null;
  searchQuery: string;
  selectedIds: Set<string>;
  onSearchChange: (query: string) => void;
  onSelect: (id: string, e: React.MouseEvent) => void;
  onNewSnippet: () => void;
  searchBarRef?: React.Ref<SearchBarHandle>;
  onOpenSettings: () => void;
  settingsActive: boolean;
}

export function Sidebar({
  snippets,
  searchResults,
  searchQuery,
  selectedIds,
  onSearchChange,
  onSelect,
  onNewSnippet,
  searchBarRef,
  onOpenSettings,
  settingsActive,
}: SidebarProps) {
  return (
    <div
      style={{
        width: 300,
        minWidth: 300,
        height: "100%",
        display: "flex",
        flexDirection: "column",
        borderRight: `1px solid ${C.borderSubtle}`,
        background: C.bgCard,
        paddingTop: 40,
        overflowX: "hidden",
      }}
    >
      <SearchBar ref={searchBarRef} value={searchQuery} onChange={onSearchChange} />
      <SnippetList
        snippets={snippets}
        searchResults={searchResults}
        selectedIds={selectedIds}
        onSelect={onSelect}
      />
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
    </div>
  );
}
