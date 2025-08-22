package simtest

import (
	"io/fs"
)

// A thin wrapper around the I/O API that the application
// uses.
type IO interface {
	CreateFile(name string) (fs.File, error)
	OpenFile(name string, flag int, mode fs.FileMode) (fs.File, error)
}