package giofs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"sync"
)

type File struct {
	fsys   FS
	r      io.Reader
	info   FileInfo
	init   sync.Once
	cmd    *exec.Cmd
	opened chan any
	err    error
}

// Open opens the named file using gio cat command
// The output of the cat command is piped into the File.
func (fsys FS) Open(name string) (fs.File, error) {
	s, err := fsys.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	f := File{
		fsys:   fsys,
		info:   s.(FileInfo),
		opened: make(chan any),
	}
	return &f, err
}

// Stat returns a FileInfo describing the file.
func (f File) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *File) open() error {
	var err error
	if err = gioCheck(); err != nil {
		return err
	}

	f.cmd = exec.CommandContext(f.fsys.ctx, "gio", "cat", f.fsys.uri+f.info.name)
	f.r, err = f.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	f.err = f.cmd.Start()
	if err != nil {
		return err
	}
	close(f.opened)

	go func() {
		f.err = f.cmd.Wait()
		if err == nil {
			f.err = io.EOF
		}
	}()
	return nil
}

// Read reads up to len(b) bytes from the `gio cat` output and stores them in b.
func (f *File) Read(b []byte) (n int, err error) {
	f.init.Do(func() {
		f.open()
	})
	<-f.opened
	if f.err != nil {
		return 0, f.err
	}
	n, err = f.r.Read(b)
	if n == 0 && err != nil && err != io.EOF {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			// Clear message caused by the the end of exec command
			if pathError.Unwrap().Error() == "file already closed" {
				return 0, io.EOF
			}
		}
	}

	return
}

// Close closes the file
func (f *File) Close() error {
	return nil
}
