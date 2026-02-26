package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/clipboard"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	var (
		copyToClipboard bool
		idOnly          bool
	)

	cmd := &cobra.Command{
		Use:   "get <id|title>",
		Short: "Output snippet content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snippet, err := resolveSnippet(cmd, args[0])
			if err != nil {
				return err
			}
			if snippet == nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "snippet %q not found\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			// Bump use count.
			_ = store.IncrementUseCount(snippet.ID)

			if idOnly {
				fmt.Fprintln(cmd.OutOrStdout(), snippet.ID)
				return nil
			}

			if copyToClipboard {
				if err := clipboard.Write(snippet.Content); err != nil {
					return fmt.Errorf("copy to clipboard: %w", err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "copied %s to clipboard\n", snippet.ID)
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), snippet.Content)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&copyToClipboard, "clipboard", "c", false, "copy to clipboard instead of stdout")
	cmd.Flags().BoolVarP(&idOnly, "id", "i", false, "output ID only")

	return cmd
}
