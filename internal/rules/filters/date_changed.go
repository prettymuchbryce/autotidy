package filters

import (
	"fmt"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/djherbis/times"
	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterFilter("date_changed", deserializeDateChanged)
}

// DateChanged is a filter that matches files by metadata change time (ctime).
type DateChanged struct {
	Before *DateSpec `yaml:"before"`
	After  *DateSpec `yaml:"after"`
}

// Evaluate checks if the file's change time matches the criteria.
func (d *DateChanged) Evaluate(path string) (bool, error) {
	t, err := times.Stat(path)
	if err != nil {
		return false, err
	}

	if !t.HasChangeTime() {
		return false, fmt.Errorf("change time is not available for %s", path)
	}

	changeTime := t.ChangeTime()

	if d.Before != nil {
		threshold, err := d.Before.ToTime()
		if err != nil {
			return false, err
		}
		if !changeTime.Before(threshold) {
			return false, nil
		}
	}

	if d.After != nil {
		threshold, err := d.After.ToTime()
		if err != nil {
			return false, err
		}
		if !changeTime.After(threshold) {
			return false, nil
		}
	}

	return true, nil
}

// deserializeDateChanged creates a DateChanged filter from YAML.
func deserializeDateChanged(node yaml.Node) (rules.Evaluable, error) {
	var d DateChanged
	if err := node.Decode(&d); err != nil {
		return nil, err
	}

	if d.Before == nil && d.After == nil {
		return nil, fmt.Errorf("date_changed filter requires at least one of: before, after")
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
