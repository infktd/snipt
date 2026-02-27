package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/gui"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui/find"
	"github.com/spf13/cobra"
)

func newFindCmd() *cobra.Command {
	var (
		pipe       bool
		langFilter string
		tagFilter  string
		pinned     bool
		idOnly     bool
		stdout     bool
	)

	cmd := &cobra.Command{
		Use:   "find [query]",
		Short: "Quick-find a snippet",
		Long: `Quick-find a snippet. Opens a floating GUI palette by default.
Use --pipe/-p for terminal output (useful for piping).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !pipe {
				return gui.LaunchGUI(store, "find", appVersion)
			}

			// TUI pipe mode (existing behavior).
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

			initialQuery := ""
			if len(args) == 1 {
				initialQuery = args[0]
			}

			result, err := find.RunFind(snippets, initialQuery, idOnly, stdout)
			if err != nil {
				return fmt.Errorf("find: %w", err)
			}

			if result == nil {
				os.Exit(model.ExitInterrupted)
			}

			_ = store.IncrementUseCount(result.ID)

			if idOnly {
				fmt.Fprintln(cmd.OutOrStdout(), result.ID)
				return nil
			}

			if stdout {
				fmt.Fprint(cmd.OutOrStdout(), result.Content)
				return nil
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&pipe, "pipe", "p", false, "output to stdout (TUI mode)")
	cmd.Flags().StringVar(&langFilter, "lang", "", "filter by language")
	cmd.Flags().StringVar(&tagFilter, "tag", "", "filter by tag")
	cmd.Flags().BoolVar(&pinned, "pinned", false, "filter pinned only")
	cmd.Flags().BoolVarP(&idOnly, "id", "i", false, "output snippet ID only")
	cmd.Flags().BoolVar(&stdout, "stdout", false, "print content to stdout instead of clipboard")

	return cmd
}
