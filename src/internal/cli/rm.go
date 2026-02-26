package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "rm <ref>",
		Short: "Delete a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "snippet %q not found\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			snippet := &results[0].Snippet

			if !force {
				// Check if stdout is a TTY.
				info, err := os.Stdout.Stat()
				if err != nil || info.Mode()&os.ModeCharDevice == 0 {
					return fmt.Errorf("non-TTY environment requires --force")
				}

				fmt.Printf("Delete %q (%s)? [y/N] ", snippet.Title, snippet.ID)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))

				if answer != "y" && answer != "yes" {
					fmt.Println("cancelled")
					return nil
				}
			}

			if err := store.Delete(snippet.ID); err != nil {
				return fmt.Errorf("delete snippet: %w", err)
			}

			fmt.Printf("deleted %s\n", snippet.ID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation")

	return cmd
}
