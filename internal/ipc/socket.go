package ipc

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// SocketPath returns the platform-appropriate socket/address for IPC.
func SocketPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return `\\.\pipe\autotidy`, nil
	case "darwin":
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(cacheDir, "autotidy", "autotidy.sock"), nil
	default:
		if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
			return filepath.Join(xdg, "autotidy", "autotidy.sock"), nil
		}
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(cacheDir, "autotidy", "autotidy.sock"), nil
	}
}

// StatePath returns the platform-appropriate state file path.
// State is stored alongside config for simplicity.
func StatePath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("APPDATA not set and cannot determine home directory: %w", err)
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "autotidy", "state.json"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".config", "autotidy", "state.json"), nil
	}
}
