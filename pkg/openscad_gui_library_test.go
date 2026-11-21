package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestBOSL2(t *testing.T, dir, version, std string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "std.scad"), []byte(std), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "version.scad"), []byte("BOSL_VERSION = ["+version+"];\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureOpenSCADGUILibraryPathCopiesManagedLibraries(t *testing.T) {
	base := t.TempDir()
	managedDir := filepath.Join(base, "managed")
	userDir := filepath.Join(base, "openSCAD", "libraries")
	writeTestBOSL2(t, filepath.Join(managedDir, "BOSL2"), "2,0,1", "// managed")
	if err := os.MkdirAll(filepath.Join(managedDir, ".openscadgen-bosl2-staging"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENSCADGEN_LIBRARY_DIR", managedDir)
	t.Setenv("OPENSCADGEN_OPENSCAD_USER_LIBRARY_DIR", userDir)

	if err := EnsureOpenSCADGUILibraryPath(); err != nil {
		t.Fatal(err)
	}

	copyDir := filepath.Join(userDir, "BOSL2")
	got, err := os.ReadFile(filepath.Join(copyDir, "std.scad"))
	if err != nil {
		t.Fatalf("read copied library: %v", err)
	}
	if string(got) != "// managed" {
		t.Fatalf("copied std.scad = %q", got)
	}
	marker, err := os.ReadFile(filepath.Join(copyDir, openSCADManagedCopyMarker))
	if err != nil {
		t.Fatalf("read managed copy marker: %v", err)
	}
	if string(marker) != "2.0.1\n" {
		t.Fatalf("marker = %q, want the library version", marker)
	}

	// A second call with the same version must not refresh the copy.
	if err := EnsureOpenSCADGUILibraryPath(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(copyDir, "std.scad")); err != nil || string(got) != "// managed" {
		t.Fatalf("copied std.scad after second call = %q, err = %v", got, err)
	}
}

func TestEnsureOpenSCADGUILibraryPathRefreshesOutdatedCopy(t *testing.T) {
	base := t.TempDir()
	managedDir := filepath.Join(base, "managed")
	userDir := filepath.Join(base, "openSCAD", "libraries")
	sourceDir := filepath.Join(managedDir, "BOSL2")
	writeTestBOSL2(t, sourceDir, "2,0,1", "// v1")
	if err := os.WriteFile(filepath.Join(sourceDir, "old-file.scad"), []byte("// stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENSCADGEN_LIBRARY_DIR", managedDir)
	t.Setenv("OPENSCADGEN_OPENSCAD_USER_LIBRARY_DIR", userDir)
	if err := EnsureOpenSCADGUILibraryPath(); err != nil {
		t.Fatal(err)
	}

	writeTestBOSL2(t, sourceDir, "2,0,2", "// v2")
	if err := os.Remove(filepath.Join(sourceDir, "old-file.scad")); err != nil {
		t.Fatal(err)
	}
	if err := EnsureOpenSCADGUILibraryPath(); err != nil {
		t.Fatal(err)
	}

	copyDir := filepath.Join(userDir, "BOSL2")
	got, err := os.ReadFile(filepath.Join(copyDir, "std.scad"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "// v2" {
		t.Fatalf("copied std.scad = %q, want refreshed contents", got)
	}
	if _, err := os.Stat(filepath.Join(copyDir, "old-file.scad")); !os.IsNotExist(err) {
		t.Fatalf("stale file should have been pruned, stat err = %v", err)
	}
	marker, err := os.ReadFile(filepath.Join(copyDir, openSCADManagedCopyMarker))
	if err != nil {
		t.Fatal(err)
	}
	if string(marker) != "2.0.2\n" {
		t.Fatalf("marker = %q, want 2.0.2", marker)
	}
}

func TestEnsureOpenSCADGUILibraryPathKeepsExistingUserLibrary(t *testing.T) {
	base := t.TempDir()
	managedDir := filepath.Join(base, "managed")
	userDir := filepath.Join(base, "openSCAD", "libraries")
	writeTestBOSL2(t, filepath.Join(managedDir, "BOSL2"), "2,0,1", "// managed")
	userCopy := filepath.Join(userDir, "BOSL2")
	if err := os.MkdirAll(userCopy, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(userCopy, "user-copy.scad")
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
	if _, err := os.Stat(filepath.Join(userCopy, openSCADManagedCopyMarker)); !os.IsNotExist(err) {
		t.Fatalf("user library should not be marked as managed, stat err = %v", err)
	}
}
