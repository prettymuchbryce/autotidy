package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/prettymuchbryce/autotidy/internal/config"
	"github.com/prettymuchbryce/autotidy/internal/ipc"
	"github.com/spf13/cobra"
)

var (
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	boldStyle      = lipgloss.NewStyle().Bold(true)
	labelStyle     = lipgloss.NewStyle().Width(12)
	boxStyle       = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("2")).
			Padding(0, 4)
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print status information (running, enabled, configuration path)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := ipc.Connect()
		if err != nil {
			return nil
		}
		defer client.Close()

		status, err := client.Status()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		// Show welcome box at top if using default config
		isDefault := config.IsDefaultConfig(status.ConfigPath)
		if isDefault {
			welcome := "👋 Welcome to autotidy\n\n" + "1. Get started by adding rules to the config file at the path below.\n" +
				"2. Reload rules with " + highlightStyle.Render("autotidy reload") + " after making changes."
			fmt.Println(boxStyle.Render(welcome))
		}

		// Build status value
		var statusValue string
		if status.Enabled {
			statusValue = "🟢 running"
		} else {
			statusValue = "🔴 disabled (run " + boldStyle.Render("autotidy enable") + " to resume)"
		}

		// Build rules value
		var rulesValue string
		if isDefault {
			rulesValue = dimStyle.Render("none")
		} else if len(status.Rules) == 0 {
			rulesValue = "⚠️ none"
		} else {
			var ruleLines []string
			for _, rule := range status.Rules {
				var icon string
				if !rule.Enabled {
					icon = "⛔️"
				} else if status.Enabled {
					icon = "🟢"
				} else {
					icon = "🔴"
				}

				// Build stats line if rule has been executed
				var statsLine string
				if rule.LastRunAt != nil && !rule.LastRunAt.IsZero() {
					statsLine = dimStyle.Render(fmt.Sprintf("  last run: %s (%s, %d files",
						formatTimeAgo(*rule.LastRunAt),
						formatDuration(*rule.LastDuration),
						*rule.FilesProcessed,
					))
					if *rule.ErrorCount > 0 {
						statsLine += fmt.Sprintf(", %d errors", *rule.ErrorCount)
					}
					statsLine += dimStyle.Render(")")
				}

				ruleLine := fmt.Sprintf("%s %s", icon, rule.Name)
				if statsLine != "" {
					ruleLine += "\n" + statsLine
				}
				ruleLines = append(ruleLines, ruleLine)
			}
			rulesValue = strings.Join(ruleLines, "\n")
		}

		// Build watching value
		var watchingValue string
		if status.Enabled && status.WatchCount > 0 {
			watchingValue = fmt.Sprintf("%d directories", status.WatchCount)
		} else {
			watchingValue = dimStyle.Render("none")
		}

		// Print status info
		fmt.Println(labelStyle.Render("status") + statusValue)
		fmt.Println(labelStyle.Render("config") + dimStyle.Render(status.ConfigPath))
		fmt.Println(labelStyle.Render("watching") + watchingValue)
		fmt.Println(labelStyle.Render("rules"))
		if rulesValue != "" {
			for _, line := range strings.Split(rulesValue, "\n") {
				fmt.Println("  " + line)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

// formatTimeAgo formats a time as a human-readable relative time.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return fmt.Sprintf("%dµs", d.Microseconds())
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	default:
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
}
