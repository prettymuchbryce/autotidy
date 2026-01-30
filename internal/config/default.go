package config

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/prettymuchbryce/autotidy/internal/pathutil"
)

//go:embed config-example.yaml
var defaultConfigContent string

// EnsureDefaultConfig creates the default config file if it doesn't exist.
// Returns the expanded path and any error encountered.
func EnsureDefaultConfig(configPath string) (string, error) {
	// Expand the path
	expanded := pathutil.ExpandTilde(configPath)

	// Check if file already exists
	if _, err := os.Stat(expanded); err == nil {
		return expanded, nil
	}

	// Create parent directory
	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	// Write default config
	if err := os.WriteFile(expanded, []byte(defaultConfigContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create default config %s: %w", expanded, err)
	}

	slog.Info("created default config", "path", expanded)
	return expanded, nil
}

// CountEnabledRules returns the number of enabled rules in the config.
func (c *Config) CountEnabledRules() int {
	count := 0
	for i := range c.Rules {
		if c.Rules[i].IsEnabled() {
			count++
		}
	}
	return count
}

// IsDefaultConfig checks if the file at the given path matches the default config.
func IsDefaultConfig(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return string(content) == defaultConfigContent
}
