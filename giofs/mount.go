package giofs

import (
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var gioChecked sync.Once

var gioCmd = struct {
	version string
	err     error
}{}

// gioCheck checks presence of gio command on the system at the first call. Return initial result
func gioCheck() error {
	gioChecked.Do(func() {
		var out []byte
		out, gioCmd.err = exec.Command("gio", "version").Output()
		if gioCmd.err == nil {
			gioCmd.version = strings.TrimSuffix(string(out), "\n")
		}
	})

	return gioCmd.err
}

var (
	reGIO        = regexp.MustCompile(`(?m)^\s+Mount\(\d+\):\s([[:print:]]*)\s->\s([[:print:]]+)$`)
	reMountPoint = regexp.MustCompile(`(\w+)://(.+)`)
)

type MountInfo struct {
	DeviceName string
	URI        string
	Protocole  string
}

// MountList return the list of removable medias mounted on the computer using `gio mount -l` command
func MountList() (MountInfos []MountInfo, err error) {
	out, err := gio("mount", "-l")
	if err != nil {
		return nil, err
	}

	ms := reGIO.FindAllStringSubmatch(string(out), -1)
	for _, m := range ms {
		if len(m) != 3 {
			continue
		}
		if m[1] == "mtp" {
			continue
		}
		p := reMountPoint.FindStringSubmatch(m[2])
		if len(p) != 3 {
			continue
		}
		uri, err := url.PathUnescape(m[2])
		if err == nil {
			MountInfos = append(MountInfos, MountInfo{
				DeviceName: m[1],
				Protocole:  p[1],
				URI:        uri,
			})
		}
	}
	return
}

// gio runs the gio command and returns the output and error
func gio(cmd string, args ...string) (out []byte, err error) {
	if err = gioCheck(); err != nil {
		return
	}
	args = append([]string{cmd}, args...)
	out, err = exec.Command("gio", args...).Output()
	return
}
