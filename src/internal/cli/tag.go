package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <ref> <tags...>",
		Short: "Add tags to a snippet",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			snippet, err := resolveSnippet(cmd, args[0])
			if err != nil {
				return err
			}
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
			snippet, err := resolveSnippet(cmd, args[0])
			if err != nil {
				return err
			}
			tags := args[1:]

			if err := store.RemoveTags(snippet.ID, tags); err != nil {
				return fmt.Errorf("remove tags: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "untagged %s from %v\n", snippet.ID, tags)
			return nil
		},
	}
}
