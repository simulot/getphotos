package main

import (
	"errors"
	"io/fs"
	"os"
	"runtime"
	"strings"
)

type RemoveFS interface {
	fs.FS
	Remove(name string) error
}

func RemoveDirFS(dir string) fs.FS {
	return &removeDirFS{
		FS:  os.DirFS(dir),
		dir: dir,
	}

}

type removeDirFS struct {
	fs.FS
	dir string
}

func (fsys removeDirFS) Remove(name string) error {
	if !fs.ValidPath(name) || runtime.GOOS == "windows" && strings.ContainsAny(name, `\:`) {
		return &os.PathError{Op: "remove", Path: name, Err: os.ErrInvalid}
	}
	return os.Remove(fsys.dir + "/" + name)
}

func Remove(fsys fs.FS, name string) error {
	if rdFS, ok := fsys.(RemoveFS); ok {
		return rdFS.Remove(name)
	}
	return errors.New("remove not implemented")
}
