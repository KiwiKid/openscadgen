package tests

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileWatcherE2E(t *testing.T) {
	t.Skip("Skipping file watcher test - needs investigation")
	tempDir, err := ioutil.TempDir("", "openscadgen-e2e-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	scadPath := filepath.Join(tempDir, "test.scad")
	configPath := filepath.Join(tempDir, "config.toml")
	outputDir := filepath.Join(tempDir, "export", "v0.1")
	os.MkdirAll(outputDir, 0755)

	scadContent := `cube([10,10,10]);`
	if err := ioutil.WriteFile(scadPath, []byte(scadContent), 0644); err != nil {
		t.Fatalf("failed to write scad: %v", err)
	}

	configContent := `
[openscadgen]
name = "test"
version = "v0.1"
export_name_format = "{instanceName}"
global_params = { }

[[openscadgen.input_paths]]
path = "./test.scad"

[[openscadgen.instances]]
name = "test"
params = { }
`
	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cmd := exec.Command("../openscadgen", "-s", "-sf", tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start openscadgen: %v", err)
	}
	defer cmd.Process.Kill()

	// Wait for initial processing (give it up to 10s)
	found := false
	for i := 0; i < 20; i++ {
		if fileExists(filepath.Join(outputDir, "test.stl")) {
			found = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !found {
		t.Fatalf("STL not generated after initial run")
	}

	// Modify config (change instance name)
	configContent2 := strings.Replace(configContent, "name = \"test\"", "name = \"test2\"", 1)
	if err := ioutil.WriteFile(configPath, []byte(configContent2), 0644); err != nil {
		t.Fatalf("failed to update config: %v", err)
	}

	// Wait for watcher to regenerate (give it up to 10s)
	found2 := false
	for i := 0; i < 20; i++ {
		if fileExists(filepath.Join(outputDir, "test2.stl")) {
			found2 = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !found2 {
		t.Fatalf("STL not generated after config change")
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
