package pkg

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestResolveInitParentDir(t *testing.T) {
	t.Parallel()
	got := ResolveInitParentDir(models.CmdFlags{
		InitProjectParentDir: "/explicit",
		ServerFolder:         "/server",
	})
	if got != "/explicit" {
		t.Fatalf("explicit init-dir: got %q", got)
	}
	got = ResolveInitParentDir(models.CmdFlags{ServerFolder: "/server"})
	if got != "/server" {
		t.Fatalf("server folder default: got %q", got)
	}
	got = ResolveInitParentDir(models.CmdFlags{})
	if got != "." {
		t.Fatalf("fallback .: got %q", got)
	}
}

func TestInitConfigInParent_smoke(t *testing.T) {
	t.Parallel()
	if err := InitLogger("memory"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := InitConfigInParentWithTemplate(dir, "my-widget", "basic"); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "my_widget", "config.toml")
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	config := string(data)
	if strings.Contains(config, "{{projectName}}") {
		t.Fatalf("basic template left project-name token unreplaced: %s", config)
	}
	if !strings.Contains(config, `name = "my_widget"`) || !strings.Contains(config, `path = "./my_widget.scad"`) {
		t.Fatalf("basic template did not render project name: %s", config)
	}
	if !strings.Contains(config, "[[openscadgen.images]]\nname = \"nice\"") {
		t.Fatalf("basic template missing nice image configuration: %s", config)
	}
	scad := filepath.Join(dir, "my_widget", "my_widget.scad")
	scadData, err := os.ReadFile(scad)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(scadData), "include <BOSL2/std.scad>;\n\n$fn = 100;") {
		t.Fatalf("basic template should start with the BOSL2 import and $fn = 100: %s", string(scadData))
	}
}

func TestInitConfigInParent_defaultTemplateWritesLibrary(t *testing.T) {
	t.Parallel()
	if err := InitLogger("memory"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := InitConfigInParentWithTemplate(dir, "my-widget", "default.toml"); err != nil {
		t.Fatal(err)
	}
	lib := filepath.Join(dir, "openscad-lib", "openscadgen-core.scad")
	if _, err := os.Stat(lib); err != nil {
		t.Fatal(err)
	}
	scad := filepath.Join(dir, "my_widget", "my_widget.scad")
	data, err := os.ReadFile(scad)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "include <../openscad-lib/openscadgen-core.scad>;") {
		t.Fatalf("default template missing shared library include: %s", string(data))
	}
	for _, option := range []string{"$fa", "$fs", "$fn"} {
		if count := strings.Count(string(data), option+" ="); count != 1 {
			t.Fatalf("default template should define %s once, found %d definitions: %s", option, count, string(data))
		}
	}
	if count := strings.Count(string(data), "include <BOSL2/std.scad>"); count != 1 {
		t.Fatalf("default template should include BOSL2 once, found %d includes: %s", count, string(data))
	}
	if !strings.HasPrefix(string(data), "include <BOSL2/std.scad>;\n\n$fn = 100;") {
		t.Fatalf("default template should start with the BOSL2 import and $fn = 100: %s", string(data))
	}
	if !strings.Contains(string(data), "$fn = 100;") {
		t.Fatalf("default template should define $fn = 100: %s", string(data))
	}
	config := filepath.Join(dir, "my_widget", "config.toml")
	configData, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(configData), "[[openscadgen.images]]\nname = \"nice\"") {
		t.Fatalf("default template missing nice image configuration: %s", string(configData))
	}
}

func TestInitConfigInParent_customTemplateFile(t *testing.T) {
	t.Parallel()
	if err := InitLogger("memory"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file path")
	}
	templateDir := filepath.Join(filepath.Dir(filepath.Dir(file)), "openscad-init-templates")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	templatePath := filepath.Join(templateDir, "custom.toml")
	templateBody := `[openscadgen]
name = "<projectNameUnderLined>"
description = ""

version = "v0.1"

export_name_format = "{designFileName}"

global_params = { }

[[openscadgen.input_paths]]
path = "<projectPath>/<projectName>.scad"
params = { }

[[openscadgen.images]]
name = "nice"
`
	if err := os.WriteFile(templatePath, []byte(templateBody), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(templatePath) })

	if err := InitConfigInParentWithTemplate(dir, "my-widget", "custom.toml"); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "my_widget", "config.toml")
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `name = "my_widget"`) {
		t.Fatalf("custom template did not render project name: %s", string(data))
	}
	if !strings.Contains(string(data), `path = "`) {
		t.Fatalf("custom template did not render project path: %s", string(data))
	}
}

func TestInitConfigInParent_rejectPathEscape(t *testing.T) {
	t.Parallel()
	if err := InitLogger("memory"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	err := InitConfigInParentWithTemplate(dir, "..", "basic")
	if err == nil {
		t.Fatal("expected error for .. base name")
	}
}

func TestInitConfigInParent_nestedRelativePath(t *testing.T) {
	t.Parallel()
	if err := InitLogger("memory"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := InitConfigInParentWithTemplate(dir, "nested/my-widget", "basic"); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, "nested", "my_widget", "config.toml")
	if _, err := os.Stat(cfg); err != nil {
		t.Fatal(err)
	}
	scad := filepath.Join(dir, "nested", "my_widget", "my_widget.scad")
	if _, err := os.Stat(scad); err != nil {
		t.Fatal(err)
	}
}

func TestInitConfig_delegatesToInParent(t *testing.T) {
	t.Parallel()
	if err := InitLogger("memory"); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := InitConfig(filepath.Join(sub, "leaf-name"), false); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(sub, "leaf_name", "config.toml")
	if _, err := os.Stat(cfg); err != nil {
		t.Fatal(err)
	}
}
