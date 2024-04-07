package util

import (
	"io"
	"os"
)

func MustLoad(open *os.File, err error) io.ReaderAt {
	if err != nil {
		panic(err)
	}
	return open
}
func MustOpen(filename string) io.ReadCloser {
	open, _ := os.Open(filename)
	return open
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
