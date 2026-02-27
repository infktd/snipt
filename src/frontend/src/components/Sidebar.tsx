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
        paddingTop: 40,
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
