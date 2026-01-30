package state

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/prettymuchbryce/autotidy/internal/ipc"
)

// RuleState tracks persistent state for a single rule.
type RuleState struct {
	LastRunAt      time.Time     `json:"last_run_at"`
	LastDuration   time.Duration `json:"last_duration"`
	FilesProcessed int           `json:"files_processed"`
	ErrorCount     int           `json:"error_count"`
}

// State tracks daemon state that persists across restarts.
type State struct {
	mu    sync.RWMutex
	path  string
	Rules map[string]RuleState `json:"rules"`
}

// Load loads state from the default state file path.
// If the file doesn't exist, returns an empty state.
func Load() (*State, error) {
	path, err := ipc.StatePath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom loads state from the specified path.
// If the file doesn't exist, returns an empty state.
func LoadFrom(path string) (*State, error) {
	s := &State{
		path:  path,
		Rules: make(map[string]RuleState),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, s); err != nil {
		// Log warning but return empty state rather than failing
		slog.Warn("failed to parse state file, starting fresh", "error", err)
		s.Rules = make(map[string]RuleState)
		return s, nil
	}

	// Ensure Rules map is initialized even if JSON had null
	if s.Rules == nil {
		s.Rules = make(map[string]RuleState)
	}

	return s, nil
}

// UpdateRuleStats records execution stats for a rule and persists to disk.
func (s *State) UpdateRuleStats(ruleName string, runAt time.Time, duration time.Duration, filesProcessed, errorCount int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Rules[ruleName] = RuleState{
		LastRunAt:      runAt,
		LastDuration:   duration,
		FilesProcessed: filesProcessed,
		ErrorCount:     errorCount,
	}
	return s.save()
}

// save persists the state to disk. Must be called with mu held.
func (s *State) save() error {
	if s.path == "" {
		return nil
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// GetRuleState returns the persisted state for a rule.
// Returns nil if the rule has never run.
func (s *State) GetRuleState(ruleName string) *RuleState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rs, ok := s.Rules[ruleName]
	if !ok {
		return nil
	}
	return &rs
}

// Clear removes all state (useful for testing).
func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Rules = make(map[string]RuleState)
}
