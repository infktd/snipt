import { useState, useRef, useEffect } from "react";
import { C, BODY, MONO } from "../styles/colors";

interface DetailHeaderProps {
  title: string;
  language: string;
  editMode: boolean;
  onTitleChange: (title: string) => void;
}

export function DetailHeader({
  title,
  language,
  editMode,
  onTitleChange,
}: DetailHeaderProps) {
  const [editing, setEditing] = useState(false);
  const [localTitle, setLocalTitle] = useState(title);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setLocalTitle(title);
    setEditing(false);
  }, [title]);

  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editing]);

  function commitTitle() {
    setEditing(false);
    if (localTitle.trim() && localTitle !== title) {
      onTitleChange(localTitle.trim());
    } else {
      setLocalTitle(title);
    }
  }

  const langColor = C.mauve;

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "0 0 16px",
        borderBottom: `1px solid ${C.border}`,
        marginBottom: 16,
      }}
    >
      {editing || editMode ? (
        <input
          ref={inputRef}
          value={localTitle}
          onChange={(e) => setLocalTitle(e.target.value)}
          onBlur={commitTitle}
          onKeyDown={(e) => {
            if (e.key === "Enter") commitTitle();
            if (e.key === "Escape") {
              setLocalTitle(title);
              setEditing(false);
            }
          }}
          style={{
            flex: 1,
            background: "transparent",
            border: "none",
            outline: "none",
            color: C.text,
            fontFamily: BODY,
            fontSize: 20,
            fontWeight: 700,
          }}
        />
      ) : (
        <h2
          onClick={() => setEditing(true)}
          style={{
            flex: 1,
            color: C.text,
            fontFamily: BODY,
            fontSize: 20,
            fontWeight: 700,
            cursor: "pointer",
          }}
        >
          {title}
        </h2>
      )}

      {language && (
        <span
          style={{
            color: langColor,
            fontFamily: MONO,
            fontSize: 12,
            padding: "3px 10px",
            borderRadius: 8,
            background: `${langColor}15`,
            marginLeft: 12,
            flexShrink: 0,
          }}
        >
          {language}
        </span>
      )}
    </div>
  );
}
