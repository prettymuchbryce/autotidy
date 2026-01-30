package cmd

import (
	"fmt"
	"log/slog"

	"github.com/prettymuchbryce/autotidy/internal/config"
	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/pathutil"
	"github.com/prettymuchbryce/autotidy/internal/report"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/spf13/cobra"

	// Import for side effects (filter/action registration)
	_ "github.com/prettymuchbryce/autotidy/internal/rules/actions"
	_ "github.com/prettymuchbryce/autotidy/internal/rules/filters"
)

var (
	runConfigPath string
	runDryRun     bool
	runVerbose    bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Perform a one-off run or dry run of rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		var configPath string
		var err error

		if cmd.Flags().Changed("config") {
			configPath = pathutil.ExpandTilde(runConfigPath)
		} else {
			configPath, err = config.EnsureDefaultConfig(runConfigPath)
			if err != nil {
				return err
			}
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		SetupLogging(cfg.Logging.Level)

		if cfg.CountEnabledRules() == 0 {
			fmt.Printf("No enabled rules found in config: %s\n", configPath)
			return nil
		}

		// Create the appropriate filesystem based on dry-run flag
		var filesystem fs.FileSystem
		if runDryRun {
			filesystem = fs.NewDryRun()
			fmt.Println("Dry-run mode enabled (pass --dry-run=false to perform a one-off run of all rules)")
		} else {
			filesystem = fs.NewReal()
			fmt.Println("Dry-run mode disabled - performing a one-off run of all rules")
		}

		// Create reporter for structured output
		reporter := report.NewStructured(runVerbose)

		for i := range cfg.Rules {
			rule := &cfg.Rules[i]
			runner := rules.NewRuleRunner(rule, filesystem, reporter)
			if _, err := runner.Execute(); err != nil {
				slog.Error("failed to execute rule", "rule", rule.Name, "error", err)
			}
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(&runConfigPath, "config", "c", pathutil.MustDefaultConfigPath(), "path to config file")
	runCmd.Flags().BoolVarP(&runDryRun, "dry-run", "n", true, "simulate changes without applying; use --dry-run=false to apply")
	runCmd.Flags().BoolVarP(&runVerbose, "verbose", "v", false, "show files that didn't match filters")
	rootCmd.AddCommand(runCmd)
}
