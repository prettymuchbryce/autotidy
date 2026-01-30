package rules

import (
	"github.com/prettymuchbryce/autotidy/internal/report"
	"gopkg.in/yaml.v3"
)

// FilterGroups defines filters for a rule as a flat list of expressions.
// All expressions are AND'd together (all must match for the file to pass).
// Use `any:` for OR logic and `not:` for negation within expressions.
type FilterGroups struct {
	Exprs []*FilterExpr
}

// Evaluate checks if a path passes all filter expressions.
// All expressions are AND'd together - all must match for the path to pass.
// Empty filters (no expressions) match all paths.
// Short-circuits only when using NullReporter (no reporting needed).
func (fg *FilterGroups) Evaluate(path string, r report.Reporter) (bool, error) {
	// Empty filters match everything
	if len(fg.Exprs) == 0 {
		return true, nil
	}

	// Short-circuit only when not reporting (NullReporter)
	_, canShortCircuit := r.(report.NullReporter)

	allPassed := true

	// All expressions must match (AND)
	for _, expr := range fg.Exprs {
		matched, err := expr.Evaluate(path, r)
		if err != nil {
			return false, err
		}
		if !matched {
			if canShortCircuit {
				return false, nil
			}
			allPassed = false
		}
	}

	return allPassed, nil
}

// UnmarshalYAML implements custom YAML unmarshaling for FilterGroups.
// It expects a sequence of filter expressions.
func (fg *FilterGroups) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return &yaml.TypeError{Errors: []string{"filters must be a list"}}
	}

	fg.Exprs = make([]*FilterExpr, 0, len(node.Content))
	for _, itemNode := range node.Content {
		expr := &FilterExpr{}
		if err := expr.UnmarshalYAML(itemNode); err != nil {
			return err
		}
		fg.Exprs = append(fg.Exprs, expr)
	}

	return nil
}
