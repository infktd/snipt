package cli

import (
	"fmt"
	"time"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/sync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync snippets with GitHub Gist",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := makeSyncEngine()
			if err != nil {
				return err
			}

			result, err := engine.Sync()
			if err != nil {
				return fmt.Errorf("sync failed: %w", err)
			}

			printSyncResult(cmd, result)
			return updateLastSync()
		},
	}

	cmd.AddCommand(newSyncSetupCmd())
	cmd.AddCommand(newSyncPushCmd())
	cmd.AddCommand(newSyncPullCmd())
	cmd.AddCommand(newSyncStatusCmd())

	return cmd
}

func newSyncSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure Gist sync with a GitHub personal access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(cmd.OutOrStdout(), "GitHub personal access token (gist scope): ")
			var token string
			fmt.Scanln(&token)
			if token == "" {
				return fmt.Errorf("token is required")
			}

			client := sync.NewGistClient(token)
			engine := sync.NewSyncEngine(store, client, &config.SyncConfig{})

			syncCfg, err := engine.Setup(token)
			if err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}

			cfg.Sync = *syncCfg
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Connected as %s\n", syncCfg.Username)
			fmt.Fprintf(cmd.OutOrStdout(), "Gist: https://gist.github.com/%s\n", syncCfg.GistID)
			return nil
		},
	}
}

func newSyncPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push local snippets to Gist",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := makeSyncEngine()
			if err != nil {
				return err
			}

			result, err := engine.Push()
			if err != nil {
				return fmt.Errorf("push failed: %w", err)
			}

			printSyncResult(cmd, result)
			return updateLastSync()
		},
	}
}

func newSyncPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull snippets from Gist",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := makeSyncEngine()
			if err != nil {
				return err
			}

			result, err := engine.Pull()
			if err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}

			printSyncResult(cmd, result)
			return updateLastSync()
		},
	}
}

func newSyncStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if cfg.Sync.GistID == "" {
				fmt.Fprintln(out, "Sync: not configured")
				fmt.Fprintln(out, "Run 'snipt sync setup' to connect a GitHub Gist.")
				return nil
			}

			fmt.Fprintf(out, "Status:    connected\n")
			fmt.Fprintf(out, "Username:  %s\n", cfg.Sync.Username)
			fmt.Fprintf(out, "Gist:      https://gist.github.com/%s\n", cfg.Sync.GistID)
			if cfg.Sync.LastSync != "" {
				fmt.Fprintf(out, "Last sync: %s\n", cfg.Sync.LastSync)
			} else {
				fmt.Fprintf(out, "Last sync: never\n")
			}
			return nil
		},
	}
}

func makeSyncEngine() (*sync.SyncEngine, error) {
	if cfg.Sync.GistID == "" {
		return nil, fmt.Errorf("sync not configured — run 'snipt sync setup' first")
	}
	client := sync.NewGistClient(cfg.Sync.Token)
	return sync.NewSyncEngine(store, client, &cfg.Sync), nil
}

func printSyncResult(cmd *cobra.Command, r *sync.SyncResult) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Pushed: %d  Pulled: %d  Deleted: %d\n", r.Pushed, r.Pulled, r.Deleted)
	if r.Conflicts > 0 {
		fmt.Fprintf(out, "Conflicts: %d\n", r.Conflicts)
	}
	for _, e := range r.Errors {
		fmt.Fprintf(out, "  error: %s\n", e)
	}
}

func updateLastSync() error {
	cfg.Sync.LastSync = time.Now().UTC().Format(time.RFC3339)
	return cfg.Save()
}
