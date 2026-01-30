package cmd

import (
	"fmt"

	"github.com/prettymuchbryce/autotidy/internal/ipc"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Resume rule execution (if it was previously disabled)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := ipc.Connect()
		if err != nil {
			return nil
		}
		defer client.Close()

		if err := client.Enable(); err != nil {
			return fmt.Errorf("failed to enable daemon: %w", err)
		}

		fmt.Println("Daemon enabled")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
}
