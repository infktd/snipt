package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <ref> <tags...>",
		Short: "Add tags to a snippet",
		Args:  cobra.MinimumNArgs(2),
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
			tags := args[1:]

			if err := store.AddTags(snippet.ID, tags); err != nil {
				return fmt.Errorf("add tags: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "tagged %s with %v\n", snippet.ID, tags)
			return nil
		},
	}
}

func newUntagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "untag <ref> <tags...>",
		Short: "Remove tags from a snippet",
		Args:  cobra.MinimumNArgs(2),
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
			tags := args[1:]

			if err := store.RemoveTags(snippet.ID, tags); err != nil {
				return fmt.Errorf("remove tags: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "untagged %s from %v\n", snippet.ID, tags)
			return nil
		},
	}
}
