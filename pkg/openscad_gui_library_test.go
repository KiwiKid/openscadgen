package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureOpenSCADGUILibraryPathLinksManagedLibraries(t *testing.T) {
	base := t.TempDir()
	managedDir := filepath.Join(base, "managed")
	userDir := filepath.Join(base, "openSCAD", "libraries")
	if err := os.MkdirAll(filepath.Join(managedDir, "BOSL2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(managedDir, "BOSL2", "std.scad"), []byte("// BOSL2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(managedDir, ".openscadgen-bosl2-staging"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENSCADGEN_LIBRARY_DIR", managedDir)
	t.Setenv("OPENSCADGEN_OPENSCAD_USER_LIBRARY_DIR", userDir)

	if err := EnsureOpenSCADGUILibraryPath(); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(userDir, "BOSL2")
	t.Cleanup(func() { _ = os.Remove(link) })

	got, err := os.ReadFile(filepath.Join(link, "std.scad"))
	if err != nil {
		t.Fatalf("read library through link: %v", err)
	}
	if string(got) != "// BOSL2" {
		t.Fatalf("std.scad through link = %q", got)
	}
	if _, err := os.Lstat(filepath.Join(userDir, ".openscadgen-bosl2-staging")); !os.IsNotExist(err) {
		t.Fatalf("staging directory should not be linked, stat err = %v", err)
	}

	// A second call must be a no-op rather than replacing the link.
	if err := EnsureOpenSCADGUILibraryPath(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(link, "std.scad")); err != nil || string(got) != "// BOSL2" {
		t.Fatalf("library through link after second call = %q, err = %v", got, err)
	}
}

func TestEnsureOpenSCADGUILibraryPathKeepsExistingUserLibrary(t *testing.T) {
	base := t.TempDir()
	managedDir := filepath.Join(base, "managed")
	userDir := filepath.Join(base, "openSCAD", "libraries")
	if err := os.MkdirAll(filepath.Join(managedDir, "BOSL2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(userDir, "BOSL2"), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(userDir, "BOSL2", "user-copy.scad")
	if err := os.WriteFile(existing, []byte("// user copy"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENSCADGEN_LIBRARY_DIR", managedDir)
	t.Setenv("OPENSCADGEN_OPENSCAD_USER_LIBRARY_DIR", userDir)

	if err := EnsureOpenSCADGUILibraryPath(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "// user copy" {
		t.Fatalf("existing user library was modified: %q", got)
	}
}
