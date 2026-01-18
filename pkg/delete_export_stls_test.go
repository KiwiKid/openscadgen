package pkg

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestFindExportSTLFiles(t *testing.T) {
	root := t.TempDir()

	// should match
	touch(t, filepath.Join(root, "a", "export", "one.stl"))
	touch(t, filepath.Join(root, "a", "export", "v0.1", "two.STL"))
	touch(t, filepath.Join(root, "export", "three.stl"))

	// should NOT match (not inside export/)
	touch(t, filepath.Join(root, "a", "not-export", "nope.stl"))
	touch(t, filepath.Join(root, "a", "exportX", "nope2.stl"))

	got, err := FindExportSTLFiles(root)
	if err != nil {
		t.Fatalf("FindExportSTLFiles err: %v", err)
	}

	want := []string{
		filepath.Join(root, "a", "export", "one.stl"),
		filepath.Join(root, "a", "export", "v0.1", "two.STL"),
		filepath.Join(root, "export", "three.stl"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch\nwant=%v\ngot =%v", want, got)
	}
}

func TestDeleteFiles(t *testing.T) {
	root := t.TempDir()
	p1 := filepath.Join(root, "export", "a.stl")
	p2 := filepath.Join(root, "export", "b.stl")
	touch(t, p1)
	touch(t, p2)

	res := DeleteFiles([]string{p1, p2})
	if len(res.Failed) != 0 {
		t.Fatalf("expected no failures, got: %v", res.Failed)
	}
	if len(res.Deleted) != 2 {
		t.Fatalf("expected 2 deleted, got %d", len(res.Deleted))
	}
	if _, err := os.Stat(p1); !os.IsNotExist(err) {
		t.Fatalf("expected p1 deleted, stat err: %v", err)
	}
	if _, err := os.Stat(p2); !os.IsNotExist(err) {
		t.Fatalf("expected p2 deleted, stat err: %v", err)
	}
}

