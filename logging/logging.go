// Package logging provides leveled console logging for goproject.
//
// Log lines go to stderr so that command output written with [Output] can be
// piped. A process-wide default logger backs the package-level functions,
// and [WithLogger] / [FromContext] carry a logger through a context.
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/fatih/color"
)

// OutputType defines the available logger output formats.
type OutputType int

const (
	// PlainOutput emits uncolored text output.
	PlainOutput OutputType = iota
	// ColorOutput emits ANSI-colored text output.
	ColorOutput
	// JSONOutput emits structured JSON from [Logger.Output].
	JSONOutput
)

// Logger writes leveled log lines and command output.
type Logger struct {
	Level      slog.Level
	OutputType OutputType
	// Quiet suppresses everything below error level.
	Quiet bool
	// ConsoleWriter receives log lines; it defaults to os.Stderr.
	ConsoleWriter io.Writer
	// OutputWriter receives [Logger.Output] data; it defaults to os.Stdout.
	OutputWriter io.Writer
}

// New returns a logger for the given level and format names. Unknown levels
// fall back to info and unknown formats to plain text. Verbose lowers the
// level to debug.
func New(level, format string, quiet, verbose bool) *Logger {
	l := &Logger{
		Level:         ParseLevel(level),
		OutputType:    ParseOutputType(format),
		Quiet:         quiet,
		ConsoleWriter: os.Stderr,
		OutputWriter:  os.Stdout,
	}
	if verbose && l.Level > slog.LevelDebug {
		l.Level = slog.LevelDebug
	}
	return l
}

// ParseLevel maps a level name to its slog equivalent, defaulting to info.
func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseOutputType maps a format name to an [OutputType], defaulting to
// [PlainOutput].
func ParseOutputType(format string) OutputType {
	switch format {
	case "json":
		return JSONOutput
	case "color":
		return ColorOutput
	default:
		return PlainOutput
	}
}

var levelColors = map[slog.Level]func(format string, a ...any) string{
	slog.LevelDebug: color.CyanString,
	slog.LevelInfo:  color.GreenString,
	slog.LevelWarn:  color.YellowString,
	slog.LevelError: color.RedString,
}

func (l *Logger) enabled(level slog.Level) bool {
	if l.Quiet {
		return level >= slog.LevelError
	}
	return level >= l.Level
}

func (l *Logger) log(level slog.Level, format string, args ...any) {
	if !l.enabled(level) || l.ConsoleWriter == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if l.OutputType == ColorOutput {
		msg = levelColors[level]("%s", msg)
	}
	_, _ = fmt.Fprintln(l.ConsoleWriter, msg)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...any) { l.log(slog.LevelDebug, format, args...) }

// Info logs an informational message.
func (l *Logger) Info(format string, args ...any) { l.log(slog.LevelInfo, format, args...) }

// Warn logs a warning.
func (l *Logger) Warn(format string, args ...any) { l.log(slog.LevelWarn, format, args...) }

// Error logs an error. Errors are shown even in quiet mode.
func (l *Logger) Error(format string, args ...any) { l.log(slog.LevelError, format, args...) }

// Output writes command output to OutputWriter, as indented JSON when the
// logger uses [JSONOutput] and as a plain line otherwise.
func (l *Logger) Output(data any) error {
	w := l.OutputWriter
	if w == nil {
		w = os.Stdout
	}
	if l.OutputType == JSONOutput {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
	_, err := fmt.Fprintln(w, data)
	return err
}

var (
	defaultLogger   = New("info", "text", false, false)
	defaultLoggerMu sync.RWMutex
)

// Default returns the process-wide logger.
func Default() *Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()
	return defaultLogger
}

// SetDefault replaces the process-wide logger. A nil logger is ignored.
func SetDefault(l *Logger) {
	if l == nil {
		return
	}
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = l
}

// Debug logs a debug message with the default logger.
func Debug(format string, args ...any) { Default().Debug(format, args...) }

// Info logs an informational message with the default logger.
func Info(format string, args ...any) { Default().Info(format, args...) }

// Warn logs a warning with the default logger.
func Warn(format string, args ...any) { Default().Warn(format, args...) }

// Error logs an error with the default logger.
func Error(format string, args ...any) { Default().Error(format, args...) }

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

// WithLogger returns a context that carries l.
func WithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns the logger stored in ctx, or the default logger.
func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		if l, ok := ctx.Value(loggerKey).(*Logger); ok && l != nil {
			return l
		}
	}
	return Default()
}
