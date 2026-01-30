package cmd

import (
	"fmt"

	"github.com/prettymuchbryce/autotidy/internal/ipc"
	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload all rules from the configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := ipc.Connect()
		if err != nil {
			return nil
		}
		defer client.Close()

		result, err := client.Reload()
		if err != nil {
			return fmt.Errorf("failed to reload config: %w", err)
		}

		fmt.Printf("Reloaded %s\n", result.ConfigPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reloadCmd)
}
