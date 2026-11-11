package pkg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenSCADLibrariesDetectsInstalledBOSL2(t *testing.T) {
	libraryRoot := t.TempDir()
	t.Setenv("OPENSCADGEN_LIBRARY_DIR", libraryRoot)
	t.Setenv("OPENSCADPATH", "")

	bosl2Dir := filepath.Join(libraryRoot, "BOSL2")
	if err := os.MkdirAll(bosl2Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bosl2Dir, "std.scad"), []byte("// BOSL2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bosl2Dir, "version.scad"), []byte("BOSL_VERSION = [2,0,785];\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	libraries := OpenSCADLibraries()
	if len(libraries) != 1 {
		t.Fatalf("libraries = %d, want 1", len(libraries))
	}
	got := libraries[0]
	if got.ID != "bosl2" || !got.Installed || got.Version != "2.0.785" {
		t.Fatalf("unexpected BOSL2 status: %+v", got)
	}
	if !OpenSCADLibraryInstalled("bosl2") || !OpenSCADLibrarySupported("bosl2") {
		t.Fatalf("expected BOSL2 to be installed and supported")
	}
}

func TestOpenSCADLibrariesReportsMissingBOSL2(t *testing.T) {
	libraryRoot := t.TempDir()
	t.Setenv("OPENSCADGEN_LIBRARY_DIR", libraryRoot)
	t.Setenv("OPENSCADPATH", "")

	libraries := OpenSCADLibraries()
	if len(libraries) != 1 {
		t.Fatalf("libraries = %d, want 1", len(libraries))
	}
	if libraries[0].Installed || libraries[0].Version != "" {
		t.Fatalf("unexpected BOSL2 status: %+v", libraries[0])
	}
}

func TestOpenSCADLibraryInstallRootUsesExistingLibraryLocation(t *testing.T) {
	libraryRoot := t.TempDir()
	t.Setenv("OPENSCADGEN_LIBRARY_DIR", libraryRoot)
	t.Setenv("OPENSCADPATH", "")

	// Install root for an existing library is the directory containing BOSL2.
	bosl2Dir := filepath.Join(libraryRoot, "BOSL2")
	if err := os.MkdirAll(bosl2Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bosl2Dir, "std.scad"), []byte("// BOSL2"), 0o644); err != nil {
		t.Fatal(err)
	}
	root, err := openSCADLibraryInstallRoot(openSCADLibraryCatalog[0])
	if err != nil {
		t.Fatal(err)
	}
	if root != libraryRoot {
		t.Fatalf("install root = %q, want %q", root, libraryRoot)
	}
}

func TestInstallOpenSCADLibraryRejectsUnknownID(t *testing.T) {
	err := InstallOpenSCADLibrary(context.Background(), "not-a-library")
	if err == nil || !strings.Contains(err.Error(), "unsupported OpenSCAD library") {
		t.Fatalf("expected unsupported library error, got %v", err)
	}
}
