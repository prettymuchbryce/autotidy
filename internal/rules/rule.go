package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/prettymuchbryce/autotidy/internal/pathutil"

	"gopkg.in/yaml.v3"
)

// TraversalMode defines how files and directories are traversed.
type TraversalMode string

const (
	TraversalDepthFirst   TraversalMode = "depth-first"
	TraversalBreadthFirst TraversalMode = "breadth-first"
)

// Rule represents a file organization rule configuration.
type Rule struct {
	Name      string        `yaml:"name"`
	Enabled   *bool         `yaml:"enabled"`   // nil defaults to true
	Recursive *bool         `yaml:"recursive"` // nil defaults to false
	Traversal TraversalMode `yaml:"traversal"` // empty defaults to depth-first
	Locations StringList    `yaml:"locations"`
	Actions   []Action      `yaml:"actions"`
	Filters   *FilterGroups `yaml:"filters"`
}

// UnmarshalYAML decodes the rule and normalizes location paths.
func (r *Rule) UnmarshalYAML(node *yaml.Node) error {
	// Use an alias type to avoid infinite recursion
	type RuleAlias Rule
	var alias RuleAlias
	if err := node.Decode(&alias); err != nil {
		return err
	}
	*r = Rule(alias)

	// Normalize all locations
	for i, loc := range r.Locations {
		// Expand tilde
		loc = pathutil.ExpandTilde(loc)

		// Require absolute paths
		if !filepath.IsAbs(loc) {
			return fmt.Errorf("location must be an absolute path: %s", loc)
		}

		// Clean the path (removes trailing slashes, resolves . and ..)
		r.Locations[i] = filepath.Clean(loc)
	}

	return nil
}

// IsEnabled returns whether the rule is enabled (defaults to true).
func (r *Rule) IsEnabled() bool {
	if r.Enabled == nil {
		return true
	}
	return *r.Enabled
}

// IsRecursive returns whether the rule is recursive (defaults to false).
func (r *Rule) IsRecursive() bool {
	if r.Recursive == nil {
		return false
	}
	return *r.Recursive
}

// GetTraversalMode returns the traversal mode (defaults to depth-first).
func (r *Rule) GetTraversalMode() TraversalMode {
	if r.Traversal == "" {
		return TraversalDepthFirst
	}
	return r.Traversal
}

// CoversPath returns true if the path is covered by this rule's locations.
// Locations are pre-expanded during unmarshaling.
func (r *Rule) CoversPath(path string) bool {
	for _, loc := range r.Locations {
		if path == loc {
			return true
		}

		// For recursive rules, match any descendant
		if r.IsRecursive() && strings.HasPrefix(path, loc+string(filepath.Separator)) {
			return true
		}

		// For non-recursive rules, only match direct children
		if !r.IsRecursive() && filepath.Dir(path) == loc {
			return true
		}
	}

	return false
}
