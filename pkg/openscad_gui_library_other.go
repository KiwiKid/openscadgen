//go:build !windows

package pkg

import (
	"os"
	"path/filepath"
	"runtime"
)

func defaultOpenSCADUserLibraryDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Documents", "OpenSCAD", "libraries"), nil
	}
	return filepath.Join(home, ".local", "share", "OpenSCAD", "libraries"), nil
}

func createDirectoryLink(target, link string) error {
	return os.Symlink(target, link)
}
