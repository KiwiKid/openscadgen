//go:build windows

package pkg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestInstallOrUpgradeBOSL2UpdatesFilesWhenDirectoryIsLocked covers the
// Windows case where another process (OpenSCAD, Explorer, a terminal) holds
// the BOSL2 directory open. Windows refuses to rename the directory, so the
// installer must fall back to replacing its contents in place.
func TestInstallOrUpgradeBOSL2UpdatesFilesWhenDirectoryIsLocked(t *testing.T) {
	libraryDir := t.TempDir()
	targetDir := filepath.Join(libraryDir, "BOSL2")
	if err := os.MkdirAll(filepath.Join(targetDir, "old"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"std.scad":     "old std",
		"stale.scad":   "stale",
		"old/stale.md": "stale nested",
	} {
		if err := os.WriteFile(filepath.Join(targetDir, filepath.FromSlash(name)), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Hold the directory open without sharing delete access, which is how
	// OpenSCAD/Explorer keep it locked on Windows.
	pathPointer, err := syscall.UTF16PtrFromString(targetDir)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(
		pathPointer,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		t.Fatalf("lock BOSL2 directory: %v", err)
	}
	defer syscall.CloseHandle(handle)

	archive := bosl2TestArchive(t, map[string]string{
		"BOSL2-main/std.scad":  "new std",
		"BOSL2-main/README.md": "BOSL2",
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	if err := installOrUpgradeBOSL2(context.Background(), libraryDir, server.URL, server.Client()); err != nil {
		t.Fatalf("install with locked BOSL2 directory: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(targetDir, "std.scad"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new std" {
		t.Fatalf("installed std.scad = %q, want %q", got, "new std")
	}
	if _, err := os.Stat(filepath.Join(targetDir, "README.md")); err != nil {
		t.Fatalf("new README.md was not installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "stale.scad")); !os.IsNotExist(err) {
		t.Fatalf("stale.scad should have been pruned, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "old")); !os.IsNotExist(err) {
		t.Fatalf("stale old/ directory should have been pruned, stat err = %v", err)
	}
}
