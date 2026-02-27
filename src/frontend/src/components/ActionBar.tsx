import { useState } from "react";
import { C, BODY } from "../styles/colors";

interface ActionBarProps {
  content: string;
  editMode: boolean;
  onEdit: () => void;
  onSave: () => void;
  onDelete: () => void;
  onCancelEdit: () => void;
}

export function ActionBar({
  content,
  editMode,
  onEdit,
  onSave,
  onDelete,
  onCancelEdit,
}: ActionBarProps) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(content);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch (err) {
      console.error("Copy failed:", err);
    }
  }

  const buttonStyle = (color: string): React.CSSProperties => ({
    padding: "6px 14px",
    borderRadius: 6,
    border: `1px solid ${C.border}`,
    background: C.bgCard,
    color,
    fontFamily: BODY,
    fontSize: 12,
    cursor: "pointer",
    transition: "background 0.1s, border-color 0.1s",
  });

  return (
    <div style={{ display: "flex", gap: 8, paddingTop: 12 }}>
      <button
        onClick={handleCopy}
        style={buttonStyle(copied ? C.green : C.textSub)}
      >
        {copied ? "Copied!" : "Copy"}
      </button>
      {editMode ? (
        <>
          <button onClick={onSave} style={buttonStyle(C.green)}>
            Save
          </button>
          <button onClick={onCancelEdit} style={buttonStyle(C.textDim)}>
            Cancel
          </button>
        </>
      ) : (
        <button onClick={onEdit} style={buttonStyle(C.textSub)}>
          Edit
        </button>
      )}
      <button onClick={onDelete} style={buttonStyle(C.red)}>
        Delete
      </button>
    </div>
  );
}
