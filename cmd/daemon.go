package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/prettymuchbryce/autotidy/daemon"
	"github.com/prettymuchbryce/autotidy/internal/config"
	"github.com/prettymuchbryce/autotidy/internal/pathutil"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	// Import for side effects (filter/action registration)
	_ "github.com/prettymuchbryce/autotidy/internal/rules/actions"
	_ "github.com/prettymuchbryce/autotidy/internal/rules/filters"
)

var daemonConfigPath string

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Hidden: true,
	Short:  "Watch files and execute rules on changes",
	Long: `Start a long-running process that watches configured locations
and executes matching rules when files change.

Includes debouncing to avoid reacting to rapid successive changes,
and graceful shutdown on SIGINT/SIGTERM.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		var configPath string
		var err error

		if cmd.Flags().Changed("config") {
			configPath = pathutil.ExpandTilde(daemonConfigPath)
		} else {
			configPath, err = config.EnsureDefaultConfig(daemonConfigPath)
			if err != nil {
				return err
			}
		}

		return daemon.Run(ctx, configPath, afero.NewOsFs(), SetupLogging)
	},
}

func init() {
	daemonCmd.Flags().StringVarP(&daemonConfigPath, "config", "c", pathutil.MustDefaultConfigPath(), "path to config file")
	rootCmd.AddCommand(daemonCmd)
}
