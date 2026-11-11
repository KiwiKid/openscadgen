package pkg

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/templates"
)

// openSCADLibraryDefinition describes a library OpenSCADGen can detect and
// install. Keep the catalog small and explicit; add entries as support is
// verified.
type openSCADLibraryDefinition struct {
	ID            string
	Name          string
	Description   string
	HomepageURL   string
	Folder        string
	MarkerFile    string
	VersionFile   string
	VersionFormat *regexp.Regexp
	Install       func(context.Context, string) error
}

var openSCADLibraryCatalog = []openSCADLibraryDefinition{
	{
		ID:          "bosl2",
		Name:        "BOSL2",
		Description: "Core helper library for many OpenSCAD models, including cuboid, attachments, shapes, and threads.",
		HomepageURL: "https://github.com/BelfrySCAD/BOSL2",
		Folder:      "BOSL2",
		MarkerFile:  "std.scad",
		VersionFile: "version.scad",
		VersionFormat: regexp.MustCompile(
			`BOSL_VERSION\s*=\s*\[\s*([0-9]+)\s*,\s*([0-9]+)\s*,\s*([0-9]+)\s*\]`,
		),
		Install: func(ctx context.Context, libraryDir string) error {
			return installOrUpgradeBOSL2(ctx, libraryDir, bosl2ArchiveURL, http.DefaultClient)
		},
	},
}

// OpenSCADLibraries returns the supported library catalog with the local
// installation state filled in.
func OpenSCADLibraries() []templates.LibraryLink {
	libraries := make([]templates.LibraryLink, 0, len(openSCADLibraryCatalog))
	for _, definition := range openSCADLibraryCatalog {
		installedPath, version := findInstalledOpenSCADLibrary(definition)
		libraries = append(libraries, templates.LibraryLink{
			ID:               definition.ID,
			Name:             definition.Name,
			Description:      definition.Description,
			URL:              definition.HomepageURL,
			Installed:        installedPath != "",
			Version:          version,
			Path:             installedPath,
			InstallSupported: true,
			InstallURL:       "/api/openscad/libraries/" + definition.ID + "/update",
		})
	}
	return libraries
}

// InstallOpenSCADLibrary installs or upgrades a supported library by ID.
func InstallOpenSCADLibrary(ctx context.Context, id string) error {
	for _, definition := range openSCADLibraryCatalog {
		if !strings.EqualFold(definition.ID, id) {
			continue
		}
		targetDir, err := openSCADLibraryInstallRoot(definition)
		if err != nil {
			return err
		}
		return definition.Install(ctx, targetDir)
	}
	return fmt.Errorf("unsupported OpenSCAD library %q", id)
}

func OpenSCADLibraryInstalled(id string) bool {
	for _, definition := range openSCADLibraryCatalog {
		if strings.EqualFold(definition.ID, id) {
			installedPath, _ := findInstalledOpenSCADLibrary(definition)
			return installedPath != ""
		}
	}
	return false
}

func OpenSCADLibrarySupported(id string) bool {
	for _, definition := range openSCADLibraryCatalog {
		if strings.EqualFold(definition.ID, id) {
			return true
		}
	}
	return false
}

func openSCADLibraryInstallRoot(definition openSCADLibraryDefinition) (string, error) {
	if installedPath, _ := findInstalledOpenSCADLibrary(definition); installedPath != "" {
		return filepath.Dir(installedPath), nil
	}
	return managedOpenSCADLibraryDir()
}

func findInstalledOpenSCADLibrary(definition openSCADLibraryDefinition) (string, string) {
	for _, root := range openSCADLibraryDirs() {
		libraryPath := filepath.Join(root, definition.Folder)
		marker := filepath.Join(libraryPath, definition.MarkerFile)
		if info, err := os.Stat(marker); err != nil || info.IsDir() {
			continue
		}
		return libraryPath, readOpenSCADLibraryVersion(libraryPath, definition)
	}
	return "", ""
}

func readOpenSCADLibraryVersion(libraryPath string, definition openSCADLibraryDefinition) string {
	if definition.VersionFile == "" || definition.VersionFormat == nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(libraryPath, definition.VersionFile))
	if err != nil {
		return ""
	}
	matches := definition.VersionFormat.FindStringSubmatch(string(data))
	if len(matches) != 4 {
		return ""
	}
	return strings.Join(matches[1:], ".")
}

// openSCADLibraryDirs returns candidate user library roots in priority order.
// The managed directory comes first; OPENSCADPATH and the conventional
// Documents location are also checked so existing installs are detected.
func openSCADLibraryDirs() []string {
	var candidates []string
	if managedDir, err := managedOpenSCADLibraryDir(); err == nil {
		candidates = append(candidates, managedDir)
	}
	for _, dir := range filepath.SplitList(os.Getenv("OPENSCADPATH")) {
		if strings.TrimSpace(dir) != "" {
			candidates = append(candidates, dir)
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, "Documents", "OpenSCAD", "libraries"),
			filepath.Join(home, ".local", "share", "OpenSCAD", "libraries"),
		)
	}

	seen := make(map[string]struct{}, len(candidates))
	dirs := make([]string, 0, len(candidates))
	for _, dir := range candidates {
		clean := filepath.Clean(dir)
		if absolute, err := filepath.Abs(clean); err == nil {
			clean = absolute
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		dirs = append(dirs, clean)
	}
	return dirs
}
