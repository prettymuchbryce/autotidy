package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/prettymuchbryce/autotidy/internal/config"
	"github.com/prettymuchbryce/autotidy/internal/ipc"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/state"
	"github.com/prettymuchbryce/autotidy/internal/watcher"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/spf13/afero"
)

// Controller manages the daemon lifecycle and implements ipc.Handler.
// All methods are called serially by the IPC server, so no locking is needed.
type Controller struct {
	configPath string
	fs         afero.Fs
	state      *state.State
	rules      []rules.Rule
	debounce   time.Duration

	watcher            *watcher.Watcher
	stopWatcher        context.CancelFunc
	chanWatcherStopped chan struct{}
}

// NewController creates a new daemon controller.
func NewController(configPath string, fs afero.Fs, st *state.State, rules []rules.Rule, debounce time.Duration) *Controller {
	return &Controller{
		configPath: configPath,
		fs:         fs,
		state:      st,
		rules:      rules,
		debounce:   debounce,
	}
}

// StartWatcher creates and starts a new watcher.
func (c *Controller) StartWatcher() error {
	w, err := watcher.New(c.rules, c.debounce, c.state)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	c.watcher = w
	c.stopWatcher = cancel
	c.chanWatcherStopped = done

	go func() {
		defer close(done)
		if err := w.Run(ctx); err != nil {
			slog.Error("watcher error", "error", err)
		}
	}()

	return nil
}

// StopWatcher stops the current watcher and waits for it to finish.
func (c *Controller) StopWatcher() {
	if c.watcher == nil {
		return
	}

	c.stopWatcher()
	<-c.chanWatcherStopped

	c.watcher = nil
	c.stopWatcher = nil
	c.chanWatcherStopped = nil
}

// Run loads config and runs the daemon until context is cancelled.
func Run(ctx context.Context, configPath string, fs afero.Fs, setupLogging func(string)) error {
	cfg, err := config.LoadWithFs(configPath, fs)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	setupLogging(cfg.Logging.Level)

	// Load persistent state
	st, err := state.Load()
	if err != nil {
		slog.Warn("failed to load state, starting fresh", "error", err)
		st, _ = state.LoadFrom("")
	}

	enabledRules := cfg.CountEnabledRules()
	slog.Info("loaded config", "rules", len(cfg.Rules), "enabled", enabledRules, "debounce", cfg.Daemon.Debounce)

	// Warn if no enabled rules, but continue running for potential reload
	if enabledRules == 0 {
		slog.Warn("no enabled rules found in config", "path", configPath)
	}

	// Create daemon controller
	controller := NewController(configPath, fs, st, cfg.Rules, cfg.Daemon.Debounce)

	// Start the watcher
	if err := controller.StartWatcher(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	// Start IPC server
	ipcServer, err := ipc.NewServer(controller)
	if err != nil {
		controller.StopWatcher()
		return fmt.Errorf("failed to create IPC server: %w", err)
	}

	// Notify systemd that we're ready (no-op on non-systemd systems)
	daemon.SdNotify(false, daemon.SdNotifyReady)
	slog.Info("daemon ready")

	// Run IPC server (blocks until context cancelled)
	if err := ipcServer.Serve(ctx); err != nil {
		slog.Error("IPC server error", "error", err)
	}

	// Notify systemd that we're stopping (no-op on non-systemd systems)
	daemon.SdNotify(false, daemon.SdNotifyStopping)

	// Clean up watcher
	controller.StopWatcher()

	return nil
}
