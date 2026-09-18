package pkg

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestWriteSourceArchiveSnapshotsConfigAndConfiguredSources(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	mainPath := filepath.Join(dir, "main.scad")
	nestedPath := filepath.Join(dir, "parts", "adapter.scad")
	if err := os.MkdirAll(filepath.Dir(nestedPath), 0o755); err != nil {
		t.Fatal(err)
	}
	for path, contents := range map[string]string{
		configPath: "[openscadgen]\nname = \"archive-test\"\n",
		mainPath:   "cube(10);\n",
		nestedPath: "sphere(5);\n",
	} {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	archivePath, err := writeSourceArchive(&models.Config{
		ConfigFile: configPath,
		Design: models.DesignConfig{
			Version:    "v1.2",
			InputPaths: []models.InputPath{{Path: "main.scad"}, {Path: "parts/adapter.scad"}},
		},
	})
	if err != nil {
		t.Fatalf("writeSourceArchive: %v", err)
	}
	wantPath := filepath.Join(dir, "export", "v1.2", sourceArchiveFilename)
	if archivePath != wantPath {
		t.Fatalf("archive path = %q, want %q", archivePath, wantPath)
	}
	if err := os.WriteFile(mainPath, []byte("cube(20);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	archivePath, err = writeSourceArchive(&models.Config{
		ConfigFile: configPath,
		Design: models.DesignConfig{
			Version:    "v1.2",
			InputPaths: []models.InputPath{{Path: "main.scad"}, {Path: "parts/adapter.scad"}},
		},
	})
	if err != nil {
		t.Fatalf("replace source archive: %v", err)
	}

	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	got := map[string]string{}
	for _, file := range archive.File {
		if file.Method != zip.Deflate {
			t.Errorf("archive entry %q method = %d, want deflate", file.Name, file.Method)
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		got[file.Name] = string(data)
	}
	want := map[string]string{
		"config.toml":        "[openscadgen]\nname = \"archive-test\"\n",
		"main.scad":          "cube(20);\n",
		"parts/adapter.scad": "sphere(5);\n",
	}
	if len(got) != len(want) {
		t.Fatalf("archive entries = %#v, want %#v", got, want)
	}
	for name, contents := range want {
		if got[name] != contents {
			t.Errorf("archive entry %q = %q, want %q", name, got[name], contents)
		}
	}
}
