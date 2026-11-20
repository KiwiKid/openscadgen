package pkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureOpenSCADGUILibraryPath makes the libraries managed by OpenSCADGen
// visible to the OpenSCAD desktop application.
//
// OpenSCAD only reads OPENSCADPATH when it starts. When OpenSCAD is already
// running, a second launch is forwarded to that instance and the environment
// is ignored, so the GUI cannot resolve include <BOSL2/std.scad>. Linking the
// managed libraries into OpenSCAD's own user library folder makes them
// resolvable even then. Existing entries in that folder are never replaced.
func EnsureOpenSCADGUILibraryPath() error {
	managedDir, err := managedOpenSCADLibraryDir()
	if err != nil {
		return fmt.Errorf("find managed OpenSCAD libraries: %w", err)
	}
	userDir, err := userOpenSCADLibraryDir()
	if err != nil {
		return fmt.Errorf("find OpenSCAD user library folder: %w", err)
	}

	managedAbs, err := filepath.Abs(managedDir)
	if err != nil {
		return fmt.Errorf("resolve managed library folder: %w", err)
	}
	userAbs, err := filepath.Abs(userDir)
	if err != nil {
		return fmt.Errorf("resolve OpenSCAD user library folder: %w", err)
	}
	if samePath(managedAbs, userAbs) {
		return nil
	}

	entries, err := os.ReadDir(managedAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read managed OpenSCAD libraries: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		// Skip staging/backup directories such as .openscadgen-bosl2-*.
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		names = append(names, entry.Name())
	}
	if len(names) == 0 {
		return nil
	}
	if err := os.MkdirAll(userAbs, 0o755); err != nil {
		return fmt.Errorf("create OpenSCAD user library folder: %w", err)
	}

	var linkErrors []error
	for _, name := range names {
		target := filepath.Join(managedAbs, name)
		link := filepath.Join(userAbs, name)
		info, err := os.Lstat(link)
		switch {
		case err == nil:
			if info.Mode()&os.ModeSymlink == 0 {
				// A real user copy wins; do not replace their library.
				continue
			}
			if _, statErr := os.Stat(link); statErr == nil {
				continue
			}
			// A stale link (app moved or library removed) is safe to refresh.
			if removeErr := os.Remove(link); removeErr != nil {
				linkErrors = append(linkErrors, fmt.Errorf("refresh stale link %s: %w", link, removeErr))
				continue
			}
		case !os.IsNotExist(err):
			linkErrors = append(linkErrors, fmt.Errorf("inspect %s: %w", link, err))
			continue
		}
		if err := createDirectoryLink(target, link); err != nil {
			linkErrors = append(linkErrors, fmt.Errorf("link %s into OpenSCAD's library folder: %w", name, err))
			continue
		}
		LogInfof("Linked %s into OpenSCAD's user library folder at %s", name, link)
	}
	return errors.Join(linkErrors...)
}

func samePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// userOpenSCADLibraryDir returns the folder OpenSCAD scans for user libraries.
// The environment override keeps tests and unusual setups isolated.
func userOpenSCADLibraryDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("OPENSCADGEN_OPENSCAD_USER_LIBRARY_DIR")); override != "" {
		return filepath.Abs(override)
	}
	return defaultOpenSCADUserLibraryDir()
}
