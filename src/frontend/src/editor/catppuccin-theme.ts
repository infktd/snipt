import { EditorView } from "@codemirror/view";
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags } from "@lezer/highlight";

const catppuccinMochaTheme = EditorView.theme(
  {
    "&": {
      backgroundColor: "#1e1e2e",
      color: "#cdd6f4",
    },
    ".cm-content": {
      fontFamily: '"Berkeley Mono", "JetBrains Mono", "Fira Code", monospace',
      fontSize: "13px",
      caretColor: "#cba6f7",
      padding: "12px 0",
    },
    ".cm-gutters": {
      backgroundColor: "#242435",
      color: "#45475a",
      border: "none",
      paddingLeft: "8px",
    },
    ".cm-activeLineGutter": {
      backgroundColor: "#2a2a3c",
    },
    ".cm-activeLine": {
      backgroundColor: "#2a2a3c40",
    },
    "&.cm-focused .cm-selectionBackground, .cm-selectionBackground": {
      backgroundColor: "#45475a !important",
    },
    ".cm-cursor, .cm-dropCursor": {
      borderLeftColor: "#cba6f7",
    },
    ".cm-lineNumbers .cm-gutterElement": {
      padding: "0 8px 0 0",
      minWidth: "32px",
    },
    "&.cm-focused": {
      outline: "none",
    },
  },
  { dark: true }
);

const catppuccinMochaHighlight = HighlightStyle.define([
  { tag: tags.keyword, color: "#cba6f7", fontWeight: "600" },
  { tag: tags.string, color: "#a6e3a1" },
  { tag: tags.comment, color: "#45475a", fontStyle: "italic" },
  { tag: tags.function(tags.variableName), color: "#89b4fa" },
  { tag: tags.number, color: "#fab387" },
  { tag: tags.typeName, color: "#f9e2af" },
  { tag: tags.bool, color: "#fab387" },
  { tag: tags.operator, color: "#89dceb" },
  { tag: tags.propertyName, color: "#89b4fa" },
  { tag: tags.punctuation, color: "#6c7086" },
  { tag: tags.className, color: "#f9e2af" },
  { tag: tags.definition(tags.variableName), color: "#cdd6f4" },
  { tag: tags.variableName, color: "#cdd6f4" },
  { tag: tags.tagName, color: "#f38ba8" },
  { tag: tags.attributeName, color: "#f9e2af" },
  { tag: tags.attributeValue, color: "#a6e3a1" },
]);

export const catppuccinMocha = [
  catppuccinMochaTheme,
  syntaxHighlighting(catppuccinMochaHighlight),
];
