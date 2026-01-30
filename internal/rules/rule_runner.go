package rules

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/report"
)

// ExecutionStats contains statistics from a rule execution.
type ExecutionStats struct {
	StartTime      time.Time
	Duration       time.Duration
	FilesProcessed int
	ErrorCount     int
}

type RuleRunners []*RuleRunner

// EachRuleLocation iterates over all rules and their locations.
// If the callback returns false, iteration stops early.
func (rrs RuleRunners) EachRuleLocation(f func(rule *Rule, location string) bool) {
	for _, rr := range rrs {
		rule := rr.Rule()
		for _, loc := range rule.Locations {
			if !f(rule, loc) {
				return
			}
		}
	}
}

// RuleRunner executes a Rule with the given dependencies.
type RuleRunner struct {
	rule              *Rule
	fs                fs.FileSystem
	reporter          report.Reporter
	lastCompletedTime time.Time
}

// NewRuleRunner creates a RuleRunner with the given dependencies.
// If reporter is nil, NullReporter is used (no output, enables short-circuit).
func NewRuleRunner(rule *Rule, fs fs.FileSystem, reporter report.Reporter) *RuleRunner {
	if reporter == nil {
		reporter = report.NullReporter{}
	}
	return &RuleRunner{
		rule:     rule,
		fs:       fs,
		reporter: reporter,
	}
}

// Rule returns the underlying rule configuration.
func (rr *RuleRunner) Rule() *Rule {
	return rr.rule
}

// LastCompletedTime returns the time when the rule last completed execution.
func (rr *RuleRunner) LastCompletedTime() time.Time {
	return rr.lastCompletedTime
}

// Execute runs the rule against all files in its locations.
// Returns execution statistics.
func (rr *RuleRunner) Execute() (*ExecutionStats, error) {
	rule := rr.rule
	stats := &ExecutionStats{
		StartTime: time.Now(),
	}

	if !rule.IsEnabled() {
		slog.Warn("rule is not enabled", "rule", rule.Name)
		stats.Duration = time.Since(stats.StartTime)
		return stats, nil
	}

	slog.Info("Executing rule", "rule", rule.Name)

	// Start reporting for this rule
	rr.reporter.StartRule(rule.Name)

	// Build snapshots for all locations first
	type locationSnapshot struct {
		loc  string
		tree *Node
	}
	var snapshots []locationSnapshot

	for _, loc := range rule.Locations {
		info, err := rr.fs.Stat(loc)
		if err != nil {
			slog.Warn("location does not exist", "rule", rule.Name, "location", loc, "error", err)
			continue
		}

		if !info.IsDir() {
			slog.Error("location is not a directory", "rule", rule.Name, "location", loc)
			continue
		}

		// Build tree from directory
		tree := BuildSnapshot(rr.fs, loc, rule.IsRecursive())
		if tree == nil {
			continue
		}

		snapshots = append(snapshots, locationSnapshot{loc: loc, tree: tree})
	}

	// Traverse and execute actions
	// Create a visitor that wraps executeOnItem and tracks stats
	visitor := func(path string) (TraverseControl, struct{}, error) {
		result, fileErr, err := rr.executeOnItem(path)
		if fileErr {
			stats.ErrorCount++
		}
		if err != nil {
			return TraverseControl{Instruction: StopTraversing}, struct{}{}, err
		}
		if result == nil {
			return TraverseControl{}, struct{}{}, nil
		}

		// File was processed (matched filters and actions were executed)
		stats.FilesProcessed++

		ctrl := TraverseControl{NewPath: result.NewPath}
		if result.Deleted {
			ctrl.Instruction = SkipChildren
		}
		return ctrl, struct{}{}, nil
	}

	for _, snap := range snapshots {
		// Traverse tree (skipping root directory)
		mode := rule.GetTraversalMode()
		var err error
		switch mode {
		case TraversalDepthFirst:
			_, err = TraverseChildrenDFS(snap.tree, filepath.Dir(snap.loc), visitor)
		case TraversalBreadthFirst:
			_, err = TraverseChildrenBFS(snap.tree, filepath.Dir(snap.loc), visitor)
		default:
			slog.Error("unknown traversal mode", "rule", rule.Name, "mode", mode)
			continue
		}
		if err != nil {
			slog.Error("error during traversal", "rule", rule.Name, "error", err)
			continue
		}
	}

	// End reporting for this rule
	rr.reporter.EndRule()

	// Record completion time and duration
	rr.lastCompletedTime = time.Now()
	stats.Duration = time.Since(stats.StartTime)

	return stats, nil
}

// executeOnItem evaluates filters and executes actions on a single item.
// Returns (result, hadError, err) where:
//   - result is nil if filters didn't match or no actions modified the file
//   - hadError is true if a filesystem error occurred (item was skipped)
//   - err is non-nil only for fatal errors that should stop traversal
func (rr *RuleRunner) executeOnItem(path string) (*ExecutionResult, bool, error) {
	rule := rr.rule
	currentPath := path
	var deleted bool

	// Start reporting for this file
	rr.reporter.StartFile(path)

	// Evaluate filters if present
	if rule.Filters != nil {
		passed, err := rule.Filters.Evaluate(currentPath, rr.reporter)
		if err != nil {
			if isFilesystemError(err) {
				slog.Warn("filesystem error during filter evaluation, skipping item", "rule", rule.Name, "path", currentPath, "error", err)
				rr.reporter.EndFile()
				return nil, true, nil
			}
			return nil, false, err
		}

		if !passed {
			rr.reporter.EndFile()
			return nil, false, nil
		}

		// Mark that filters passed (for hasMatchOrAction in reporting)
		rr.reporter.MarkFiltersPassed()
	}

	// Execute all actions, tracking path changes
	for _, action := range rule.Actions {
		result, err := action.Execute(currentPath, rr.fs)
		if err != nil {
			if isFilesystemError(err) {
				slog.Warn("filesystem error during action, skipping item", "rule", rule.Name, "action", action.Name, "path", currentPath, "error", err)
				rr.reporter.ReportAction(action.Name, report.ActionResult{
					Outcome: report.OutcomeFailed,
					Error:   err.Error(),
				})
				rr.reporter.EndFile()
				return nil, true, nil
			}
			return nil, false, err
		}

		// Report action result
		if result == nil {
			// Action completed but no changes
			rr.reporter.ReportAction(action.Name, report.ActionResult{Outcome: report.OutcomeSuccess})
		} else if result.ConflictAlreadyExists {
			rr.reporter.ReportAction(action.Name, report.ActionResult{Outcome: report.OutcomeSkipped})
		} else if result.Deleted {
			rr.reporter.ReportAction(action.Name, report.ActionResult{Outcome: report.OutcomeDeleted})
		} else if result.NewPath != "" {
			rr.reporter.ReportAction(action.Name, report.ActionResult{
				Outcome: report.OutcomeMoved,
				NewPath: result.NewPath,
			})
		}

		if result != nil {
			if result.Deleted {
				deleted = true
				break // Stop processing actions on deleted file
			}
			if result.ConflictAlreadyExists {
				break // Stop processing actions when destination exists
			}
			if result.NewPath != "" {
				currentPath = result.NewPath
			}
		}
	}

	// End reporting for this file
	rr.reporter.EndFile()

	// Return nil if nothing changed
	if currentPath == path && !deleted {
		return nil, false, nil
	}

	return &ExecutionResult{
		NewPath: currentPath,
		Deleted: deleted,
	}, false, nil
}

// isFilesystemError returns true if the error is a filesystem-related error.
// These errors should be logged as warnings rather than stopping execution.
func isFilesystemError(err error) bool {
	var pathErr *os.PathError
	var linkErr *os.LinkError
	return errors.As(err, &pathErr) || errors.As(err, &linkErr)
}
