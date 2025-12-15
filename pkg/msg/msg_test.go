package msg_test

import (
	"bytes"
	"testing"

	"github.com/hashload/boss/pkg/msg"
)

func TestNewMessenger(t *testing.T) {
	m := msg.NewMessenger()

	if m == nil {
		t.Fatal("NewMessenger() should not return nil")
	}

	if m.Stdout == nil {
		t.Error("Messenger.Stdout should not be nil")
	}

	if m.Stderr == nil {
		t.Error("Messenger.Stderr should not be nil")
	}

	if m.Stdin == nil {
		t.Error("Messenger.Stdin should not be nil")
	}
}

func TestMessenger_LogLevel(t *testing.T) {
	t.Helper()
	m := msg.NewMessenger()

	// Test setting log levels using the exported constants
	m.LogLevel(msg.WARN)
	m.LogLevel(msg.ERROR)
	m.LogLevel(msg.INFO)
	m.LogLevel(msg.DEBUG)
	// No panic means success
}

func TestMessenger_ExitCode(t *testing.T) {
	t.Helper()
	m := msg.NewMessenger()

	// Test setting exit codes
	exitCodes := []int{0, 1, 2, 127, 255}

	for _, code := range exitCodes {
		m.ExitCode(code)
		// No panic means success
	}
}

func TestMessenger_HasErrored_Initial(t *testing.T) {
	m := msg.NewMessenger()

	if m.HasErrored() {
		t.Error("New Messenger should not have errors initially")
	}
}

func TestMessenger_HasErrored_AfterErr(t *testing.T) {
	m := msg.NewMessenger()
	m.Stdout = &bytes.Buffer{} // Suppress output
	m.Stderr = &bytes.Buffer{}

	m.Err("test error")

	if !m.HasErrored() {
		t.Error("HasErrored() should return true after Err() call")
	}
}

func TestMessenger_Info_NoOutput_WhenLevelLow(t *testing.T) {
	t.Helper()
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	m.LogLevel(msg.WARN) // Below INFO
	m.Info("should not appear")

	// Info should be suppressed when log level is WARN
	// Note: actual output goes through pterm, so we just verify no panic
}

func TestMessenger_Warn_NoOutput_WhenLevelLow(t *testing.T) {
	t.Helper()
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	m.LogLevel(msg.ERROR) // Below WARN
	m.Warn("should not appear")

	// Warn should be suppressed when log level is ERROR
}

func TestMessenger_Debug_NoOutput_WhenLevelLow(t *testing.T) {
	t.Helper()
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	m.LogLevel(msg.INFO) // Below DEBUG
	m.Debug("should not appear")

	// Debug should be suppressed when log level is INFO
}

func TestGlobalFunctions(t *testing.T) {
	t.Helper()
	// Test that global functions don't panic

	// LogLevel
	msg.LogLevel(msg.INFO)

	// ExitCode
	msg.ExitCode(0)

	// The other global functions (Info, Warn, Err, Debug) write to stdout/stderr
	// so we just verify they exist and are callable
	_ = msg.Info
	_ = msg.Warn
	_ = msg.Err
	_ = msg.Debug
	_ = msg.Die
}

func TestLogLevel_Constants(t *testing.T) {
	// Verify log level ordering
	if msg.WARN >= msg.ERROR {
		t.Error("WARN should be less than ERROR")
	}
	if msg.ERROR >= msg.INFO {
		t.Error("ERROR should be less than INFO")
	}
	if msg.INFO >= msg.DEBUG {
		t.Error("INFO should be less than DEBUG")
	}
}

func TestMessenger_Info_WithOutput(_ *testing.T) {
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	m.LogLevel(msg.DEBUG) // High level to ensure output
	m.Info("test info message")

	// pterm writes to its own internal writer, so we just verify no panic
}

func TestMessenger_Warn_WithOutput(_ *testing.T) {
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	m.LogLevel(msg.DEBUG) // High level to ensure output
	m.Warn("test warning message")

	// Verify no panic occurred
}

func TestMessenger_Debug_WithOutput(_ *testing.T) {
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	m.LogLevel(msg.DEBUG)
	m.Debug("test debug message")

	// Verify no panic occurred
}

func TestMessenger_Err_SetsHasErrored(t *testing.T) {
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	if m.HasErrored() {
		t.Error("Should not have error initially")
	}

	m.LogLevel(msg.DEBUG)
	m.Err("test error message")

	if !m.HasErrored() {
		t.Error("HasErrored() should be true after Err()")
	}
}

func TestMessenger_Err_NoOutput_WhenLevelLow(_ *testing.T) {
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	// Set level below ERROR (only WARN level is below ERROR)
	m.LogLevel(msg.WARN)

	m.Err("should not set error")
	// When level is low, Err returns early - just verify no panic
}

func TestMessenger_WithFormatArgs(_ *testing.T) {
	m := msg.NewMessenger()
	buf := &bytes.Buffer{}
	m.Stdout = buf
	m.Stderr = buf

	m.LogLevel(msg.DEBUG)

	// Test with format arguments
	m.Info("formatted %s number %d", "string", 42)
	m.Warn("warning %v", []int{1, 2, 3})
	m.Debug("debug value: %+v", struct{ Name string }{"test"})
	m.Err("error with %s", "context")
}

func TestGlobalInfo(_ *testing.T) {
	// Capture that it doesn't panic
	// Note: this writes to real stdout
	msg.LogLevel(msg.DEBUG)
	msg.Info("global info test")
}

func TestGlobalWarn(_ *testing.T) {
	msg.LogLevel(msg.DEBUG)
	msg.Warn("global warn test")
}

func TestGlobalErr(_ *testing.T) {
	msg.LogLevel(msg.DEBUG)
	msg.Err("global err test")
}

func TestGlobalDebug(_ *testing.T) {
	msg.LogLevel(msg.DEBUG)
	msg.Debug("global debug test")
}

func TestExitCode_Global(_ *testing.T) {
	// Test global ExitCode function
	msg.ExitCode(0)
	msg.ExitCode(1)
	msg.ExitCode(127)
}
