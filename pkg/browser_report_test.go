package pkg

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestBuildBrowserBundleUsesAppliedParamsAndNormalizesPaths(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "box.scad")
	if err := os.WriteFile(sourcePath, []byte("cube([10, 20, 5]);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := &models.Config{
		ConfigFile:    filepath.Join(dir, "config.toml"),
		RawConfigFile: "[openscadgen]\nname = \"browser\"\n",
		Design: models.DesignConfig{
			Name:      "browser",
			InputPath: "./box.scad",
			Version:   "v1.0",
		},
	}
	instance := models.InstanceConfig{
		UniqueID:        "box-10",
		AutoName:        "box-10",
		Name:            "box",
		InputPath:       models.InputPath{Path: "./box.scad"},
		RunOutputPathV3: filepath.Join("export", "v1.0", "box-10.stl"),
		Params:          map[string]interface{}{"width": 10, "ignored": "x", "version": "v1.0"},
		STLResults:      []models.GenerateSTLResult{{AppliedParams: map[string]interface{}{"width": 10, "version": "v1.0"}}},
		OutputPathV2:    "export/v1.0/box-10.stl",
		IsComplete:      true,
	}

	bundle, sources, err := buildBrowserBundle(config, []models.InstanceConfig{instance})
	if err != nil {
		t.Fatalf("buildBrowserBundle: %v", err)
	}
	if len(sources) != 1 || sources[0].Path != "box.scad" {
		t.Fatalf("sources = %+v, want box.scad", sources)
	}
	if !strings.Contains(sources[0].Text, "cube") || sources[0].SHA256 == "" {
		t.Fatalf("source text/hash not captured: %+v", sources[0])
	}
	if len(bundle.Instances) != 1 {
		t.Fatalf("instances = %d, want 1", len(bundle.Instances))
	}
	got := bundle.Instances[0]
	if got.OutputPath != "box-10.stl" || got.SourcePath != "box.scad" {
		t.Fatalf("instance paths = %+v", got)
	}
	if _, ok := got.Definitions["version"]; ok {
		t.Fatal("version must not be sent to OpenSCAD in browser mode")
	}
	if got.Definitions["width"] != 10 {
		t.Fatalf("width = %#v, want 10", got.Definitions["width"])
	}
	if bundle.Config.TOML != config.RawConfigFile {
		t.Fatal("raw config TOML was not preserved")
	}
}

func TestBrowserSourcePathFallsBackForOutsideConfigDirectory(t *testing.T) {
	base := filepath.Join("some", "project")
	if got := browserSourcePath(base, filepath.Join("elsewhere", "part.scad")); got != "part.scad" {
		t.Fatalf("browserSourcePath = %q, want part.scad", got)
	}
	if got := browserSourcePath(base, filepath.Join(base, "parts", "part.scad")); got != "parts/part.scad" {
		t.Fatalf("browserSourcePath = %q, want parts/part.scad", got)
	}
}

func TestPrepareBrowserReportCopiesAssetsAndWritesBundle(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"og-browser.js", "og-browser-worker.js", "openscad.js", "openscad.wasm"} {
		if err := os.WriteFile(filepath.Join(assets, name), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv(browserAssetsEnv, assets)
	t.Setenv("OPENSCADPATH", "")

	sourcePath := filepath.Join(dir, "box.scad")
	if err := os.WriteFile(sourcePath, []byte("cube(1);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := &models.Config{
		ConfigFile:    filepath.Join(dir, "config.toml"),
		RawConfigFile: "[openscadgen]\nname = \"browser\"\n",
		Design:        models.DesignConfig{Name: "browser", InputPath: "./box.scad", Version: "v1.0"},
	}
	instance := models.InstanceConfig{
		UniqueID:        "box",
		AutoName:        "box",
		InputPath:       models.InputPath{Path: "./box.scad"},
		RunOutputPathV3: "export/v1.0/box.stl",
		OutputPathV2:    "export/v1.0/box.stl",
		Params:          map[string]interface{}{"size": 1},
	}

	bundleJSON, err := PrepareBrowserReport(config, []models.InstanceConfig{instance}, dir)
	if err != nil {
		t.Fatalf("PrepareBrowserReport: %v", err)
	}
	for _, name := range []string{"og-browser.js", "og-browser-worker.js", "openscad.js", "openscad.wasm", "bundle.json"} {
		if _, err := os.Stat(filepath.Join(dir, browserOutputDir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	var bundle BrowserBundle
	if err := json.Unmarshal([]byte(bundleJSON), &bundle); err != nil {
		t.Fatalf("bundle JSON: %v", err)
	}
	if bundle.Schema != browserBundleSchema || len(bundle.Instances) != 1 {
		t.Fatalf("bundle = %+v", bundle)
	}
}

func TestGenerateOutputReportEmbedsBrowserBundle(t *testing.T) {
	dir := t.TempDir()
	assets := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assets, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"og-browser.js", "og-browser-worker.js", "openscad.js", "openscad.wasm"} {
		if err := os.WriteFile(filepath.Join(assets, name), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv(browserAssetsEnv, assets)
	t.Setenv("OPENSCADPATH", "")

	if err := os.WriteFile(filepath.Join(dir, "box.scad"), []byte("cube(1);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	config := &models.Config{
		ConfigFile:    filepath.Join(dir, "config.toml"),
		RawConfigFile: "[openscadgen]\nname = \"browser\"\n",
		BrowserReport: true,
		Design:        models.DesignConfig{Name: "browser", InputPath: "./box.scad", Version: "v1.0"},
	}
	instance := models.InstanceConfig{
		UniqueID:        "box",
		AutoName:        "box",
		InputPath:       models.InputPath{Path: "./box.scad"},
		RunOutputPathV3: "box.stl",
		OutputPathV2:    "box.stl",
		Params:          map[string]interface{}{"size": 1},
	}

	_, outputFile, err := GenerateOutputReport(config, []models.InstanceConfig{instance}, nil, nil, dir, true, "", 0)
	if err != nil {
		t.Fatalf("GenerateOutputReport: %v", err)
	}
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, want := range []string{"og-browser-bundle", "og-browser.js", "og-browser.css", "&#34;schema&#34;:1"} {
		if !strings.Contains(html, want) {
			t.Fatalf("report is missing %q", want)
		}
	}
	if !strings.Contains(html, "./browser") {
		t.Fatal("report does not point at the browser assets")
	}
}

func TestWriteBrowserLibrariesBuildsBOSL2Archive(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "lib", "BOSL2", "std.scad")); err != nil {
		t.Skip("bundled BOSL2 checkout is not present")
	}
	t.Setenv("OPENSCADPATH", filepath.Join("..", "lib"))

	dir := t.TempDir()
	libraries, err := writeBrowserLibraries(dir, []BrowserSource{{Text: "include <BOSL2/std.scad>;"}})
	if err != nil {
		t.Fatalf("writeBrowserLibraries: %v", err)
	}
	if len(libraries) != 1 || libraries[0].Missing {
		t.Fatalf("libraries = %+v", libraries)
	}
	archivePath := filepath.Join(dir, filepath.FromSlash(libraries[0].Archive))
	compressed, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("open archive: %v", err)
	}
	defer compressed.Close()
	reader, err := gzip.NewReader(compressed)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	var archive browserLibraryArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		t.Fatalf("archive JSON: %v", err)
	}
	if archive.Name != "BOSL2" || !strings.Contains(archive.Files["std.scad"], "BOSL") {
		t.Fatalf("archive does not contain BOSL2/std.scad")
	}
}
