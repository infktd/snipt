import { useState } from "react";
import { C, MONO } from "../styles/colors";

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

  const secondaryStyle: React.CSSProperties = {
    background: "transparent",
    color: C.textSub,
    fontFamily: MONO,
    fontSize: 12,
    fontWeight: 500,
    padding: "8px 18px",
    borderRadius: 6,
    border: `1px solid ${C.border}`,
    cursor: "pointer",
    transition: "all 0.12s ease",
  };

  const deleteStyle: React.CSSProperties = {
    background: "transparent",
    color: C.textDim,
    fontFamily: MONO,
    fontSize: 12,
    fontWeight: 500,
    padding: "8px 18px",
    borderRadius: 6,
    border: `1px solid ${C.border}`,
    cursor: "pointer",
    transition: "all 0.12s ease",
  };

  return (
    <div style={{ display: "flex", gap: 8, padding: "12px 24px 20px" }}>
      <button
        onClick={handleCopy}
        style={{
          background: copied
            ? C.green
            : `linear-gradient(135deg, ${C.pink}, ${C.mauve})`,
          color: C.bg,
          fontFamily: MONO,
          fontSize: 12,
          fontWeight: 600,
          padding: "8px 18px",
          borderRadius: 6,
          border: "none",
          cursor: "pointer",
          transition: "transform 0.12s ease, box-shadow 0.12s ease",
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.transform = "translateY(-1px)";
          e.currentTarget.style.boxShadow = "0 4px 16px rgba(203, 166, 247, 0.2)";
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.transform = "translateY(0)";
          e.currentTarget.style.boxShadow = "none";
        }}
      >
        {copied ? "Copied!" : "Copy"}
      </button>
      {editMode ? (
        <>
          <button
            onClick={onSave}
            style={{
              background: `linear-gradient(135deg, ${C.green}, ${C.teal})`,
              color: C.bg,
              fontFamily: MONO,
              fontSize: 12,
              fontWeight: 600,
              padding: "8px 18px",
              borderRadius: 6,
              border: "none",
              cursor: "pointer",
              transition: "transform 0.12s ease, box-shadow 0.12s ease",
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = "translateY(-1px)";
              e.currentTarget.style.boxShadow = "0 4px 16px rgba(166, 227, 161, 0.2)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = "translateY(0)";
              e.currentTarget.style.boxShadow = "none";
            }}
          >
            Save
          </button>
          <button onClick={onCancelEdit} style={secondaryStyle}>
            Cancel
          </button>
        </>
      ) : (
        <button
          onClick={onEdit}
          style={secondaryStyle}
          onMouseEnter={(e) => {
            e.currentTarget.style.borderColor = "rgba(203, 166, 247, 0.3)";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.borderColor = C.border;
          }}
        >
          Edit
        </button>
      )}
      <button
        onClick={onDelete}
        style={deleteStyle}
        onMouseEnter={(e) => {
          e.currentTarget.style.color = C.red;
          e.currentTarget.style.borderColor = "rgba(243, 139, 168, 0.3)";
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.color = C.textDim;
          e.currentTarget.style.borderColor = C.border;
        }}
      >
        Delete
      </button>
    </div>
  );
}
