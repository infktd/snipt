import { useState } from "react";
import { C, MONO } from "../styles/colors";

interface MetadataFooterProps {
  tags: string[];
  pinned: boolean;
  useCount: number;
  createdAt: string;
  updatedAt: string;
  editMode: boolean;
  onTagsChange: (tags: string[]) => void;
  onTogglePin: () => void;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return "";
  if (dateStr.startsWith("0001-")) return "";
  try {
    // SQLite datetime('now') gives "YYYY-MM-DD HH:MM:SS" (no T, no Z)
    const normalized = dateStr.includes("T") ? dateStr : dateStr.replace(" ", "T") + "Z";
    const d = new Date(normalized);
    if (isNaN(d.getTime())) return "";
    return d.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  } catch {
    return "";
  }
}

export function MetadataFooter({
  tags,
  pinned,
  useCount,
  createdAt,
  updatedAt,
  editMode,
  onTagsChange,
  onTogglePin,
}: MetadataFooterProps) {
  const [tagInput, setTagInput] = useState("");

  function addTag() {
    const tag = tagInput.trim().toLowerCase().replace(/^#/, "");
    if (tag && !tags.includes(tag)) {
      onTagsChange([...tags, tag]);
    }
    setTagInput("");
  }

  function removeTag(tag: string) {
    onTagsChange(tags.filter((t) => t !== tag));
  }

  return (
    <div
      style={{
        padding: "16px 24px",
        display: "flex",
        flexDirection: "column",
        gap: 12,
      }}
    >
      {/* Tags */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 8,
          flexWrap: "wrap",
        }}
      >
        {tags.map((tag) => (
          <span
            key={tag}
            style={{
              color: C.textSub,
              fontFamily: MONO,
              fontSize: 11,
              padding: "3px 10px",
              borderRadius: 6,
              background: C.bgSurface,
              border: `1px solid ${C.border}`,
              display: "inline-flex",
              alignItems: "center",
              gap: 4,
            }}
          >
            #{tag}
            {editMode && (
              <span
                onClick={() => removeTag(tag)}
                style={{
                  cursor: "pointer",
                  color: C.textMuted,
                  marginLeft: 2,
                }}
              >
                ×
              </span>
            )}
          </span>
        ))}
        {editMode && (
          <input
            value={tagInput}
            onChange={(e) => setTagInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") addTag();
            }}
            placeholder="add tag..."
            style={{
              background: "transparent",
              border: `1px solid ${C.border}`,
              borderRadius: 6,
              padding: "3px 10px",
              color: C.text,
              fontFamily: MONO,
              fontSize: 11,
              outline: "none",
              width: 100,
            }}
          />
        )}
      </div>

      {/* Meta row */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 16,
          fontFamily: MONO,
          fontSize: 12,
          color: C.textMuted,
          marginTop: 8,
        }}
      >
        <span
          onClick={onTogglePin}
          style={{
            cursor: "pointer",
            color: pinned ? C.yellow : C.textMuted,
            transition: "color 0.12s ease",
          }}
        >
          {pinned ? "● Pinned" : "○ Not pinned"}
        </span>
        <span>Used {useCount}x</span>
        {createdAt && <span>Created {formatDate(createdAt)}</span>}
        {updatedAt && <span>Updated {formatDate(updatedAt)}</span>}
      </div>
    </div>
  );
}
