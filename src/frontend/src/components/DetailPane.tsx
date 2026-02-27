import { useState, useEffect } from "react";
import { DetailHeader } from "./DetailHeader";
import { CodeEditor } from "./CodeEditor";
import { MetadataFooter } from "./MetadataFooter";
import { ActionBar } from "./ActionBar";
import { ConfirmDialog } from "./ConfirmDialog";
import { C, BODY } from "../styles/colors";
import type { Snippet } from "../state/types";

interface DetailPaneProps {
  snippet: Snippet | null;
  editMode: boolean;
  createMode: boolean;
  onUpdate: (snippet: Snippet) => void;
  onUpdateTags: (id: string, tags: string[]) => void;
  onDelete: (id: string) => void;
  onTogglePin: (id: string, pinned: boolean) => void;
  onCreate: (snippet: Partial<Snippet>) => void;
  onSetEditMode: (editing: boolean) => void;
}

export function DetailPane({
  snippet,
  editMode,
  createMode,
  onUpdate,
  onUpdateTags,
  onDelete,
  onTogglePin,
  onCreate,
  onSetEditMode,
}: DetailPaneProps) {
  const [editedContent, setEditedContent] = useState("");
  const [editedTitle, setEditedTitle] = useState("");
  const [editedLanguage, setEditedLanguage] = useState("");
  const [editedTags, setEditedTags] = useState<string[]>([]);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  // Sync local state when snippet changes
  useEffect(() => {
    if (snippet) {
      setEditedContent(snippet.Content);
      setEditedTitle(snippet.Title);
      setEditedLanguage(snippet.Language);
      setEditedTags([...snippet.Tags]);
    }
  }, [snippet]);

  // Reset for create mode
  useEffect(() => {
    if (createMode) {
      setEditedContent("");
      setEditedTitle("");
      setEditedLanguage("");
      setEditedTags([]);
    }
  }, [createMode]);

  if (!snippet && !createMode) {
    return (
      <div
        style={{
          flex: 1,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          paddingTop: 40,
          color: C.textMuted,
          fontFamily: BODY,
          fontSize: 14,
        }}
      >
        Select a snippet or create a new one
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
      });
    } else if (snippet) {
      onUpdate({
        ...snippet,
        Title: editedTitle,
        Content: editedContent,
        Language: editedLanguage,
      });
      if (JSON.stringify(editedTags) !== JSON.stringify(snippet.Tags)) {
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
        padding: "40px 24px 16px 24px",
        overflow: "hidden",
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
        <div style={{ marginBottom: 12 }}>
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
        onContentChange={setEditedContent}
        onDoubleClick={() => !editMode && onSetEditMode(true)}
      />

      {!createMode && snippet && (
        <MetadataFooter
          tags={editMode ? editedTags : snippet.Tags}
          pinned={snippet.Pinned}
          useCount={snippet.UseCount}
          createdAt={snippet.CreatedAt}
          updatedAt={snippet.UpdatedAt}
          editMode={editMode}
          onTagsChange={setEditedTags}
          onTogglePin={() => onTogglePin(snippet.ID, !snippet.Pinned)}
        />
      )}

      {createMode && editMode && (
        <div style={{ paddingTop: 16 }}>
          <MetadataFooter
            tags={editedTags}
            pinned={false}
            useCount={0}
            createdAt=""
            updatedAt=""
            editMode={true}
            onTagsChange={setEditedTags}
            onTogglePin={() => {}}
          />
        </div>
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
            onDelete(snippet.ID);
            setShowDeleteConfirm(false);
          }}
          onCancel={() => setShowDeleteConfirm(false)}
        />
      )}
    </div>
  );
}
