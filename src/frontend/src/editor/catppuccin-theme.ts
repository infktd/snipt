import { EditorView } from "@codemirror/view";
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags } from "@lezer/highlight";

const catppuccinMacchiatoTheme = EditorView.theme(
  {
    "&": {
      backgroundColor: "#24273a",
      color: "#cad3f5",
    },
    ".cm-content": {
      fontFamily: '"Berkeley Mono", "JetBrains Mono", "Fira Code", monospace',
      fontSize: "13px",
      caretColor: "#c6a0f6",
      padding: "12px 0",
    },
    ".cm-gutters": {
      backgroundColor: "#1e2030",
      color: "#5b6078",
      borderRight: "1px solid #363a4f",
      paddingLeft: "8px",
    },
    ".cm-activeLineGutter": {
      backgroundColor: "#363a4f",
    },
    ".cm-activeLine": {
      backgroundColor: "rgba(54, 58, 79, 0.4)",
    },
    "&.cm-focused .cm-selectionBackground, .cm-selectionBackground": {
      backgroundColor: "#494d64 !important",
    },
    ".cm-cursor, .cm-dropCursor": {
      borderLeftColor: "#c6a0f6",
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

const catppuccinMacchiatoHighlight = HighlightStyle.define([
  { tag: tags.keyword, color: "#c6a0f6", fontWeight: "600" },
  { tag: tags.string, color: "#a6da95" },
  { tag: tags.comment, color: "#5b6078", fontStyle: "italic" },
  { tag: tags.function(tags.variableName), color: "#8aadf4" },
  { tag: tags.number, color: "#f5a97f" },
  { tag: tags.typeName, color: "#eed49f" },
  { tag: tags.bool, color: "#f5a97f" },
  { tag: tags.operator, color: "#91d7e3" },
  { tag: tags.propertyName, color: "#8aadf4" },
  { tag: tags.punctuation, color: "#8087a2" },
  { tag: tags.className, color: "#eed49f" },
  { tag: tags.definition(tags.variableName), color: "#cad3f5" },
  { tag: tags.variableName, color: "#cad3f5" },
  { tag: tags.tagName, color: "#ed8796" },
  { tag: tags.attributeName, color: "#eed49f" },
  { tag: tags.attributeValue, color: "#a6da95" },
]);

export const catppuccinMocha = [
  catppuccinMacchiatoTheme,
  syntaxHighlighting(catppuccinMacchiatoHighlight),
];
