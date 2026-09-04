package pkg

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const bosl2ArchiveURL = "https://github.com/BelfrySCAD/BOSL2/archive/refs/heads/master.zip"

// InstallOrUpgradeBOSL2 installs the current BOSL2 source tree into the
// OpenSCADGen-managed user library on macOS. This avoids macOS Documents-folder
// privacy restrictions that can prevent the headless OpenSCAD process reading
// libraries which are otherwise available to the GUI.
func InstallOrUpgradeBOSL2(ctx context.Context) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("BOSL2 installation is currently supported on macOS only")
	}

	libraryDir, err := managedOpenSCADLibraryDir()
	if err != nil {
		return err
	}
	return installOrUpgradeBOSL2(ctx, libraryDir, bosl2ArchiveURL, http.DefaultClient)
}

func managedOpenSCADLibraryDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory for managed OpenSCAD library: %w", err)
	}
	return filepath.Join(home, ".local", "share", "OpenSCAD", "libraries"), nil
}

func managedOpenSCADEnvironment() []string {
	managedDir, err := managedOpenSCADLibraryDir()
	if err != nil {
		return os.Environ()
	}
	const prefix = "OPENSCADPATH="
	var existing string
	env := make([]string, 0, len(os.Environ())+1)
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, prefix) {
			existing = strings.TrimPrefix(value, prefix)
			continue
		}
		env = append(env, value)
	}
	pathValue := managedDir
	if existing != "" {
		pathValue += string(os.PathListSeparator) + existing
	}
	return append(env, prefix+pathValue)
}

func installOrUpgradeBOSL2(ctx context.Context, libraryDir, archiveURL string, client *http.Client) error {
	if client == nil {
		client = http.DefaultClient
	}
	if err := os.MkdirAll(libraryDir, 0o755); err != nil {
		return fmt.Errorf("create OpenSCAD library directory: %w", err)
	}

	stagingDir, err := os.MkdirTemp(libraryDir, ".openscadgen-bosl2-")
	if err != nil {
		return fmt.Errorf("create BOSL2 staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	archivePath := filepath.Join(stagingDir, "bosl2.zip")
	if err := downloadBOSL2Archive(ctx, client, archiveURL, archivePath); err != nil {
		return err
	}
	if err := extractBOSL2Archive(archivePath, filepath.Join(stagingDir, "source")); err != nil {
		return err
	}

	sourceDir, err := findBOSL2Source(filepath.Join(stagingDir, "source"))
	if err != nil {
		return err
	}
	targetDir := filepath.Join(libraryDir, "BOSL2")
	backupDir := filepath.Join(libraryDir, fmt.Sprintf(".BOSL2-backup-%d", time.Now().UnixNano()))
	if err := replaceBOSL2Directory(sourceDir, targetDir, backupDir); err != nil {
		return err
	}
	return nil
}

func downloadBOSL2Archive(ctx context.Context, client *http.Client, archiveURL, destination string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return fmt.Errorf("create BOSL2 download request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download BOSL2: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download BOSL2: unexpected HTTP status %s", resp.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create BOSL2 archive: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, io.LimitReader(resp.Body, 100<<20)); err != nil {
		return fmt.Errorf("save BOSL2 archive: %w", err)
	}
	return nil
}

func extractBOSL2Archive(archivePath, destination string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open BOSL2 archive: %w", err)
	}
	defer archive.Close()
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("create BOSL2 extraction directory: %w", err)
	}

	for _, entry := range archive.File {
		name := filepath.Clean(entry.Name)
		if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("BOSL2 archive contains unsafe path %q", entry.Name)
		}
		if entry.FileInfo().Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("BOSL2 archive contains unsupported symbolic link %q", entry.Name)
		}
		path := filepath.Join(destination, name)
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("create BOSL2 archive directory: %w", err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create BOSL2 archive parent: %w", err)
		}
		in, err := entry.Open()
		if err != nil {
			return fmt.Errorf("open BOSL2 archive entry: %w", err)
		}
		out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err == nil {
			_, err = io.Copy(out, in)
			closeErr := out.Close()
			if err == nil {
				err = closeErr
			}
		}
		in.Close()
		if err != nil {
			return fmt.Errorf("extract BOSL2 archive entry %q: %w", entry.Name, err)
		}
	}
	return nil
}

func findBOSL2Source(root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", fmt.Errorf("read extracted BOSL2 archive: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(root, entry.Name())
		if info, err := os.Stat(filepath.Join(candidate, "std.scad")); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("BOSL2 archive does not contain std.scad")
}

func replaceBOSL2Directory(sourceDir, targetDir, backupDir string) error {
	if err := os.Rename(sourceDir, targetDir); err == nil {
		return nil
	} else if !os.IsExist(err) {
		return fmt.Errorf("install BOSL2: %w", err)
	}

	info, err := os.Lstat(targetDir)
	if err != nil {
		return fmt.Errorf("inspect existing BOSL2 library: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("existing BOSL2 library is not a directory")
	}
	if err := os.Rename(targetDir, backupDir); err != nil {
		return fmt.Errorf("back up existing BOSL2 library: %w", err)
	}
	if err := os.Rename(sourceDir, targetDir); err != nil {
		if restoreErr := os.Rename(backupDir, targetDir); restoreErr != nil {
			return fmt.Errorf("install BOSL2: %w (also failed to restore previous library: %v)", err, restoreErr)
		}
		return fmt.Errorf("install BOSL2: %w", err)
	}
	if err := os.RemoveAll(backupDir); err != nil {
		return fmt.Errorf("remove BOSL2 backup after successful install: %w", err)
	}
	return nil
}
