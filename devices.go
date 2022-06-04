package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type PhotoDevice struct {
	Name string
	Path string
}

func SearchDCIM(ctx context.Context) (devs []PhotoDevice, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("can't search photo device: %w", err)
		}
	}()

	gvfsMount, err := getGVFSMountPoint()
	if err != nil {
		return
	}

	mounts, err := parseGIOMount(gvfsMount)
	if err != nil {
		return nil, err
	}

	for _, m := range mounts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			fs, err := filepath.Glob(m.Path)
			if err != nil {
				continue
			}
			fmt.Println("Found device", m.Name)
			for _, s := range fs {
				m.Path = s
				devs = append(devs, m)
			}
		}
	}
	return
}

var (
	reGIO        = regexp.MustCompile(`(?m)^\s*Mount\(\d+\):\s([[:print:]]*)\s->\s([[:print:]]+)$`)
	reMountPoint = regexp.MustCompile(`(\w+)://(.+)`)
)

func parseGIOMount(gvfsMountPoint string) (path []PhotoDevice, err error) {
	buf := bytes.NewBuffer(nil)
	c := exec.Command("gio", "mount", "-l")
	c.Stdout = buf
	err = c.Run()
	if err != nil {
		return nil, fmt.Errorf("can't get 'gio mount -l': %w", err)
	}

	ms := reGIO.FindAllStringSubmatch(buf.String(), -1)
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
		switch p[1] {
		case "file":
			path = append(path, PhotoDevice{Name: m[1], Path: filepath.Join(p[2], "DCIM")})
		case "mtp":
			path = append(path, PhotoDevice{Name: m[1], Path: filepath.Join(gvfsMountPoint, p[1]+":host="+p[2], "*", "DCIM")})
		default:
			path = append(path, PhotoDevice{Name: m[1], Path: filepath.Join(gvfsMountPoint, p[1]+":host="+p[2], "DCIM")})
		}
	}
	return
}

var reFuseMount = regexp.MustCompile(`(?m)^gvfsd-fuse\s(\S+)`)

func getGVFSMountPoint() (string, error) {
	b, err := os.ReadFile("/proc/self/mounts")
	if err != nil {
		return "", err
	}
	m := reFuseMount.FindStringSubmatch(string(b))
	if len(m) == 0 {
		return "", fmt.Errorf("gvfs mount point not found: %w", os.ErrNotExist)
	}
	return m[1], nil
}
