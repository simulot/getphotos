package giofs

import (
	"bufio"
	"bytes"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FileInfo describes a GIO file
type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

// Name give the basename of the file
func (fi FileInfo) Name() string {
	return fi.name
}

// Size gives length in bytes for regular files;
func (fi FileInfo) Size() int64 {
	return fi.size
}

// Mode gives file mode bits
func (fi FileInfo) Mode() fs.FileMode {
	return fi.mode
}

// ModTime give modification time
func (fi FileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi FileInfo) Sys() any {
	return nil
}

// IsDir is true when the file is a directory (DirEntry implementation)
func (fi FileInfo) IsDir() bool {
	return fi.isDir
}

// Type (DirEntry implementation)
func (fi FileInfo) Type() fs.FileMode {
	return fi.mode
}

// Info give the FileInfo at the time of ReadDir (DirEntry implementation)
func (fi FileInfo) Info() (fs.FileInfo, error) {
	return fi, nil
}

const (
	info_ATTRIBUTES = "time::modified,access::can-write,standard::size,standard::type"
	list_ATTRIBUTES = "time::modified,access::can-write"
)

// Stat returns a FileInfo describing the named file using `gio info` command
func (fsys *FS) Stat(name string) (fs.FileInfo, error) {
	out, err := gio("info", "-a", info_ATTRIBUTES, fsys.uri+name)
	if err != nil {
		return nil, err
	}

	info := FileInfo{
		name: name,
	}

	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		row := strings.Split(s.Text(), ": ")
		if len(row) == 2 {
			att, val := strings.TrimSpace(row[0]), strings.TrimSpace(row[1])
			info.setAttribute(att, val)
		}
	}
	err = s.Err()
	return info, err

}

var (
	splitList       = regexp.MustCompile(`(?m)^([^\t]+)\t(\d+)\t\((\w+)\)\t(.*)$`)
	splitAttributes = regexp.MustCompile(`([^=]+)=(\S+)\s?`)
)

// ReadDir reads the named directory using `gio list` command, returning all its directory entries sorted by filename.
func (fsys FS) ReadDir(name string) (entries []fs.DirEntry, err error) {
	out, err := gio("list", "-a", info_ATTRIBUTES, fsys.uri+name)
	if err != nil {
		return nil, err
	}

	rows := splitList.FindAllStringSubmatch(string(out), -1)
	for _, r := range rows {
		entry := FileInfo{}
		if len(r) >= 4 {
			entry.name = r[1]
			if s, err := strconv.ParseInt(r[2], 0, 64); err == nil {
				entry.size = s
			}
			entry.isDir = r[3] == "directory"
		}
		if len(r) >= 5 {
			aa := splitAttributes.FindAllStringSubmatch(r[4], -1)
			for _, a := range aa {
				if len(a) >= 3 {
					entry.setAttribute(a[1], a[2])
				}
			}
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	return
}

// setAttribute set the FileInfo field using attribute passed by gio command
func (fi *FileInfo) setAttribute(attribute, value string) {
	switch attribute {
	case "standard::display-name":
		fi.name = value
	case "time::modified":
		if s, err := strconv.ParseInt(value, 0, 64); err == nil {
			fi.modTime = time.Unix(s, 0)
		}
	case "access::can-read":
		if value == "TRUE" {
			fi.mode = fi.mode | 0b100
		}
	case "access::can-write":
		if value == "TRUE" {
			fi.mode = fi.mode | 0b010
		}
	case "access::can-execute":
		if value == "TRUE" {
			fi.mode = fi.mode | 0b001
		}
	case "standard::size":
		if s, err := strconv.ParseInt(value, 10, 64); err == nil {
			fi.size = s
		}
	case "standard::type":
		fi.isDir = value == "2"
	case "type":
		fi.isDir = value == "directory"
	}
}
