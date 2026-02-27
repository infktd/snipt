import { SnippetRow } from "./SnippetRow";
import type { Snippet, SearchResult } from "../state/types";

interface SnippetListProps {
  snippets: Snippet[];
  searchResults: SearchResult[] | null;
  selectedIds: Set<string>;
  onSelect: (id: string, e: React.MouseEvent) => void;
}

export function SnippetList({ snippets, searchResults, selectedIds, onSelect }: SnippetListProps) {
  if (searchResults) {
    return (
      <div
        style={{ flex: 1, overflowY: "auto", userSelect: "none" }}
        onMouseDown={(e) => { if (e.shiftKey) e.preventDefault(); }}
      >
        {searchResults.map((result) => (
          <SnippetRow
            key={result.Snippet.ID}
            snippet={result.Snippet}
            selected={selectedIds.has(result.Snippet.ID)}
            titleIndices={result.TitleIndices}
            onClick={(e) => onSelect(result.Snippet.ID, e)}
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
    <div
        style={{ flex: 1, overflowY: "auto", userSelect: "none" }}
        onMouseDown={(e) => { if (e.shiftKey) e.preventDefault(); }}
      >
      {snippets.map((snippet) => (
        <SnippetRow
          key={snippet.ID}
          snippet={snippet}
          selected={selectedIds.has(snippet.ID)}
          onClick={(e) => onSelect(snippet.ID, e)}
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
