package msg

import (
	"io"
	"sync"
)

//provenied by glide https://github.com/Masterminds/glide

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
	ecode      int
	hasErrored bool
}

func NewMessenger() *Messenger {
	m := &Messenger{
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Stdin:      os.Stdin,
		PanicOnDie: false,
		ecode:      1,
	}

	return m
}

var Default = NewMessenger()

func (m *Messenger) Info(msg string, args ...interface{}) {
	m.Msg("[INFO]\t"+msg, args...)
}

func Info(msg string, args ...interface{}) {
	Default.Info(msg, args...)
}

func (m *Messenger) Debug(msg string, args ...interface{}) {
	m.Msg("[DEBUG]\t"+msg, args...)
}

func Debug(msg string, args ...interface{}) {
	Default.Debug(msg, args...)
}

func (m *Messenger) Warn(msg string, args ...interface{}) {
	m.Msg("[WARN]\t"+msg, args...)
}

func Warn(msg string, args ...interface{}) {
	Default.Warn(msg, args...)
}

func (m *Messenger) Err(msg string, args ...interface{}) {
	m.Msg("[ERROR]\t"+msg, args...)
	m.hasErrored = true
}

func Err(msg string, args ...interface{}) {
	Default.Err(msg, args...)
}

func (m *Messenger) Die(msg string, args ...interface{}) {
	m.Err(msg, args...)
	if m.PanicOnDie {
		panic("trapped a Die() call")
	}
	os.Exit(m.ecode)
}

func Die(msg string, args ...interface{}) {
	Default.Die(msg, args...)
}

func (m *Messenger) ExitCode(e int) int {
	m.Lock()
	old := m.ecode
	m.ecode = e
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
		fmt.Fprint(m.Stderr, msg)
	} else {
		fmt.Fprintf(m.Stderr, msg, args...)
	}
}

func Msg(msg string, args ...interface{}) {
	Default.Msg(msg, args...)
}

func (m *Messenger) Puts(msg string, args ...interface{}) {
	m.Lock()
	defer m.Unlock()

	fmt.Fprintf(m.Stdout, msg, args...)
	fmt.Fprintln(m.Stdout)
}

func Puts(msg string, args ...interface{}) {
	Default.Puts(msg, args...)
}

func (m *Messenger) Print(msg string) {
	m.Lock()
	defer m.Unlock()

	fmt.Fprintln(m.Stdout, msg)
}

func Print(msg string) {
	Default.Print(msg)
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

func PromptUntil(opts []string) (string, error) {
	return Default.PromptUntil(opts)
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

func PromptUntilYorN() bool {
	return Default.PromptUntilYorN()
}
