package pkg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFolderForConfigFiles_PrimaryImagePathPrefersNice(t *testing.T) {
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "proj1")
	if err := os.MkdirAll(filepath.Join(projectDir, "export", "v0.1"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Needs to include "[openscadgen]" within the first 1024 bytes to be detected.
	configPath := filepath.Join(projectDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("[openscadgen]\nname='x'\nexport_name_format='x'\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Two images; "nice" should win regardless of sort order.
	if err := os.WriteFile(filepath.Join(projectDir, "export", "v0.1", "aaa.png"), []byte("x"), 0644); err != nil {
		t.Fatalf("write image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "export", "v0.1", "my_nice_view.png"), []byte("x"), 0644); err != nil {
		t.Fatalf("write image: %v", err)
	}

	configFiles, err := ScanFolderForConfigFiles(tmpDir)
	if err != nil {
		t.Fatalf("ScanFolderForConfigFiles: %v", err)
	}

	var got string
	for _, cf := range configFiles {
		if cf.Path == filepath.Join("proj1", "config.toml") {
			got = cf.PrimaryImagePath
			break
		}
	}
	if got == "" {
		t.Fatalf("expected PrimaryImagePath to be set for proj1/config.toml")
	}

	want := filepath.ToSlash(filepath.Join("proj1", "export", "v0.1", "my_nice_view.png"))
	if got != want {
		t.Fatalf("PrimaryImagePath mismatch: got %q want %q", got, want)
	}
}

func TestScanFolderForConfigFiles_PrimaryImagePathFallsBackToFirstImage(t *testing.T) {
	tmpDir := t.TempDir()

	projectDir := filepath.Join(tmpDir, "proj2")
	if err := os.MkdirAll(filepath.Join(projectDir, "export", "v1.0"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	configPath := filepath.Join(projectDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("[openscadgen]\nname='x'\nexport_name_format='x'\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// No "nice" in names; deterministic fallback is the first path in sorted order.
	if err := os.WriteFile(filepath.Join(projectDir, "export", "v1.0", "z.png"), []byte("x"), 0644); err != nil {
		t.Fatalf("write image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "export", "v1.0", "a.png"), []byte("x"), 0644); err != nil {
		t.Fatalf("write image: %v", err)
	}

	configFiles, err := ScanFolderForConfigFiles(tmpDir)
	if err != nil {
		t.Fatalf("ScanFolderForConfigFiles: %v", err)
	}

	var got string
	for _, cf := range configFiles {
		if cf.Path == filepath.Join("proj2", "config.toml") {
			got = cf.PrimaryImagePath
			break
		}
	}
	if got == "" {
		t.Fatalf("expected PrimaryImagePath to be set for proj2/config.toml")
	}

	want := filepath.ToSlash(filepath.Join("proj2", "export", "v1.0", "a.png"))
	if got != want {
		t.Fatalf("PrimaryImagePath mismatch: got %q want %q", got, want)
	}
}

