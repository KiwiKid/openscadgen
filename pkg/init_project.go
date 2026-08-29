package pkg

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kiwikid/openscadgen/pkg/models"
)

const initTemplateDir = "openscad-init-templates"
const initDefaultTemplateFile = "default.toml"

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

func hasExistingProjectConfig(parentDir string) (bool, error) {
	found := false
	err := filepath.WalkDir(parentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == "config.toml" {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found, err
}

func promptForExplainerUpgrade() bool {
	reader := bufio.NewReader(os.Stdin)
	answerCh := make(chan string, 1)
	go func() {
		line, _ := reader.ReadString('\n')
		answerCh <- strings.TrimSpace(line)
	}()

	select {
	case answer := <-answerCh:
		answer = strings.ToLower(answer)
		return answer == "" || answer == "y" || answer == "yes"
	case <-time.After(10 * time.Second):
		return true
	}
}

func ChooseInitTemplate(f models.CmdFlags, parentDir string) (bool, error) {
	if f.NoInput {
		return false, nil
	}
	existing, err := hasExistingProjectConfig(parentDir)
	if err != nil {
		return false, err
	}
	if existing {
		return false, nil
	}
	fmt.Printf("No existing project config found under %s.\n", parentDir)
	fmt.Printf("Create the explainer starter project first? [Y/n] (auto-yes in 10s): ")
	if promptForExplainerUpgrade() {
		fmt.Println("Y")
		return true, nil
	}
	fmt.Println("n")
	return false, nil
}

func resolveInitTemplatePath(templateName string) (string, error) {
	name := strings.TrimSpace(templateName)
	if name == "" || name == "basic" {
		return "", nil
	}
	if name == "extended" {
		name = initDefaultTemplateFile
	}
	cleanName := filepath.Clean(name)
	_, file, _, ok := runtime.Caller(0)
	searchDir := "."
	if ok {
		searchDir = filepath.Dir(filepath.Dir(file))
	}
	for {
		templatePath := filepath.Join(searchDir, initTemplateDir, cleanName)
		if st, err := os.Stat(templatePath); err == nil {
			if st.IsDir() {
				return "", fmt.Errorf("template path is a directory: %s", templatePath)
			}
			return templatePath, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat template file %s: %w", templatePath, err)
		}
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break
		}
		searchDir = parent
	}
	return "", fmt.Errorf("template file not found: %s/%s", initTemplateDir, cleanName)
}

func readInitTemplate(templateName string) (string, error) {
	templatePath, err := resolveInitTemplatePath(templateName)
	if err != nil {
		return "", err
	}
	if templatePath == "" {
		return configTemplate, nil
	}
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read template file %s: %w", templatePath, err)
	}
	return string(data), nil
}

func renderInitTemplate(templateBody string, tokens map[string]string) string {
	out := templateBody
	for token, value := range tokens {
		out = strings.ReplaceAll(out, token, value)
	}
	return out
}

// InitConfigInParent creates a project directory from projectName inside parentDir.
// It preserves the original boolean API used by existing call sites.
func InitConfigInParent(parentDir, projectName string, extended bool) error {
	template := ""
	if extended {
		template = initDefaultTemplateFile
	}
	return InitConfigInParentWithTemplate(parentDir, projectName, template)
}

// InitConfigInParentWithTemplate creates a project directory from projectName inside parentDir
// using the requested template. projectName may be a single leaf ("my-part") or a relative
// path ("examples/spoon-holder"); each path segment is sanitized. parentDir is left as-is.
func InitConfigInParentWithTemplate(parentDir, projectName string, template string) error {
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

	useExtendedTemplate := filepath.Base(strings.TrimSpace(template)) == initDefaultTemplateFile || strings.TrimSpace(template) == "extended"

	configBody := configTemplate
	if strings.TrimSpace(template) != "" {
		var err error
		configBody, err = readInitTemplate(template)
		if err != nil {
			logError(fmt.Sprintf("Failed to load init template: %s", err))
			return err
		}
	}
	configBody = renderInitTemplate(configBody, map[string]string{
		"{{projectName}}":         projectNameUnderLined,
		"<projectName>":           projectNameUnderLined,
		"<projectNameUnderLined>": projectNameUnderLined,
		"<projectPath>":           projectPath,
	})

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

	if useExtendedTemplate {
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

	if useExtendedTemplate {
		libDir := filepath.Join(parentDir, "openscad-lib")
		if err := os.MkdirAll(libDir, 0o755); err != nil {
			logError(fmt.Sprintf("Failed to create library directory: %s", err))
			return fmt.Errorf("create library dir: %w", err)
		}
		libPath := filepath.Join(libDir, "openscadgen-core.scad")
		if err := os.WriteFile(libPath, []byte(openScadCoreLibrary), 0o644); err != nil {
			logError(fmt.Sprintf("Failed to write library file: %s", err))
			return fmt.Errorf("write library: %w", err)
		}
		logCreation(fmt.Sprintf("Wrote shared OpenSCAD library: %s", libPath))
	}

	logCreation(fmt.Sprintf("Project Successfully Initialized: %s", sanitizedRel))
	LogKeyValuePair("Project Path", projectPath)
	logCreation("\nUse the command:\n\n\topenscadgen -c " + configPath + "\n\nto generate the STL files")
	return nil
}
