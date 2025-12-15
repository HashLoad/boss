// Package msg provides logging and messaging functionality with support for different log levels.
// It handles informational messages, warnings, errors, and debug output.
package msg

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/pterm/pterm"
)

type logLevel int

const (
	_ logLevel = iota
	WARN
	ERROR
	INFO
	DEBUG
)

// Stoppable is an interface for anything that can be stopped.
// This is used to stop progress trackers when errors occur.
type Stoppable interface {
	Stop()
}

// Messenger handles CLI output and logging.
// For testable code, create instances with NewMessenger() and inject as dependency.
// Package-level functions (Info, Err, Die, etc.) use the global defaultMsg instance.
//
// Usage patterns:
//   - Production: Use package functions (Info, Err, etc.)
//   - Testing: Create Messenger instance and inject to functions under test
type Messenger struct {
	sync.Mutex
	Stdout          io.Writer
	Stderr          io.Writer
	Stdin           io.Reader
	exitStatus      int
	hasError        bool
	quietMode       bool
	progressTracker Stoppable

	logLevel logLevel
}

// NewMessenger creates a new Messenger instance.
func NewMessenger() *Messenger {
	m := &Messenger{
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Stdin:      os.Stdin,
		exitStatus: 1,
		logLevel:   INFO,
	}

	return m
}

// ARCHITECTURAL DEBT: Global messenger singleton
// This global variable creates hidden dependencies and makes testing difficult.
// However, logging is often acceptable as global state in CLI applications.
// For testable code, consider using Messenger instances passed as dependencies.
//
//nolint:gochecknoglobals // Global logger is acceptable for CLI apps
var defaultMsg = NewMessenger()

// Die prints an error message and exits the program.
func Die(msg string, args ...any) {
	defaultMsg.Die(msg, args...)
}

// Info prints an informational message.
func Info(msg string, args ...any) {
	defaultMsg.Info(msg, args...)
}

// Success prints a success message.
func Success(msg string, args ...any) {
	defaultMsg.Success(msg, args...)
}

// Debug prints a debug message.
func Debug(msg string, args ...any) {
	defaultMsg.Debug(msg, args...)
}

// Warn prints a warning message.
func Warn(msg string, args ...any) {
	defaultMsg.Warn(msg, args...)
}

// Err prints an error message.
func Err(msg string, args ...any) {
	defaultMsg.Err(msg, args...)
}

// LogLevel sets the global log level.
func LogLevel(level logLevel) {
	defaultMsg.LogLevel(level)
}

// LogLevel sets the log level for the messenger.
func (m *Messenger) LogLevel(level logLevel) {
	m.Lock()
	m.logLevel = level
	m.Unlock()
}

// IsDebugMode returns true if the log level is set to DEBUG.
func IsDebugMode() bool {
	return defaultMsg.IsDebugMode()
}

// IsDebugMode returns true if the log level is set to DEBUG.
func (m *Messenger) IsDebugMode() bool {
	m.Lock()
	defer m.Unlock()
	return m.logLevel >= DEBUG
}

// Err prints an error message.
func (m *Messenger) Err(msg string, args ...any) {
	if m.logLevel < ERROR {
		return
	}

	if m.progressTracker != nil {
		m.progressTracker.Stop()
		m.progressTracker = nil
	}

	m.quietMode = false

	m.print(pterm.Error.MessageStyle, msg, args...)
	m.hasError = true
}

// Warn prints a warning message.
func (m *Messenger) Warn(msg string, args ...any) {
	if m.logLevel < WARN {
		return
	}

	wasQuiet := m.quietMode
	m.quietMode = false

	m.print(pterm.Warning.MessageStyle, msg, args...)

	m.quietMode = wasQuiet
}

// Info prints an informational message.
func (m *Messenger) Info(msg string, args ...any) {
	if m.logLevel < INFO {
		return
	}
	if m.quietMode && m.logLevel < DEBUG {
		return
	}
	m.print(nil, msg, args...)
}

// Success prints a success message.
func (m *Messenger) Success(msg string, args ...any) {
	if m.logLevel < INFO {
		return
	}
	if m.quietMode && m.logLevel < DEBUG {
		return
	}
	m.print(pterm.Success.MessageStyle, msg, args...)
}

// Debug prints a debug message.
func (m *Messenger) Debug(msg string, args ...any) {
	if m.logLevel < DEBUG {
		return
	}
	m.print(pterm.Debug.MessageStyle, msg, args...)
}

// Die prints an error message and exits the program.
func (m *Messenger) Die(msg string, args ...any) {
	m.Err(msg, args...)
	os.Exit(m.exitStatus)
}

// ExitCode sets the exit code for the program.
func (m *Messenger) ExitCode(exitStatus int) {
	m.Lock()
	m.exitStatus = exitStatus
	m.Unlock()
}

// ExitCode sets the exit code for the program.
func ExitCode(exitStatus int) {
	defaultMsg.ExitCode(exitStatus)
}

// print prints a message with the given style.
func (m *Messenger) print(style *pterm.Style, msg string, args ...any) {
	m.Lock()
	defer m.Unlock()
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	if style == nil {
		pterm.Printf(msg, args...)
		return
	}

	style.Printf(msg, args...)
}

// HasErrored returns true if an error has occurred.
func (m *Messenger) HasErrored() bool {
	return m.hasError
}

// SetQuietMode sets the quiet mode flag.
func SetQuietMode(quiet bool) {
	defaultMsg.SetQuietMode(quiet)
}

// SetQuietMode sets the quiet mode flag.
func (m *Messenger) SetQuietMode(quiet bool) {
	m.Lock()
	m.quietMode = quiet
	m.Unlock()
}

// SetProgressTracker sets the progress tracker.
func SetProgressTracker(tracker Stoppable) {
	defaultMsg.SetProgressTracker(tracker)
}

// SetProgressTracker sets the progress tracker.
func (m *Messenger) SetProgressTracker(tracker Stoppable) {
	m.Lock()
	m.progressTracker = tracker
	m.Unlock()
}
