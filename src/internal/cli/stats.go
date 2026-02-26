package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show collection statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := store.Stats()
			if err != nil {
				return fmt.Errorf("get stats: %w", err)
			}

			fmt.Printf("Total snippets: %d\n", stats.TotalSnippets)
			fmt.Printf("Total tags:     %d\n", stats.TotalTags)

			if len(stats.Languages) > 0 {
				// Sort languages by count descending.
				type langCount struct {
					lang  string
					count int
				}
				var langs []langCount
				for l, c := range stats.Languages {
					langs = append(langs, langCount{l, c})
				}
				sort.Slice(langs, func(i, j int) bool {
					return langs[i].count > langs[j].count
				})

				var parts []string
				for _, lc := range langs {
					parts = append(parts, fmt.Sprintf("%s(%d)", lc.lang, lc.count))
				}
				fmt.Printf("Languages:      %s\n", strings.Join(parts, ", "))
			}

			if stats.MostUsed != nil {
				title := stats.MostUsed.Title
				if title == "" {
					title = stats.MostUsed.ID
				}
				fmt.Printf("Most used:      %s (%d uses)\n", title, stats.MostUsed.UseCount)
			}

			if len(stats.RecentlyAdded) > 0 {
				fmt.Println("Recent:")
				for _, s := range stats.RecentlyAdded {
					title := s.Title
					if title == "" {
						title = s.ID
					}
					fmt.Printf("  %s  %s  %s\n", s.ID, title, s.CreatedAt.Format("2006-01-02"))
				}
			}

			return nil
		},
	}
}
