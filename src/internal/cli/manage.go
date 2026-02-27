package cli

import (
	"fmt"

	"github.com/infktd/snipt/src/internal/gui"
	"github.com/infktd/snipt/src/internal/tui/manage"
	"github.com/spf13/cobra"
)

func newManageCmd() *cobra.Command {
	var useTUI bool

	cmd := &cobra.Command{
		Use:   "manage",
		Short: "Full-screen snippet manager",
		Long:  "Browse, create, edit, and delete snippets in a full-screen TUI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if useTUI {
				editor := cfg.ResolveEditor()
				if err := manage.RunManage(store, editor); err != nil {
					return fmt.Errorf("manage: %w", err)
				}
				return nil
			}
			return gui.LaunchGUI(store, "manage")
		},
	}

	cmd.Flags().BoolVar(&useTUI, "tui", false, "use terminal UI instead of GUI")
	return cmd
}
