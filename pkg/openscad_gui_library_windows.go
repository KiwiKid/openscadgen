//go:build windows

package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// defaultOpenSCADUserLibraryDir returns [Documents]/OpenSCAD/libraries. The
// Documents folder is resolved through the shell known folder because it can be
// redirected (for example to OneDrive or another drive).
func defaultOpenSCADUserLibraryDir() (string, error) {
	documents, err := windowsDocumentsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(documents, "OpenSCAD", "libraries"), nil
}

func windowsDocumentsDir() (string, error) {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		defer key.Close()
		if value, _, valueErr := key.GetStringValue("Personal"); valueErr == nil && strings.TrimSpace(value) != "" {
			if expanded, expandErr := expandWindowsEnvironment(value); expandErr == nil {
				return expanded, nil
			}
			return value, nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents"), nil
}

// expandWindowsEnvironment expands REG_EXPAND_SZ values such as
// %USERPROFILE%\Documents.
func expandWindowsEnvironment(value string) (string, error) {
	source, err := windows.UTF16PtrFromString(value)
	if err != nil {
		return "", err
	}
	buffer := make([]uint16, 4096)
	for {
		length, err := windows.ExpandEnvironmentStrings(source, &buffer[0], uint32(len(buffer)))
		if err != nil {
			return "", err
		}
		if length == 0 {
			return "", fmt.Errorf("ExpandEnvironmentStrings returned an empty result")
		}
		if int(length) <= len(buffer) {
			return windows.UTF16ToString(buffer[:length-1]), nil
		}
		buffer = make([]uint16, length)
	}
}
