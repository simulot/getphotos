package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/simulot/lib/myflag"
)

func GetXDGDirectory(dir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("can't get user's home directory: %w", err)
	}
	config := os.Getenv("XDG_CONFIG_HOME")
	if config == "" {
		config = filepath.Join(home, ".config")
	}
	config = filepath.Join(config, "user-dirs.dirs")

	kv, err := myflag.ReadKVFile(config)
	if err != nil {
		return "", fmt.Errorf("can't get user's home directory: %w", err)
	}

	r, ok := kv[dir]
	if !ok {
		r = os.ExpandEnv(filepath.Join(home, dir))
		return r, fmt.Errorf("desktop directory '%s' not found, defaulted to '%s'", dir, r)
	}
	r = os.ExpandEnv(r)
	return r, nil
}
