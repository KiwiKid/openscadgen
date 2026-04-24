package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

var initProjectNameReplacements = map[string]string{
	" ":  "_",
	"-":  "_",
	".":  "_",
	"__": "_",
}

func sanitizeInitProjectSegment(name string) string {
	out := name
	for old, new := range initProjectNameReplacements {
		out = strings.ReplaceAll(out, old, new)
	}
	return filepath.Clean(out)
}

// sanitizeInitRelPath applies sanitizeInitProjectSegment to each path element of rel
// (after filepath.Clean), preserving directory structure for nested -i paths.
func sanitizeInitRelPath(rel string) string {
	rel = filepath.Clean(rel)
	if rel == "." || rel == ".." {
		return rel
	}
	dir, base := filepath.Dir(rel), filepath.Base(rel)
	if dir == "." || dir == "" {
		return sanitizeInitProjectSegment(base)
	}
	return filepath.Join(sanitizeInitRelPath(dir), sanitizeInitProjectSegment(base))
}

// ResolveInitParentDir returns the directory where a new project folder should be created.
// Explicit -init-dir wins; otherwise if ServerFolder is set it is used; else ".".
func ResolveInitParentDir(f models.CmdFlags) string {
	if strings.TrimSpace(f.InitProjectParentDir) != "" {
		return filepath.Clean(f.InitProjectParentDir)
	}
	if strings.TrimSpace(f.ServerFolder) != "" {
		return filepath.Clean(f.ServerFolder)
	}
	return "."
}

// InitConfigInParent creates a project directory from projectName inside parentDir
// (same templates as -i / -ie). projectName may be a single leaf ("my-part") or a relative
// path ("examples/spoon-holder"); each path segment is sanitized. parentDir is left as-is.
func InitConfigInParent(parentDir, projectName string, extended bool) error {
	parentDir = filepath.Clean(parentDir)
	rel := filepath.Clean(strings.TrimSpace(projectName))
	if rel == "" || rel == "." || rel == ".." {
		return fmt.Errorf("invalid project name %q", projectName)
	}
	leaf := filepath.Base(rel)
	if leaf == "" || leaf == "." || leaf == ".." {
		return fmt.Errorf("invalid project name %q", projectName)
	}
	if st, err := os.Stat(parentDir); err != nil {
		return fmt.Errorf("parent directory %s: %w", parentDir, err)
	} else if !st.IsDir() {
		return fmt.Errorf("parent is not a directory: %s", parentDir)
	}

	sanitizedRel := sanitizeInitRelPath(rel)
	projectPath := filepath.Join(parentDir, sanitizedRel)
	projectPath = filepath.Clean(projectPath)

	absParent, err := filepath.Abs(parentDir)
	if err != nil {
		return fmt.Errorf("abs parent: %w", err)
	}
	absProject, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("abs project path: %w", err)
	}
	if absProject != absParent && !strings.HasPrefix(absProject, absParent+string(filepath.Separator)) {
		return fmt.Errorf("project path escapes parent directory")
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		if err := os.MkdirAll(projectPath, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", projectPath, err)
		}
	} else {
		logError(fmt.Sprintf("Project name already taken: %s", projectPath))
		return fmt.Errorf("project name already taken")
	}

	sanitizedLeaf := filepath.Base(sanitizedRel)
	projectNameUnderLined := strings.NewReplacer(" ", "_").Replace(sanitizedLeaf)

	var configBody string
	if extended {
		configBody = configTemplateExtended
	} else {
		configBody = configTemplate
	}

	configBody = strings.ReplaceAll(configBody, "{{projectName}}", projectNameUnderLined)

	configPath := filepath.Join(projectPath, "config.toml")
	configFile, err := os.Create(configPath)
	if err != nil {
		logError(fmt.Sprintf("Failed to create config file: %s", err))
		return fmt.Errorf("create config: %w", err)
	}
	defer configFile.Close()
	_, err = configFile.WriteString(configBody)
	if err != nil {
		logError(fmt.Sprintf("Failed to write template to config file: %s", err))
		return fmt.Errorf("write config: %w", err)
	}
	logCreation(fmt.Sprintf("Created config file: %s", configPath))
	logCreation(fmt.Sprintf("Wrote template to config file: %s", configPath))

	scadPath := filepath.Join(projectPath, projectNameUnderLined+".scad")
	scadFile, err := os.Create(scadPath)
	if err != nil {
		logError(fmt.Sprintf("Failed to create scad file: %s", err))
		return fmt.Errorf("create scad: %w", err)
	}
	defer scadFile.Close()

	if extended {
		_, err = scadFile.WriteString(openScadTemplateExtended(projectNameUnderLined))
		if err != nil {
			logError(fmt.Sprintf("Failed to write template to scad file: %s", err))
			return fmt.Errorf("write scad: %w", err)
		}
	} else {
		_, err = scadFile.WriteString(openScadTemplate(projectNameUnderLined))
		if err != nil {
			logError(fmt.Sprintf("Failed to write template to scad file: %s", err))
			return fmt.Errorf("write scad: %w", err)
		}
	}
	logCreation(fmt.Sprintf("Wrote template to scad file: %s", scadPath))

	logCreation(fmt.Sprintf("Project Successfully Initialized: %s", sanitizedRel))
	LogKeyValuePair("Project Path", projectPath)
	logCreation("\nUse the command:\n\n\topenscadgen -c " + configPath + "\n\nto generate the STL files")
	return nil
}
