package cli

import (
	"context"
	"fmt"

	"dunk/internal/app"
	"github.com/spf13/cobra"
)

var dryRun, yes, allowSecrets bool

func Execute() error { return rootCmd().Execute() }

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dunk <agent>",
		Short: "Run terminal coding agents in persistent cloud sandboxes",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			if len(args) > 1 {
				return fmt.Errorf("expected one agent name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return app.New(dryRun, yes, allowSecrets).RunAgent(context.Background(), args[0])
		},
	}
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what Dunk would do without creating/updating a sandbox")
	cmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "accept safe prompts")
	cmd.PersistentFlags().BoolVar(&allowSecrets, "allow-secrets", false, "allow explicitly selected secret-looking files to be uploaded")
	cmd.AddCommand(&cobra.Command{Use: "stop", Short: "Kill the active sandbox for this repo worktree", RunE: func(cmd *cobra.Command, args []string) error {
		return app.New(dryRun, yes, allowSecrets).Stop(context.Background())
	}})
	return cmd
}
