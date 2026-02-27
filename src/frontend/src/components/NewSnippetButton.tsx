import { C, BODY } from "../styles/colors";

interface NewSnippetButtonProps {
  onClick: () => void;
}

export function NewSnippetButton({ onClick }: NewSnippetButtonProps) {
  return (
    <button
      onClick={onClick}
      style={{
        width: "100%",
        padding: "12px 16px",
        background: C.bgCard,
        border: "none",
        borderTop: `1px solid ${C.border}`,
        color: C.textDim,
        fontFamily: BODY,
        fontSize: 13,
        cursor: "pointer",
        textAlign: "left",
        transition: "color 0.1s",
      }}
      onMouseEnter={(e) => (e.currentTarget.style.color = C.text)}
      onMouseLeave={(e) => (e.currentTarget.style.color = C.textDim)}
    >
      + New snippet
    </button>
  );
}
