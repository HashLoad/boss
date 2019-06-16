package msg

import (
	"io"
	"sync"
)

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Messenger struct {
	sync.Mutex
	Stdout     io.Writer
	Stderr     io.Writer
	Stdin      io.Reader
	PanicOnDie bool
	eCode      int
	hasErrored bool
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

func Die(msg string, args ...interface{}) {
	Default.Die(msg, args...)
}

func Msg(msg string, args ...interface{}) {
	Default.Msg(msg, args...)
}

func Puts(msg string, args ...interface{}) {
	Default.Puts(msg, args...)
}

func Print(msg string) {
	Default.Print(msg)
}

func PromptUntil(opts []string) (string, error) {
	return Default.PromptUntil(opts)
}

func PromptUntilYorN() bool {
	return Default.PromptUntilYorN()
}

func Info(msg string, args ...interface{}) {
	Default.Info(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	Default.Debug(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	Default.Warn(msg, args...)
}

func Err(msg string, args ...interface{}) {
	Default.Err(msg, args...)
}

func (m *Messenger) Err(msg string, args ...interface{}) {
	m.Msg("[ERROR]\t"+msg, args...)
	m.hasErrored = true
}

func (m *Messenger) Warn(msg string, args ...interface{}) {
	m.Msg("[WARN ]\t"+msg, args...)
}

func (m *Messenger) Info(msg string, args ...interface{}) {
	m.Msg("[INFO ]\t"+msg, args...)
}

func (m *Messenger) Debug(msg string, args ...interface{}) {
	if !DebugEnable {
		return
	}
	m.Msg("[DEBUG]\t"+msg, args...)
}

func (m *Messenger) Die(msg string, args ...interface{}) {
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

func (m *Messenger) Msg(msg string, args ...interface{}) {
	m.Lock()
	defer m.Unlock()
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	if len(args) == 0 {
		if _, err := fmt.Fprint(m.Stderr, msg); err != nil {
			println("[Fault] %s", err.Error())
		}
	} else {
		if _, err := fmt.Fprintf(m.Stderr, msg, args...); err != nil {
			println("[Fault] %s", err.Error())
		}
	}
}

func (m *Messenger) Puts(msg string, args ...interface{}) {
	m.Lock()
	defer m.Unlock()
	if _, err := fmt.Fprintf(m.Stderr, msg, args...); err != nil {
		println("[Fault] %s", err.Error())
	}
	if _, err := fmt.Fprintln(m.Stderr); err != nil {
		println("[Fault] %s", err.Error())
	}
}

func (m *Messenger) Print(msg string) {
	m.Puts(msg)
}

func (m *Messenger) HasErrored() bool {
	return m.hasErrored
}

func (m *Messenger) PromptUntil(opts []string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		for _, c := range opts {
			if strings.EqualFold(c, strings.TrimSpace(text)) {
				return c, nil
			}
		}
	}
}

func (m *Messenger) PromptUntilYorN() bool {
	res, err := m.PromptUntil([]string{"y", "yes", "n", "no"})
	if err != nil {
		m.Die("Error processing response: %s", err)
	}

	if res == "y" || res == "yes" {
		return true
	}

	return false
}
