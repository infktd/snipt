package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/infktd/snipt/src/internal/clipboard"
	"github.com/infktd/snipt/src/internal/lang"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var (
		title         string
		language      string
		tags          string
		desc          string
		source        string
		fromClipboard bool
	)

	cmd := &cobra.Command{
		Use:   "add [file]",
		Short: "Create a new snippet from file, stdin, or clipboard",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string
			var detectedLang string
			var defaultTitle string

			switch {
			case len(args) == 1:
				// Read from file.
				data, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				content = string(data)
				detectedLang = lang.FromExtension(args[0])
				defaultTitle = filepath.Base(args[0])

			case fromClipboard:
				data, err := clipboard.Read()
				if err != nil {
					return fmt.Errorf("read clipboard: %w", err)
				}
				if strings.TrimSpace(data) == "" {
					return fmt.Errorf("clipboard is empty")
				}
				content = data

			default:
				// Check if stdin has data (piped).
				info, err := os.Stdin.Stat()
				if err != nil {
					return fmt.Errorf("stat stdin: %w", err)
				}
				if info.Mode()&os.ModeCharDevice != 0 {
					return fmt.Errorf("no input: provide a file, pipe data, or use --from-clipboard")
				}
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				if strings.TrimSpace(string(data)) == "" {
					return fmt.Errorf("stdin is empty")
				}
				content = string(data)
			}

			// Resolve title.
			if title == "" {
				title = defaultTitle
			}

			// Resolve language: flag > detected from extension > config default.
			snippetLang := language
			if snippetLang == "" {
				snippetLang = detectedLang
			}
			if snippetLang == "" && cfg != nil {
				snippetLang = cfg.DefaultLanguage
			}

			// Parse tags.
			var tagList []string
			if tags != "" {
				for _, t := range strings.Split(tags, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						tagList = append(tagList, t)
					}
				}
			}

			snippet := &model.Snippet{
				ID:          model.NewID(),
				Title:       title,
				Content:     content,
				Language:    snippetLang,
				Description: desc,
				Source:      source,
				Tags:        tagList,
			}

			if err := store.Create(snippet); err != nil {
				return fmt.Errorf("save snippet: %w", err)
			}

			displayTitle := snippet.Title
			if displayTitle == "" {
				displayTitle = snippet.ID
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved %s (%s)\n", snippet.ID, displayTitle)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "snippet title")
	cmd.Flags().StringVar(&language, "lang", "", "programming language")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	cmd.Flags().StringVar(&desc, "desc", "", "snippet description")
	cmd.Flags().StringVar(&source, "source", "", "source URL or reference")
	cmd.Flags().BoolVar(&fromClipboard, "from-clipboard", false, "read content from clipboard")

	return cmd
}
