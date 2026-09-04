package pkg

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallOrUpgradeBOSL2ReplacesExistingLibraryAfterValidation(t *testing.T) {
	libraryDir := t.TempDir()
	targetDir := filepath.Join(libraryDir, "BOSL2")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "std.scad"), []byte("broken"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive := bosl2TestArchive(t, map[string]string{
		"BOSL2-main/std.scad":  "module cuboid() {}",
		"BOSL2-main/README.md": "BOSL2",
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	if err := installOrUpgradeBOSL2(context.Background(), libraryDir, server.URL, server.Client()); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(targetDir, "std.scad"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "module cuboid() {}" {
		t.Fatalf("installed std.scad = %q", got)
	}
}

func TestInstallOrUpgradeBOSL2DoesNotReplaceLibraryWithoutStdScad(t *testing.T) {
	libraryDir := t.TempDir()
	targetDir := filepath.Join(libraryDir, "BOSL2")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "std.scad"), []byte("known-good"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive := bosl2TestArchive(t, map[string]string{"BOSL2-main/README.md": "missing std"})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	if err := installOrUpgradeBOSL2(context.Background(), libraryDir, server.URL, server.Client()); err == nil {
		t.Fatal("expected archive validation error")
	}
	got, err := os.ReadFile(filepath.Join(targetDir, "std.scad"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "known-good" {
		t.Fatalf("existing std.scad changed to %q", got)
	}
}

func TestManagedOpenSCADEnvironmentPrioritizesManagedLibrary(t *testing.T) {
	managedDir, err := managedOpenSCADLibraryDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range managedOpenSCADEnvironment() {
		if strings.HasPrefix(value, "OPENSCADPATH=") {
			if !strings.HasPrefix(strings.TrimPrefix(value, "OPENSCADPATH="), managedDir) {
				t.Fatalf("OPENSCADPATH does not prioritize managed library: %s", value)
			}
			return
		}
	}
	t.Fatal("OPENSCADPATH was not configured")
}

func bosl2TestArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	archivePath := filepath.Join(t.TempDir(), "bosl2.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, body := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	archive, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	return archive
}
