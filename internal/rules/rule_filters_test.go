package rules

import (
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/report"
	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

// newMockFilterExpr creates a FilterExpr with a single mock filter.
func newMockFilterExpr(result bool) *FilterExpr {
	return &FilterExpr{
		Filters: []Filter{{Name: "mock", Inner: &mockEvaluable{result: result}}},
	}
}

func TestFilterGroups_EmptyFilters(t *testing.T) {
	fg := &FilterGroups{}

	passed, err := fg.Evaluate(testutil.Path("/", "any", "path"), report.NullReporter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !passed {
		t.Error("empty filters should match all paths")
	}
}

func TestFilterGroups_SingleFilter(t *testing.T) {
	tests := []struct {
		name     string
		match    bool
		expected bool
	}{
		{"match", true, true},
		{"no match", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fg := &FilterGroups{
				Exprs: []*FilterExpr{newMockFilterExpr(tt.match)},
			}
			passed, err := fg.Evaluate(testutil.Path("/", "test", "path"), report.NullReporter{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if passed != tt.expected {
				t.Errorf("got %v, want %v", passed, tt.expected)
			}
		})
	}
}

func TestFilterGroups_MultipleFilters_AND(t *testing.T) {
	tests := []struct {
		name     string
		filters  []bool
		expected bool
	}{
		{"all match", []bool{true, true}, true},
		{"first fails", []bool{false, true}, false},
		{"second fails", []bool{true, false}, false},
		{"all fail", []bool{false, false}, false},
		{"three all match", []bool{true, true, true}, true},
		{"three one fails", []bool{true, false, true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var exprs []*FilterExpr
			for _, r := range tt.filters {
				exprs = append(exprs, newMockFilterExpr(r))
			}

			fg := &FilterGroups{Exprs: exprs}
			passed, err := fg.Evaluate(testutil.Path("/", "test", "path"), report.NullReporter{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if passed != tt.expected {
				t.Errorf("got %v, want %v", passed, tt.expected)
			}
		})
	}
}

func TestFilterExpr_Any_OR(t *testing.T) {
	tests := []struct {
		name     string
		children []bool
		expected bool
	}{
		{"single match", []bool{true}, true},
		{"single no match", []bool{false}, false},
		{"first matches", []bool{true, false}, true},
		{"second matches", []bool{false, true}, true},
		{"all match", []bool{true, true}, true},
		{"none match", []bool{false, false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var children []*FilterExpr
			for _, r := range tt.children {
				children = append(children, newMockFilterExpr(r))
			}

			expr := &FilterExpr{Any: children}
			result, err := expr.Evaluate(testutil.Path("/", "test", "path"), report.NullReporter{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterExpr_Not(t *testing.T) {
	tests := []struct {
		name     string
		children []bool // Children are AND'd, then negated
		expected bool
	}{
		{"single match negated", []bool{true}, false},
		{"single no match negated", []bool{false}, true},
		{"all match negated", []bool{true, true}, false},
		{"one fails (AND fails) negated", []bool{true, false}, true},
		{"all fail (AND fails) negated", []bool{false, false}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var children []*FilterExpr
			for _, r := range tt.children {
				children = append(children, newMockFilterExpr(r))
			}

			expr := &FilterExpr{Not: children}
			result, err := expr.Evaluate(testutil.Path("/", "test", "path"), report.NullReporter{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterExpr_Combined(t *testing.T) {
	// Test: filter AND any AND not combinations
	tests := []struct {
		name     string
		expr     *FilterExpr
		expected bool
	}{
		{
			name: "filter AND any: true AND (true OR false) = true",
			expr: &FilterExpr{
				Filters: []Filter{{Name: "mock", Inner: &mockEvaluable{result: true}}},
				Any: []*FilterExpr{
					newMockFilterExpr(true),
					newMockFilterExpr(false),
				},
			},
			expected: true,
		},
		{
			name: "filter AND any: true AND (false OR false) = false",
			expr: &FilterExpr{
				Filters: []Filter{{Name: "mock", Inner: &mockEvaluable{result: true}}},
				Any: []*FilterExpr{
					newMockFilterExpr(false),
					newMockFilterExpr(false),
				},
			},
			expected: false,
		},
		{
			name: "filter AND not: true AND NOT(true) = false",
			expr: &FilterExpr{
				Filters: []Filter{{Name: "mock", Inner: &mockEvaluable{result: true}}},
				Not:     []*FilterExpr{newMockFilterExpr(true)},
			},
			expected: false,
		},
		{
			name: "filter AND not: true AND NOT(false) = true",
			expr: &FilterExpr{
				Filters: []Filter{{Name: "mock", Inner: &mockEvaluable{result: true}}},
				Not:     []*FilterExpr{newMockFilterExpr(false)},
			},
			expected: true,
		},
		{
			name: "any AND not: (true OR false) AND NOT(false) = true",
			expr: &FilterExpr{
				Any: []*FilterExpr{
					newMockFilterExpr(true),
					newMockFilterExpr(false),
				},
				Not: []*FilterExpr{newMockFilterExpr(false)},
			},
			expected: true,
		},
		{
			name: "any passes but not fails: (true OR false) AND NOT(true) = false",
			expr: &FilterExpr{
				Any: []*FilterExpr{
					newMockFilterExpr(true),
					newMockFilterExpr(false),
				},
				Not: []*FilterExpr{newMockFilterExpr(true)},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(testutil.Path("/", "test", "path"), report.NullReporter{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterExpr_Nested(t *testing.T) {
	// Test nested operators
	tests := []struct {
		name     string
		expr     *FilterExpr
		expected bool
	}{
		{
			name: "any containing not: (NOT(true) OR true) = true",
			expr: &FilterExpr{
				Any: []*FilterExpr{
					{Not: []*FilterExpr{newMockFilterExpr(true)}},  // NOT(true) = false
					newMockFilterExpr(true),                         // true
				},
			},
			expected: true,
		},
		{
			name: "not containing any: NOT(true OR false) = false",
			expr: &FilterExpr{
				Not: []*FilterExpr{
					{Any: []*FilterExpr{
						newMockFilterExpr(true),
						newMockFilterExpr(false),
					}},
				},
			},
			expected: false,
		},
		{
			name: "not containing any: NOT(false OR false) = true",
			expr: &FilterExpr{
				Not: []*FilterExpr{
					{Any: []*FilterExpr{
						newMockFilterExpr(false),
						newMockFilterExpr(false),
					}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.expr.Evaluate(testutil.Path("/", "test", "path"), report.NullReporter{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
