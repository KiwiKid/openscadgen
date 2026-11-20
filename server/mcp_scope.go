package server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

// projectRef is one project folder discovered under an MCP root.
type projectRef struct {
	Label      string
	Root       string
	ConfigPath string
}

// cleanPath normalises a path for comparisons.
func cleanPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return filepath.Clean(path)
}

// pathWithinRoot reports whether target is root or lives underneath it. Both
// arguments are expected to be already canonical.
func pathWithinRoot(root string, target string) bool {
	if root == "" || target == "" {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(target))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// evalAllowMissing resolves symlinks for path, tolerating components that do
// not exist yet (write_project_file creates them). Existing ancestors are fully
// resolved, so a symlinked parent cannot be used to escape a root.
func evalAllowMissing(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)

	current := abs
	missing := make([]string, 0, 4)
	for {
		resolved, evalErr := filepath.EvalSymlinks(current)
		if evalErr == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return resolved, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", evalErr
		}
		missing = append(missing, filepath.Base(current))
		current = parent
		if len(missing) > 64 {
			return "", fmt.Errorf("path %q is too deep to resolve", path)
		}
	}
}

// containsGitDir reports whether a root-relative path reaches into .git. Git
// internals are off limits: .git/config can redirect git commands and .git/hooks
// can execute code, so file tools must not write them.
func containsGitDir(relPath string) bool {
	for _, part := range strings.Split(filepath.ToSlash(relPath), "/") {
		if strings.EqualFold(part, ".git") {
			return true
		}
	}
	return false
}

// resolveToolPath resolves a requested tool path against the configured roots
// and rejects anything that escapes them, including through symlinks.
//
// With one root, relative paths resolve from that root. With several roots a
// relative path must identify exactly one root (either "<root>/..." from the
// common base, or a path that exists in only one root); ambiguous requests are
// refused with the candidates listed.
func resolveToolPath(roots []string, requestedPath string, allowRoot bool) (absPath string, relPath string, root string, err error) {
	if len(roots) == 0 {
		return "", "", "", fmt.Errorf("no MCP root is configured; pass -mcp-root or -sf")
	}

	cleanRequested := strings.TrimSpace(requestedPath)
	if cleanRequested == "" {
		if !allowRoot {
			return "", "", "", fmt.Errorf("path is required")
		}
		resolved, resolveErr := evalAllowMissing(roots[0])
		if resolveErr != nil {
			return "", "", "", resolveErr
		}
		return resolved, ".", roots[0], nil
	}

	if filepath.IsAbs(cleanRequested) {
		candidate := filepath.Clean(cleanRequested)
		for _, candidateRoot := range roots {
			resolvedRoot, rootErr := evalAllowMissing(candidateRoot)
			if rootErr != nil {
				continue
			}
			resolvedCandidate, candidateErr := evalAllowMissing(candidate)
			if candidateErr != nil {
				continue
			}
			if pathWithinRoot(resolvedRoot, resolvedCandidate) {
				rel, relErr := filepath.Rel(resolvedRoot, resolvedCandidate)
				if relErr != nil {
					continue
				}
				if rel == "." && !allowRoot {
					continue
				}
				if containsGitDir(rel) {
					return "", "", "", fmt.Errorf("path %q points into .git, which is not accessible to MCP tools", requestedPath)
				}
				return resolvedCandidate, rel, candidateRoot, nil
			}
		}
		return "", "", "", fmt.Errorf("path %q is outside the configured MCP roots (%s)", requestedPath, strings.Join(roots, ", "))
	}

	type match struct {
		abs  string
		rel  string
		root string
	}
	matches := make([]match, 0, len(roots))

	for _, candidateRoot := range roots {
		resolvedRoot, rootErr := evalAllowMissing(candidateRoot)
		if rootErr != nil {
			continue
		}
		candidate := filepath.Clean(filepath.Join(candidateRoot, cleanRequested))
		resolvedCandidate, candidateErr := evalAllowMissing(candidate)
		if candidateErr != nil {
			continue
		}
		if !pathWithinRoot(resolvedRoot, resolvedCandidate) {
			// A path such as "..\outside" is always an error, even if another
			// root would technically contain it.
			if strings.HasPrefix(cleanRequested, "..") {
				return "", "", "", fmt.Errorf("path %q is outside the configured MCP roots", requestedPath)
			}
			continue
		}
		rel, relErr := filepath.Rel(resolvedRoot, resolvedCandidate)
		if relErr != nil {
			continue
		}
		if rel == "." && !allowRoot {
			continue
		}
		if containsGitDir(rel) {
			return "", "", "", fmt.Errorf("path %q points into .git, which is not accessible to MCP tools", requestedPath)
		}
		matches = append(matches, match{abs: resolvedCandidate, rel: rel, root: candidateRoot})
	}

	switch len(matches) {
	case 0:
		return "", "", "", fmt.Errorf("path %q is outside the configured MCP roots (%s)", requestedPath, strings.Join(roots, ", "))
	case 1:
		return matches[0].abs, matches[0].rel, matches[0].root, nil
	default:
		// Several roots contain the same relative path. If exactly one of them
		// actually has the file, prefer it; otherwise the caller must qualify.
		existing := make([]match, 0, len(matches))
		for _, candidate := range matches {
			if _, statErr := os.Lstat(candidate.abs); statErr == nil {
				existing = append(existing, candidate)
			}
		}
		if len(existing) == 1 {
			return existing[0].abs, existing[0].rel, existing[0].root, nil
		}
		candidates := make([]string, 0, len(matches))
		for _, candidate := range matches {
			candidates = append(candidates, filepath.ToSlash(candidate.rel))
		}
		return "", "", "", fmt.Errorf(
			"path %q is ambiguous with %d MCP roots; qualify it as one of: %s",
			requestedPath, len(roots), strings.Join(candidates, ", "))
	}
}

// gitRoot returns the repository/project directory used by git tools. Git needs
// a single project, so multiple roots are refused unless the requested path
// identifies one.
func gitRoot(roots []string) (string, error) {
	switch len(roots) {
	case 0:
		return "", fmt.Errorf("no MCP root is configured")
	case 1:
		return roots[0], nil
	default:
		return "", fmt.Errorf("git tools need a single MCP root; configure -mcp-root with one folder (%d configured)", len(roots))
	}
}

// discoverProjects scans the roots for folders containing config.toml.
func discoverProjects(roots []string) []projectRef {
	projects := make([]projectRef, 0)
	seen := make(map[string]bool)

	for _, root := range roots {
		files, err := pkg.ScanFolderForConfigFiles(root)
		if err != nil && len(files) == 0 {
			continue
		}
		base := cleanPath(root)
		for _, file := range files {
			configPath := strings.TrimSpace(file.Path)
			if configPath == "" {
				continue
			}
			if !filepath.IsAbs(configPath) {
				configPath = filepath.Join(base, configPath)
			}
			projectRoot := cleanPath(filepath.Dir(configPath))
			key := strings.ToLower(projectRoot)
			if projectRoot == "" || seen[key] {
				continue
			}
			seen[key] = true
			label := mcpProjectLabel(base, projectRoot)
			projects = append(projects, projectRef{
				Label:      label,
				Root:       projectRoot,
				ConfigPath: cleanPath(configPath),
			})
		}
	}

	sortProjects(projects)
	return projects
}

// mcpProjectLabel is the project label used in listings and settings output.
func mcpProjectLabel(base string, projectRoot string) string {
	if base != "" {
		if rel, err := filepath.Rel(base, projectRoot); err == nil {
			switch {
			case rel == ".":
				return filepath.Base(base)
			case rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)):
				return filepath.ToSlash(rel)
			}
		}
	}
	if label := filepath.Base(projectRoot); label != "" && label != "." {
		return label
	}
	return projectRoot
}

func sortProjects(projects []projectRef) {
	for i := 1; i < len(projects); i++ {
		for j := i; j > 0 && strings.ToLower(projects[j].Label) < strings.ToLower(projects[j-1].Label); j-- {
			projects[j], projects[j-1] = projects[j-1], projects[j]
		}
	}
}

// shouldSkipDir keeps listings focused on project content.
func shouldSkipDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", "node_modules", ".svn", ".hg":
		return true
	default:
		return false
	}
}

// isSymlink reports whether a directory entry is a link that must not be
// traversed by the listing walker.
func isSymlink(entry fs.DirEntry) bool {
	return entry.Type()&fs.ModeSymlink != 0
}

// projectSummaryFor builds a short one-line description of a project.
func projectSummaryFor(project projectRef) models.MCPProjectInfo {
	return models.MCPProjectInfo{
		Label:      project.Label,
		Root:       project.Root,
		ConfigPath: project.ConfigPath,
	}
}
