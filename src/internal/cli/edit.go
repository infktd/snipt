package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <ref>",
		Short: "Edit snippet content in your editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snippet, err := resolveSnippet(cmd, args[0])
			if err != nil {
				return err
			}

			// Determine file extension for editor syntax highlighting.
			ext := ".txt"
			if snippet.Language != "" {
				ext = langToExt(snippet.Language)
			}

			tmpFile, err := os.CreateTemp("", "snipt-*"+ext)
			if err != nil {
				return fmt.Errorf("create temp file: %w", err)
			}
			tmpPath := tmpFile.Name()
			defer os.Remove(tmpPath)

			if _, err := tmpFile.WriteString(snippet.Content); err != nil {
				tmpFile.Close()
				return fmt.Errorf("write temp file: %w", err)
			}
			tmpFile.Close()

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

			snippet.Content = string(data)
			if err := store.Update(snippet); err != nil {
				if db.IsNotFound(err) {
					fmt.Fprintf(cmd.ErrOrStderr(), "snippet %q not found\n", snippet.ID)
					os.Exit(model.ExitNotFound)
				}
				return fmt.Errorf("update snippet: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "updated %s\n", snippet.ID)
			return nil
		},
	}
}

// langToExt returns a file extension for a language name, for editor syntax highlighting.
func langToExt(language string) string {
	m := map[string]string{
		"go":         ".go",
		"python":     ".py",
		"typescript": ".ts",
		"javascript": ".js",
		"rust":       ".rs",
		"lua":        ".lua",
		"bash":       ".sh",
		"sql":        ".sql",
		"nix":        ".nix",
		"ruby":       ".rb",
		"java":       ".java",
		"c":          ".c",
		"cpp":        ".cpp",
		"markdown":   ".md",
		"toml":       ".toml",
		"yaml":       ".yaml",
		"json":       ".json",
	}
	if ext, ok := m[language]; ok {
		return ext
	}
	return ".txt"
}
