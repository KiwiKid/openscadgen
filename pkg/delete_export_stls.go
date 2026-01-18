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

