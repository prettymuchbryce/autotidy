package filters

import (
	"fmt"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/djherbis/times"
	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterFilter("date_created", deserializeDateCreated)
}

// DateCreated is a filter that matches files by creation/birth time.
type DateCreated struct {
	Before *DateSpec `yaml:"before"`
	After  *DateSpec `yaml:"after"`
}

// Evaluate checks if the file's creation time matches the criteria.
func (d *DateCreated) Evaluate(path string) (bool, error) {
	t, err := times.Stat(path)
	if err != nil {
		return false, err
	}

	if !t.HasBirthTime() {
		return false, fmt.Errorf("creation time is not available for %s", path)
	}

	birthTime := t.BirthTime()

	if d.Before != nil {
		threshold, err := d.Before.ToTime()
		if err != nil {
			return false, err
		}
		if !birthTime.Before(threshold) {
			return false, nil
		}
	}

	if d.After != nil {
		threshold, err := d.After.ToTime()
		if err != nil {
			return false, err
		}
		if !birthTime.After(threshold) {
			return false, nil
		}
	}

	return true, nil
}

// deserializeDateCreated creates a DateCreated filter from YAML.
func deserializeDateCreated(node yaml.Node) (rules.Evaluable, error) {
	var d DateCreated
	if err := node.Decode(&d); err != nil {
		return nil, err
	}

	if d.Before == nil && d.After == nil {
		return nil, fmt.Errorf("date_created filter requires at least one of: before, after")
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
