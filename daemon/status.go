package daemon

import "github.com/prettymuchbryce/autotidy/internal/ipc"

// HandleStatus returns the current daemon status.
func (c *Controller) HandleStatus() ipc.StatusData {
	ruleStatuses := make([]ipc.RuleStatus, len(c.rules))
	for i, rule := range c.rules {
		rs := ipc.RuleStatus{
			Name:      rule.Name,
			Enabled:   rule.IsEnabled(),
			Locations: rule.Locations,
		}
		if ruleState := c.state.GetRuleState(rule.Name); ruleState != nil {
			rs.LastRunAt = &ruleState.LastRunAt
			rs.LastDuration = &ruleState.LastDuration
			rs.FilesProcessed = &ruleState.FilesProcessed
			rs.ErrorCount = &ruleState.ErrorCount
		}
		ruleStatuses[i] = rs
	}

	var watchCount int
	if c.watcher != nil {
		watchCount = c.watcher.WatchCount()
	}

	return ipc.StatusData{
		ConfigPath:  c.configPath,
		ConfigValid: true,
		Enabled:     c.watcher != nil,
		WatchCount:  watchCount,
		Rules:       ruleStatuses,
	}
}
