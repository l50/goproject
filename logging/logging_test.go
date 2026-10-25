package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func newTestLogger(level, format string, quiet, verbose bool) (*Logger, *bytes.Buffer, *bytes.Buffer) {
	var console, output bytes.Buffer
	l := New(level, format, quiet, verbose)
	l.ConsoleWriter = &console
	l.OutputWriter = &output
	return l, &console, &output
}

func logAll(l *Logger) {
	l.Debug("debug %d", 1)
	l.Info("info %d", 2)
	l.Warn("warn %d", 3)
	l.Error("error %d", 4)
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		quiet   bool
		verbose bool
		want    []string
		notWant []string
	}{
		{name: "info", level: "info", want: []string{"info 2", "warn 3", "error 4"}, notWant: []string{"debug 1"}},
		{name: "debug", level: "debug", want: []string{"debug 1", "info 2", "warn 3", "error 4"}},
		{name: "warn", level: "warn", want: []string{"warn 3", "error 4"}, notWant: []string{"debug 1", "info 2"}},
		{name: "error", level: "error", want: []string{"error 4"}, notWant: []string{"debug 1", "info 2", "warn 3"}},
		{name: "verbose lowers to debug", level: "warn", verbose: true, want: []string{"debug 1", "info 2"}},
		{name: "quiet shows only errors", level: "debug", quiet: true, want: []string{"error 4"}, notWant: []string{"debug 1", "info 2", "warn 3"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l, console, _ := newTestLogger(tc.level, "text", tc.quiet, tc.verbose)
			logAll(l)
			got := console.String()
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Errorf("output %q missing %q", got, w)
				}
			}
			for _, nw := range tc.notWant {
				if strings.Contains(got, nw) {
					t.Errorf("output %q unexpectedly contains %q", got, nw)
				}
			}
		})
	}
}

func TestLoggerColor(t *testing.T) {
	prev := color.NoColor
	color.NoColor = false
	t.Cleanup(func() { color.NoColor = prev })

	l, console, _ := newTestLogger("info", "color", false, false)
	l.Info("hello")
	if !strings.Contains(console.String(), "\x1b[") {
		t.Errorf("color output %q has no ANSI escape", console.String())
	}

	l, console, _ = newTestLogger("info", "text", false, false)
	l.Info("hello")
	if strings.Contains(console.String(), "\x1b[") {
		t.Errorf("plain output %q has an ANSI escape", console.String())
	}
}

func TestLoggerNilConsoleWriter(t *testing.T) {
	l := New("debug", "text", false, false)
	l.ConsoleWriter = nil
	l.Error("must not panic")
}

func TestOutput(t *testing.T) {
	data := map[string]string{"key": "value"}

	l, _, out := newTestLogger("info", "json", false, false)
	if err := l.Output(data); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if got := out.String(); got != "{\n  \"key\": \"value\"\n}\n" {
		t.Errorf("JSON output = %q", got)
	}

	l, _, out = newTestLogger("info", "text", false, false)
	if err := l.Output("plain"); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if got := out.String(); got != "plain\n" {
		t.Errorf("plain output = %q, want %q", got, "plain\n")
	}
}

func TestParseLevel(t *testing.T) {
	tests := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"warn":    slog.LevelWarn,
		"error":   slog.LevelError,
		"":        slog.LevelInfo,
		"verbose": slog.LevelInfo,
	}
	for in, want := range tests {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseOutputType(t *testing.T) {
	tests := map[string]OutputType{
		"json":  JSONOutput,
		"color": ColorOutput,
		"text":  PlainOutput,
		"":      PlainOutput,
	}
	for in, want := range tests {
		if got := ParseOutputType(in); got != want {
			t.Errorf("ParseOutputType(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestDefaultLogger(t *testing.T) {
	prev := Default()
	t.Cleanup(func() { SetDefault(prev) })

	l, console, _ := newTestLogger("debug", "text", false, false)
	SetDefault(l)
	SetDefault(nil)
	if Default() != l {
		t.Fatal("SetDefault(nil) replaced the default logger")
	}

	Debug("d")
	Info("i")
	Warn("w")
	Error("e")
	if got := console.String(); got != "d\ni\nw\ne\n" {
		t.Errorf("package-level output = %q", got)
	}
}

func TestContextLogger(t *testing.T) {
	l, _, _ := newTestLogger("info", "text", false, false)
	ctx := WithLogger(context.Background(), l)
	if FromContext(ctx) != l {
		t.Error("FromContext did not return the stored logger")
	}
	if FromContext(context.Background()) != Default() {
		t.Error("FromContext without a logger should return the default")
	}
	//nolint:staticcheck // SA1012: a nil context is exactly what this checks.
	if FromContext(nil) != Default() {
		t.Error("FromContext(nil) should return the default")
	}
}
