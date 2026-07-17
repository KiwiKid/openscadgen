package pkg

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func pathHasSegment(p string, segment string) bool {
	p = filepath.Clean(p)
	parts := strings.Split(p, string(filepath.Separator))
	for _, part := range parts {
		if part == segment {
			return true
		}
	}
	return false
}

// FindExportSTLFiles walks rootDir and returns all files with `.stl` extension (case-insensitive)
// that are inside any directory named `export` (at any depth), including export subfolders.
func FindExportSTLFiles(rootDir string) ([]string, error) {
	if strings.TrimSpace(rootDir) == "" {
		return nil, errors.New("rootDir is empty")
	}
	rootDir = filepath.Clean(rootDir)
	info, err := os.Stat(rootDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", rootDir)
	}

	var out []string
	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".stl" {
			return nil
		}
		// Must be within a directory segment named "export"
		if !pathHasSegment(filepath.Dir(path), "export") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(out)
	return out, nil
}

type DeleteFilesResult struct {
	Deleted []string
	Failed  map[string]error
}

type VersionFilePreview struct {
	Version string
	Files   []string
	Count   int
}

func DeleteFiles(paths []string) DeleteFilesResult {
	res := DeleteFilesResult{
		Deleted: []string{},
		Failed:  map[string]error{},
	}
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		err := os.Remove(p)
		if err != nil {
			res.Failed[p] = err
			continue
		}
		res.Deleted = append(res.Deleted, p)
	}
	sort.Strings(res.Deleted)
	return res
}

func FindFilesInDir(rootDir string) ([]string, error) {
	if strings.TrimSpace(rootDir) == "" {
		return nil, errors.New("rootDir is empty")
	}
	rootDir = filepath.Clean(rootDir)
	info, err := os.Stat(rootDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", rootDir)
	}
	var out []string
	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

func PreviewVersionFiles(outputRootDir string, version string, includeOtherVersions bool) (current []string, other map[string][]string, err error) {
	if strings.TrimSpace(outputRootDir) == "" {
		return nil, nil, errors.New("outputRootDir is empty")
	}
	version = strings.TrimSpace(version)
	outputRootDir = filepath.Clean(outputRootDir)
	entries, err := os.ReadDir(outputRootDir)
	if err != nil {
		return nil, nil, err
	}
	other = map[string][]string{}
	currentDir := filepath.Join(outputRootDir, version)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(outputRootDir, entry.Name())
		files, err := FindFilesInDir(dirPath)
		if err != nil {
			return nil, nil, err
		}
		if dirPath == currentDir {
			current = files
			continue
		}
		if includeOtherVersions {
			other[entry.Name()] = files
		}
	}
	sort.Strings(current)
	return current, other, nil
}
