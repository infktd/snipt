import { C, BODY, MONO } from "../styles/colors";
import { langColors, highlightTitle } from "../utils/snippetDisplay";
import type { Snippet } from "../state/types";

interface SnippetRowProps {
  snippet: Snippet;
  selected: boolean;
  titleIndices?: number[] | null;
  onClick: (e: React.MouseEvent) => void;
}

export function SnippetRow({ snippet, selected, titleIndices, onClick }: SnippetRowProps) {
  const badgeColor = langColors[snippet.Language.toLowerCase()] ?? C.textDim;

  return (
    <div
      onClick={onClick}
      style={{
        padding: "10px 14px",
        cursor: "pointer",
        userSelect: "none",
        borderLeft: `3px solid ${selected ? C.pink : "transparent"}`,
        background: selected ? C.bgSurface : "transparent",
        transition: "background 0.12s ease, border-color 0.12s ease",
      }}
      onMouseEnter={(e) => {
        if (!selected) e.currentTarget.style.background = "rgba(203, 166, 247, 0.04)";
      }}
      onMouseLeave={(e) => {
        if (!selected) e.currentTarget.style.background = "transparent";
      }}
    >
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
          <span style={{ color: C.yellow, fontSize: 8, flexShrink: 0 }}>●</span>
        )}
        <span style={{ overflow: "hidden", textOverflow: "ellipsis" }}>
          {highlightTitle(snippet.Title, titleIndices)}
        </span>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 6,
          marginTop: 4,
          fontSize: 11,
          overflow: "hidden",
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
              fontSize: 11,
            }}
          >
            #{tag}
          </span>
        ))}
      </div>
    </div>
  );
}
