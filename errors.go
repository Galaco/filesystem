package filesystem

import (
	"fmt"
	"strings"
)

// FileNotFoundError
type FileNotFoundError struct {
	fileName string
}

// Error
func (err FileNotFoundError) Error() string {
	return fmt.Sprintf("%s not found in filesystem", err.fileName)
}

// NewFileNotFoundError
func NewFileNotFoundError(filename string) *FileNotFoundError {
	return &FileNotFoundError{
		fileName: filename,
	}
}

// InvalidResourcePathCollectionError
type InvalidResourcePathCollectionError struct {
	paths []string
}

// Error will return a list of paths that could not be added.
// This list is a pipe-separated(|) string
func (err InvalidResourcePathCollectionError) Error() string {
	msg := ""
	for _,p := range err.paths {
		msg += p + "|"
	}
	return strings.Trim(msg, "|")
}

// AddPath adds a new path to this error colleciton
func (err InvalidResourcePathCollectionError) AddPath(path string) {
	err.paths = append(err.paths, path)
}

// NewInvalidResourcePathCollectionError
func NewInvalidResourcePathCollectionError() *InvalidResourcePathCollectionError {
	return &InvalidResourcePathCollectionError{
		paths: make([]string, 0),
	}
}
