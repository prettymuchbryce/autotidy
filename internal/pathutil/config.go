package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultConfigPath returns the platform-appropriate default config file path.
func DefaultConfigPath() (string, error) {
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
		return filepath.Join(appData, "autotidy", "config.yaml"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		return filepath.Join(home, ".config", "autotidy", "config.yaml"), nil
	}
}

// MustDefaultConfigPath returns DefaultConfigPath or panics on error.
// Use this only for flag defaults where error handling isn't possible.
func MustDefaultConfigPath() string {
	path, err := DefaultConfigPath()
	if err != nil {
		panic(fmt.Sprintf("failed to determine default config path: %v", err))
	}
	return path
}
