package report

// Reporter provides structured output for rule execution.
// Implementations can format output as tree-style text, JSON, etc.
// All methods are nil-safe - they are no-ops when called on nil.
type Reporter interface {
	// StartRule begins reporting for a rule
	StartRule(name string)

	// EndRule finishes reporting for current rule.
	// Returns the number of files that matched filters.
	EndRule() int

	// StartFile begins reporting for a file
	StartFile(path string)

	// RecordFilter records a single filter evaluation result.
	// Called during filter evaluation to build up the filter tree.
	RecordFilter(name string, matched bool, detail string)

	// PushOperator starts a new operator group (e.g., "any", "not").
	// Subsequent RecordFilter calls add to this group until PopOperator.
	PushOperator(op string)

	// PopOperator ends the current operator group with its final result.
	PopOperator(op string, matched bool)

	// ReportAction records an action execution result
	ReportAction(name string, result ActionResult)

	// MarkFiltersPassed marks that all filters passed for the current file.
	// Called after filter evaluation when all filters passed.
	MarkFiltersPassed()

	// EndFile finishes reporting for current file.
	// Returns true if anything was reported (filters matched or actions ran).
	EndFile() bool
}

// ActionOutcome represents the type of action result.
type ActionOutcome int

const (
	OutcomeSuccess ActionOutcome = iota // Action completed successfully
	OutcomeMoved                        // File was moved/renamed to NewPath
	OutcomeDeleted                      // File was deleted
	OutcomeSkipped                      // Skipped due to conflict (destination exists)
	OutcomeFailed                       // Action failed with error
)

// ActionResult describes the outcome of an action execution.
type ActionResult struct {
	Outcome ActionOutcome
	NewPath string // Destination path for move/copy/rename
	Error   string // Error message for failed actions
}

// FilterDetail describes a filter or operator result with optional children.
type FilterDetail struct {
	Name     string         // Filter name (e.g., "extension") or operator ("any", "not")
	Matched  bool           // Whether this filter/operator matched
	Detail   string         // Human-readable detail
	Children []FilterDetail // For operators, the nested filter results
}
