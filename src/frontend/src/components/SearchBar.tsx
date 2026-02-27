import { useRef, forwardRef, useImperativeHandle } from "react";
import { C, BODY } from "../styles/colors";

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
}

export interface SearchBarHandle {
  focus: () => void;
}

export const SearchBar = forwardRef<SearchBarHandle, SearchBarProps>(
  function SearchBar({ value, onChange }, ref) {
    const inputRef = useRef<HTMLInputElement>(null);

    useImperativeHandle(ref, () => ({
      focus: () => inputRef.current?.focus(),
    }));

    return (
      <div
        className="no-select"
        style={{
          padding: "12px 16px",
          borderBottom: `1px solid ${C.border}`,
          display: "flex",
          alignItems: "center",
          gap: 12,
        }}
      >
        <span
          style={{
            background: `linear-gradient(135deg, ${C.pink}, ${C.mauve})`,
            color: C.bg,
            fontFamily: BODY,
            fontWeight: 700,
            fontSize: 11,
            padding: "3px 10px",
            borderRadius: 12,
            letterSpacing: "0.5px",
            flexShrink: 0,
          }}
        >
          SNIPT
        </span>
        <input
          ref={inputRef}
          type="text"
          placeholder="Search snippets..."
          value={value}
          onChange={(e) => onChange(e.target.value)}
          style={{
            flex: 1,
            background: "transparent",
            border: "none",
            outline: "none",
            color: C.text,
            fontFamily: BODY,
            fontSize: 13,
          }}
        />
      </div>
    );
  }
);
