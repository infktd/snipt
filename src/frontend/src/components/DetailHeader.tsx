import { useState, useRef, useEffect } from "react";
import { C, BODY, MONO } from "../styles/colors";

interface DetailHeaderProps {
  title: string;
  language: string;
  editMode: boolean;
  onTitleChange: (title: string) => void;
}

const langColors: Record<string, string> = {
  go: C.sky,
  javascript: C.yellow,
  js: C.yellow,
  typescript: C.blue,
  ts: C.blue,
  python: C.green,
  py: C.green,
  bash: C.green,
  sh: C.green,
  sql: C.yellow,
  html: C.red,
  css: C.sky,
  json: C.peach,
  markdown: C.teal,
  md: C.teal,
  rust: C.peach,
  ruby: C.red,
  yaml: C.peach,
  toml: C.lavender,
  nix: C.mauve,
};

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

  const langColor = langColors[language.toLowerCase()] ?? C.mauve;

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "20px 24px 16px",
        borderBottom: `1px solid ${C.borderSubtle}`,
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
            fontSize: 22,
            fontWeight: 600,
          }}
        />
      ) : (
        <h2
          onClick={() => setEditing(true)}
          style={{
            flex: 1,
            color: C.text,
            fontFamily: BODY,
            fontSize: 22,
            fontWeight: 600,
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
            fontWeight: 500,
            padding: "4px 12px",
            borderRadius: 6,
            background: `${langColor}1F`,
            border: `1px solid ${langColor}33`,
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
