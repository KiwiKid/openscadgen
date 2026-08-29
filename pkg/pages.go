package pkg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var pagesSourceRoots = []string{
	"examples",
	"bols2otropolis",
}

var pagesCopyExtensions = map[string]bool{
	".html": true,
	".htm":  true,
	".md":   true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".stl":  true,
	".3mf":  true,
	".pdf":  true,
	".zip":  true,
	".toml": true,
	".scad": true,
	".txt":  true,
	".log":  true,
}

func BuildPagesSite(repoRoot string, outputDir string) error {
	if repoRoot == "" {
		repoRoot = "."
	}
	if outputDir == "" {
		outputDir = "pages"
	}

	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolve pages output dir: %w", err)
	}

	if err := os.RemoveAll(absOutputDir); err != nil {
		return fmt.Errorf("clear pages output dir: %w", err)
	}
	if err := os.MkdirAll(absOutputDir, 0o755); err != nil {
		return fmt.Errorf("create pages output dir: %w", err)
	}

	for _, sourceRoot := range pagesSourceRoots {
		sourceRootAbs := filepath.Join(absRepoRoot, sourceRoot)
		if _, err := os.Stat(sourceRootAbs); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("stat %s: %w", sourceRoot, err)
		}
		if err := filepath.WalkDir(sourceRootAbs, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel(absRepoRoot, path)
			if err != nil {
				return err
			}
			destPath := filepath.Join(absOutputDir, rel)
			if d.IsDir() {
				return os.MkdirAll(destPath, 0o755)
			}
			if !shouldCopyForPages(path) {
				return nil
			}
			return copyPagesFile(absRepoRoot, path, destPath)
		}); err != nil {
			return fmt.Errorf("copy %s: %w", sourceRoot, err)
		}
	}

	return nil
}

func shouldCopyForPages(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return pagesCopyExtensions[ext]
}

func copyPagesFile(repoRoot, src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if strings.EqualFold(filepath.Ext(src), ".html") {
		data = rewritePagesHTML(repoRoot, data)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func rewritePagesHTML(repoRoot string, data []byte) []byte {
	content := string(data)
	prefixes := []string{
		"file:///" + filepath.ToSlash(repoRoot) + "/",
		"file://" + filepath.ToSlash(repoRoot) + "/",
	}
	for _, prefix := range prefixes {
		content = strings.ReplaceAll(content, prefix, "/")
	}
	return []byte(content)
}
