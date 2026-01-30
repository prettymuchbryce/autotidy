package rules

import (
	"fmt"
	"sort"

	"github.com/prettymuchbryce/autotidy/internal/report"
	"gopkg.in/yaml.v3"
)

// FilterExpr represents a filter expression that can be:
// - A single filter (leaf node)
// - Multiple filters AND'd together (implicit AND)
// - A boolean operator (any/not) with children
type FilterExpr struct {
	// For leaf nodes (regular filters) - AND'd together if multiple
	Filters []Filter

	// For boolean operators
	Any []*FilterExpr // OR: at least one must match
	Not []*FilterExpr // NOT: none must match (children AND'd, then negated)
}

// Evaluate evaluates the filter expression against a path.
// The reporter records filter results for output.
func (fe *FilterExpr) Evaluate(path string, r report.Reporter) (bool, error) {
	// Short-circuit when not reporting
	_, canShortCircuit := r.(report.NullReporter)

	filtersMatched := true
	anyMatched := true
	notMatched := false

	// Evaluate regular filters (AND'd together)
	for _, f := range fe.Filters {
		matched, err := f.Evaluate(path)
		if err != nil {
			return false, err
		}
		r.RecordFilter(f.Name, matched, "")
		filtersMatched = filtersMatched && matched
		if !filtersMatched && canShortCircuit {
			return false, nil
		}
	}

	// Evaluate any: (OR - at least one must match)
	if len(fe.Any) > 0 {
		r.PushOperator("any")
		anyMatched = false
		for _, expr := range fe.Any {
			matched, err := expr.Evaluate(path, r)
			if err != nil {
				return false, err
			}
			anyMatched = anyMatched || matched
			if anyMatched && canShortCircuit {
				break
			}
		}
		r.PopOperator("any", anyMatched)
		if !anyMatched && canShortCircuit {
			return false, nil
		}
	}

	// Evaluate not: (children AND'd, then negated)
	if len(fe.Not) > 0 {
		r.PushOperator("not")
		notMatched = true
		for _, expr := range fe.Not {
			matched, err := expr.Evaluate(path, r)
			if err != nil {
				return false, err
			}
			notMatched = notMatched && matched
			if !notMatched && canShortCircuit {
				break
			}
		}
		r.PopOperator("not", !notMatched)
		if notMatched && canShortCircuit {
			return false, nil
		}
	}

	return filtersMatched && anyMatched && !notMatched, nil
}

// UnmarshalYAML implements custom YAML unmarshaling for FilterExpr.
// It handles:
// - Regular filter keys (extension, name, etc.) -> leaf nodes
// - any: key -> OR operator
// - not: key -> NOT operator
// - Multiple keys in mapping -> implicit AND
func (fe *FilterExpr) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("filter expression must be a mapping, got %v", node.Kind)
	}

	// Process key-value pairs in the mapping
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		var key string
		if err := keyNode.Decode(&key); err != nil {
			return fmt.Errorf("failed to decode key: %w", err)
		}

		switch key {
		case "any":
			// Parse any: as OR operator
			children, err := parseFilterExprList(valueNode)
			if err != nil {
				return fmt.Errorf("failed to parse 'any' operator: %w", err)
			}
			fe.Any = children

		case "not":
			// Parse not: as NOT operator
			children, err := parseFilterExprList(valueNode)
			if err != nil {
				return fmt.Errorf("failed to parse 'not' operator: %w", err)
			}
			fe.Not = children

		default:
			// Regular filter - check if it's registered
			if _, ok := filterRegistry[key]; !ok {
				available := make([]string, 0, len(filterRegistry))
				for k := range filterRegistry {
					available = append(available, k)
				}
				sort.Strings(available)
				return fmt.Errorf("unknown filter %q, available: %v", key, available)
			}

			// Create a filter from this key-value pair
			filter := Filter{}
			// Build a mini mapping node with just this key-value
			filterNode := &yaml.Node{
				Kind:    yaml.MappingNode,
				Content: []*yaml.Node{keyNode, valueNode},
			}
			if err := filter.UnmarshalYAML(filterNode); err != nil {
				return err
			}
			fe.Filters = append(fe.Filters, filter)
		}
	}

	return nil
}

// parseFilterExprList parses a YAML sequence node into a slice of FilterExpr.
func parseFilterExprList(node *yaml.Node) ([]*FilterExpr, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected a list, got %v", node.Kind)
	}

	result := make([]*FilterExpr, 0, len(node.Content))
	for _, itemNode := range node.Content {
		expr := &FilterExpr{}
		if err := expr.UnmarshalYAML(itemNode); err != nil {
			return nil, err
		}
		result = append(result, expr)
	}

	return result, nil
}
