//go:build integration

package testutil

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/spf13/afero"
	"github.com/prettymuchbryce/autotidy/daemon"
)

// FileEntry describes a file or directory with all relevant properties.
type FileEntry struct {
	Path       string        // relative path using forward slashes (e.g., "source/file.txt")
	IsDir      bool          // true for directories
	Content    string        // file content (mutually exclusive with Size)
	Size       int64         // create file with this many zero bytes
	ModTime    time.Duration // relative to now, e.g., -48*time.Hour means "2 days ago"
	AccessTime time.Duration // relative to now
}

// TestCase is a complete data-driven integration test.
type TestCase struct {
	Name    string        // test name (used for t.Run)
	Config  string        // YAML config with {{.TmpDir}} template variable
	Before  []FileEntry   // files/dirs to create BEFORE daemon starts
	Trigger []FileEntry   // files to create AFTER daemon starts (triggers rules)
	Expect  []FileEntry   // files that SHOULD exist after processing
	Missing []string      // paths that should NOT exist after processing
	Timeout time.Duration // how long to wait for expected state (default: 2s)
}

// Harness manages the test environment.
type Harness struct {
	t      *testing.T
	tmpDir string
	fs     afero.Fs
	cancel context.CancelFunc
	errCh  chan error
}

// Run executes a single test case.
func Run(t *testing.T, tc TestCase) {
	t.Helper()

	h := &Harness{
		t:      t,
		tmpDir: t.TempDir(),
		fs:     afero.NewOsFs(),
		errCh:  make(chan error, 1),
	}

	// Create initial directories and files
	h.createEntries(tc.Before)

	// Start daemon with templated config
	h.startDaemon(tc.Config)

	// Create files that trigger rules
	h.createEntries(tc.Trigger)

	// Wait for expected state
	h.waitAndVerify(tc)

	// Cleanup
	h.cleanup()
}

// RunTable executes multiple test cases as subtests.
func RunTable(t *testing.T, cases []TestCase) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			Run(t, tc)
		})
	}
}

// createEntries creates files and directories from FileEntry specs.
func (h *Harness) createEntries(entries []FileEntry) {
	h.t.Helper()

	for _, e := range entries {
		// Convert forward slashes to OS-specific separator for Windows compatibility
		path := filepath.Join(h.tmpDir, filepath.FromSlash(e.Path))

		if e.IsDir {
			if err := os.MkdirAll(path, 0755); err != nil {
				h.t.Fatalf("failed to create directory %s: %v", e.Path, err)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			h.t.Fatalf("failed to create parent directory for %s: %v", e.Path, err)
		}

		// Create file with content or size
		var content []byte
		if e.Content != "" {
			content = []byte(e.Content)
		} else if e.Size > 0 {
			content = make([]byte, e.Size)
		}

		if err := os.WriteFile(path, content, 0644); err != nil {
			h.t.Fatalf("failed to create file %s: %v", e.Path, err)
		}

		// Set timestamps if specified
		if e.ModTime != 0 || e.AccessTime != 0 {
			now := time.Now()
			mtime := now.Add(e.ModTime)
			atime := now.Add(e.AccessTime)
			if e.AccessTime == 0 {
				atime = mtime
			}
			if err := os.Chtimes(path, atime, mtime); err != nil {
				h.t.Fatalf("failed to set timestamps for %s: %v", e.Path, err)
			}
		}
	}
}

// startDaemon starts the daemon with the given config template.
func (h *Harness) startDaemon(configTemplate string) {
	h.t.Helper()

	// Render config template with helper functions
	tmpl, err := template.New("config").Funcs(template.FuncMap{
		// join creates OS-native paths: {{join .TmpDir "source" "subdir"}}
		"join": filepath.Join,
	}).Parse(configTemplate)
	if err != nil {
		h.t.Fatalf("failed to parse config template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{
		"TmpDir": h.tmpDir,
	}); err != nil {
		h.t.Fatalf("failed to execute config template: %v", err)
	}

	// Write config file
	configPath := filepath.Join(h.tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		h.t.Fatalf("failed to write config file: %v", err)
	}

	// Start daemon
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	go func() {
		h.errCh <- daemon.Run(ctx, configPath, h.fs, func(level string) {
			var logLevel slog.Level
			switch level {
			case "debug":
				logLevel = slog.LevelDebug
			case "info":
				logLevel = slog.LevelInfo
			case "warn":
				logLevel = slog.LevelWarn
			case "error":
				logLevel = slog.LevelError
			default:
				logLevel = slog.LevelInfo
			}
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
		})
	}()

	// Wait for daemon to start
	time.Sleep(100 * time.Millisecond)

	// Check for immediate failure
	select {
	case err := <-h.errCh:
		h.t.Fatalf("daemon failed to start: %v", err)
	default:
	}
}

// waitAndVerify waits for the expected state and verifies it.
func (h *Harness) waitAndVerify(tc TestCase) {
	h.t.Helper()

	timeout := tc.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if h.checkState(tc.Expect, tc.Missing) {
			return // Success
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Final check with error reporting
	h.assertState(tc.Expect, tc.Missing)
}

// checkState checks if the expected state is achieved (no error reporting).
func (h *Harness) checkState(expect []FileEntry, missing []string) bool {
	for _, e := range expect {
		path := filepath.Join(h.tmpDir, filepath.FromSlash(e.Path))
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		if e.IsDir && !info.IsDir() {
			return false
		}
		if !e.IsDir && info.IsDir() {
			return false
		}
	}

	for _, p := range missing {
		path := filepath.Join(h.tmpDir, filepath.FromSlash(p))
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// assertState checks the expected state and reports errors.
func (h *Harness) assertState(expect []FileEntry, missing []string) {
	h.t.Helper()

	for _, e := range expect {
		path := filepath.Join(h.tmpDir, filepath.FromSlash(e.Path))
		info, err := os.Stat(path)
		if err != nil {
			h.t.Errorf("expected %s to exist, but it doesn't", e.Path)
			continue
		}
		if e.IsDir && !info.IsDir() {
			h.t.Errorf("expected %s to be a directory, but it's a file", e.Path)
		}
		if !e.IsDir && info.IsDir() {
			h.t.Errorf("expected %s to be a file, but it's a directory", e.Path)
		}
	}

	for _, p := range missing {
		path := filepath.Join(h.tmpDir, filepath.FromSlash(p))
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			h.t.Errorf("expected %s to NOT exist, but it does", p)
		}
	}
}

// cleanup stops the daemon gracefully.
func (h *Harness) cleanup() {
	h.t.Helper()

	if h.cancel != nil {
		h.cancel()
	}

	// Wait for daemon to stop
	select {
	case err := <-h.errCh:
		if err != nil {
			h.t.Errorf("daemon returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		h.t.Error("daemon did not stop within timeout")
	}
}
