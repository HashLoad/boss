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

type Messenger struct {
	sync.Mutex
	Stdout          io.Writer
	Stderr          io.Writer
	Stdin           io.Reader
	exitStatus      int
	hasError        bool
	quietMode       bool
	progressTracker any

	logLevel logLevel
}

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

//nolint:gochecknoglobals // This is a global variable
var defaultMsg = NewMessenger()

func Die(msg string, args ...any) {
	defaultMsg.Die(msg, args...)
}

func Info(msg string, args ...any) {
	defaultMsg.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	defaultMsg.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultMsg.Warn(msg, args...)
}

func Err(msg string, args ...any) {
	defaultMsg.Err(msg, args...)
}

func LogLevel(level logLevel) {
	defaultMsg.LogLevel(level)
}

func (m *Messenger) LogLevel(level logLevel) {
	m.Lock()
	m.logLevel = level
	m.Unlock()
}

func (m *Messenger) Err(msg string, args ...any) {
	if m.logLevel < ERROR {
		return
	}

	// Stop progress tracker if running
	if m.progressTracker != nil {
		if tracker, ok := m.progressTracker.(interface{ Stop() }); ok {
			tracker.Stop()
		}
		m.progressTracker = nil
	}

	// Disable quiet mode to show errors
	m.quietMode = false

	m.print(pterm.Error, msg, args...)
	m.hasError = true
}

func (m *Messenger) Warn(msg string, args ...any) {
	if m.logLevel < WARN {
		return
	}

	// Warnings don't stop the progress tracker, only errors do
	// But we need to temporarily disable quiet mode to show the warning
	wasQuiet := m.quietMode
	m.quietMode = false

	m.print(pterm.Warning, msg, args...)

	// Restore quiet mode after printing warning
	m.quietMode = wasQuiet
}

func (m *Messenger) Info(msg string, args ...any) {
	if m.quietMode || m.logLevel < INFO {
		return
	}
	m.print(pterm.Info, msg, args...)
}

func (m *Messenger) Debug(msg string, args ...any) {
	if m.quietMode || m.logLevel < DEBUG {
		return
	}

	m.print(pterm.Debug, msg, args...)
}

func (m *Messenger) Die(msg string, args ...any) {
	m.Err(msg, args...)
	os.Exit(m.exitStatus)
}

func (m *Messenger) ExitCode(exitStatus int) {
	m.Lock()
	m.exitStatus = exitStatus
	m.Unlock()
}

func ExitCode(exitStatus int) {
	defaultMsg.ExitCode(exitStatus)
}

func (m *Messenger) print(printer pterm.PrefixPrinter, msg string, args ...any) {
	m.Lock()
	defer m.Unlock()
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	printer.Printf(msg, args...)
}

func (m *Messenger) HasErrored() bool {
	return m.hasError
}

func SetQuietMode(quiet bool) {
	defaultMsg.SetQuietMode(quiet)
}

func (m *Messenger) SetQuietMode(quiet bool) {
	m.Lock()
	m.quietMode = quiet
	m.Unlock()
}

func SetProgressTracker(tracker any) {
	defaultMsg.SetProgressTracker(tracker)
}

func (m *Messenger) SetProgressTracker(tracker any) {
	m.Lock()
	m.progressTracker = tracker
	m.Unlock()
}
