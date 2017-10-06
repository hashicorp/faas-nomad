package util

import (
	"io"
	"os"
)

// NewFile is a convenience function which creates and opens a file
func NewFile(filename string) io.Writer {
	if err, _ := os.Open(filename); err != nil {
		os.Remove(filename)
	}

	file, _ := os.Create(filename)

	return file
}
