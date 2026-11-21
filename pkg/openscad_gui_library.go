package pkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// openSCADManagedCopyMarker identifies library folders that OpenSCADGen copied
// into OpenSCAD's own user library folder.
const openSCADManagedCopyMarker = ".openscadgen-copy"

// EnsureOpenSCADGUILibraryPath makes the libraries managed by OpenSCADGen
// visible to the OpenSCAD desktop application.
//
// OpenSCAD only reads OPENSCADPATH when it starts. When OpenSCAD is already
// running, a second launch is forwarded to that instance and the environment
// is ignored, so the GUI cannot resolve include <BOSL2/std.scad>. OpenSCAD also
// crashes on symlinked library folders, so a managed copy is placed in
// OpenSCAD's own user library folder instead. User-owned folders are never
// replaced.
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

	var copyErrors []error
	for _, definition := range openSCADLibraryCatalog {
		sourceDir := filepath.Join(managedAbs, definition.Folder)
		marker := filepath.Join(sourceDir, definition.MarkerFile)
		if info, statErr := os.Stat(marker); statErr != nil || info.IsDir() {
			continue
		}
		version := readOpenSCADLibraryVersion(sourceDir, definition)
		if err := syncOpenSCADLibraryCopy(sourceDir, filepath.Join(userAbs, definition.Folder), version); err != nil {
			copyErrors = append(copyErrors, fmt.Errorf("copy %s into OpenSCAD's user library folder: %w", definition.Name, err))
		}
	}
	return errors.Join(copyErrors...)
}

// syncOpenSCADLibraryCopy refreshes OpenSCADGen's copy of a library when it is
// missing or out of date. Folders without the marker belong to the user and are
// left untouched.
func syncOpenSCADLibraryCopy(sourceDir, targetDir, version string) error {
	markerPath := filepath.Join(targetDir, openSCADManagedCopyMarker)
	var existingMarker []byte
	hasExistingMarker := false
	info, err := os.Lstat(targetDir)
	switch {
	case err == nil:
		if !info.IsDir() {
			return nil
		}
		existingMarker, err = os.ReadFile(markerPath)
		if err != nil {
			// Not created by OpenSCADGen; leave the user's library alone.
			return nil
		}
		hasExistingMarker = true
		if version != "" && strings.TrimSpace(string(existingMarker)) == version {
			return nil
		}
	case !os.IsNotExist(err):
		return err
	}

	if err := replaceDirectoryContents(sourceDir, targetDir); err != nil {
		// A failed refresh must not orphan the copy: without its marker the
		// next attempt would treat it as a user-owned library and skip it.
		if hasExistingMarker {
			_ = os.WriteFile(markerPath, existingMarker, 0o644)
		}
		return err
	}
	if err := os.WriteFile(markerPath, []byte(version+"\n"), 0o644); err != nil {
		return err
	}
	LogInfof("Updated OpenSCAD's copy of %s at %s", filepath.Base(targetDir), targetDir)
	return nil
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
