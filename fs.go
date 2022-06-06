package main

import (
	"errors"
	"io/fs"
)

/*
	FS helpers

*/

type RemoveFS interface {
	fs.FS
	Remove(name string) error
}

func Remove(fsys fs.FS, name string) error {
	if rdFS, ok := fsys.(RemoveFS); ok {
		return rdFS.Remove(name)
	}
	return errors.New("remove not implemented")
}
