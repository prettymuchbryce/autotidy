package cmd

import (
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
)

// SetupLogging configures slog with charmbracelet/log for colorful output.
func SetupLogging(levelStr string) {
	var level log.Level
	switch levelStr {
	case "debug":
		level = log.DebugLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	default:
		level = log.InfoLevel
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{
		Level:           level,
		ReportTimestamp: true,
	})

	slog.SetDefault(slog.New(logger))
}
