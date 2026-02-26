package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		langFilter string
		tagFilter  string
		pinned     bool
		sort       string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := db.ListOpts{
				Language: langFilter,
				Tag:      tagFilter,
				Sort:     sort,
			}

			if cmd.Flags().Changed("pinned") {
				opts.Pinned = &pinned
			}

			snippets, err := store.List(opts)
			if err != nil {
				return fmt.Errorf("list snippets: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(snippets)
			}

			if len(snippets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no snippets found")
				return nil
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tLANG\tTAGS\tPIN\tUSES")

			for _, s := range snippets {
				title := s.Title
				if len(title) > 40 {
					title = title[:37] + "..."
				}

				pin := ""
				if s.Pinned {
					pin = "*"
				}

				tags := strings.Join(s.Tags, ",")

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n",
					s.ID, title, s.Language, tags, pin, s.UseCount)
			}

			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&langFilter, "lang", "", "filter by language")
	cmd.Flags().StringVar(&tagFilter, "tag", "", "filter by tag")
	cmd.Flags().BoolVar(&pinned, "pinned", false, "filter pinned snippets")
	cmd.Flags().StringVar(&sort, "sort", "created", "sort by: created, updated, usage, title")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}
