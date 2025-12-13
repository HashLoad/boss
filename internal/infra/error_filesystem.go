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

func (l *ErrorFileSystem) ReadFile(path string) ([]byte, error) {
	return nil, errors.New("IO operation not allowed in domain: ReadFile")
}

func (l *ErrorFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return errors.New("IO operation not allowed in domain: WriteFile")
}

func (l *ErrorFileSystem) Stat(path string) (os.FileInfo, error) {
	return nil, errors.New("IO operation not allowed in domain: Stat")
}

func (l *ErrorFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return errors.New("IO operation not allowed in domain: MkdirAll")
}

func (l *ErrorFileSystem) Remove(path string) error {
	return errors.New("IO operation not allowed in domain: Remove")
}

func (l *ErrorFileSystem) RemoveAll(path string) error {
	return errors.New("IO operation not allowed in domain: RemoveAll")
}

func (l *ErrorFileSystem) Rename(oldpath, newpath string) error {
	return errors.New("IO operation not allowed in domain: Rename")
}

func (l *ErrorFileSystem) Open(name string) (io.ReadCloser, error) {
	return nil, errors.New("IO operation not allowed in domain: Open")
}

func (l *ErrorFileSystem) Create(name string) (io.WriteCloser, error) {
	return nil, errors.New("IO operation not allowed in domain: Create")
}

func (l *ErrorFileSystem) Exists(name string) bool {
	return false
}

func (l *ErrorFileSystem) IsDir(name string) bool {
	return false
}
