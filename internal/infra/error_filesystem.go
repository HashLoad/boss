// Package infra provides error-returning filesystem implementation.
// ErrorFileSystem prevents accidental I/O in tests by returning errors for all operations.
package infra

import (
	"errors"
	"io"
	"os"
)

// ErrorFileSystem is a FileSystem implementation that returns errors for all operations.
// This is used as a default in the domain layer to prevent implicit I/O.
type ErrorFileSystem struct{}

// NewErrorFileSystem creates a new ErrorFileSystem.
func NewErrorFileSystem() *ErrorFileSystem {
	return &ErrorFileSystem{}
}

func (l *ErrorFileSystem) ReadFile(_ string) ([]byte, error) {
	return nil, errors.New("IO operation not allowed in domain: ReadFile")
}

func (l *ErrorFileSystem) WriteFile(_ string, _ []byte, _ os.FileMode) error {
	return errors.New("IO operation not allowed in domain: WriteFile")
}

func (l *ErrorFileSystem) Stat(_ string) (os.FileInfo, error) {
	return nil, errors.New("IO operation not allowed in domain: Stat")
}

func (l *ErrorFileSystem) MkdirAll(_ string, _ os.FileMode) error {
	return errors.New("IO operation not allowed in domain: MkdirAll")
}

func (l *ErrorFileSystem) Remove(_ string) error {
	return errors.New("IO operation not allowed in domain: Remove")
}

func (l *ErrorFileSystem) RemoveAll(_ string) error {
	return errors.New("IO operation not allowed in domain: RemoveAll")
}

func (l *ErrorFileSystem) Rename(_, _ string) error {
	return errors.New("IO operation not allowed in domain: Rename")
}

func (l *ErrorFileSystem) Open(_ string) (io.ReadCloser, error) {
	return nil, errors.New("IO operation not allowed in domain: Open")
}

func (l *ErrorFileSystem) Create(_ string) (io.WriteCloser, error) {
	return nil, errors.New("IO operation not allowed in domain: Create")
}

func (l *ErrorFileSystem) Exists(_ string) bool {
	return false
}

func (l *ErrorFileSystem) IsDir(_ string) bool {
	return false
}
