package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// appName names the per-application config directory.
const appName = "goproject"

func getConfigHome() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config")
	}
	return ""
}

// GetConfigDirs returns all config directories to search, in priority order:
// $XDG_CONFIG_HOME/goproject, ~/.goproject, then $XDG_CONFIG_DIRS (or
// /etc/xdg) on Unix-like systems.
func GetConfigDirs() []string {
	var dirs []string

	if configHome := getConfigHome(); configHome != "" {
		dirs = append(dirs, filepath.Join(configHome, appName))
	}

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "."+appName))
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "freebsd" || runtime.GOOS == "openbsd" {
		if xdgConfigDirs := os.Getenv("XDG_CONFIG_DIRS"); xdgConfigDirs != "" {
			for _, dir := range filepath.SplitList(xdgConfigDirs) {
				if dir != "" {
					dirs = append(dirs, filepath.Join(dir, appName))
				}
			}
		} else {
			dirs = append(dirs, filepath.Join("/etc", "xdg", appName))
		}
	}

	return dirs
}
