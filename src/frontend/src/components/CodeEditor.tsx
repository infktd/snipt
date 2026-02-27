import { useEffect, useRef } from "react";
import { EditorState, Compartment } from "@codemirror/state";
import { EditorView, lineNumbers } from "@codemirror/view";
import { catppuccinMocha } from "../editor/catppuccin-theme";
import { loadLanguage } from "../editor/languages";

const readOnlyCompartment = new Compartment();
const languageCompartment = new Compartment();

interface CodeEditorProps {
  content: string;
  language: string;
  readOnly: boolean;
  onContentChange?: (content: string) => void;
  onDoubleClick?: () => void;
}

export function CodeEditor({
  content,
  language,
  readOnly,
  onContentChange,
  onDoubleClick,
}: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onContentChangeRef = useRef(onContentChange);
  onContentChangeRef.current = onContentChange;

  // Initialize editor
  useEffect(() => {
    if (!containerRef.current) return;

    const state = EditorState.create({
      doc: content,
      extensions: [
        lineNumbers(),
        ...catppuccinMocha,
        readOnlyCompartment.of(EditorState.readOnly.of(readOnly)),
        languageCompartment.of([]),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            onContentChangeRef.current?.(update.state.doc.toString());
          }
        }),
      ],
    });

    const view = new EditorView({
      state,
      parent: containerRef.current,
    });

    viewRef.current = view;

    loadLanguage(language).then((lang) => {
      if (lang && viewRef.current) {
        viewRef.current.dispatch({
          effects: languageCompartment.reconfigure(lang),
        });
      }
    });

    return () => {
      view.destroy();
      viewRef.current = null;
    };
  }, [content, language, readOnly]);

  // Toggle read-only via compartment reconfiguration (for changes after mount)
  useEffect(() => {
    if (viewRef.current) {
      viewRef.current.dispatch({
        effects: readOnlyCompartment.reconfigure(
          EditorState.readOnly.of(readOnly)
        ),
      });
    }
  }, [readOnly]);

  return (
    <div
      ref={containerRef}
      onDoubleClick={onDoubleClick}
      style={{
        flex: 1,
        overflow: "auto",
        borderRadius: 8,
        border: "1px solid var(--border)",
      }}
    />
  );
}
