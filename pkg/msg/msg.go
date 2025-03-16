package msg

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/pterm/pterm"
)

type Messenger struct {
	sync.Mutex
	Stdout     io.Writer
	Stderr     io.Writer
	Stdin      io.Reader
	PanicOnDie bool
	eCode      int
	hasError   bool
}

func NewMessenger() *Messenger {
	m := &Messenger{
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Stdin:      os.Stdin,
		PanicOnDie: false,
		eCode:      1,
	}

	return m
}

var DebugEnable bool
var Default = NewMessenger()

func Die(msg string, args ...any) {
	Default.Die(msg, args...)
}

func Info(msg string, args ...any) {
	Default.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	Default.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Default.Warn(msg, args...)
}

func Err(msg string, args ...any) {
	Default.Err(msg, args...)
}

func (m *Messenger) Err(msg string, args ...any) {
	m.print(pterm.Error, msg, args...)
	m.hasError = true
}

func (m *Messenger) Warn(msg string, args ...any) {
	m.print(pterm.Warning, msg, args...)
}

func (m *Messenger) Info(msg string, args ...any) {
	m.print(pterm.Info, msg, args...)
}

func (m *Messenger) Debug(msg string, args ...any) {
	if !DebugEnable {
		return
	}

	m.print(pterm.Debug, msg, args...)
}

func (m *Messenger) Die(msg string, args ...any) {
	m.Err(msg, args...)
	if m.PanicOnDie {
		panic("trapped a Die() call")
	}
	os.Exit(m.eCode)
}

func (m *Messenger) ExitCode(e int) int {
	m.Lock()
	old := m.eCode
	m.eCode = e
	m.Unlock()
	return old
}

func ExitCode(e int) int {
	return Default.ExitCode(e)
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
