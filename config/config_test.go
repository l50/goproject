package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// isolate points HOME and the XDG directories at a temporary root and
// changes into an empty directory so no real config file is found.
func isolate(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("HOME", filepath.Join(root, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_CONFIG_DIRS", filepath.Join(root, "xdg"))
	t.Chdir(t.TempDir())
	return root
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Log.Level != "info" || cfg.Log.Format != "text" {
		t.Errorf("Defaults().Log = %+v, want info/text", cfg.Log)
	}
	if cfg.Quiet || cfg.Verbose {
		t.Errorf("Defaults() quiet=%v verbose=%v, want both false", cfg.Quiet, cfg.Verbose)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		file      string // path relative to the isolated root; empty means no file
		body      string
		env       map[string]string
		wantLevel string
	}{
		{name: "no config file uses defaults", wantLevel: "info"},
		{name: "xdg config home", file: "config/goproject/config.yaml", body: "log:\n  level: debug\n", wantLevel: "debug"},
		{name: "legacy home dir", file: "home/.goproject/config.yaml", body: "log:\n  level: warn\n", wantLevel: "warn"},
		{
			name:      "env overrides file",
			file:      "config/goproject/config.yaml",
			body:      "log:\n  level: debug\n",
			env:       map[string]string{"GOPROJECT_LOG_LEVEL": "error"},
			wantLevel: "error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := isolate(t)
			if tc.file != "" {
				writeFile(t, filepath.Join(root, tc.file), tc.body)
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg, v, err := Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if v == nil {
				t.Fatal("Load returned a nil viper instance")
			}
			if cfg.Log.Level != tc.wantLevel {
				t.Errorf("log level = %q, want %q", cfg.Log.Level, tc.wantLevel)
			}
		})
	}
}

func TestLoadInvalidFile(t *testing.T) {
	root := isolate(t)
	writeFile(t, filepath.Join(root, "config", "goproject", "config.yaml"), "log: [unclosed\n")
	if _, _, err := Load(); err == nil {
		t.Fatal("expected an error for an unparsable config file")
	}
}

func TestLoadFromPath(t *testing.T) {
	isolate(t)
	dir := t.TempDir()

	good := filepath.Join(dir, "good.yaml")
	writeFile(t, good, "log:\n  format: json\nquiet: true\n")
	cfg, _, err := LoadFromPath(good)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}
	if cfg.Log.Format != "json" || !cfg.Quiet {
		t.Errorf("LoadFromPath = %+v, want json format and quiet", cfg)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("unset log level = %q, want default info", cfg.Log.Level)
	}

	if _, _, err := LoadFromPath(filepath.Join(dir, "missing.yaml")); err == nil {
		t.Error("expected an error for a missing explicit config file")
	}
}

func TestGetConfigDirs(t *testing.T) {
	root := isolate(t)
	dirs := GetConfigDirs()

	for _, want := range []string{
		filepath.Join(root, "config", "goproject"),
		filepath.Join(root, "home", ".goproject"),
	} {
		if !slices.Contains(dirs, want) {
			t.Errorf("GetConfigDirs() = %v, missing %q", dirs, want)
		}
	}
	if dirs[0] != filepath.Join(root, "config", "goproject") {
		t.Errorf("first config dir = %q, want XDG_CONFIG_HOME first", dirs[0])
	}
}
