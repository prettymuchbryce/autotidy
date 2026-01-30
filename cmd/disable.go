package cmd

import (
	"fmt"

	"github.com/prettymuchbryce/autotidy/internal/ipc"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Temporarily pause rule execution",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := ipc.Connect()
		if err != nil {
			return nil
		}
		defer client.Close()

		if err := client.Disable(); err != nil {
			return fmt.Errorf("failed to disable daemon: %w", err)
		}

		fmt.Println("Daemon disabled")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}
