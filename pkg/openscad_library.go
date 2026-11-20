package pkg

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

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

// OpenSCADLibraryAccess reports whether the OpenSCAD executable can resolve a
// supported library's entry point.
type OpenSCADLibraryAccess struct {
	Checked    bool
	Accessible bool
	Error      string
}

type openSCADLibraryAccessCacheEntry struct {
	access    OpenSCADLibraryAccess
	checkedAt time.Time
}

var (
	openSCADLibraryAccessCacheMu sync.Mutex
	openSCADLibraryAccessCache   = map[string]openSCADLibraryAccessCacheEntry{}
)

// openSCADLibraryAccessCacheTTL keeps repeated tools-page loads from launching
// OpenSCAD on every request while still noticing a repaired installation.
const openSCADLibraryAccessCacheTTL = 30 * time.Second

// InvalidateOpenSCADLibraryAccessCache drops cached probe results so the next
// check reflects a library that was just installed or repaired.
func InvalidateOpenSCADLibraryAccessCache() {
	openSCADLibraryAccessCacheMu.Lock()
	defer openSCADLibraryAccessCacheMu.Unlock()
	openSCADLibraryAccessCache = map[string]openSCADLibraryAccessCacheEntry{}
}

// CheckOpenSCADLibraryAccess probes each installed supported library with the
// given OpenSCAD executable and returns a copy with the access fields filled
// in. An empty executable path leaves every library's access state unchecked.
func CheckOpenSCADLibraryAccess(libraries []templates.LibraryLink, openscadPath string) []templates.LibraryLink {
	result := make([]templates.LibraryLink, len(libraries))
	copy(result, libraries)
	openscadPath = strings.TrimSpace(openscadPath)
	if openscadPath == "" {
		return result
	}
	for i := range result {
		if !result[i].Installed {
			continue
		}
		access := cachedOpenSCADLibraryAccess(result[i].ID, result[i].Path, openscadPath)
		result[i].AccessChecked = access.Checked
		result[i].Accessible = access.Accessible
		result[i].AccessError = access.Error
	}
	return result
}

func cachedOpenSCADLibraryAccess(id, libraryPath, openscadPath string) OpenSCADLibraryAccess {
	key := strings.ToLower(id) + "\x00" + libraryPath + "\x00" + openscadPath
	now := time.Now()

	openSCADLibraryAccessCacheMu.Lock()
	if entry, ok := openSCADLibraryAccessCache[key]; ok && now.Sub(entry.checkedAt) < openSCADLibraryAccessCacheTTL {
		openSCADLibraryAccessCacheMu.Unlock()
		return entry.access
	}
	openSCADLibraryAccessCacheMu.Unlock()

	access := probeOpenSCADLibraryAccess(id, openscadPath)

	openSCADLibraryAccessCacheMu.Lock()
	openSCADLibraryAccessCache[key] = openSCADLibraryAccessCacheEntry{access: access, checkedAt: now}
	openSCADLibraryAccessCacheMu.Unlock()
	return access
}

func probeOpenSCADLibraryAccess(id, openscadPath string) OpenSCADLibraryAccess {
	for _, definition := range openSCADLibraryCatalog {
		if !strings.EqualFold(definition.ID, id) {
			continue
		}
		includePath := definition.Folder + "/" + definition.MarkerFile
		if err := probeOpenSCADInclude(openscadPath, includePath); err != nil {
			return OpenSCADLibraryAccess{Checked: true, Error: err.Error()}
		}
		return OpenSCADLibraryAccess{Checked: true, Accessible: true}
	}
	return OpenSCADLibraryAccess{
		Checked: true,
		Error:   fmt.Sprintf("unsupported OpenSCAD library %q", id),
	}
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
	if userDir, err := userOpenSCADLibraryDir(); err == nil {
		candidates = append(candidates, userDir)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "Documents", "OpenSCAD", "libraries"))
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
