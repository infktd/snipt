import { C, BODY } from "../styles/colors";

interface NewSnippetButtonProps {
  onClick: () => void;
}

export function NewSnippetButton({ onClick }: NewSnippetButtonProps) {
  return (
    <button
      onClick={onClick}
      style={{
        flex: 1,
        padding: "10px 14px",
        background: "transparent",
        border: "none",
        color: C.textDim,
        fontFamily: BODY,
        fontSize: 13,
        cursor: "pointer",
        textAlign: "left",
        transition: "color 0.12s ease",
      }}
      onMouseEnter={(e) => (e.currentTarget.style.color = C.textSub)}
      onMouseLeave={(e) => (e.currentTarget.style.color = C.textDim)}
    >
      + New snippet
    </button>
  );
}
