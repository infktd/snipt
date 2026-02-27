import { C, BODY, MONO } from "../styles/colors";

interface StatusBarProps {
  snippetCount: number;
  searching: boolean;
  editMode: boolean;
}

export function StatusBar({ snippetCount, searching, editMode }: StatusBarProps) {
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
        borderTop: `1px solid ${C.border}`,
        fontFamily: BODY,
        fontSize: 11,
        color: C.textMuted,
      }}
    >
      <div style={{ display: "flex", gap: 12 }}>
        <span>
          {snippetCount} snippet{snippetCount !== 1 ? "s" : ""}
        </span>
        {searching && <span style={{ color: C.pink }}>searching...</span>}
      </div>
      <div style={{ display: "flex", gap: 16, fontFamily: MONO, fontSize: 10 }}>
        {editMode ? (
          <>
            <span>Cmd+S save</span>
            <span>Esc cancel</span>
          </>
        ) : (
          <>
            <span>Cmd+N new</span>
            <span>Cmd+F search</span>
            <span>Up/Down navigate</span>
            <span>Cmd+P pin</span>
          </>
        )}
      </div>
    </div>
  );
}
