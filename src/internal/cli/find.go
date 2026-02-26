package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/clipboard"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui"
	"github.com/spf13/cobra"
)

func newFindCmd() *cobra.Command {
	var (
		clipOutput bool
		langFilter string
		tagFilter  string
		pinned     bool
		idOnly     bool
	)

	cmd := &cobra.Command{
		Use:   "find [query]",
		Short: "Fuzzy search snippets",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load snippets from store with filters.
			opts := db.ListOpts{
				Language: langFilter,
				Tag:      tagFilter,
			}
			if cmd.Flags().Changed("pinned") {
				opts.Pinned = &pinned
			}

			snippets, err := store.List(opts)
			if err != nil {
				return fmt.Errorf("load snippets: %w", err)
			}

			// Get initial query if provided.
			initialQuery := ""
			if len(args) == 1 {
				initialQuery = args[0]
			}

			// Launch TUI.
			result, err := tui.RunFind(snippets, initialQuery, idOnly, clipOutput)
			if err != nil {
				return fmt.Errorf("find: %w", err)
			}

			if result == nil {
				os.Exit(model.ExitInterrupted)
			}

			// Bump use count.
			_ = store.IncrementUseCount(result.ID)

			// Output.
			if idOnly {
				fmt.Fprintln(cmd.OutOrStdout(), result.ID)
				return nil
			}

			if clipOutput {
				if err := clipboard.Write(result.Content); err != nil {
					return fmt.Errorf("copy to clipboard: %w", err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "copied %s to clipboard\n", result.ID)
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), result.Content)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&clipOutput, "clipboard", "c", false, "copy to clipboard")
	cmd.Flags().StringVar(&langFilter, "lang", "", "filter by language")
	cmd.Flags().StringVar(&tagFilter, "tag", "", "filter by tag")
	cmd.Flags().BoolVar(&pinned, "pinned", false, "filter pinned only")
	cmd.Flags().BoolVarP(&idOnly, "id", "i", false, "output ID instead of content")

	return cmd
}
