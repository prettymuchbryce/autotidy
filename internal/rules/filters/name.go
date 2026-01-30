package filters

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterFilter("name", deserializeName)
}

// Name is a filter that matches files by name pattern.
// Either Glob or Regex should be set, not both.
type Name struct {
	Glob  string
	Regex *regexp.Regexp
}

// Evaluate checks if the file path matches the pattern.
// The pattern is matched against the base filename only.
func (n *Name) Evaluate(path string) (bool, error) {
	filename := filepath.Base(path)

	if n.Regex != nil {
		return n.Regex.MatchString(filename), nil
	}

	matched, err := doublestar.Match(n.Glob, filename)
	if err != nil {
		return false, fmt.Errorf("invalid glob pattern %q: %w", n.Glob, err)
	}
	return matched, nil
}

// deserializeName creates a Name filter from YAML.
// Supports:
//   - "name: foo.jpg" (glob shorthand)
//   - "name: {glob: foo.jpg}"
//   - "name: {regex: '^foo\d+\.jpg$'}"
func deserializeName(node yaml.Node) (rules.Evaluable, error) {
	// Try as plain string first (treated as glob)
	if node.Kind == yaml.ScalarNode {
		var glob string
		if err := node.Decode(&glob); err != nil {
			return nil, err
		}
		return &Name{Glob: glob}, nil
	}

	// Otherwise expect a mapping with "glob" or "regex" key
	var m struct {
		Glob  string `yaml:"glob"`
		Regex string `yaml:"regex"`
	}
	if err := node.Decode(&m); err != nil {
		return nil, err
	}

	if m.Glob != "" && m.Regex != "" {
		return nil, fmt.Errorf("name filter cannot have both glob and regex")
	}

	if m.Regex != "" {
		re, err := regexp.Compile(m.Regex)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern %q: %w", m.Regex, err)
		}
		return &Name{Regex: re}, nil
	}

	return &Name{Glob: m.Glob}, nil
}
