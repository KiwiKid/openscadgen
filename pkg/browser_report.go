package pkg

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

const (
	browserBundleSchema = 1
	browserOutputDir    = "browser"
	browserAssetsFolder = "static/browser"
	browserAssetsEnv    = "OPENSCADGEN_BROWSER_ASSETS"
)

// BrowserSource is one OpenSCAD source file exposed to the browser renderer.
type BrowserSource struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Text   string `json:"text"`
}

// BrowserInstance is a single resolved render request. Definitions are the
// same applied parameters used by the CLI, so browser output matches host
// output for the common case.
type BrowserInstance struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	AutoName    string                 `json:"autoName"`
	Description string                 `json:"description,omitempty"`
	SourcePath  string                 `json:"sourcePath"`
	Definitions map[string]interface{} `json:"definitions"`
	OutputPath  string                 `json:"outputPath"`
	Render      bool                   `json:"render"`
}

// BrowserLibrary describes a library archive that the browser may fetch on
// demand. Archives are gzipped JSON maps of relative path to source text.
type BrowserLibrary struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Archive string `json:"archive,omitempty"`
	SHA256  string `json:"sha256,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Missing bool   `json:"missing,omitempty"`
}

// BrowserCapabilities tells the UI which parts of the host contract the
// browser engine can currently satisfy. Image rendering is deliberately
// unsupported because OpenSCAD-WASM has no OpenGL/Cairo build.
type BrowserCapabilities struct {
	STL      bool `json:"stl"`
	Images   bool `json:"images"`
	Preview  bool `json:"preview"`
	Manifold bool `json:"manifold"`
}

// BrowserBundle is the versioned data contract shared by the report template,
// the browser worker and any future remote render endpoint.
type BrowserBundle struct {
	Schema       int                 `json:"schema"`
	Generator    BrowserGenerator    `json:"generator"`
	Config       BrowserConfig       `json:"config"`
	Sources      []BrowserSource     `json:"sources"`
	Libraries    []BrowserLibrary    `json:"libraries"`
	Instances    []BrowserInstance   `json:"instances"`
	Capabilities BrowserCapabilities `json:"capabilities"`
}

type BrowserGenerator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type BrowserConfig struct {
	Path string `json:"path"`
	TOML string `json:"toml"`
}

type browserLibraryArchive struct {
	Name    string            `json:"name"`
	Version string            `json:"version,omitempty"`
	Files   map[string]string `json:"files"`
}

// PrepareBrowserReport writes the browser runtime and library archives beside
// report.html and returns the JSON bundle to embed in the page.
func PrepareBrowserReport(config *models.Config, instances []models.InstanceConfig, outputDir string) (string, error) {
	if config == nil {
		return "", fmt.Errorf("browser report: config is nil")
	}

	browserDir := filepath.Join(outputDir, browserOutputDir)
	if err := copyBrowserAssets(browserDir); err != nil {
		return "", err
	}

	bundle, sources, err := buildBrowserBundle(config, instances)
	if err != nil {
		return "", err
	}
	libraries, err := writeBrowserLibraries(browserDir, sources)
	if err != nil {
		return "", err
	}
	bundle.Libraries = libraries

	data, err := json.Marshal(bundle)
	if err != nil {
		return "", fmt.Errorf("browser report: marshal bundle: %w", err)
	}
	if err := os.WriteFile(filepath.Join(browserDir, "bundle.json"), data, 0o644); err != nil {
		return "", fmt.Errorf("browser report: write bundle.json: %w", err)
	}
	return string(data), nil
}

func buildBrowserBundle(config *models.Config, instances []models.InstanceConfig) (*BrowserBundle, []BrowserSource, error) {
	tomlText := config.RawConfigFile
	if strings.TrimSpace(tomlText) == "" && config.ConfigFile != "" {
		data, err := os.ReadFile(config.ConfigFile)
		if err != nil {
			return nil, nil, fmt.Errorf("browser report: read config: %w", err)
		}
		tomlText = string(data)
	}
	if strings.TrimSpace(tomlText) == "" {
		return nil, nil, fmt.Errorf("browser report: config TOML is empty")
	}

	baseDir := filepath.Dir(config.ConfigFile)
	if baseDir == "." || baseDir == "" {
		baseDir = "."
	}
	sourceMap := make(map[string]string)
	sources := make([]BrowserSource, 0, len(config.GetInputPaths()))
	seen := make(map[string]bool)

	for _, input := range config.GetInputPaths() {
		fullPath := input.Path
		if fullPath == "" {
			fullPath = input.RawOpenSCADFileName
		}
		if fullPath == "" {
			continue
		}
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(baseDir, fullPath)
		}
		fullPath = filepath.Clean(fullPath)

		if seen[fullPath] {
			continue
		}
		seen[fullPath] = true

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, nil, fmt.Errorf("browser report: read source %s: %w", input.Path, err)
		}
		browserPath := browserSourcePath(baseDir, fullPath)
		sourceMap[fullPath] = browserPath
		sources = append(sources, BrowserSource{
			Path:   browserPath,
			SHA256: sha256Hex(data),
			Text:   string(data),
		})
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })

	renderInstances := make([]BrowserInstance, 0, len(instances))
	for index, instance := range instances {
		if instance.ImageOnly {
			continue
		}
		if instance.OutputPathV2 == "" && instance.RunOutputPathV3 == "" {
			continue
		}

		inputPath := instance.InputPath.Path
		if inputPath == "" {
			inputPath = instance.InputPath.RawOpenSCADFileName
		}
		if inputPath == "" {
			continue
		}
		if !filepath.IsAbs(inputPath) {
			inputPath = filepath.Join(baseDir, inputPath)
		}
		inputPath = filepath.Clean(inputPath)
		browserPath, ok := sourceMap[inputPath]
		if !ok {
			continue
		}

		definitions := instance.Params
		for _, result := range instance.STLResults {
			if len(result.AppliedParams) > 0 {
				definitions = result.AppliedParams
				break
			}
		}
		definitions = withoutVersionParam(definitions)

		outputPath := filepath.Base(instance.RunOutputPathV3)
		if outputPath == "" || outputPath == "." {
			continue
		}
		id := instance.UniqueID
		if id == "" {
			id = fmt.Sprintf("instance-%d", index+1)
		}
		renderInstances = append(renderInstances, BrowserInstance{
			ID:          id,
			Name:        instance.Name,
			AutoName:    instance.AutoName,
			Description: instance.Description,
			SourcePath:  browserPath,
			Definitions: definitions,
			OutputPath:  filepath.ToSlash(outputPath),
			Render:      !config.SkipRender,
		})
	}

	bundle := &BrowserBundle{
		Schema: browserBundleSchema,
		Generator: BrowserGenerator{
			Name:    "openscadgen",
			Version: config.OpenScadGenVersion,
		},
		Config: BrowserConfig{
			Path: filepath.Base(config.ConfigFile),
			TOML: tomlText,
		},
		Sources:   sources,
		Instances: renderInstances,
		Capabilities: BrowserCapabilities{
			STL:      true,
			Images:   false,
			Preview:  true,
			Manifold: true,
		},
	}
	return bundle, sources, nil
}

func withoutVersionParam(params map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for key, value := range params {
		if key == "version" {
			continue
		}
		out[key] = value
	}
	return out
}

func browserSourcePath(baseDir, fullPath string) string {
	relative, err := filepath.Rel(baseDir, fullPath)
	if err == nil {
		relative = filepath.ToSlash(relative)
		if relative != "." && relative != ".." && !strings.HasPrefix(relative, "../") {
			return relative
		}
	}
	return filepath.Base(fullPath)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func writeBrowserLibraries(browserDir string, sources []BrowserSource) ([]BrowserLibrary, error) {
	usesBOSL2 := false
	for _, source := range sources {
		if strings.Contains(source.Text, "BOSL2/") {
			usesBOSL2 = true
			break
		}
	}
	if !usesBOSL2 {
		return nil, nil
	}

	libraryDir := filepath.Join(browserDir, "libraries")
	if err := os.MkdirAll(libraryDir, 0o755); err != nil {
		return nil, fmt.Errorf("browser report: create library directory: %w", err)
	}

	libraryPath, version := findInstalledOpenSCADLibrary(openSCADLibraryCatalog[0])
	if libraryPath == "" {
		return []BrowserLibrary{{Name: "BOSL2", Missing: true}}, nil
	}

	files := make(map[string]string)
	err := filepath.WalkDir(libraryPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".scad") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(libraryPath, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = string(data)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("browser report: read BOSL2 library: %w", err)
	}
	if len(files) == 0 {
		return []BrowserLibrary{{Name: "BOSL2", Version: version, Missing: true}}, nil
	}

	archiveName := "libraries/BOSL2.json.gz"
	archivePath := filepath.Join(browserDir, filepath.FromSlash(archiveName))
	compressed, err := writeGzipJSON(archivePath, browserLibraryArchive{Name: "BOSL2", Version: version, Files: files})
	if err != nil {
		return nil, err
	}
	return []BrowserLibrary{{
		Name:    "BOSL2",
		Version: version,
		Archive: filepath.ToSlash(archiveName),
		SHA256:  sha256Hex(compressed),
		Size:    int64(len(compressed)),
	}}, nil
}

func writeGzipJSON(path string, value interface{}) ([]byte, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("browser report: create library directory: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("browser report: create %s: %w", path, err)
	}
	writer := gzip.NewWriter(file)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		_ = writer.Close()
		_ = file.Close()
		return nil, fmt.Errorf("browser report: write %s: %w", path, err)
	}
	if err := writer.Close(); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("browser report: close %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("browser report: close %s: %w", path, err)
	}
	return os.ReadFile(path)
}

func copyBrowserAssets(destinationDir string) error {
	sourceDir, err := locateBrowserAssets()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return fmt.Errorf("browser report: create assets directory: %w", err)
	}
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("browser report: list assets: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == ".gitignore" || strings.HasPrefix(name, ".") {
			continue
		}
		if err := copyFileIfChanged(filepath.Join(sourceDir, name), filepath.Join(destinationDir, name)); err != nil {
			return fmt.Errorf("browser report: copy %s: %w", name, err)
		}
	}
	return nil
}

func locateBrowserAssets() (string, error) {
	candidates := make([]string, 0, 4)
	if configured := strings.TrimSpace(os.Getenv(browserAssetsEnv)); configured != "" {
		candidates = append(candidates, configured)
	}
	candidates = append(candidates, browserAssetsFolder)
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), browserAssetsFolder))
	}
	for _, candidate := range candidates {
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if browserAssetsPresent(absolute) {
			return absolute, nil
		}
	}
	return "", fmt.Errorf("browser report: OpenSCAD WebAssembly assets not found; run `node scripts/fetch-browser-assets.mjs` or set %s", browserAssetsEnv)
}

func browserAssetsPresent(dir string) bool {
	for _, name := range []string{"og-browser.js", "og-browser-worker.js", "openscad.js", "openscad.wasm"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.IsDir() || info.Size() == 0 {
			return false
		}
	}
	return true
}

func copyFileIfChanged(source, destination string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if destinationInfo, err := os.Stat(destination); err == nil &&
		destinationInfo.Size() == sourceInfo.Size() &&
		!destinationInfo.ModTime().Before(sourceInfo.ModTime()) {
		// The report directory can be regenerated many times; avoid copying the
		// ~10 MB OpenSCAD WASM file when the pinned asset has not changed.
		return nil
	}

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return err
	}
	return output.Close()
}
