package daemon

import "log/slog"

// HandleDisable stops the watcher if running.
func (c *Controller) HandleDisable() {
	if c.watcher == nil {
		return
	}

	c.StopWatcher()
	slog.Info("daemon disabled")
}
