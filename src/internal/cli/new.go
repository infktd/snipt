package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/infktd/snipt/src/internal/model"
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

			if err := store.Create(snippet); err != nil {
				return fmt.Errorf("save snippet: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "saved %s\n", snippet.ID)
			return nil
		},
	}
}
