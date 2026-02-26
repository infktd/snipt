package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui"
	"github.com/spf13/cobra"
)

func newFindCmd() *cobra.Command {
	var (
		langFilter string
		tagFilter  string
		pinned     bool
		idOnly     bool
		stdout     bool
	)

	cmd := &cobra.Command{
		Use:   "find [query]",
		Short: "Fuzzy search snippets",
		Long: `Fuzzy search snippets and copy the selected one to clipboard.

By default, the selected snippet is copied to the system clipboard
via OSC52 and the command exits silently. Use --stdout to print
the content to stdout instead (useful for piping).`,
		Args: cobra.MaximumNArgs(1),
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

			// Launch TUI. Clipboard copy happens inside via OSC52.
			result, err := tui.RunFind(snippets, initialQuery, idOnly, stdout)
			if err != nil {
				return fmt.Errorf("find: %w", err)
			}

			if result == nil {
				os.Exit(model.ExitInterrupted)
			}

			// Bump use count.
			_ = store.IncrementUseCount(result.ID)

			// Output based on flags.
			if idOnly {
				fmt.Fprintln(cmd.OutOrStdout(), result.ID)
				return nil
			}

			if stdout {
				fmt.Fprint(cmd.OutOrStdout(), result.Content)
				return nil
			}

			// Default: silent. Clipboard was already set by the TUI.
			return nil
		},
	}

	cmd.Flags().StringVar(&langFilter, "lang", "", "filter by language")
	cmd.Flags().StringVar(&tagFilter, "tag", "", "filter by tag")
	cmd.Flags().BoolVar(&pinned, "pinned", false, "filter pinned only")
	cmd.Flags().BoolVarP(&idOnly, "id", "i", false, "output snippet ID only")
	cmd.Flags().BoolVar(&stdout, "stdout", false, "print content to stdout instead of clipboard")

	return cmd
}
