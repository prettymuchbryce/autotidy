package daemon

import "log/slog"

// HandleEnable starts the watcher if not running.
func (c *Controller) HandleEnable() {
	if c.watcher != nil {
		return
	}

	if err := c.StartWatcher(); err != nil {
		slog.Error("failed to start watcher", "error", err)
		return
	}
	slog.Info("daemon enabled")
}
