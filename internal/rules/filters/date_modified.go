package filters

import (
	"fmt"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/djherbis/times"
	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterFilter("date_modified", deserializeDateModified)
}

// DateModified is a filter that matches files by modification time.
type DateModified struct {
	Before *DateSpec `yaml:"before"`
	After  *DateSpec `yaml:"after"`
}

// Evaluate checks if the file's modification time matches the criteria.
func (d *DateModified) Evaluate(path string) (bool, error) {
	t, err := times.Stat(path)
	if err != nil {
		return false, err
	}

	modTime := t.ModTime()

	if d.Before != nil {
		threshold, err := d.Before.ToTime()
		if err != nil {
			return false, err
		}
		if !modTime.Before(threshold) {
			return false, nil
		}
	}

	if d.After != nil {
		threshold, err := d.After.ToTime()
		if err != nil {
			return false, err
		}
		if !modTime.After(threshold) {
			return false, nil
		}
	}

	return true, nil
}

// deserializeDateModified creates a DateModified filter from YAML.
func deserializeDateModified(node yaml.Node) (rules.Evaluable, error) {
	var d DateModified
	if err := node.Decode(&d); err != nil {
		return nil, err
	}

	if d.Before == nil && d.After == nil {
		return nil, fmt.Errorf("date_modified filter requires at least one of: before, after")
	}

	if d.Before != nil {
		if err := d.Before.Validate(); err != nil {
			return nil, fmt.Errorf("before: %w", err)
		}
	}

	if d.After != nil {
		if err := d.After.Validate(); err != nil {
			return nil, fmt.Errorf("after: %w", err)
		}
	}

	return &d, nil
}
