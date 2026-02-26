package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pin <ref>",
		Short: "Pin a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snippet, err := resolveSnippet(cmd, args[0])
			if err != nil {
				return err
			}
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
			snippet, err := resolveSnippet(cmd, args[0])
			if err != nil {
				return err
			}
			if err := store.SetPinned(snippet.ID, false); err != nil {
				return fmt.Errorf("unpin snippet: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "unpinned %s\n", snippet.ID)
			return nil
		},
	}
}
