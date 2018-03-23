package base

import (
	"net/http"
	"os"
)

type HideDirFS struct {
	FileSystem http.FileSystem
}

type hideFile struct {
	http.File
}

func (fs HideDirFS) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return hideFile{File: f}, nil
}

func (f hideFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}
