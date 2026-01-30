package filters

import (
	"fmt"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/djherbis/times"
	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterFilter("date_accessed", deserializeDateAccessed)
}

// DateAccessed is a filter that matches files by access time.
type DateAccessed struct {
	Before *DateSpec `yaml:"before"`
	After  *DateSpec `yaml:"after"`
}

// Evaluate checks if the file's access time matches the criteria.
func (d *DateAccessed) Evaluate(path string) (bool, error) {
	t, err := times.Stat(path)
	if err != nil {
		return false, err
	}

	accessTime := t.AccessTime()

	if d.Before != nil {
		threshold, err := d.Before.ToTime()
		if err != nil {
			return false, err
		}
		if !accessTime.Before(threshold) {
			return false, nil
		}
	}

	if d.After != nil {
		threshold, err := d.After.ToTime()
		if err != nil {
			return false, err
		}
		if !accessTime.After(threshold) {
			return false, nil
		}
	}

	return true, nil
}

// deserializeDateAccessed creates a DateAccessed filter from YAML.
func deserializeDateAccessed(node yaml.Node) (rules.Evaluable, error) {
	var d DateAccessed
	if err := node.Decode(&d); err != nil {
		return nil, err
	}

	if d.Before == nil && d.After == nil {
		return nil, fmt.Errorf("date_accessed filter requires at least one of: before, after")
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
