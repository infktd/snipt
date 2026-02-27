import { C, MONO } from "../styles/colors";

interface StatusBarProps {
  snippetCount: number;
  selectedCount: number;
  searching: boolean;
  editMode: boolean;
}

export function StatusBar({ snippetCount, selectedCount, searching, editMode }: StatusBarProps) {
  const multiSelected = selectedCount > 1;

  return (
    <div
      className="no-select"
      style={{
        height: 32,
        minHeight: 32,
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "0 16px",
        background: C.bgCard,
        borderTop: `1px solid ${C.borderSubtle}`,
        fontFamily: MONO,
        fontSize: 11,
      }}
    >
      <div style={{ display: "flex", gap: 12 }}>
        <span style={{ color: C.textDim }}>
          {snippetCount} snippet{snippetCount !== 1 ? "s" : ""}
        </span>
        {multiSelected && (
          <span style={{ color: C.mauve }}>
            {selectedCount} selected
          </span>
        )}
        {searching && <span style={{ color: C.pink }}>searching...</span>}
      </div>
      <div style={{ display: "flex", gap: 16, color: C.textMuted }}>
        {editMode ? (
          <>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>⌘S</span> save</span>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>Esc</span> cancel</span>
          </>
        ) : multiSelected ? (
          <>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>⌘C</span> copy all</span>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>⌘⌫</span> delete</span>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>Esc</span> deselect</span>
          </>
        ) : (
          <>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>⌘N</span> new</span>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>⌘F</span> search</span>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>↑↓</span> navigate</span>
            <span><span style={{ color: C.textDim, fontWeight: 500 }}>⌘P</span> pin</span>
          </>
        )}
      </div>
    </div>
  );
}
