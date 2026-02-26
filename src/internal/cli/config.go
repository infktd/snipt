package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Open config file in editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			editor := c.ResolveEditor()
			path := config.ConfigPath()

			parts := strings.Fields(editor)
			name := parts[0]
			editorArgs := append(parts[1:], path)

			proc := exec.Command(name, editorArgs...)
			proc.Stdin = os.Stdin
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			return proc.Run()
		},
	}

	cmd.AddCommand(newConfigPathCmd())
	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), config.ConfigPath())
		},
	}
}
