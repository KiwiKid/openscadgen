package pkg

import (
	"os"
	"path/filepath"
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
	if _, err := os.Stat(cfg); err != nil {
		t.Fatal(err)
	}
	scad := filepath.Join(dir, "my_widget", "my_widget.scad")
	if _, err := os.Stat(scad); err != nil {
		t.Fatal(err)
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
