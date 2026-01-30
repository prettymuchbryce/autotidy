package watcher

import (
	"context"
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/state"
)

// Watcher monitors filesystem changes and executes matching rules.
type Watcher struct {
	fsWatcher *fsnotify.Watcher
	runners   rules.RuleRunners
	state     *state.State
	// Debounce delay for triggering rule execution after events
	debounceDelay time.Duration
	// How long after rule completion to ignore events
	eventCooldown time.Duration

	// Per-runner timers for debounced execution
	runnerTimers       map[*rules.RuleRunner]*time.Timer
	runnerDebounceChan chan *rules.RuleRunner

	// Channel for timestamped events from event goroutine
	eventChan chan TimestampedEvent

	watchManager        *WatchedDirs
	watchDebounceChan   chan string
	watchRootsRecreated chan bool

	// Closed when the watcher is stopping to unblock goroutines
	done chan struct{}
}

// TimestampedEvent wraps an fsnotify event with its receive time.
type TimestampedEvent struct {
	Event fsnotify.Event
	Time  time.Time
}

// New creates a new Watcher for the given rules.
// Disabled rules are filtered out automatically.
// If st is provided, execution stats will be persisted after each rule run.
func New(ruleList []rules.Rule, debounce time.Duration, st *state.State) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Convert enabled rules to runners
	realFs := fs.NewReal()
	var runners []*rules.RuleRunner
	for i := range ruleList {
		rule := &ruleList[i]
		if rule.IsEnabled() {
			runners = append(runners, rules.NewRuleRunner(rule, realFs, nil))
		} else {
			slog.Info("skipping disabled rule", "rule", rule.Name)
		}
	}

	var watchDebounceChan = make(chan string)
	var watchRootsRecreated = make(chan bool, 1)
	var done = make(chan struct{})

	w := &Watcher{
		fsWatcher:           fsw,
		runners:             runners,
		state:               st,
		debounceDelay:       debounce,
		eventCooldown:       1 * time.Second,
		runnerTimers:        make(map[*rules.RuleRunner]*time.Timer),
		runnerDebounceChan:  make(chan *rules.RuleRunner),
		eventChan:           make(chan TimestampedEvent, 100),
		watchManager:        NewWatchedDirs(realFs, fsw, debounce, watchDebounceChan, watchRootsRecreated, done),
		watchDebounceChan:   watchDebounceChan,
		watchRootsRecreated: watchRootsRecreated,
		done:                done,
	}

	// Initialize per-runner timers
	w.initRunnerTimers()

	// Add watches for all root locations
	w.runners.EachRuleLocation(func(rule *rules.Rule, loc string) bool {
		w.watchManager.AddRoot(loc, rule.IsRecursive())
		return true
	})

	return w, nil
}

// eventLoop reads from fsnotify and timestamps events before forwarding.
func (w *Watcher) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			te := TimestampedEvent{
				Event: event,
				Time:  time.Now(),
			}
			select {
			case w.eventChan <- te:
			case <-w.done:
				return
			}
		}
	}
}

// Run starts the watcher and blocks until context is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	slog.Info("watcher started", "debounce", w.debounceDelay)

	// Start event goroutine to timestamp incoming events
	go w.eventLoop(ctx)

	for {
		// Execute all rules with pending timers before returning.
		// This ensures rule execution takes priority over event processing.
		// When multiple rules share the same location, their timers fire
		// nearly simultaneously.
		// Without draining, Go's select picks randomly between runnerDebounceChan and
		// eventChan. If an event is picked between rule executions, it might reset
		// a timer for a rule that has been enqueued in runnerDebounceChan, but not yet
		// executed.
		//
		// By draining all pending timers first, we ensure all queued rules
		// complete before any events can trigger additional timer resets.
	drain:
		for {
			select {
			case runner := <-w.runnerDebounceChan:
				w.executeRunner(runner)
			default:
				break drain
			}
		}

		select {
		case <-ctx.Done():
			slog.Info("watcher stopping")
			// Signal all goroutines to stop
			close(w.done)
			// Stop all timers
			for _, timer := range w.runnerTimers {
				timer.Stop()
			}
			w.watchManager.Destroy()
			return w.fsWatcher.Close()

		case event := <-w.eventChan:
			w.watchManager.ProcessEvent(event)

			fsEvent := event.Event
			path := fsEvent.Name
			w.scheduleRulesForPath(path, event.Time)
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return nil
			}
			slog.Error("watcher error", "error", err)
		case path := <-w.watchDebounceChan:
			w.watchManager.EvaluateDebounced(path)
		case <-w.watchRootsRecreated:
			rootsRecreated := w.watchManager.GetRecreatedRoots()
			for _, root := range rootsRecreated {
				slog.Info("watch root recreated", "path", root.Path)
				w.scheduleRulesForPath(root.Path, root.Time)
			}
		case runner := <-w.runnerDebounceChan:
			w.executeRunner(runner)
		}
	}
}

// initRunnerTimers initializes the per-runner timers for debounced execution.
func (w *Watcher) initRunnerTimers() {
	for _, runner := range w.runners {
		r := runner // capture for closure
		// Create timer (initially stopped)
		timer := time.AfterFunc(time.Hour, func() {
			select {
			case w.runnerDebounceChan <- r:
			case <-w.done:
			}
		})
		timer.Stop()
		w.runnerTimers[runner] = timer
	}
}

// WatchCount returns the number of directories currently being watched.
func (w *Watcher) WatchCount() int {
	return w.watchManager.WatchCount()
}

// executeRunner runs a single runner and persists execution stats.
func (w *Watcher) executeRunner(runner *rules.RuleRunner) {
	rule := runner.Rule()
	stats, err := runner.Execute()
	if err != nil {
		slog.Error("rule execution failed", "rule", rule.Name, "error", err)
	}

	// Persist execution stats
	if w.state != nil && stats != nil {
		if err := w.state.UpdateRuleStats(rule.Name, stats.StartTime, stats.Duration, stats.FilesProcessed, stats.ErrorCount); err != nil {
			slog.Warn("failed to persist rule stats", "rule", rule.Name, "error", err)
		}
	}
}

// scheduleRulesForPath schedules execution for all rules that cover the given path.
// Events that occurred before the cooldown period after rule completion are ignored.
func (w *Watcher) scheduleRulesForPath(path string, eventTime time.Time) {
	for _, runner := range w.runners {
		rule := runner.Rule()
		if !rule.CoversPath(path) {
			continue
		}

		// Time-based filtering: ignore events during cooldown period after rule execution.
		// This prevents cascading re-triggers from the rule's own filesystem changes.
		filterUntil := runner.LastCompletedTime().Add(w.eventCooldown)
		if eventTime.Before(filterUntil) {
			slog.Debug("ignoring event during cooldown", "path", path, "rule", rule.Name)
			continue
		}

		// Schedule rule execution after debounce
		slog.Debug("scheduling rule execution", "path", path, "rule", rule.Name)
		w.runnerTimers[runner].Reset(w.debounceDelay)
	}
}
