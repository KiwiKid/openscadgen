package pkg

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

const sourceArchiveFilename = "source.zip"

// sourceArchivePath is deliberately kept in export/<version>, rather than a
// generated design directory. It is one small, replaceable source snapshot per
// version and ScanFolderForConfigFiles already ignores export directories.
func sourceArchivePath(config *models.Config) string {
	return filepath.Join(filepath.Dir(config.ConfigFile), "export", config.Design.Version, sourceArchiveFilename)
}

// writeSourceArchive records the exact config file and each configured OpenSCAD
// source in a compact archive. It intentionally includes direct inputs only:
// libraries and imported assets are external dependencies, not project source.
func writeSourceArchive(config *models.Config) (string, error) {
	if config == nil {
		return "", fmt.Errorf("create source archive: config is nil")
	}

	archivePath := sourceArchivePath(config)
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return "", fmt.Errorf("create source archive directory: %w", err)
	}

	temporary, err := os.CreateTemp(filepath.Dir(archivePath), ".source-*.zip")
	if err != nil {
		return "", fmt.Errorf("create source archive: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	writer := zip.NewWriter(temporary)
	if err := addSourceArchiveFile(writer, "config.toml", config.ConfigFile); err != nil {
		_ = writer.Close()
		_ = temporary.Close()
		return "", err
	}

	usedNames := map[string]bool{"config.toml": true}
	for index, input := range config.GetInputPaths() {
		inputPath := input.Path
		if !filepath.IsAbs(inputPath) {
			inputPath = filepath.Join(filepath.Dir(config.ConfigFile), inputPath)
		}
		entryName := sourceArchiveEntryName(config.ConfigFile, inputPath, index)
		for usedNames[entryName] {
			entryName = fmt.Sprintf("sources/%d-%s", index+1, filepath.Base(inputPath))
		}
		usedNames[entryName] = true
		if err := addSourceArchiveFile(writer, entryName, inputPath); err != nil {
			_ = writer.Close()
			_ = temporary.Close()
			return "", err
		}
	}

	if err := writer.Close(); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("finish source archive: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close source archive: %w", err)
	}
	// Windows does not allow Rename to replace an existing file. Removing only
	// after the complete temporary archive has been closed keeps one archive per
	// version without retaining stale snapshots.
	if err := os.Remove(archivePath); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("replace source archive: %w", err)
	}
	if err := os.Rename(temporaryPath, archivePath); err != nil {
		return "", fmt.Errorf("publish source archive: %w", err)
	}
	return archivePath, nil
}

func sourceArchiveEntryName(configPath, inputPath string, index int) string {
	relativePath, err := filepath.Rel(filepath.Dir(configPath), inputPath)
	if err == nil && relativePath != "." && relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(relativePath)
	}
	return fmt.Sprintf("sources/%d-%s", index+1, filepath.Base(inputPath))
}

func addSourceArchiveFile(writer *zip.Writer, entryName, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("read source archive file %s: %w", filePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("read source archive file %s: is a directory", filePath)
	}

	source, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open source archive file %s: %w", filePath, err)
	}
	defer source.Close()

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("create source archive entry %s: %w", entryName, err)
	}
	header.Name = filepath.ToSlash(entryName)
	header.Method = zip.Deflate
	destination, err := writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create source archive entry %s: %w", entryName, err)
	}
	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("write source archive entry %s: %w", entryName, err)
	}
	return nil
}
