package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newPinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pin <ref>",
		Short: "Pin a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "snippet %q not found\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			snippet := &results[0].Snippet
			if err := store.SetPinned(snippet.ID, true); err != nil {
				return fmt.Errorf("pin snippet: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "pinned %s\n", snippet.ID)
			return nil
		},
	}
}

func newUnpinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpin <ref>",
		Short: "Unpin a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "snippet %q not found\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			snippet := &results[0].Snippet
			if err := store.SetPinned(snippet.ID, false); err != nil {
				return fmt.Errorf("unpin snippet: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "unpinned %s\n", snippet.ID)
			return nil
		},
	}
}
