package ipc

import "time"

// Empty is used for RPC methods that don't need arguments or return values.
type Empty struct{}

// StatusData is returned by Daemon.Status.
type StatusData struct {
	ConfigPath  string       `json:"config_path"`
	ConfigValid bool         `json:"config_valid"`
	ConfigError string       `json:"config_error,omitempty"`
	LogPath     string       `json:"log_path,omitempty"`
	Enabled     bool         `json:"enabled"`
	WatchCount  int          `json:"watch_count"`
	Rules       []RuleStatus `json:"rules"`
}

// RuleStatus shows per-rule status information.
type RuleStatus struct {
	Name           string         `json:"name"`
	Enabled        bool           `json:"enabled"`
	Locations      []string       `json:"locations"`
	LastRunAt      *time.Time     `json:"last_run_at,omitempty"`
	LastDuration   *time.Duration `json:"last_duration,omitempty"`
	FilesProcessed *int           `json:"files_processed,omitempty"`
	ErrorCount     *int           `json:"error_count,omitempty"`
}

// ReloadResult is returned by Daemon.Reload.
type ReloadResult struct {
	ConfigPath string `json:"config_path"`
}
