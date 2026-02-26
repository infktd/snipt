package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui/form"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Create a new snippet in your editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			tmpFile, err := os.CreateTemp("", "snipt-new-*.txt")
			if err != nil {
				return fmt.Errorf("create temp file: %w", err)
			}
			tmpPath := tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpPath)

			editor := cfg.ResolveEditor()
			parts := strings.Fields(editor)
			name := parts[0]
			editorArgs := append(parts[1:], tmpPath)

			proc := exec.Command(name, editorArgs...)
			proc.Stdin = os.Stdin
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			if err := proc.Run(); err != nil {
				return fmt.Errorf("editor: %w", err)
			}

			data, err := os.ReadFile(tmpPath)
			if err != nil {
				return fmt.Errorf("read temp file: %w", err)
			}

			content := string(data)
			if strings.TrimSpace(content) == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "empty content, nothing saved")
				return nil
			}

			snippet := &model.Snippet{
				ID:       model.NewID(),
				Content:  content,
				Language: cfg.DefaultLanguage,
			}

			// Launch the metadata form to collect title, language, tags, and description.
			formResult, err := form.RunForm(cfg.DefaultLanguage)
			if err != nil {
				return fmt.Errorf("metadata form: %w", err)
			}

			if !formResult.Cancelled {
				if formResult.Title != "" {
					snippet.Title = formResult.Title
				}
				if formResult.Language != "" {
					snippet.Language = formResult.Language
				}
				if formResult.Description != "" {
					snippet.Description = formResult.Description
				}
				if formResult.Tags != "" {
					snippet.Tags = parseTags(formResult.Tags)
				}
			}

			if err := store.Create(snippet); err != nil {
				return fmt.Errorf("save snippet: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "saved %s\n", snippet.ID)
			return nil
		},
	}
}

// parseTags splits a comma-separated tag string into a slice of trimmed, non-empty tags.
func parseTags(s string) []string {
	raw := strings.Split(s, ",")
	tags := make([]string, 0, len(raw))
	for _, t := range raw {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
