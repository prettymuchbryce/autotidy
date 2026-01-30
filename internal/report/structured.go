package report

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/xlab/treeprint"
)

// Styles for the structured reporter
var (
	ruleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")) // Cyan
	pathStyle   = lipgloss.NewStyle().Bold(true)
	passStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	failStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
	skipStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	detailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
)

const (
	passIcon = "✓"
	failIcon = "✗"
	skipIcon = "⊘"
)

// actionEntry stores action data for deferred formatting.
type actionEntry struct {
	name   string
	result ActionResult
}

// StructuredReporter outputs a tree-style report of rule execution.
type StructuredReporter struct {
	w       io.Writer
	verbose bool

	// Current rule state
	currentRule  string
	matchedFiles int

	// Current file state
	currentPath      string
	filterRoot       []FilterDetail   // top-level filter results
	filterStack      []*FilterDetail  // stack for building nested operator groups
	actions          []actionEntry
	hasMatchOrAction bool
}

// NewStructured creates a new StructuredReporter.
func NewStructured(verbose bool) *StructuredReporter {
	return &StructuredReporter{
		w:       os.Stdout,
		verbose: verbose,
	}
}

// NewStructuredWithWriter creates a StructuredReporter writing to a custom writer.
func NewStructuredWithWriter(w io.Writer, verbose bool) *StructuredReporter {
	return &StructuredReporter{
		w:       w,
		verbose: verbose,
	}
}

// StartRule begins reporting for a rule.
func (r *StructuredReporter) StartRule(name string) {
	r.currentRule = name
	r.matchedFiles = 0
	fmt.Fprintf(r.w, "\n%s\n", ruleStyle.Render("━━━ Rule: "+name+" ━━━"))
}

// EndRule finishes reporting for current rule.
// Returns the number of files that matched filters.
func (r *StructuredReporter) EndRule() int {
	if r.matchedFiles == 0 && !r.verbose {
		fmt.Fprintf(r.w, "%s\n", detailStyle.Render("  No files matched. Use --verbose for more info."))
	}
	return r.matchedFiles
}

// StartFile begins reporting for a file.
func (r *StructuredReporter) StartFile(path string) {
	if r == nil {
		return
	}
	r.currentPath = path
	r.filterRoot = nil
	r.filterStack = nil
	r.actions = nil
	r.hasMatchOrAction = false
}

// RecordFilter records a single filter evaluation result.
func (r *StructuredReporter) RecordFilter(name string, matched bool, detail string) {
	if r == nil {
		return
	}
	d := FilterDetail{Name: name, Matched: matched, Detail: detail}

	if len(r.filterStack) > 0 {
		// Add to current operator's children
		parent := r.filterStack[len(r.filterStack)-1]
		parent.Children = append(parent.Children, d)
	} else {
		// Add to root
		r.filterRoot = append(r.filterRoot, d)
	}
}

// PushOperator starts a new operator group (e.g., "any", "not").
func (r *StructuredReporter) PushOperator(op string) {
	if r == nil {
		return
	}
	d := &FilterDetail{Name: op}
	r.filterStack = append(r.filterStack, d)
}

// PopOperator ends the current operator group with its final result.
func (r *StructuredReporter) PopOperator(op string, matched bool) {
	if r == nil {
		return
	}
	if len(r.filterStack) == 0 {
		return
	}

	// Pop from stack
	n := len(r.filterStack)
	current := r.filterStack[n-1]
	current.Matched = matched
	r.filterStack = r.filterStack[:n-1]

	if len(r.filterStack) > 0 {
		// Nest under parent operator
		parent := r.filterStack[len(r.filterStack)-1]
		parent.Children = append(parent.Children, *current)
	} else {
		// Add to root
		r.filterRoot = append(r.filterRoot, *current)
	}
}

// ReportAction records an action execution result.
func (r *StructuredReporter) ReportAction(name string, result ActionResult) {
	if r == nil {
		return
	}
	r.actions = append(r.actions, actionEntry{name: name, result: result})
	r.hasMatchOrAction = true
}

// MarkFiltersPassed marks that all filters passed (called after filter evaluation).
func (r *StructuredReporter) MarkFiltersPassed() {
	if r == nil {
		return
	}
	r.hasMatchOrAction = true
}

// EndFile finishes reporting for current file.
// Returns true if anything was reported.
func (r *StructuredReporter) EndFile() bool {
	if r == nil || r.currentPath == "" {
		return false
	}

	// Decide what to print based on verbose mode
	if r.hasMatchOrAction {
		r.matchedFiles++
		r.printFile()
		return true
	}

	// In verbose mode, show files that didn't match filters
	if r.verbose && len(r.filterRoot) > 0 {
		r.printFile()
		return true
	}

	return false
}

// printFile outputs the full file report with tree connectors.
func (r *StructuredReporter) printFile() {
	tree := treeprint.NewWithRoot(pathStyle.Render(r.currentPath))

	// Calculate max name width and depth for alignment
	maxWidth := r.calculateMaxWidth()
	maxDepth := r.calculateMaxDepth()

	if len(r.filterRoot) > 0 {
		filtersBranch := tree.AddBranch("filters:")
		r.addFilterDetails(filtersBranch, r.filterRoot, maxWidth, 0, maxDepth)
	}

	if len(r.actions) > 0 {
		actionsBranch := tree.AddBranch("actions:")
		// Actions are at depth 0, add padding to align with deepest filters
		extraPadding := maxDepth * 4
		for _, a := range r.actions {
			actionsBranch.AddNode(r.formatActionWithPadding(a, maxWidth, extraPadding))
		}
	}

	fmt.Fprint(r.w, tree.String())
}

// calculateMaxWidth calculates the maximum name width for alignment.
func (r *StructuredReporter) calculateMaxWidth() int {
	maxWidth := 0

	// Check filter tree
	var checkDetails func(details []FilterDetail)
	checkDetails = func(details []FilterDetail) {
		for _, d := range details {
			if len(d.Name) > maxWidth {
				maxWidth = len(d.Name)
			}
			if len(d.Children) > 0 {
				checkDetails(d.Children)
			}
		}
	}
	checkDetails(r.filterRoot)

	// Check actions
	for _, a := range r.actions {
		if len(a.name) > maxWidth {
			maxWidth = len(a.name)
		}
	}

	return maxWidth
}

// calculateMaxDepth calculates the maximum nesting depth of filter details.
func (r *StructuredReporter) calculateMaxDepth() int {
	var maxDepth func(details []FilterDetail, depth int) int
	maxDepth = func(details []FilterDetail, depth int) int {
		max := depth
		for _, d := range details {
			if len(d.Children) > 0 {
				childMax := maxDepth(d.Children, depth+1)
				if childMax > max {
					max = childMax
				}
			}
		}
		return max
	}
	return maxDepth(r.filterRoot, 0)
}

// addFilterDetails adds hierarchical filter details to a tree branch.
func (r *StructuredReporter) addFilterDetails(branch treeprint.Tree, details []FilterDetail, maxWidth int, depth int, maxDepth int) {
	for _, d := range details {
		if d.Name == "" && len(d.Children) > 0 {
			// Anonymous group (implicit AND) - add children directly
			r.addFilterDetails(branch, d.Children, maxWidth, depth, maxDepth)
		} else if len(d.Children) > 0 {
			// Operator with children - create a sub-branch
			subBranch := branch.AddBranch(r.formatFilterDetailAtDepth(d, maxWidth, depth, maxDepth))
			r.addFilterDetails(subBranch, d.Children, maxWidth, depth+1, maxDepth)
		} else if d.Name != "" {
			// Leaf filter
			branch.AddNode(r.formatFilterDetailAtDepth(d, maxWidth, depth, maxDepth))
		}
	}
}

// formatFilterDetailAtDepth formats a FilterDetail with depth-aware padding for alignment.
// Each tree depth level adds 4 chars of indentation, so we add extra padding for shallower items.
func (r *StructuredReporter) formatFilterDetailAtDepth(d FilterDetail, maxWidth int, depth int, maxDepth int) string {
	var icon string
	if d.Matched {
		icon = passStyle.Render(passIcon)
	} else {
		icon = failStyle.Render(failIcon)
	}

	// Add extra padding for shallower items to align with deeper nested items
	// Each depth level adds 4 chars of tree indentation
	extraPadding := (maxDepth - depth) * 4
	totalWidth := maxWidth + 1 + extraPadding

	result := fmt.Sprintf("%-*s %s", totalWidth, d.Name+":", icon)
	if d.Detail != "" {
		result += " " + detailStyle.Render(d.Detail)
	}
	return result
}

// formatActionWithPadding formats an action entry with extra padding for depth alignment.
func (r *StructuredReporter) formatActionWithPadding(a actionEntry, maxWidth int, extraPadding int) string {
	var icon, status string

	switch a.result.Outcome {
	case OutcomeSuccess:
		icon = passStyle.Render(passIcon)
		status = "done"

	case OutcomeMoved:
		icon = passStyle.Render(passIcon)
		status = "→ " + a.result.NewPath

	case OutcomeDeleted:
		icon = passStyle.Render(passIcon)
		status = "deleted"

	case OutcomeSkipped:
		icon = skipStyle.Render(skipIcon)
		status = "skipped " + detailStyle.Render("(destination exists)")

	case OutcomeFailed:
		icon = failStyle.Render(failIcon)
		status = "failed"
		if a.result.Error != "" {
			status += " " + detailStyle.Render("("+a.result.Error+")")
		}
	}

	totalWidth := maxWidth + 1 + extraPadding
	return fmt.Sprintf("%-*s %s %s", totalWidth, a.name+":", icon, status)
}

// NullReporter is a no-op reporter for when reporting is disabled.
// It allows filter evaluation code to call methods unconditionally.
type NullReporter struct{}

func (NullReporter) StartRule(name string)                              {}
func (NullReporter) EndRule() int                                       { return 0 }
func (NullReporter) StartFile(path string)                              {}
func (NullReporter) RecordFilter(name string, matched bool, detail string) {}
func (NullReporter) PushOperator(op string)                             {}
func (NullReporter) PopOperator(op string, matched bool)                {}
func (NullReporter) ReportAction(name string, result ActionResult)      {}
func (NullReporter) MarkFiltersPassed()                                 {}
func (NullReporter) EndFile() bool                                      { return false }
