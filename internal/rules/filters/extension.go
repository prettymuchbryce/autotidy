package filters

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/rules"
)

func init() {
	rules.RegisterFilter("extension", deserializeExtension)
}

// Extension is a filter that matches files by extension.
type Extension struct {
	Extensions []string
}

// Evaluate checks if the file's extension matches any of the patterns.
func (e *Extension) Evaluate(path string) (bool, error) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	// Check against each pattern
	for _, pattern := range e.Extensions {
		// Normalize pattern (remove leading dot if present)
		pattern = strings.TrimPrefix(pattern, ".")

		matched, err := doublestar.Match(pattern, ext)
		if err != nil {
			return false, fmt.Errorf("invalid extension pattern %q: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

// deserializeExtension creates an Extension filter from YAML.
// Supports:
//   - "extension: txt"
//   - "extension: .txt"
//   - "extension: [txt, md, json]"
//   - "extension: {extensions: txt}"
//   - "extension: {extensions: [txt, md]}"
func deserializeExtension(node yaml.Node) (rules.Evaluable, error) {
	// Try as plain string first
	if node.Kind == yaml.ScalarNode {
		var ext string
		if err := node.Decode(&ext); err != nil {
			return nil, err
		}
		return &Extension{Extensions: []string{ext}}, nil
	}

	// Try as sequence of strings
	if node.Kind == yaml.SequenceNode {
		var exts []string
		if err := node.Decode(&exts); err != nil {
			return nil, err
		}
		return &Extension{Extensions: exts}, nil
	}

	// Otherwise expect a mapping with "extensions" key
	var m struct {
		Extensions rules.StringList `yaml:"extensions"`
	}
	if err := node.Decode(&m); err != nil {
		return nil, err
	}

	if len(m.Extensions) == 0 {
		return nil, fmt.Errorf("extension filter requires at least one extension pattern")
	}

	return &Extension{Extensions: m.Extensions}, nil
}
