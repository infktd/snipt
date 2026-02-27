package cli

import (
	"fmt"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/gui"
	"github.com/spf13/cobra"
)

var (
	dbPath  string
	noColor bool
	store   *db.Store
	cfg     *config.Config
)

// NewRootCmd creates the root snipt command with all subcommands registered.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:          "snipt",
		Short:        "A CLI snippet manager",
		Version:      version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip DB init for config commands.
			if cmd.Name() == "config" || (cmd.Parent() != nil && cmd.Parent().Name() == "config") {
				return nil
			}

			var err error
			cfg, err = config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			resolvedPath := config.DBPath(dbPath)
			store, err = db.Open(resolvedPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if store != nil {
				store.Close()
			}
		},
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "", "path to database file")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")

	root.AddCommand(newConfigCmd())
	root.AddCommand(newAddCmd())
	root.AddCommand(newGetCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newEditCmd())
	root.AddCommand(newSetCmd())
	root.AddCommand(newNewCmd())
	root.AddCommand(newTagCmd())
	root.AddCommand(newUntagCmd())
	root.AddCommand(newPinCmd())
	root.AddCommand(newUnpinCmd())
	root.AddCommand(newRmCmd())
	root.AddCommand(newStatsCmd())
	root.AddCommand(newExportCmd())
	root.AddCommand(newImportCmd())
	root.AddCommand(newFindCmd())
	root.AddCommand(newManageCmd())

	// Default: launch GUI manage window when no subcommand is given.
	root.RunE = func(cmd *cobra.Command, args []string) error {
		return gui.LaunchGUI(store, "manage")
	}

	return root
}
