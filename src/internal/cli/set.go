package cli

import (
	"fmt"
	"os"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	var (
		title    string
		language string
		desc     string
		source   string
	)

	cmd := &cobra.Command{
		Use:   "set <ref>",
		Short: "Modify snippet metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Require at least one flag.
			if !cmd.Flags().Changed("title") &&
				!cmd.Flags().Changed("lang") &&
				!cmd.Flags().Changed("desc") &&
				!cmd.Flags().Changed("source") {
				return fmt.Errorf("at least one flag is required: --title, --lang, --desc, --source")
			}

			results, err := store.ResolveRef(args[0])
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintf(os.Stderr, "snippet %q not found\n", args[0])
				os.Exit(model.ExitNotFound)
			}

			snippet := &results[0].Snippet

			if cmd.Flags().Changed("title") {
				snippet.Title = title
			}
			if cmd.Flags().Changed("lang") {
				snippet.Language = language
			}
			if cmd.Flags().Changed("desc") {
				snippet.Description = desc
			}
			if cmd.Flags().Changed("source") {
				snippet.Source = source
			}

			if err := store.Update(snippet); err != nil {
				if db.IsNotFound(err) {
					fmt.Fprintf(os.Stderr, "snippet %q not found\n", snippet.ID)
					os.Exit(model.ExitNotFound)
				}
				return fmt.Errorf("update snippet: %w", err)
			}

			fmt.Printf("updated %s\n", snippet.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "set title")
	cmd.Flags().StringVar(&language, "lang", "", "set language")
	cmd.Flags().StringVar(&desc, "desc", "", "set description")
	cmd.Flags().StringVar(&source, "source", "", "set source URL or reference")

	return cmd
}
