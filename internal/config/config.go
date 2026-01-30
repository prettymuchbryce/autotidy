package config

import (
	"time"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/pathutil"
	"github.com/prettymuchbryce/autotidy/internal/rules"
)

// Config represents the top-level configuration.
type Config struct {
	Rules   []rules.Rule  `yaml:"rules"`
	Daemon  DaemonConfig  `yaml:"daemon"`
	Logging LoggingConfig `yaml:"logging"`
}

// DaemonConfig represents daemon-specific configuration.
type DaemonConfig struct {
	Debounce time.Duration `yaml:"debounce"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// DefaultDaemonConfig returns the default daemon configuration.
func DefaultDaemonConfig() DaemonConfig {
	return DaemonConfig{
		Debounce: 500 * time.Millisecond,
	}
}

// DefaultLoggingConfig returns the default logging configuration.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level: "warn",
	}
}

// Load reads and parses a configuration file using the real filesystem.
func Load(path string) (*Config, error) {
	return LoadWithFs(path, afero.NewOsFs())
}

// LoadWithFs reads and parses a configuration file using the provided filesystem.
// Note: This only uses the fs for reading the config file. Rule.Fs must be set
// separately after loading (e.g., to enable dry-run mode).
func LoadWithFs(path string, afs afero.Fs) (*Config, error) {
	expanded := pathutil.ExpandTilde(path)

	data, err := afero.ReadFile(afs, expanded)
	if err != nil {
		return nil, err
	}

	// Start with defaults
	config := &Config{
		Daemon:  DefaultDaemonConfig(),
		Logging: DefaultLoggingConfig(),
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}
