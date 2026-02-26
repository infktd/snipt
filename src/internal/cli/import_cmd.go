package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var (
		overwrite bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import snippets from a JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			snippets, err := parseImportData(data)
			if err != nil {
				return fmt.Errorf("parse import data: %w", err)
			}

			var imported, skipped, overwritten int

			for _, s := range snippets {
				// Generate new ID if missing.
				if s.ID == "" {
					s.ID = model.NewID()
				}

				// Check if snippet already exists.
				existing, _ := store.Get(s.ID)

				if existing != nil {
					if overwrite {
						if dryRun {
							fmt.Printf("[dry-run] would overwrite %s (%s)\n", s.ID, s.Title)
							overwritten++
							continue
						}
						s.CreatedAt = existing.CreatedAt
						if err := store.Update(&s); err != nil {
							return fmt.Errorf("update snippet %s: %w", s.ID, err)
						}
						overwritten++
					} else {
						if dryRun {
							fmt.Printf("[dry-run] would skip %s (%s) — already exists\n", s.ID, s.Title)
						}
						skipped++
						continue
					}
				} else {
					if dryRun {
						fmt.Printf("[dry-run] would import %s (%s)\n", s.ID, s.Title)
						imported++
						continue
					}
					if err := store.Create(&s); err != nil {
						return fmt.Errorf("create snippet %s: %w", s.ID, err)
					}
					imported++
				}
			}

			prefix := ""
			if dryRun {
				prefix = "[dry-run] "
			}
			fmt.Printf("%simported: %d, skipped: %d, overwritten: %d\n", prefix, imported, skipped, overwritten)
			return nil
		},
	}

	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing snippets")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be imported without saving")

	return cmd
}

// parseImportData auto-detects JSON envelope format vs flat array format.
func parseImportData(data []byte) ([]model.Snippet, error) {
	// Try envelope format first: {version, snippets: [...]}
	var envelope exportEnvelope
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Snippets != nil {
		return envelope.Snippets, nil
	}

	// Try flat array format: [{...}, {...}]
	var snippets []model.Snippet
	if err := json.Unmarshal(data, &snippets); err == nil {
		return snippets, nil
	}

	return nil, fmt.Errorf("unrecognized format: expected JSON envelope or array of snippets")
}
