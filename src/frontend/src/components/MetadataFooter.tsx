import { useState } from "react";
import { C, BODY, MONO } from "../styles/colors";

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

  function formatDate(dateStr: string): string {
    if (!dateStr) return "";
    try {
      return new Date(dateStr).toLocaleDateString("en-US", {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    } catch {
      return dateStr;
    }
  }

  return (
    <div
      style={{
        padding: "16px 0 0",
        borderTop: `1px solid ${C.border}`,
        marginTop: 16,
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
        <span style={{ color: C.textDim, fontFamily: BODY, fontSize: 12 }}>
          Tags:
        </span>
        {tags.map((tag) => (
          <span
            key={tag}
            style={{
              color: C.textSub,
              fontFamily: MONO,
              fontSize: 11,
              padding: "2px 8px",
              borderRadius: 4,
              background: C.bgSurface,
              display: "flex",
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
                x
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
              borderRadius: 4,
              padding: "2px 8px",
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
          fontFamily: BODY,
          fontSize: 11,
          color: C.textMuted,
        }}
      >
        <span
          onClick={onTogglePin}
          style={{
            cursor: "pointer",
            color: pinned ? C.yellow : C.textMuted,
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
