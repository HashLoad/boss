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

func Success(msg string, args ...any) {
	defaultMsg.Success(msg, args...)
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

	if m.progressTracker != nil {
		m.progressTracker.Stop()
		m.progressTracker = nil
	}

	m.quietMode = false

	m.print(pterm.Error.MessageStyle, msg, args...)
	m.hasError = true
}

func (m *Messenger) Warn(msg string, args ...any) {
	if m.logLevel < WARN {
		return
	}

	wasQuiet := m.quietMode
	m.quietMode = false

	m.print(pterm.Warning.MessageStyle, msg, args...)

	m.quietMode = wasQuiet
}

func (m *Messenger) Info(msg string, args ...any) {
	if m.quietMode || m.logLevel < INFO {
		return
	}
	m.print(pterm.Info.MessageStyle, msg, args...)
}

func (m *Messenger) Success(msg string, args ...any) {
	if m.quietMode || m.logLevel < INFO {
		return
	}
	m.print(pterm.Success.MessageStyle, msg, args...)
}

func (m *Messenger) Debug(msg string, args ...any) {
	if m.quietMode || m.logLevel < DEBUG {
		return
	}

	m.print(pterm.Debug.MessageStyle, msg, args...)
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

func SetProgressTracker(tracker Stoppable) {
	defaultMsg.SetProgressTracker(tracker)
}

func (m *Messenger) SetProgressTracker(tracker Stoppable) {
	m.Lock()
	m.progressTracker = tracker
	m.Unlock()
}
