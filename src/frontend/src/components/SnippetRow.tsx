import { C, BODY, MONO } from "../styles/colors";
import type { Snippet } from "../state/types";

interface SnippetRowProps {
  snippet: Snippet;
  selected: boolean;
  titleIndices?: number[] | null;
  onClick: () => void;
}

const langColors: Record<string, string> = {
  go: C.blue,
  javascript: C.yellow,
  js: C.yellow,
  typescript: C.blue,
  ts: C.blue,
  python: C.green,
  py: C.green,
  bash: C.peach,
  sh: C.peach,
  sql: C.mauve,
  html: C.red,
  css: C.sky,
  json: C.peach,
  markdown: C.teal,
  md: C.teal,
  rust: C.peach,
  ruby: C.red,
  yaml: C.lavender,
  toml: C.lavender,
};

function highlightTitle(title: string, indices: number[] | null | undefined) {
  if (!indices || indices.length === 0) return title;

  const indexSet = new Set(indices);
  const parts: JSX.Element[] = [];
  let current = "";
  let inMatch = false;

  for (let i = 0; i < title.length; i++) {
    const isMatch = indexSet.has(i);
    if (isMatch !== inMatch) {
      if (current) {
        parts.push(
          inMatch ? <mark key={i}>{current}</mark> : <span key={i}>{current}</span>
        );
      }
      current = "";
      inMatch = isMatch;
    }
    current += title[i];
  }
  if (current) {
    parts.push(
      inMatch ? <mark key="last">{current}</mark> : <span key="last">{current}</span>
    );
  }

  return <>{parts}</>;
}

export function SnippetRow({ snippet, selected, titleIndices, onClick }: SnippetRowProps) {
  const badgeColor = langColors[snippet.Language.toLowerCase()] ?? C.textDim;

  return (
    <div
      onClick={onClick}
      style={{
        padding: "10px 16px",
        cursor: "pointer",
        borderLeft: selected ? `3px solid ${C.pink}` : "3px solid transparent",
        background: selected ? C.bgSurface : "transparent",
        transition: "background 0.1s",
      }}
      onMouseEnter={(e) => {
        if (!selected) e.currentTarget.style.background = C.bgCard;
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
          <span style={{ color: C.yellow, fontSize: 10, flexShrink: 0 }}>●</span>
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
        }}
      >
        {snippet.Language && (
          <span
            style={{
              color: badgeColor,
              fontFamily: MONO,
              fontSize: 10,
              padding: "1px 6px",
              borderRadius: 4,
              background: `${badgeColor}15`,
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
              fontFamily: BODY,
              fontSize: 10,
            }}
          >
            #{tag}
          </span>
        ))}
      </div>
    </div>
  );
}
