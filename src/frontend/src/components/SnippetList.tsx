import { SnippetRow } from "./SnippetRow";
import type { Snippet, SearchResult } from "../state/types";

interface SnippetListProps {
  snippets: Snippet[];
  searchResults: SearchResult[] | null;
  selectedId: string | null;
  onSelect: (id: string) => void;
}

export function SnippetList({ snippets, searchResults, selectedId, onSelect }: SnippetListProps) {
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
