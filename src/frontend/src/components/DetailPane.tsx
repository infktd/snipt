import { useState, useEffect } from "react";
import { DetailHeader } from "./DetailHeader";
import { CodeEditor } from "./CodeEditor";
import { MetadataFooter } from "./MetadataFooter";
import { ActionBar } from "./ActionBar";
import { ConfirmDialog } from "./ConfirmDialog";
import { C, BODY, MONO } from "../styles/colors";
import type { Snippet } from "../state/types";

interface DetailPaneProps {
  snippet: Snippet | null;
  selectedCount: number;
  editMode: boolean;
  createMode: boolean;
  onUpdate: (snippet: Snippet) => void;
  onUpdateTags: (id: string, tags: string[]) => void;
  onDelete: () => void;
  onTogglePin: () => void;
  onBulkCopy: () => void;
  onCreate: (snippet: Partial<Snippet>) => void;
  onSetEditMode: (editing: boolean) => void;
}

export function DetailPane({
  snippet,
  selectedCount,
  editMode,
  createMode,
  onUpdate,
  onUpdateTags,
  onDelete,
  onTogglePin,
  onBulkCopy,
  onCreate,
  onSetEditMode,
}: DetailPaneProps) {
  const [editedContent, setEditedContent] = useState("");
  const [editedTitle, setEditedTitle] = useState("");
  const [editedLanguage, setEditedLanguage] = useState("");
  const [editedTags, setEditedTags] = useState<string[]>([]);
  const [editedPinned, setEditedPinned] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  // Sync local state when snippet changes
  useEffect(() => {
    if (snippet) {
      setEditedContent(snippet.Content);
      setEditedTitle(snippet.Title);
      setEditedLanguage(snippet.Language);
      setEditedTags([...(snippet.Tags ?? [])]);
    }
  }, [snippet]);

  // Reset for create mode
  useEffect(() => {
    if (createMode) {
      setEditedContent("");
      setEditedTitle("");
      setEditedLanguage("");
      setEditedTags([]);
      setEditedPinned(false);
    }
  }, [createMode]);

  // Multi-select summary view
  if (!snippet && !createMode && selectedCount > 1) {
    return (
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          paddingTop: 40,
          gap: 16,
          background: C.bg,
        }}
      >
        <span style={{ color: C.mauve, fontFamily: MONO, fontSize: 32, fontWeight: 700 }}>
          {selectedCount}
        </span>
        <span style={{ color: C.textSub, fontFamily: BODY, fontSize: 14 }}>
          snippets selected
        </span>
        <div style={{ display: "flex", gap: 8, marginTop: 8 }}>
          <button
            onClick={onBulkCopy}
            style={{
              background: `linear-gradient(135deg, ${C.pink}, ${C.mauve})`,
              color: C.bg,
              fontFamily: MONO,
              fontSize: 12,
              fontWeight: 600,
              padding: "8px 18px",
              borderRadius: 6,
              border: "none",
              cursor: "pointer",
            }}
          >
            Copy {selectedCount}
          </button>
          <button
            onClick={() => setShowDeleteConfirm(true)}
            style={{
              background: "transparent",
              color: C.textDim,
              fontFamily: MONO,
              fontSize: 12,
              fontWeight: 500,
              padding: "8px 18px",
              borderRadius: 6,
              border: `1px solid ${C.border}`,
              cursor: "pointer",
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.color = C.red;
              e.currentTarget.style.borderColor = C.red;
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.color = C.textDim;
              e.currentTarget.style.borderColor = C.border;
            }}
          >
            Delete {selectedCount}
          </button>
        </div>

        {showDeleteConfirm && (
          <ConfirmDialog
            title={`Delete ${selectedCount} snippets?`}
            message="This is permanent. All selected snippets will be removed."
            onConfirm={() => {
              onDelete();
              setShowDeleteConfirm(false);
            }}
            onCancel={() => setShowDeleteConfirm(false)}
          />
        )}
      </div>
    );
  }

  // Empty state
  if (!snippet && !createMode) {
    return (
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          paddingTop: 40,
          color: C.textMuted,
          fontFamily: MONO,
          fontSize: 13,
          gap: 8,
        }}
      >
        <span style={{ fontSize: 32 }}>✂</span>
        <span>Select a snippet</span>
        <span>Cmd+N to create</span>
      </div>
    );
  }

  function handleSave() {
    if (createMode) {
      onCreate({
        Title: editedTitle || "Untitled",
        Content: editedContent,
        Language: editedLanguage,
        Tags: editedTags,
        Pinned: editedPinned,
      });
    } else if (snippet) {
      onUpdate({
        ...snippet,
        Title: editedTitle,
        Content: editedContent,
        Language: editedLanguage,
      });
      if (JSON.stringify(editedTags) !== JSON.stringify(snippet.Tags ?? [])) {
        onUpdateTags(snippet.ID, editedTags);
      }
    }
    onSetEditMode(false);
  }

  return (
    <div
      style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        paddingTop: 40,
        overflow: "hidden",
        background: C.bg,
      }}
    >
      <DetailHeader
        title={createMode ? editedTitle : (snippet?.Title ?? "")}
        language={createMode ? editedLanguage : (snippet?.Language ?? "")}
        editMode={editMode}
        onTitleChange={setEditedTitle}
      />

      {/* Language input in create/edit mode */}
      {editMode && (
        <div style={{ margin: "0 24px 12px" }}>
          <input
            value={editedLanguage}
            onChange={(e) => setEditedLanguage(e.target.value)}
            placeholder="Language (go, python, bash...)"
            style={{
              background: "transparent",
              border: `1px solid ${C.border}`,
              borderRadius: 6,
              padding: "6px 12px",
              color: C.text,
              fontFamily: BODY,
              fontSize: 12,
              outline: "none",
              width: 240,
            }}
          />
        </div>
      )}

      <CodeEditor
        content={createMode ? "" : (snippet?.Content ?? "")}
        language={editMode ? editedLanguage : (snippet?.Language ?? "")}
        readOnly={!editMode}
        editMode={editMode}
        onContentChange={setEditedContent}
        onDoubleClick={() => !editMode && onSetEditMode(true)}
      />

      {!createMode && snippet && (
        <MetadataFooter
          tags={editMode ? editedTags : (snippet.Tags ?? [])}
          pinned={snippet.Pinned}
          useCount={snippet.UseCount}
          createdAt={snippet.CreatedAt}
          updatedAt={snippet.UpdatedAt}
          editMode={editMode}
          onTagsChange={setEditedTags}
          onTogglePin={onTogglePin}
        />
      )}

      {createMode && editMode && (
        <MetadataFooter
          tags={editedTags}
          pinned={editedPinned}
          useCount={0}
          createdAt=""
          updatedAt=""
          editMode={true}
          onTagsChange={setEditedTags}
          onTogglePin={() => setEditedPinned((p) => !p)}
        />
      )}

      <ActionBar
        content={editMode ? editedContent : (snippet?.Content ?? "")}
        editMode={editMode}
        onEdit={() => onSetEditMode(true)}
        onSave={handleSave}
        onDelete={() => setShowDeleteConfirm(true)}
        onCancelEdit={() => onSetEditMode(false)}
      />

      {showDeleteConfirm && snippet && (
        <ConfirmDialog
          title={`Delete "${snippet.Title}"?`}
          message="This is permanent. The snippet will be removed from your database."
          onConfirm={() => {
            onDelete();
            setShowDeleteConfirm(false);
          }}
          onCancel={() => setShowDeleteConfirm(false)}
        />
      )}
    </div>
  );
}
