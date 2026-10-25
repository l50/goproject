package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l50/goproject/config"
	"github.com/spf13/cobra"
)

// TestMain points the XDG directories at a temporary root so tests never
// read the developer's real config.
func TestMain(m *testing.M) {
	root, err := os.MkdirTemp("", "goproject-test-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create test XDG root: %v\n", err)
		os.Exit(1)
	}
	for env, sub := range map[string]string{
		"HOME":            "home",
		"XDG_CONFIG_HOME": "config",
		"XDG_CONFIG_DIRS": "xdg",
	} {
		if err := os.Setenv(env, filepath.Join(root, sub)); err != nil {
			fmt.Fprintf(os.Stderr, "failed to set %s: %v\n", env, err)
			os.Exit(1)
		}
	}
	code := m.Run()
	_ = os.RemoveAll(root)
	os.Exit(code)
}

// execute runs the root command with args and a probe subcommand that
// captures the config initConfig resolved.
func execute(t *testing.T, args ...string) (*config.Config, string, error) {
	t.Helper()

	var got *config.Config
	root := NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, _ []string) error {
			got = configFromContext(cmd)
			return nil
		},
	})

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return got, out.String(), err
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestVersionCmd(t *testing.T) {
	_, out, err := execute(t, "version")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	for _, want := range []string{"goproject version " + version, "commit:", "built:"} {
		if !strings.Contains(out, want) {
			t.Errorf("version output %q missing %q", out, want)
		}
	}
}

func TestCompletionCmd(t *testing.T) {
	tests := []struct {
		shell   string
		wantErr bool
	}{
		{shell: "bash"},
		{shell: "zsh"},
		{shell: "fish"},
		{shell: "powershell"},
		{shell: "tcsh", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.shell, func(t *testing.T) {
			_, out, err := execute(t, "completion", tc.shell)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error for an unsupported shell")
				}
				return
			}
			if err != nil {
				t.Fatalf("completion %s: %v", tc.shell, err)
			}
			if !strings.Contains(out, "goproject") {
				t.Errorf("completion %s output does not mention goproject", tc.shell)
			}
		})
	}
}

func TestInitConfigPrecedence(t *testing.T) {
	path := writeConfig(t, "log:\n  level: warn\n  format: json\n")

	tests := []struct {
		name       string
		env        map[string]string
		args       []string
		wantLevel  string
		wantFormat string
	}{
		{
			name:       "defaults",
			args:       []string{"probe"},
			wantLevel:  "info",
			wantFormat: "text",
		},
		{
			name:       "config file",
			args:       []string{"--config", path, "probe"},
			wantLevel:  "warn",
			wantFormat: "json",
		},
		{
			name:       "env overrides file",
			env:        map[string]string{"GOPROJECT_LOG_LEVEL": "error"},
			args:       []string{"--config", path, "probe"},
			wantLevel:  "error",
			wantFormat: "json",
		},
		{
			name:       "flag overrides env",
			env:        map[string]string{"GOPROJECT_LOG_LEVEL": "error"},
			args:       []string{"--config", path, "--log-level", "debug", "probe"},
			wantLevel:  "debug",
			wantFormat: "json",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			cfg, _, err := execute(t, tc.args...)
			if err != nil {
				t.Fatalf("execute: %v", err)
			}
			if cfg == nil {
				t.Fatal("config was not stored on the command context")
			}
			if cfg.Log.Level != tc.wantLevel {
				t.Errorf("log level = %q, want %q", cfg.Log.Level, tc.wantLevel)
			}
			if cfg.Log.Format != tc.wantFormat {
				t.Errorf("log format = %q, want %q", cfg.Log.Format, tc.wantFormat)
			}
		})
	}
}

func TestInitConfigErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "quiet and verbose", args: []string{"--quiet", "--verbose", "probe"}},
		{name: "missing config file", args: []string{"--config", filepath.Join(t.TempDir(), "nope.yaml"), "probe"}},
		{name: "invalid config file", args: []string{"--config", writeConfig(t, "log: [unclosed\n"), "probe"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := execute(t, tc.args...); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestConfigFromContextUnset(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	if cfg := configFromContext(cmd); cfg != nil {
		t.Errorf("configFromContext = %+v, want nil", cfg)
	}
}
