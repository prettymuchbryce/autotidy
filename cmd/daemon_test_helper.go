//go:build integration

package cmd

import (
	"context"

	"github.com/spf13/afero"
	"github.com/prettymuchbryce/autotidy/daemon"
)

// RunDaemon is a test helper that wraps daemon.Run with proper logging setup.
func RunDaemon(ctx context.Context, configPath string, fs afero.Fs) error {
	return daemon.Run(ctx, configPath, fs, SetupLogging)
}
