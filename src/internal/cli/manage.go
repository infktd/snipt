package cli

import (
	"fmt"

	"github.com/infktd/snipt/src/internal/tui/manage"
	"github.com/spf13/cobra"
)

func newManageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "Full-screen snippet manager",
		Long:  "Browse, create, edit, and delete snippets in a full-screen TUI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := cfg.ResolveEditor()
			if err := manage.RunManage(store, editor); err != nil {
				return fmt.Errorf("manage: %w", err)
			}
			return nil
		},
	}
}
