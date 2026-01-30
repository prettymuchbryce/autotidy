package daemon

import (
	"fmt"
	"log/slog"

	"github.com/prettymuchbryce/autotidy/internal/config"
	"github.com/prettymuchbryce/autotidy/internal/ipc"
)

// HandleReload reloads the configuration file.
func (c *Controller) HandleReload() (ipc.ReloadResult, error) {
	cfg, err := config.LoadWithFs(c.configPath, c.fs)
	if err != nil {
		return ipc.ReloadResult{}, fmt.Errorf("failed to load config: %w", err)
	}

	wasEnabled := c.watcher != nil

	c.rules = cfg.Rules
	c.debounce = cfg.Daemon.Debounce

	// Restart watcher with new config if it was running
	if wasEnabled {
		c.StopWatcher()
		if err := c.StartWatcher(); err != nil {
			return ipc.ReloadResult{}, fmt.Errorf("failed to restart watcher: %w", err)
		}
	}

	enabledRules := cfg.CountEnabledRules()
	slog.Info("reloaded config", "path", c.configPath, "rules", len(cfg.Rules), "enabled", enabledRules)

	if enabledRules == 0 {
		slog.Warn("no enabled rules found in config", "path", c.configPath)
	}

	return ipc.ReloadResult{ConfigPath: c.configPath}, nil
}
