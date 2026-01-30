package actions

import (
	"fmt"
	"log/slog"

	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/utils"

	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterAction("log", deserializeLog)
}

// Log is an action that logs a message.
type Log struct {
	Msg   utils.Template
	Level string // debug, info, warn, error
}

// Execute logs the message at the specified level.
func (l *Log) Execute(path string, _ fs.FileSystem) (*rules.ExecutionResult, error) {
	msg := l.Msg.ExpandTilde().ExpandWithNameExt(path).ExpandWithTime().String()

	switch l.Level {
	case "debug":
		slog.Debug(msg, "path", path)
	case "warn":
		slog.Warn(msg, "path", path)
	case "error":
		slog.Error(msg, "path", path)
	default:
		slog.Info(msg, "path", path)
	}

	return nil, nil
}

// deserializeLog creates a Log action from YAML.
// Supports both "log: message" and "log: {msg: message, level: warn}".
func deserializeLog(node yaml.Node) (rules.Executable, error) {
	// Try as plain string first
	if node.Kind == yaml.ScalarNode {
		var msg string
		if err := node.Decode(&msg); err != nil {
			return nil, err
		}
		return &Log{Msg: utils.Template(msg), Level: "info"}, nil
	}

	// Otherwise expect a mapping with "msg" key and optional "level"
	var m struct {
		Msg   utils.Template `yaml:"msg"`
		Level string         `yaml:"level"`
	}
	if err := node.Decode(&m); err != nil {
		return nil, err
	}

	if m.Msg == "" {
		return nil, fmt.Errorf("log action requires msg")
	}

	level := m.Level
	if level == "" {
		level = "info"
	}

	// Validate level
	switch level {
	case "debug", "info", "warn", "error":
		// valid
	default:
		return nil, fmt.Errorf("log action has invalid level: %s (must be debug, info, warn, or error)", level)
	}

	return &Log{Msg: m.Msg, Level: level}, nil
}
