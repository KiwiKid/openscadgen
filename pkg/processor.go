package pkg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
	"github.com/pkg/xattr"
)

func ShowMan() {
	flag.PrintDefaults()
}
func getOutputPaths(config *models.Config) models.OutputPaths {
	// Get the directory containing the config file - this is our anchor point
	configDir := filepath.Dir(config.ConfigFile)

	// Use version as-is for paths (no normalization)
	versionPath := config.Design.Version

	// Get the design name from the first input path
	var designName string
	if len(config.GetInputPaths()) > 0 {
		designName = strings.TrimSuffix(filepath.Base(config.GetInputPaths()[0].Path), ".scad")
	} else {
		designName = "test_design"
	}

	exportNameFormat := config.Design.ExportNameFormat
	hasExportPrefix := strings.HasPrefix(exportNameFormat, "export/") || strings.HasPrefix(exportNameFormat, "/export")

	// Determine the base directory for exports
	var baseDir string
	if len(config.GetInputPaths()) > 0 {
		inputPath := config.GetInputPaths()[0].Path
		if filepath.IsAbs(inputPath) {
			// For absolute input paths, use the input file's directory
			baseDir = filepath.Dir(inputPath)
		} else {
			// For relative input paths, always use the config file's directory
			// This ensures exports always go to a sibling 'export' folder relative to the config file
			baseDir = configDir
		}
	} else {
		baseDir = configDir
	}

	// Construct paths relative to the base directory
	var exportFolderPath, baseExportPath string
	if hasExportPrefix {
		exportFolderPath = baseDir
		baseExportPath = baseDir
	} else {
		exportFolderPath = filepath.Join(baseDir, "export", versionPath)
		baseExportPath = filepath.Join(baseDir, "export", versionPath, designName)
	}

	outputPath := filepath.Join(baseExportPath, "export", versionPath)

	output := models.OutputPaths{
		OutputPath:            outputPath,
		ExportFolderPath:      exportFolderPath,
		LowQualityWarningPath: filepath.Join(baseExportPath, "LOW_QUALITY_WARNING.md"),
		ReadmePath:            filepath.Join(baseExportPath, "README.md"),
		LogOutputPath:         filepath.Join(baseExportPath, "export_log.log"),
		ReportPath:            filepath.Join(baseExportPath, "report.html"),
	}

	if config.Debug {
		log.Printf("ExportNameFormat: %+v", config.Design.ExportNameFormat)
		log.Printf("configDir: %+v", configDir)
		log.Printf("baseDir: %+v", baseDir)
		log.Printf("outputPath: %+v", outputPath)
		log.Printf("exportFolderPath: %+v", exportFolderPath)
		log.Printf("baseExportPath: %+v", baseExportPath)

		log.Printf("OutputPaths:\n\n %+v", output)

	}
	if output.ExportFolderPath == "" || output.ExportFolderPath == "." {
		log.Panicf("2 hmm empty exportFolderPath folder path %s", output.ExportFolderPath)
	}

	if output.OutputPath == "" || output.OutputPath == "." {
		log.Panicf("2 hmm empty OutputPath folder path: %s", output.OutputPath)
	}

	return output
}

const OPENSCAD_VERSION_WARN_IF_OLDER_THAN = 2024

var config models.Config

var logger *log.Logger

var logBuffer bytes.Buffer
var logToMemory bool

var invalidChars = "?'\""

// BurntSushi/toml decode errors include "line N".
var tomlDecodeLineRE = regexp.MustCompile(`line (\d+)`)

// tomlDecodeErrorSnippet returns numbered source lines around the line mentioned in a TOML decode error.
func tomlDecodeErrorSnippet(configData string, decodeErr error) string {
	if decodeErr == nil {
		return ""
	}
	m := tomlDecodeLineRE.FindStringSubmatch(decodeErr.Error())
	if len(m) < 2 {
		return ""
	}
	lineNum, aerr := strconv.Atoi(m[1])
	if aerr != nil || lineNum < 1 {
		return ""
	}
	lines := strings.Split(configData, "\n")
	if lineNum > len(lines) {
		return ""
	}
	const radius = 3
	start := lineNum - radius
	if start < 1 {
		start = 1
	}
	end := lineNum + radius
	if end > len(lines) {
		end = len(lines)
	}
	var b strings.Builder
	for i := start; i <= end; i++ {
		marker := " "
		if i == lineNum {
			marker = ">"
		}
		fmt.Fprintf(&b, "%s %4d | %s\n", marker, i, lines[i-1])
	}
	return strings.TrimSuffix(b.String(), "\n")
}

const (
	colorReset  = "\033[0m"
	colorOrange = "\033[38;5;208m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

/*
```sh
git commit -m "New and improved version"
git tag "v2.7.24"
git push && git push --tags
```

To create a new version:

```sh
git commit -m "New and improved version"
git tag "v[NEW_VERSION_HERE]-alpha"
*/
const VERSION = "v2.8.2"

type Version struct {
	OpenSCADGen string
	OpenSCAD    string
	IsOutOfDate bool
}

type RunErrorCode int

const (
	RunErrorCode_NoInputPaths RunErrorCode = iota
	RunErrorCode_NoExportNameFormat
	RunErrorCode_DuplicateExportPath
)

var runErrorCodeName = map[RunErrorCode]string{
	RunErrorCode_NoInputPaths:        "No input paths found in config",
	RunErrorCode_NoExportNameFormat:  "No export name format found in config",
	RunErrorCode_DuplicateExportPath: "Duplicate export path found in config",
}

func GetVersion() Version {
	openSCADVersion, err := findOpenSCAD()
	if err != nil {
		log.Fatalf("GetVersion Error: %v", err)
	}
	return Version{
		OpenSCADGen: VERSION,
		OpenSCAD:    openSCADVersion.Version,
		IsOutOfDate: false,
	}
}

func GenerateInstanceConfigs(config *models.Config) ([]models.InstanceConfig, error) {
	if config == nil {
		return nil, fmt.Errorf("GenerateInstanceConfigs: config is nil")
	}
	if len(config.Design.ConfiguredInstanceConfig) == 0 {
		config.Design.ConfiguredInstanceConfig = []models.ConfiguredInstanceConfig{
			{
				Name:   "default",
				Params: map[string]interface{}{},
			},
		}
	}
	var instances []models.InstanceConfig
	for _, dynamicInstance := range config.Design.ConfiguredInstanceConfig {
		for _, inputPath := range config.GetInputPaths() {
			var err error
			var newInstances []models.InstanceConfig
			newInstances, _, err = GenerateInstances(config, dynamicInstance, inputPath)
			if err != nil {
				if config.StopOnError {
					logError(fmt.Sprintf("Warning: failed to generate instances: %v", err))
					continue
				}
				return newInstances, fmt.Errorf("failed to generate instances: %w", err)
			}

			newInstancesWithImages, err := populateExportImages(config, newInstances)
			if err != nil {
				if config.StopOnError {
					log.Printf("Warning: failed to generate export images: %v", err)
					continue
				}
				return newInstances, fmt.Errorf("failed to generate export images: %w", err)
			}

			if config.Debug {
				log.Printf("[DEBUG] Generated %d instances for inputPath %s", len(newInstances), inputPath.Path)
				for i, inst := range newInstances {
					log.Printf("[DEBUG] Instance %d: OutputPathV2=%s, SkippedReason=%s", i, inst.OutputPathV2, inst.SkippedReason)
				}
			}
			instances = append(instances, newInstancesWithImages...)
		}
	}

	return instances, nil
}

type Operations struct {
	GenerateReport bool
}

func Process(config *models.Config, progress ProgressReporter, cancel <-chan struct{}, operations Operations, isServerMode bool) (models.ProcessResult, error) {
	start := time.Now()
	if config.Debug {
		logStage("=== Processing === ")
	}

	// Get output paths
	outputPaths := getOutputPaths(config)

	// Create export folder if it doesn't exist
	if err := os.MkdirAll(outputPaths.ExportFolderPath, 0755); err != nil {
		return models.ProcessResult{}, fmt.Errorf("Process: failed to create export folder '%s': %w", outputPaths.ExportFolderPath, err)
	}

	// Check if export folder has existing files
	//clearExportFolder(config, outputPaths)

	// Generate instances
	var instances []models.InstanceConfig
	var stlResults []models.GenerateSTLResult
	var allImageResults []models.GenerateImageResult
	if config.Debug {
		log.Printf("[DEBUG] Generating instances for %d configured instances and %d input paths", len(config.Design.ConfiguredInstanceConfig), len(config.GetInputPaths()))
	}

	/*
		for _, dynamicInstance := range config.Design.ConfiguredInstanceConfig {
			for _, inputPath := range config.GetInputPaths() {
				if config.Debug {
					logStage(fmt.Sprintf("Generating instance %s", dynamicInstance.Name))
				}
				var err error
				var newInstances []models.InstanceConfig
				newInstances, _, err = GenerateInstances(config, dynamicInstance, inputPath)
				if err != nil {
					if config.StopOnError {
						logError(fmt.Sprintf("Warning: failed to generate instances: %v", err))
						continue
					}
					return models.ProcessResult{}, fmt.Errorf("failed to generate instances: %w", err)
				}

				newInstancesWithImages, err := populateExportImages(config, newInstances)
				if err != nil {
					if config.StopOnError {
						log.Printf("Warning: failed to generate export images: %v", err)
						continue
					}
					return models.ProcessResult{}, fmt.Errorf("failed to generate export images: %w", err)
				}

				if config.Debug {
					log.Printf("[DEBUG] Generated %d instances for inputPath %s", len(newInstances), inputPath.Path)
					for i, inst := range newInstances {
						log.Printf("[DEBUG] Instance %d: OutputPathV2=%s, SkippedReason=%s", i, inst.OutputPathV2, inst.SkippedReason)
					}
				}
				instances = append(instances, newInstancesWithImages...)
			}
		}*/

	instances, err := GenerateInstanceConfigs(config)
	if err != nil {
		return models.ProcessResult{}, fmt.Errorf("failed to generate instances: %w", err)
	}

	if config.Debug {
		log.Printf("[DEBUG] Total instances generated: %d", len(instances))
	}

	errors := validateInstances(instances, config)
	if len(errors) > 0 {
		logError("Validation of generated instances failed:")
		for _, error := range errors {
			logError(fmt.Sprintf("%d: %s\n%s", error.ErrorCode, runErrorCodeName[error.ErrorCode], error.Message))
			for k, v := range error.KVPs {
				LogKeyValuePair(k, v)
			}
		}
		if !config.StopOnError {
			if !config.Server {
				os.Exit(1)
			}
		}
	}

	nonSkippedInstances := 0
	for _, instance := range instances {
		if instance.SkippedReason == "" {
			nonSkippedInstances++
		}
	}

	completedInstances := 0

	// Construct progress for all instances
	progress.Construct(instances, nonSkippedInstances)

	for i := range instances {
		// Check for cancellation
		select {
		case <-cancel:
			return models.ProcessResult{}, fmt.Errorf("processing cancelled")
		default:
		}

		// Set PartIDLetter
		instances[i].PartIDLetter = getPartIDLetter(i)

		if instances[i].SkippedReason == "" {
			// Start processing this instance
			progress.StartInstance(instances[i].ID, instances[i].AutoName, i, nonSkippedInstances)

			if config.Debug {
				logCreation(fmt.Sprintf("Generated instance %s", instances[i].AutoName))
				log.Printf("[DEBUG] Generating STL for instance %d: OutputPathV2=%s", i, instances[i].OutputPathV2)
			}
			var result models.GenerateSTLResult
			var genErr error
			if !config.OnlyImages {
				result, genErr = generateSTL(&instances[i], config)
				if genErr != nil {

					result.OutputLog += genErr.Error()

					logError(fmt.Sprintf("Warning: failed to generate STL for instance %s:\n Error:\n%+v", instances[i].AutoName, genErr))
					stlResults = append(stlResults, result)
					completedInstances++
					if config.StopOnError {
						return models.ProcessResult{}, fmt.Errorf("failed to generate STL: %w", genErr)
					}
					instances[i].IsSuccessful = false
					continue

				} else {

					result.OutputLog = result.CommandOutput
					instances[i].STLResults = append(instances[i].STLResults, result)
					instances[i].IsSuccessful = true

					stlResults = append(stlResults, result)
				}
			}

			// Progress - STL complete
			//progress.ProgressInstance(instances[i].ID, 100)

			if !config.OnlyExport {
				genImageResult, err := processImage(config, &instances[i], progress)
				if err != nil {
					if config.StopOnError && !config.Server {
						logError(fmt.Sprintf("Warning: failed to generate image for instance %s:\n Error:\n%+v", instances[i].Name, err))
						continue
					}
				} else {
					instances[i].ImageResults = genImageResult
					for _, imageResult := range genImageResult {
						allImageResults = append(allImageResults, imageResult)
					}
				}
			} else if config.Debug {
				logSkip(fmt.Sprintf("Skipping image for instance %s", instances[i].AutoName))
			}

			// Progress - complete
			//progress.ProgressInstance(instances[i].ID, 100)

		} else {
			if config.Debug {
				logSkip(fmt.Sprintf("STL skipped %s %s", instances[i].AutoName, instances[i].SkippedReason))
			}
		}

		instances[i].IsComplete = true
		instances[i].CompletedAt = time.Now()

		// Finish this instance
		progress.FinishInstance()

		completedInstances++

		if config.Debug {
			LogKeyValuePair("ImageResults Count", fmt.Sprintf("%d", len(instances[i].ImageResults)))
		}
	}

	configDirectory := filepath.Dir(config.ConfigFile)
	exportLoc := filepath.Join(configDirectory, "export", config.Design.Version)

	if config.Debug {
		LogKeyValuePair("Process complete - ExportLocation:", exportLoc)
	}

	totalTime := time.Since(start)

	// Store processing time for next run
	storeProcessingTime(config, totalTime)

	// Processing complete

	// Signal completion
	progress.Done()

	models.SortInstanceConfigsByAutoName(instances)

	if operations.GenerateReport {
		_, location, genReportErr := GenerateOutputReport(config, instances, stlResults, allImageResults, exportLoc, operations.GenerateReport, totalTime)
		if genReportErr != nil {
			if config.StopOnError {
				log.Printf("Warning: failed to generate output report: %v", genReportErr)
			} else {
				log.Fatalf("failed to generate output report: %v", genReportErr)
			}
		} else if config.Debug {
			LogKeyValuePair("Output report generated at", location)
		}
	}

	return models.ProcessResult{
		ConfigFile:     config.ConfigFile,
		ExportLocation: exportLoc,
		Instances:      instances,
		STLResults:     stlResults,
		ImageResults:   allImageResults,
		TotalTimeTaken: totalTime,
	}, nil
}

type RunError struct {
	ErrorCode RunErrorCode
	Message   string
	KVPs      map[string]string
}

func validateInstances(instances []models.InstanceConfig, config *models.Config) []RunError {
	if config.Debug {
		logStage("Validating instances")
	}

	instanceParamCount := make(map[string]int)
	errors := []RunError{}
	exportPaths := make(map[string]bool)
	for _, instance := range instances {
		if _, exists := exportPaths[instance.RunOutputPathV3]; exists {

			var paramStr string = "\n"
			for k, v := range instance.Params {
				paramStr += fmt.Sprintf("%s=%v\n", k, v)
				instanceParamCount[k]++
			}

			LogKeyValuePair("Validation", fmt.Sprintf("%s: %+v", instance.RunOutputPathV3, exists))

			errors = append(errors, RunError{
				ErrorCode: RunErrorCode_DuplicateExportPath,
				Message:   fmt.Sprintf("Run Stopped as the export_name_format in the config file will result in duplicate export path: \n\n\t%s. \n\n Ensure the export_name_format includes all parameters (in {curlyBrackets}) that are different between instances. \n\nAnother common cause is including commas in parameters by mistake. \n\n If you param value has commas you would like to ignore, add `ignore_comma_in_params = [\"content\"]` to this [[openscadgen.instances]] tag ", instance.RunOutputPathV3),
				KVPs: map[string]string{
					"DuplicatePath":    instance.RunOutputPathV3,
					"ConfigFile":       config.ConfigFile,
					"InstanceCount":    fmt.Sprintf("%d", len(instances)),
					"exportNameFormat": config.Design.ExportNameFormat,
					"prams":            paramStr,
					"paramCount":       fmt.Sprintf("%+v", instanceParamCount),
				},
			})
		}
		exportPaths[instance.RunOutputPathV3] = true
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

func clearExportFolder(config *models.Config, outputPaths models.OutputPaths) {
	if files, err := os.ReadDir(outputPaths.ExportFolderPath); err == nil && len(files) > 0 {
		filesStr := ""
		for i, file := range files {
			if i < 5 {
				filesStr += fmt.Sprintf("\t- %s\n", file.Name())
			} else {
				filesStr += fmt.Sprintf("\tand %d other files ...\n", len(files)-5)
				break
			}
		}

		if config.NoProcessing {
			log.Printf(colorBlue + "No processing requested, skipping export folder actions" + colorReset)
			return
		} else if config.Debug {
			LogKeyValuePair("[processing] Export folder", outputPaths.ExportFolderPath)
		}

		// Skip deletion if run_type is appendOrOverwrite
		if config.Design.RunType == "appendOrOverwrite" {
			if config.Debug {
				LogKeyValuePair("run_type", config.Design.RunType)
				log.Printf(colorBlue + "Run type is appendOrOverwrite, skipping deletion of existing files" + colorReset)
			}
			return
		}
		if config.Debug {
			LogKeyValuePair("OverwriteExisting set, skipping check", outputPaths.OutputPath)
		}
		if config.Server {
			LogKeyValuePair("Server mode, skipping check", outputPaths.OutputPath)
		} else if !config.OverwriteExisting {
			LogWarn(fmt.Sprintf("\nThe export folder (%s) has %d existing files: \n%s\n\n(the '-ow' flag will skip this check)\n\n(tip: if you want to keep the existing stl export files, cancel this run and update the 'version' in the config file, this will generate a new folder and keep the existing files)", outputPaths.ExportFolderPath, len(files), filesStr), false)
			printDirectoryContents(outputPaths.ExportFolderPath)

			LogWarn(fmt.Sprintf("\n\n the files above can be overwritten: \n\n\t%s\n\nDo you want to continue? (y/n):", outputPaths.ExportFolderPath), true)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if response != "y\n" && response != "Y\n" {
				fmt.Println("Aborting operation.")
				os.Exit(1)
			}
			LogWarn("Prepare to recieve", false)
		} else if !strings.HasPrefix(outputPaths.ExportFolderPath, "export") {
			log.Printf("Export folder path does not start with export, skipping deletion")
			return
		}

		/*err := os.RemoveAll(outputPaths.ExportFolderPath)
		if err != nil {
			log.Panicf(colorRed+"Clear export folder Failed to delete export folder (outputPaths.ExportFolderPath): '%s' %s", outputPaths.ExportFolderPath, err)
		}*/
	}
}

func GetNiceName(path string) string {
	// Get the directory name from the path
	dir := filepath.Dir(path)
	// Get the base name of the directory (the parent folder name)
	return filepath.Base(dir)
}

func isImageFileExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	default:
		return false
	}
}

// listSortedImageAbsPathsInExportDir returns absolute paths to all image files under exportDir (recursive), sorted.
func listSortedImageAbsPathsInExportDir(exportDir string) []string {
	var images []string
	_ = filepath.Walk(exportDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info == nil || info.IsDir() {
			return nil
		}
		if !isImageFileExt(filepath.Ext(info.Name())) {
			return nil
		}
		images = append(images, p)
		return nil
	})
	sort.Strings(images)
	return images
}

func findPrimaryImageInExportDir(exportDir string) string {
	images := listSortedImageAbsPathsInExportDir(exportDir)
	if len(images) == 0 {
		return ""
	}
	for _, p := range images {
		if strings.Contains(strings.ToLower(filepath.Base(p)), "nice") {
			return p
		}
	}
	return images[0]
}

func ScanFolderForConfigFiles(folder string) ([]models.ConfigFile, error) {
	log.Printf("Scanning folder for config.toml files: %s", folder)
	var configFiles []models.ConfigFile
	maxSize := int64(2 * 1024 * 1024) // 2MB

	err := filepath.Walk(folder, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".toml") {
			return nil
		}
		if info.Size() > maxSize {
			return nil
		}

		// Skip files in export folders
		if strings.Contains(filePath, "/export/") || strings.Contains(filePath, "\\export\\") {
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return nil // skip unreadable files
		}
		defer f.Close()

		buf := make([]byte, 1024)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return nil // skip unreadable files
		}
		content := string(buf[:n])
		if !strings.Contains(content, "[openscadgen]") {
			return nil
		}

		subPath, err := filepath.Rel(folder, filePath)
		if err != nil {
			return err
		}

		dateModified, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		primaryImagePath := ""
		{
			configDir := filepath.Dir(filePath)
			exportDir := filepath.Join(configDir, "export")
			if st, err := os.Stat(exportDir); err == nil && st.IsDir() {
				primaryAbs := findPrimaryImageInExportDir(exportDir)
				if primaryAbs != "" {
					if rel, err := filepath.Rel(folder, primaryAbs); err == nil {
						primaryImagePath = filepath.ToSlash(rel)
					} else {
						// Best-effort fallback: absolute path (still servable by /images).
						primaryImagePath = filepath.ToSlash(primaryAbs)
					}
				}
			}
		}

		configFiles = append(configFiles, models.ConfigFile{
			Path:             subPath,
			NiceName:         GetNiceName(subPath),
			PrimaryImagePath: primaryImagePath,
			DateModified:     dateModified.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(configFiles, func(i, j int) bool {
		return configFiles[i].DateModified.After(configFiles[j].DateModified)
	})
	return configFiles, nil
}

func CleanDirectory(dryRun bool, cleanOldVersions bool, folder string) ([]models.CleanResult, error) {
	var results []models.CleanResult

	// Find all config.toml files
	configFiles, err := ScanFolderForConfigFiles(folder)
	if err != nil {
		return nil, err
	}

	for _, configFile := range configFiles {
		configPath := filepath.Join(folder, configFile.Path)
		configDir := filepath.Dir(configPath)
		exportDir := filepath.Join(configDir, "export")

		// Check if export directory exists
		if _, err := os.Stat(exportDir); os.IsNotExist(err) {
			continue
		}

		// Find all STL files in export directory
		var stlFiles []string
		err := filepath.Walk(exportDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".stl") {
				stlFiles = append(stlFiles, path)
			}
			return nil
		})
		if err != nil {
			continue
		}

		// Find old version folders
		if cleanOldVersions {
			err := filepath.Walk(exportDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() && path != exportDir {
					// Check if this is an old version folder (contains version pattern like v0.1, v1.0, etc.)
					dirName := filepath.Base(path)
					if matched, _ := regexp.MatchString(`^v\d+\.\d+`, dirName); matched {
						result := models.CleanResult{
							Path:               path,
							IsDeleted:          false,
							IsOldVersionFolder: true,
						}

						if !dryRun {
							err := os.RemoveAll(path)
							if err == nil {
								result.IsDeleted = true
							}
						}

						results = append(results, result)
					}
				}
				return nil
			})
			if err != nil {
				continue
			}
		}

		// Add STL files to results
		for _, stlFile := range stlFiles {
			result := models.CleanResult{
				Path:               stlFile,
				IsDeleted:          false,
				IsOldVersionFolder: false,
			}

			if !dryRun {
				err := os.Remove(stlFile)
				if err == nil {
					result.IsDeleted = true
				}
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func CleanConfig(dryRun bool, cleanOldVersions bool, folder string) (models.CleanResult, error) {
	// Find config.toml file in the folder
	configFiles, err := ScanFolderForConfigFiles(folder)
	if err != nil {
		return models.CleanResult{}, err
	}

	if len(configFiles) == 0 {
		return models.CleanResult{}, fmt.Errorf("no config.toml files found in folder")
	}

	// Use the first config file found
	configFile := configFiles[0]
	configPath := filepath.Join(folder, configFile.Path)
	configDir := filepath.Dir(configPath)
	exportDir := filepath.Join(configDir, "export")

	result := models.CleanResult{
		Path:               exportDir,
		IsDeleted:          false,
		IsOldVersionFolder: false,
	}

	// Check if export directory exists
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		return result, nil
	}

	// Find all STL files in export directory
	var stlFiles []string
	err = filepath.Walk(exportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".stl") {
			stlFiles = append(stlFiles, path)
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	// Find old version folders
	if cleanOldVersions {
		err = filepath.Walk(exportDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && path != exportDir {
				// Check if this is an old version folder (contains version pattern like v0.1, v1.0, etc.)
				dirName := filepath.Base(path)
				if matched, _ := regexp.MatchString(`^v\d+\.\d+`, dirName); matched {
					if !dryRun {
						err := os.RemoveAll(path)
						if err == nil {
							result.IsDeleted = true
						}
					}
				}
			}
			return nil
		})
		if err != nil {
			return result, err
		}
	}

	// Remove STL files
	for _, stlFile := range stlFiles {
		if !dryRun {
			err := os.Remove(stlFile)
			if err == nil {
				result.IsDeleted = true
			}
		}
	}

	return result, nil
}

var PRESET_images = []models.ExportCameraCoordinates{
	{
		CameraName:        "top",
		CameraCoordinates: "0,0,0,0,0,0,300",
	},
	{
		CameraName:        "down",
		CameraCoordinates: "0,0,0,0,0,0,300",
	},
	{
		CameraName:        "top-far",
		CameraCoordinates: "0,0,0,0,0,0,800",
	},
	{
		CameraName:        "bottom",
		CameraCoordinates: "0,0,0,180,0,0,300",
	},
	{
		CameraName:        "up",
		CameraCoordinates: "0,0,0,180,0,0,300",
	},
	{
		CameraName:        "bottom-far",
		CameraCoordinates: "0,0,0,180,0,0,800",
	},
	{
		CameraName:        "front",
		CameraCoordinates: "0,0,0,90,0,0,300",
	},
	{
		CameraName:        "front-far",
		CameraCoordinates: "0,0,0,90,0,0,800",
	},
	{
		CameraName:        "back",
		CameraCoordinates: "0,0,0,270,0,0,300",
	},
	{
		CameraName:        "back-far",
		CameraCoordinates: "0,0,0,270,0,0,800",
	},
	{
		CameraName:        "side", // (aka left)
		CameraCoordinates: "0,0,0,0,90,0,300",
	},
	{
		CameraName:        "side-near", // (aka left)
		CameraCoordinates: "0,0,0,0,90,0,150",
	},
	{
		CameraName:        "left",
		CameraCoordinates: "0,0,0,0,90,0,300",
	},
	{
		CameraName:        "left-far",
		CameraCoordinates: "0,0,0,0,90,0,800",
	},
	{
		CameraName:        "left-near",
		CameraCoordinates: "0,0,0,0,90,0,150",
	},
	{
		CameraName:        "right",
		CameraCoordinates: "0,0,0,0,0,0,300",
	},
	{
		CameraName:        "left-far",
		CameraCoordinates: "0,0,0,0,90,0,800",
	},
	{
		CameraName:        "right-far",
		CameraCoordinates: "0,0,0,0,0,0,800",
	},
	// "nice" - a diagonal downward looking view
	{
		CameraName:        "nice-far",
		CameraCoordinates: "0,0,0,45,0,45,800",
	},
	{
		CameraName:        "nice",
		CameraCoordinates: "0,0,0,45,0,45,350",
	},
	{
		CameraName:        "nice-near",
		CameraCoordinates: "0,0,0,45,0,45,150",
	},
	{
		CameraName:        "nice-far",
		CameraCoordinates: "0,0,0,45,0,45,800",
	},
	{
		CameraName:        "nice-100",
		CameraCoordinates: "0,0,0,45,0,45,100",
	},
	{
		CameraName:        "nice-200",
		CameraCoordinates: "0,0,0,45,0,45,200",
	},
	{
		CameraName:        "nice-300",
		CameraCoordinates: "0,0,0,45,0,45,300",
	},
	{
		CameraName:        "nice-400",
		CameraCoordinates: "0,0,0,45,0,45,400",
	},
	{
		CameraName:        "nice-500",
		CameraCoordinates: "0,0,0,45,0,45,500",
	},
	{
		CameraName:        "nice-600",
		CameraCoordinates: "0,0,0,45,0,45,600",
	},
	{
		CameraName:        "nice-700",
		CameraCoordinates: "0,0,0,45,0,45,700",
	},
	{
		CameraName:        "nice-800",
		CameraCoordinates: "0,0,0,45,0,45,800",
	},
	{
		CameraName:        "nice-900",
		CameraCoordinates: "0,0,0,45,0,45,900",
	},
	{
		CameraName:        "nice-1000",
		CameraCoordinates: "0,0,0,45,0,45,1000",
	},
}

func MakeExportImage(instance *models.InstanceConfig, camera models.ExportCameraCoordinates) models.ExportCameraCoordinates {
	return models.ExportCameraCoordinates{
		CameraName:         camera.CameraName,
		CameraCoordinates:  camera.CameraCoordinates,
		ImageSize:          camera.ImageSize,
		ParamFilter:        camera.ParamFilter,
		RunOutputImagePath: GetImagePath(instance.RunOutputImagePath, camera.CameraName),
	}
}

func interfaceToFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	default:
		return 0, false
	}
}

func interfaceToBool(v interface{}) (bool, bool) {
	switch b := v.(type) {
	case bool:
		return b, true
	default:
		return false, false
	}
}

func interfaceToStringList(v interface{}) []string {
	switch t := v.(type) {
	case string:
		parts := strings.Split(t, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			out = append(out, p)
		}
		if len(out) == 0 {
			return []string{strings.TrimSpace(t)}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, el := range t {
			out = append(out, interfaceToStringList(el)...)
		}
		return out
	default:
		return []string{strings.TrimSpace(fmt.Sprintf("%v", v))}
	}
}

func paramFilterValueMatches(filterVal interface{}, instanceVal interface{}) bool {
	if filterVal == nil {
		return true
	}
	// Strong-typed matches first (avoid "1" vs "1.0" issues).
	if fNum, ok := interfaceToFloat64(filterVal); ok {
		if iNum, ok2 := interfaceToFloat64(instanceVal); ok2 {
			return fNum == iNum
		}
	}
	if fBool, ok := interfaceToBool(filterVal); ok {
		if iBool, ok2 := interfaceToBool(instanceVal); ok2 {
			return fBool == iBool
		}
	}

	// Fallback: string/list matching (supports comma-delimited strings).
	filterVals := interfaceToStringList(filterVal)
	instanceVals := interfaceToStringList(instanceVal)
	for _, fv := range filterVals {
		for _, iv := range instanceVals {
			if fv == iv {
				return true
			}
		}
	}
	return false
}

func paramFilterMatchesAll(filter map[string]interface{}, instanceParams map[string]interface{}) (bool, string) {
	if filter == nil {
		return true, ""
	}
	for k, v := range filter {
		instanceVal, ok := instanceParams[k]
		if !ok {
			return false, fmt.Sprintf("Skipping image because param %s is not set", k)
		}
		if !paramFilterValueMatches(v, instanceVal) {
			return false, fmt.Sprintf("Skipping image because param %s does not match %v", k, v)
		}
	}
	return true, ""
}

func populateExportImages(config *models.Config, instances []models.InstanceConfig) ([]models.InstanceConfig, error) {
	if config.Debug {
		log.Printf("populateExportImages")
	}
	if config.OnlyExport {
		return instances, nil
	}

	for i := range instances {
		allExportImages := []models.ExportCameraCoordinates{}
		skippedReasons := []string{}

		if instances[i].SkipImages {
			if config.Debug {
				logSkip(fmt.Sprintf("Skipping images for instance %s", instances[i].AutoName))
			}
			instances[i].SkippedImageReason = "Skipping images because of skip_images flag"
			continue
		}

		for _, exportImage := range config.Design.ExportImages {
			if ok, reason := paramFilterMatchesAll(exportImage.ParamFilter, instances[i].Params); !ok {
				if config.Debug {
					logSkip(fmt.Sprintf("Skipping image for instance %s: %s", instances[i].AutoName, reason))
				}
				skippedReasons = append(skippedReasons, fmt.Sprintf("Skipping image for instance %s: %s", instances[i].AutoName, reason))
				continue
			}

			exportImages := makePresetReplacement(instances[i].RunOutputPathV3, exportImage)
			if len(exportImages) > 0 {
				allExportImages = append(allExportImages, exportImages...)
			}
		}

		// Then add instance-specific export images if they exist
		for _, configuredInstance := range config.Design.ConfiguredInstanceConfig {
			if configuredInstance.Name == instances[i].Name && len(configuredInstance.ExportImages) > 0 {
				allExportImages = append(allExportImages, configuredInstance.ExportImages...)
			}
		}

		instances[i].ExportImages = allExportImages
		if len(allExportImages) == 0 && instances[i].SkippedImageReason == "" && len(skippedReasons) > 0 {
			// Only show a skip reason when we actually ended up with zero images.
			instances[i].SkippedImageReason = skippedReasons[0]
		}
	}

	return instances, nil
}

// CameraPreset defines the base camera settings for different views
type CameraPreset struct {
	Direction string
	RotX      float64
	RotY      float64
	RotZ      float64
}

// CameraDistance defines the distance presets
type CameraDistance struct {
	Name     string
	Distance float64
}

// Camera presets for different directions
var cameraPresets = map[string]CameraPreset{
	"top":    {Direction: "top", RotX: 0, RotY: 0, RotZ: 0},
	"bottom": {Direction: "bottom", RotX: 180, RotY: 0, RotZ: 0},
	"front":  {Direction: "front", RotX: 90, RotY: 0, RotZ: 0},
	"back":   {Direction: "back", RotX: 270, RotY: 0, RotZ: 0},
	"left":   {Direction: "left", RotX: 0, RotY: 90, RotZ: 0},
	"right":  {Direction: "right", RotX: 0, RotY: 270, RotZ: 0},
}

// Camera distance presets
var cameraDistances = map[string]CameraDistance{
	"near": {Name: "near", Distance: 150},
	"":     {Name: "default", Distance: 300},
	"far":  {Name: "far", Distance: 800},
}

// parseCameraName parses a camera name to extract direction and distance
func parseCameraName(cameraName string) (string, string) {
	parts := strings.Split(cameraName, "-")

	if len(parts) == 1 {
		log.Printf("parseCameraName 1: %s", parts[0])
		return parts[0], ""
	}

	if len(parts) == 2 {
		log.Printf("parseCameraName 2: %s", parts[0])
		// Check if the second part is a distance keyword
		if _, ok := cameraDistances[parts[1]]; ok {
			return parts[0], parts[1]
		}
	}

	LogWarn(fmt.Sprintf("Camera name '%s' is not a preset and has no coordinates specified", cameraName), false)

	// If we can't parse it properly, return the original name and empty distance
	return cameraName, ""
}

// generateCameraCoordinates creates camera coordinates based on direction and distance
func generateCameraCoordinates(direction, distanceKey string) string {
	// Default values
	x, y, z := 0.0, 0.0, 0.0
	rotX, rotY, rotZ := 0.0, 0.0, 0.0
	camDistance := 300.0 // Default distance

	// Get direction preset
	if preset, ok := cameraPresets[direction]; ok {
		rotX, rotY, rotZ = preset.RotX, preset.RotY, preset.RotZ
	}

	// Get distance preset
	if dist, ok := cameraDistances[distanceKey]; ok {
		camDistance = dist.Distance
	}

	// Format: x,y,z,rotx,roty,rotz,distance
	return fmt.Sprintf("%.0f,%.0f,%.0f,%.0f,%.0f,%.0f,%.0f",
		x, y, z, rotX, rotY, rotZ, camDistance)
}

func makePresetReplacement(runOutputPath string, exportImage models.ExportCameraCoordinates) []models.ExportCameraCoordinates {
	if exportImage.CameraName == "all" {
		return PRESET_images
	} else if strings.HasPrefix(exportImage.CameraName, "all") && len(strings.Split(exportImage.CameraName, "-")) == 2 {
		suffix := strings.Split(exportImage.CameraName, "-")[1]
		nearPresetImages := make([]models.ExportCameraCoordinates, 0)
		for _, cm := range PRESET_images {
			if strings.HasSuffix(cm.CameraName, suffix) {
				nearPresetImages = append(nearPresetImages, cm)
			}
		}
		return nearPresetImages
	}
	// If custom coordinates are provided, use them directly
	if exportImage.CameraCoordinates != "" {
		return []models.ExportCameraCoordinates{
			{
				CameraName:         exportImage.CameraName,
				CameraCoordinates:  exportImage.CameraCoordinates,
				ImageSize:          exportImage.ImageSize,
				ParamFilter:        exportImage.ParamFilter,
				RunOutputImagePath: GetImagePath(runOutputPath, exportImage.CameraName),
			},
		}
	}

	// Parse the camera name to extract direction and distance
	direction, distance := parseCameraName(exportImage.CameraName)

	// Check if we have a valid direction
	if _, ok := cameraPresets[direction]; ok {
		// Generate coordinates based on the parsed direction and distance
		coordinates := generateCameraCoordinates(direction, distance)

		// Create a new camera preset with the generated coordinates
		return []models.ExportCameraCoordinates{
			{
				CameraName:         exportImage.CameraName,
				CameraCoordinates:  coordinates,
				ImageSize:          exportImage.ImageSize,
				ParamFilter:        exportImage.ParamFilter,
				RunOutputImagePath: GetImagePath(runOutputPath, exportImage.CameraName),
			},
		}
	}

	// Look for matching preset camera in the static list
	for _, preset := range PRESET_images {
		if preset.CameraName == exportImage.CameraName {
			return []models.ExportCameraCoordinates{
				{
					CameraName:         exportImage.CameraName,
					CameraCoordinates:  preset.CameraCoordinates,
					ImageSize:          preset.ImageSize,
					ParamFilter:        exportImage.ParamFilter,
					RunOutputImagePath: GetImagePath(runOutputPath, exportImage.CameraName),
				},
			}
		}
	}

	/*log.Printf("Preset Export Camera Names:")
	for _, preset := range PRESET_images {
		log.Printf(preset.CameraName)
	}*/

	// If no preset found and no coordinates provided, log an error
	/*log.Panicf(`Camera '%s' is not a preset and has no coordinates specified.

	Options are listed above
	`, exportImage.CameraName)*/

	return []models.ExportCameraCoordinates{}
}

/*
func getPresetExportImages(config *models.Config) []models.ExportCameraCoordinates {
	allExportImages := []models.ExportCameraCoordinates{}

	// First, handle any "all" camera requests
	for _, exportImage := range config.Design.ExportImages {
		if exportImage.CameraName == "all" {
			// If "all" is specified, add all preset cameras
			allExportImages = append(allExportImages, PRESET_images...)
			break
		}
	}

	// Then process global export images
	for _, exportImage := range config.Design.ExportImages {
		// Skip "all" as it's already handled
		if exportImage.CameraName == "all" {
			continue
		}

		if exportImage.ParamFilter != nil {
			var match bool
			for _, param := range exportImage.ParamFilter {

			}
		}

		// If custom coordinates are provided, use them directly
		if exportImage.CameraCoordinates != "" {
			allExportImages = append(allExportImages, exportImage)
			continue
		}

		// Parse the camera name to extract direction and distance
		direction, distance := parseCameraName(exportImage.CameraName)

		// Check if we have a valid direction
		if _, ok := cameraPresets[direction]; ok {
			// Generate coordinates based on the parsed direction and distance
			coordinates := generateCameraCoordinates(direction, distance)

			// Create a new camera preset with the generated coordinates
			allExportImages = append(allExportImages, models.ExportCameraCoordinates{
				CameraName:        exportImage.CameraName,
				CameraCoordinates: coordinates,
				ImageSize:         exportImage.ImageSize,
				ParamFilter:       exportImage.ParamFilter,
			})
			continue
		}

		// Look for matching preset camera in the static list
		found := false
		for _, preset := range PRESET_images {
			if preset.CameraName == exportImage.CameraName {
				allExportImages = append(allExportImages, preset)
				found = true
				break
			}
		}

		// If no preset found and no coordinates provided, log an error
		if !found {
			log.Panicf("Camera '%s' is not a preset and has no coordinates specified", exportImage.CameraName)
		}

		// instance configured export images
		for _, instance := range config.Design.ConfiguredInstanceConfig {
			if instance.ExportImages != nil && len(instance.ExportImages) > 0 {
				allExportImages = append(allExportImages, instance.ExportImages...)
			}
		}
	}

	return allExportImages
}*/

func LogInit() {
	log.Printf(colorGreen + `

   ___                                     _                  
  / _ \ _ __   ___ _ __  ___  ___ __ _  __| | __ _  ___ _ __  
 | | | | '_ \ / _ \ '_ \/ __|/ __/ _` + "`" + ` |/ _` + "`" + ` |/ _` + "`" + ` |/ _ \ '_ \ 
 | |_| | |_) |  __/ | | \__ \ (_| (_| | (_| | (_| |  __/ | | |
  \___/| .__/ \___|_| |_|___/\___\__,_|\__,_|\__, |\___|_| |_|
       |_|                                   |___/            

Welcome to openscadgen!

This software is under active development - feedback welcome at https://github.com/kiwikid/openscadgen/issues

` + colorReset)
}

// LoadConfigFromFile reads the configuration file and populates the Config struct.
// Return order is (config, warning, err). If err != nil, config is always nil.
// warning is non-fatal (e.g. unrecognised export image camera) when err is nil.
func LoadConfigFromFile(flags models.CmdFlags) (*models.Config, error, error) {
	if flags.ConfigFile == "" {
		log.Printf(colorRed + "Run directly with a config file use '-c' like '-c you-project/config.toml'\n\nUse -s to run in Server mode\n\n Or Server Folder Scan mode:  -sf parent-project-folder " + colorReset)
		return nil, nil, fmt.Errorf("no config file provided")
	}

	// Resolve config file path relative to current working directory
	configPath := flags.ConfigFile
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			log.Printf(colorRed+"Failed to resolve config file path '%s': %v", configPath, err)
			return nil, nil, err
		}
		configPath = absPath
	}
	if flags.Debug {
		log.Printf("ReadFile config for %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Try fallback to serverFolder + configFile if serverFolder is set
		if flags.ServerFolder != "" {
			fallbackPath := filepath.Join(flags.ServerFolder, flags.ConfigFile)
			if flags.Debug {
				log.Printf("Trying fallback config path: %s", fallbackPath)
			}
			fallbackData, fallbackErr := os.ReadFile(fallbackPath)
			if fallbackErr == nil {
				configPath = fallbackPath
				data = fallbackData
				err = nil
				if flags.Debug {
					log.Printf("Successfully read config from fallback path: %s", fallbackPath)
				}
			} else {
				log.Printf(colorRed+"Failed to read config file at path '%s'\n\n Error: %v\n\nTried fallback path '%s' but also failed: %v"+colorReset, configPath, err, fallbackPath, fallbackErr)
				return nil, nil, err
			}
		} else {
			log.Printf(colorRed+"Failed to read config file at path '%s'\n\n Error: %v"+colorReset, configPath, err)
			return nil, nil, err
		}
	} else if flags.Debug {
		log.Printf("ReadFile config for %s", configPath)
	}

	// Use the new LoadConfig method with the file content
	return LoadConfig(string(data), flags, configPath)
}

// LoadConfig parses and validates a configuration string and populates the Config struct.
// Return order is (config, warning, err). If err != nil, config is always nil.
func LoadConfig(configData string, flags models.CmdFlags, configPath string) (*models.Config, error, error) {
	var conf models.Config

	// First decode into a map to check for unmapped fields
	var metadata toml.MetaData
	metadata, err := toml.Decode(configData, &conf)
	if err != nil {
		log.Printf(colorRed+"Config file is not valid toml [%s]: %v"+colorReset, configPath, err)
		if snip := tomlDecodeErrorSnippet(configData, err); snip != "" {
			log.Printf(colorRed + "Source context near TOML error:\n" + snip + colorReset)
		}
		return nil, nil, fmt.Errorf("config file is not valid toml [%s]: %w", configPath, err)
	}

	// Check for undecoded keys
	undecoded := metadata.Undecoded()
	if len(undecoded) > 0 {
		LogKeyValuePair("Config file", flags.ConfigFile)
		for _, key := range undecoded {
			logError(fmt.Sprintf("Invalid field: %s", key.String()))
		}
		if flags.StopOnError {
			log.Printf(colorYellow + "Continuing on error" + colorReset)
		} else {
			return nil, nil, fmt.Errorf("invalid fields in config file [%s]: %v", configPath, undecoded)
		}
	}

	// Validate the config
	validate := validator.New()
	err = validate.Struct(conf)
	if err != nil {
		log.Printf(colorRed+"Failed to validate config [%s]: %v"+colorReset, configPath, err)
		return nil, nil, fmt.Errorf("failed to validate config file [%s]: %w", configPath, err)
	}

	var warning error
	for _, exportImage := range conf.Design.ExportImages {
		cameraName, distance := parseCameraName(exportImage.CameraName)
		if distance == "" {
			// Check if it's a known preset camera name
			isPreset := false
			for _, preset := range PRESET_images {
				if preset.CameraName == cameraName {
					isPreset = true
					break
				}
			}
			// Also check if it's in the cameraPresets map
			if _, ok := cameraPresets[cameraName]; ok {
				isPreset = true
			}
			if !isPreset {
				log.Printf(colorRed+"Failed to parse camera name: %v"+colorReset, cameraName)
				warning = fmt.Errorf("failed to parse camera name: %v", cameraName)
			}
		}
	}

	if flags.Debug {
		log.Printf("Loaded config")
		LogKeyValuePair("Config", conf.ConfigFile)
	}

	conf.RawConfigFile = configData

	// Merge command-line flags into the config
	conf.Quiet = flags.Quiet
	conf.Debug = flags.Debug
	conf.NoProcessing = flags.NoProcessing
	conf.RegexPattern = flags.RegexPattern
	conf.MaxInstances = flags.MaxInstances
	conf.SkipRender = flags.SkipRender
	conf.SkipReadme = flags.SkipReadme
	conf.OverwriteExisting = flags.OverwriteExisting
	conf.CustomOpenSCADCommand = flags.CustomOpenSCADCommand
	conf.MaxConcurrentRequests = flags.MaxConcurrentRequests
	conf.IncludePartIDLetter = flags.IncludePartIDLetter
	conf.SetBuildInfoInFileAttributes = flags.SetBuildInfoInFileAttributes

	conf.StopOnError = flags.StopOnError
	conf.ConfigFile = configPath
	conf.IncludeExportLog = flags.IncludeExportLog
	conf.OverwriteExisting = flags.OverwriteExisting
	conf.Server = flags.Server
	conf.ServerFolder = flags.ServerFolder
	conf.ServerModeConfigFile = flags.ServerModeConfigFile
	conf.OnlyImages = flags.OnlyImages
	conf.OnlyExport = flags.OnlyExport
	conf.Debug = flags.Debug
	conf.Quiet = flags.Quiet
	conf.NoProcessing = flags.NoProcessing
	conf.RegexPattern = flags.RegexPattern
	conf.MaxInstances = flags.MaxInstances
	conf.SkipRender = flags.SkipRender
	conf.SkipReadme = flags.SkipReadme
	conf.OverwriteExisting = flags.OverwriteExisting
	conf.CustomOpenSCADCommand = flags.CustomOpenSCADCommand

	conf.TotalQueuedInstances = 0
	if flags.CustomOpenSCADOutputFormat != "" {
		conf.Design.CustomOpenSCADOutputFormat = flags.CustomOpenSCADOutputFormat
	}

	conf.Debug = flags.Debug

	conf.Quality = ""
	if flags.OverrideFN > 0 {
		conf.OverrideFN = flags.OverrideFN
		conf.Quality = fmt.Sprintf("fn-%d", flags.OverrideFN)
	} else if flags.HighQuality {
		conf.OverrideFN = 200
		conf.Quality = "high"
	} else if flags.LowQuality {
		conf.OverrideFN = 20
		conf.Quality = "low"
	}

	if conf.Design.Version == "" {
		conf.Design.Version = "v0.1"
	}

	if conf.Design.RunType == "" {
		conf.Design.RunType = "clearAndCreate"
	}

	exportNameFormat := getExportNameFormat(&conf)

	exportNameFormatParams := getExportNameFormatParams(exportNameFormat)

	if flags.Debug {
		logStage("DEBUG Validating: export_name_format params")
	}
	for _, paramName := range exportNameFormatParams {
		if flags.Debug {
			LogKeyValuePair("Param name to confirm", paramName)
			LogKeyValuePair("ExportNameFormat", exportNameFormat)
		}
		if !strings.Contains(conf.Design.ExportNameFormat, paramName) {
			LogWarn(fmt.Sprintf("ExportNameFormat contains param: \n\n -\t(%s)\n\n that is not in the params. Include every param in the export_name_format (in the format '{param_name}') to ensure all instances are generated to unique files.", paramName), true)
		}
	}

	openSCADVersion, err := findOpenSCAD()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find OpenSCAD version: %w", err)
	}
	conf.OpenSCADVersion = openSCADVersion.Version
	conf.OpenScadGenVersion = VERSION
	/* (temp disabled instance validation for now)
		if conf.Design.ConfiguredInstanceConfig != nil {
			if len(conf.Design.ConfiguredInstanceConfig) > 0 {
				for dynamicInstanceIndex, dynamicInstance := range conf.Design.ConfiguredInstanceConfig {
					for paramName, paramValue := range dynamicInstance.Params {
						if config.Debug {
							LogKeyValuePair("LoadConfig:Param name", paramName)
							LogKeyValuePair("LoadConfig:Param value", fmt.Sprintf("%v", paramValue))
						}

						paramHasMoreThanOneValue := false
						if reflect.TypeOf(paramValue).Kind() == reflect.Slice {
							paramHasMoreThanOneValue = true
						} else {
							paramHasMoreThanOneValue = strings.Contains(fmt.Sprintf("%v", paramValue), ",") || strings.Contains(fmt.Sprintf("%v", paramValue), "-")
						}

						if !paramHasMoreThanOneValue {
							continue
						}

						if len(conf.Design.InputPaths) > 1 {
							nameHasDesignFileName := strings.Contains(exportNameFormat, "{designFileName}")
							if !nameHasDesignFileName {
								LogWarn("If more than one input is specified, the export_name_format need to include designFileName (add {designFileName} to the export_name_format)", true)
								LogKeyValuePair("ExportNameFormat missing {designFileName}", exportNameFormat)
								LogKeyValuePair("from config file:", flags.ConfigFile)
								os.Exit(1)
							}
						}

						nameIsNumberated := false
						for _, key := range strings.Split(dynamicInstance.ParamNumberationKeys, ",") {
							if key == paramName {
								nameIsNumberated = true
								break
							}
						}

						nameHasParams := strings.Contains(exportNameFormat, fmt.Sprintf("{%s}", paramName))
						if !nameHasParams && paramHasMoreThanOneValue && !nameIsNumberated {
							LogKeyValuePair("Dynamic instance index 1", fmt.Sprintf("%d", dynamicInstanceIndex))
							LogKeyValuePair("Export Name Format", exportNameFormat)
							LogKeyValuePair("Missing Param name", paramName)
							LogKeyValuePair("Param value", fmt.Sprintf("%v", paramValue))
							LogKeyValuePair("Config file", flags.ConfigFile)
							LogWarn(fmt.Sprintf(`Export instance name:
							   %s

							   does not contain param:
							    {%s}

							   Include every param in the export_name_format (in the format '{param_name}') to ensure all instances are generated to unique files.`, dynamicInstance.Name, paramName), true)
							os.Exit(1)
						}
					}
				}
			}
		}

	// confirm all params in the export_name_format are in the params
	for _, paramName := range strings.Split(exportNameFormat, "{") {
		if flags.Debug {
			LogKeyValuePair("Param name to confirm", paramName)
			LogKeyValuePair("ExportNameFormat", conf.Design.ExportNameFormat)
		}
		name := strings.Split(paramName, "}")[0]
		if !strings.Contains(conf.Design.ExportNameFormat, name) {
			LogWarn(fmt.Sprintf("ExportNameFormat contains param (%s) that is not in the params", name), true)
		}
	}

	// Add validation for export_name_format
	if err := validateExportNameFormat(&conf); err != nil {
		return nil, fmt.Errorf("export name format validation failed: %w", err)
	}	*/

	return &conf, warning, nil
}

func ProcessFolder(folder string, cmdFlags models.CmdFlags) ([]models.ProcessResult, error) {
	configs, err := ScanFolderForConfigFiles(folder)
	if err != nil {
		return []models.ProcessResult{}, fmt.Errorf("failed to scan folder for config files: %w", err)
	}

	processResults := []models.ProcessResult{}
	for _, config := range configs {
		cmdFlags.ConfigFile = config.Path
		config, _, err := LoadConfigFromFile(cmdFlags)
		if err != nil {
			return []models.ProcessResult{}, fmt.Errorf("failed to load config: %w", err)
		}
		processResult, err := Process(config, &NoopProgress{}, nil, Operations{
			GenerateReport: true,
		}, cmdFlags.Server)
		if err != nil {
			return []models.ProcessResult{}, fmt.Errorf("failed to process config: %w", err)
		}
		processResults = append(processResults, processResult)
	}

	return processResults, nil
}

func validateExportNameFormat(config *models.Config) error {
	// Get all parameters that have multiple values
	multiValueParams := make(map[string]bool)

	// Check dynamic instances for multi-value parameters
	for _, instance := range config.Design.ConfiguredInstanceConfig {
		for paramName, paramValue := range instance.Params {
			if reflect.TypeOf(paramValue).Kind() == reflect.Slice {
				multiValueParams[paramName] = true
			} else if strings.Contains(fmt.Sprintf("%v", paramValue), ",") {
				multiValueParams[paramName] = true
			}
		}
	}

	// Get the export name format to validate
	exportNameFormat := config.Design.ExportNameFormat

	// Check each input path's export name format if specified
	for _, inputPath := range config.Design.InputPaths {
		if inputPath.ExportNameFormat != "" {
			exportNameFormat = inputPath.ExportNameFormat
		}

		// Validate that all multi-value parameters are included in the export name format
		for paramName := range multiValueParams {
			if strings.Contains(inputPath.IgnoreParamsWhenProcessing, paramName) {
				continue
			}
			if !strings.Contains(exportNameFormat, fmt.Sprintf("{%s}", paramName)) {
				return fmt.Errorf("parameter '%s' has multiple values but is not included in export_name_format '%s'. "+
					"This would cause files to overwrite each other", paramName, exportNameFormat)
			}
		}
	}

	return nil
}

func generateLowQualityWarningFile(config *models.Config, outputPath string) {

	if config.OverrideFN == 0 {
		return
	}

	if config.OverrideFN > 100 {
		return
	}

	if !config.Quiet {
		logStage("Generating LOW_QUALITY_WARNING.md")
	}

	contents := fmt.Sprintf("[WARNING: This model was generated with a low quality (fn = %d)]", config.OverrideFN)

	lowQualityWarningFile, err := os.Create(outputPath)
	if err != nil {
		log.Panicf(colorRed+"Failed to create LOW_QUALITY_WARNING.md file: %s", err)
	}
	defer lowQualityWarningFile.Close()

	_, err = lowQualityWarningFile.WriteString(contents)
	if err != nil {
		log.Panicf(colorRed+"Failed to write to LOW_QUALITY_WARNING.md file: %s", err)
	} else if !config.Quiet {
		LogKeyValuePair("LOW_QUALITY_WARNING.md written to", outputPath)
	}

}

func generateReadme(config *models.Config, dynamicInstances []*models.InstanceConfig, version string, openscadVersion string, readmePath string) {
	if config.SkipReadme {
		log.Printf(colorYellow + "Skipping readme generation" + colorReset)
		return
	}

	if !config.Quiet {
		logStage("Generating README.md")
	}

	contents := fmt.Sprintf("# %s\n\n%s\n\n", config.Design.Name, config.Design.Description)
	contents += "## Contents \n"

	for _, instance := range dynamicInstances {
		paths := instance.GetInstancePaths(config)
		contents += fmt.Sprintf("- [%s](.%s)\n", paths.OutputPath, strings.ToLower(strings.ReplaceAll(paths.OutputPath, " ", "-")))

		// Use relative path for input path in README
		relInputPath := paths.InputPathRelative
		if relInputPath == "" {
			relInputPath = filepath.Base(paths.InputPath)
		}
		contents += fmt.Sprintf("\t- **%s**: %v\n", "InputPath", relInputPath)

		for name, value := range instance.Params {
			contents += fmt.Sprintf("\t- **%s**: %v\n", name, value)
		}
		contents += "\n\n"
	}

	// Optionally add a footer or additional information
	contents += "## Additional Information\n"
	contents += fmt.Sprintf("This README was generated by [openscadgen](https://github.com/kiwikid/openscadgen) %s %s. The free, local, open source openscad stl release generator.\n", version, openscadVersion)

	//readmePath := path.Join(config.Design.OutputPath, config.Design.Version, "README.md")

	if config.NoProcessing {
		log.Printf(colorBlue + "(as requested) README.md was not generated" + colorReset)
		return
	}

	readmeFile, err := os.Create(readmePath)
	if err != nil {
		log.Panicf(colorRed+"Failed to create README.md file: %s", err)
	} else {
		if !config.Quiet {
			LogKeyValuePair("README.md written to", readmePath)
		}
	}
	defer readmeFile.Close()

	_, err = readmeFile.WriteString(contents)
	if err != nil {
		log.Panicf(colorRed+"Failed to write to README.md file: %s", err)
	}

}

func getExportNameFormat(config *models.Config) string {

	exportNameFormat := config.Design.ExportNameFormat
	if exportNameFormat == "" {
		log.Printf("Export name format is not set, defaulting to '{instanceName}'")
		exportNameFormat = "{instanceName}"
	}

	return exportNameFormat
}

func getExportNameFormatParams(exportNameFormat string) []string {
	var params []string
	parts := strings.Split(exportNameFormat, "{")

	// Skip first part before any {
	for i := 1; i < len(parts); i++ {
		// Split on } to get just the param name
		paramPart := strings.Split(parts[i], "}")
		if len(paramPart) > 0 {
			params = append(params, paramPart[0])
		}
	}

	return params
}

// Helper function to parse a single parameter value
func parseParamValue(value interface{}) ([]interface{}, error) {
	var parsedValues []interface{}

	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case string:
		// Handle string values that contain ranges or comma-separated values
		values := strings.Split(v, ",")
		for _, val := range values {
			val = strings.TrimSpace(val)
			if val == "true" || val == "false" {
				parsedValues = append(parsedValues, val == "true")
			} else if num, err := strconv.ParseFloat(val, 64); err == nil {
				parsedValues = append(parsedValues, num)
			} else {
				parsedValues = append(parsedValues, val)
			}
		}
		if len(parsedValues) == 0 {
			parsedValues = append(parsedValues, value)
		}
	default:
		// For non-slice, non-string values, try to convert to float64 if numeric
		if num, ok := v.(int); ok {
			parsedValues = append(parsedValues, float64(num))
		} else if num, ok := v.(float64); ok {
			parsedValues = append(parsedValues, num)
		} else if b, ok := v.(bool); ok {
			parsedValues = append(parsedValues, b)
		} else {
			parsedValues = append(parsedValues, value)
		}
	}

	return parsedValues, nil
}

// Helper function to parse direct array parameters
func parseDirectArrayParam(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		// Already parsed as array from TOML
		return v, nil
	case string:
		// Parse string representation of array (JSON-like format)
		// Remove outer brackets and split by array elements
		trimmed := strings.TrimSpace(v)
		if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
			return nil, fmt.Errorf("invalid array format: %s", v)
		}

		// Remove outer brackets
		inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		if inner == "" {
			return []interface{}{}, nil
		}

		// Parse nested arrays by finding matching brackets
		var result []interface{}
		var current strings.Builder
		var bracketCount int
		var inString bool
		var escapeNext bool

		for _, char := range inner {
			if escapeNext {
				current.WriteRune(char)
				escapeNext = false
				continue
			}

			if char == '\\' {
				escapeNext = true
				current.WriteRune(char)
				continue
			}

			if char == '"' && !escapeNext {
				inString = !inString
			}

			if !inString {
				if char == '[' {
					bracketCount++
				} else if char == ']' {
					bracketCount--
				} else if char == ',' && bracketCount == 0 {
					// Found array element boundary
					elementStr := strings.TrimSpace(current.String())
					if elementStr != "" {
						parsed, err := parseArrayElement(elementStr)
						if err != nil {
							return nil, fmt.Errorf("error parsing array element '%s': %w", elementStr, err)
						}
						result = append(result, parsed)
					}
					current.Reset()
					continue
				}
			}

			current.WriteRune(char)
		}

		// Handle the last element
		elementStr := strings.TrimSpace(current.String())
		if elementStr != "" {
			parsed, err := parseArrayElement(elementStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing array element '%s': %w", elementStr, err)
			}
			result = append(result, parsed)
		}

		return result, nil
	default:
		// For other types, wrap in array
		return []interface{}{v}, nil
	}
}

// Helper function to parse individual array elements
func parseArrayElement(elementStr string) (interface{}, error) {
	elementStr = strings.TrimSpace(elementStr)

	// Check if it's a nested array
	if strings.HasPrefix(elementStr, "[") && strings.HasSuffix(elementStr, "]") {
		// Recursively parse nested array
		return parseDirectArrayParam(elementStr)
	}

	// Try to parse as number
	if num, err := strconv.ParseFloat(elementStr, 64); err == nil {
		return num, nil
	}

	// Try to parse as boolean
	if elementStr == "true" {
		return true, nil
	}
	if elementStr == "false" {
		return false, nil
	}

	// Return as string (remove quotes if present)
	if strings.HasPrefix(elementStr, "\"") && strings.HasSuffix(elementStr, "\"") && len(elementStr) >= 2 {
		return elementStr[1 : len(elementStr)-1], nil
	}

	return elementStr, nil
}

// Helper function to convert a map of parameters to a map of parameter combinations
func convertToParamCombinations(params map[string]interface{}, ignoredParams map[string]bool, directArrayParams []string) (map[string][]interface{}, error) {
	result := make(map[string][]interface{})
	for k, v := range params {
		// Skip ignored parameters
		if ignoredParams[k] {
			continue
		}
		// Skip direct array parameters - they should not be processed as combinations
		if slices.Contains(directArrayParams, k) {
			continue
		}
		parsedValues, err := parseParamValue(v)
		if err != nil {
			return nil, fmt.Errorf("error parsing parameter %s: %w", k, err)
		}
		result[k] = parsedValues
	}
	return result, nil
}

// Helper function to check if all numbers in a slice are whole numbers
func areAllWholeNumbers(values []float64) bool {
	for _, v := range values {
		if v != float64(int(v)) {
			return false
		}
	}
	return true
}

func getAllParams(dynamicInstance models.ConfiguredInstanceConfig, globalParams map[string]interface{}, paramSets []models.ParamSet, inputPath models.InputPath) (map[string]interface{}, map[string][]interface{}, []string) {
	params := make(map[string]interface{})
	globalParamsMap := make(map[string][]interface{})
	var ignoredKeys []string

	// Parse ignored parameters
	if inputPath.IgnoreParamsWhenProcessing != "" {
		ignoredKeys = strings.Split(inputPath.IgnoreParamsWhenProcessing, ",")
		for i, key := range ignoredKeys {
			ignoredKeys[i] = strings.TrimSpace(key)
		}
	}

	for _, paramSet := range paramSets {
		if slices.Contains(strings.Split(inputPath.ParamSets, ","), paramSet.Name) {
			for k, v := range paramSet.Params {
				if slices.Contains(ignoredKeys, k) {
					continue
				}
				params[k] = v
			}
		}

		if slices.Contains(strings.Split(dynamicInstance.ParamSets, ","), paramSet.Name) {
			for k, v := range paramSet.Params {
				if slices.Contains(ignoredKeys, k) {
					continue
				}
				params[k] = v
			}
		}

	}

	// Process global parameters
	for key, value := range globalParams {
		// Skip if this parameter should be ignored
		shouldSkip := false
		for _, ignoredKey := range ignoredKeys {
			if key == ignoredKey {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		shouldNotSplitOnComma := false
		if slices.Contains(dynamicInstance.IgnoreCommaInParams, key) {
			shouldNotSplitOnComma = true
		}

		if strValue, ok := value.(string); ok && strings.Contains(strValue, ",") && !shouldNotSplitOnComma {
			values := strings.Split(strValue, ",")
			var parsedValues []interface{}
			for _, val := range values {
				val = strings.TrimSpace(val)
				if num, err := strconv.ParseFloat(val, 64); err == nil {
					parsedValues = append(parsedValues, num)
				} else if val == "true" || val == "false" {
					parsedValues = append(parsedValues, val == "true")
				} else {
					parsedValues = append(parsedValues, val)
				}
			}
			globalParamsMap[key] = parsedValues
		} else {
			switch v := value.(type) {
			case int:
				globalParamsMap[key] = []interface{}{float64(v)}
			case float64:
				globalParamsMap[key] = []interface{}{v}
			case bool:
				globalParamsMap[key] = []interface{}{v}
			case []string:
				sl := make([]interface{}, len(v))
				for i, s := range v {
					sl[i] = s
				}
				globalParamsMap[key] = sl
			case []interface{}:
				out := make([]interface{}, 0, len(v))
				for _, el := range v {
					switch e := el.(type) {
					case int:
						out = append(out, float64(e))
					case float64:
						out = append(out, e)
					case bool:
						out = append(out, e)
					case string:
						if num, err := strconv.ParseFloat(e, 64); err == nil {
							out = append(out, num)
						} else if e == "true" || e == "false" {
							out = append(out, e == "true")
						} else {
							out = append(out, e)
						}
					default:
						out = append(out, el)
					}
				}
				globalParamsMap[key] = out
			case string:
				if num, err := strconv.ParseFloat(v, 64); err == nil {
					globalParamsMap[key] = []interface{}{num}
				} else if v == "true" || v == "false" {
					globalParamsMap[key] = []interface{}{v == "true"}
				} else {
					globalParamsMap[key] = []interface{}{v}
				}
			default:
				globalParamsMap[key] = []interface{}{v}
			}
		}
	}

	paramSetsKeys := append(strings.Split(dynamicInstance.ParamSets, ","), strings.Split(inputPath.ParamSets, ",")...)

	for _, paramSet := range paramSets {
		if slices.Contains(paramSetsKeys, paramSet.Name) {
			for k, v := range paramSet.Params {
				shouldSkip := false
				for _, ignoredKey := range ignoredKeys {
					if k == ignoredKey {
						shouldSkip = true
						break
					}
				}
				if shouldSkip {
					continue
				}
				params[k] = v
			}
		}
	}
	/*
		// Process parameters from ParamSets
		if len(paramSets) > 0 {
			for _, paramSet := range paramSets {
				for k, v := range paramSet.Params {
					// Skip if this parameter should be ignored

					if shouldSkip {
						continue
					}
					params[k] = v
				}
			}
		}
	*/
	// Process instance-specific parameters
	for k, v := range dynamicInstance.Params {
		// Skip if this parameter should be ignored
		shouldSkip := false
		for _, ignoredKey := range ignoredKeys {
			if k == ignoredKey {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		// Check if this parameter should be treated as a direct array
		shouldUseDirectArray := false
		if slices.Contains(dynamicInstance.DirectArrayParams, k) {
			shouldUseDirectArray = true
		}

		if shouldUseDirectArray {
			// Parse as direct array without comma splitting
			parsedArray, err := parseDirectArrayParam(v)
			if err != nil {
				// If parsing fails, fall back to treating as regular parameter
				params[k] = v
			} else {
				params[k] = parsedArray
			}
		} else {
			// Use existing comma splitting logic
			shouldNotSplitOnComma := false
			if slices.Contains(dynamicInstance.IgnoreCommaInParams, k) {
				shouldNotSplitOnComma = true
			}

			if strValue, ok := v.(string); ok && strings.Contains(strValue, ",") && !shouldNotSplitOnComma {
				values := strings.Split(strValue, ",")
				var parsedValues []interface{}
				for _, val := range values {
					val = strings.TrimSpace(val)
					parsedValues = append(parsedValues, val)
				}
				params[k] = parsedValues
			} else {
				params[k] = v
			}
		}
	}

	for k, v := range inputPath.Params {
		// Skip if this parameter should be ignored
		shouldSkip := false
		for _, ignoredKey := range ignoredKeys {
			if k == ignoredKey {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}
		if strValue, ok := v.(string); ok && strings.Contains(strValue, ",") {
			values := strings.Split(strValue, ",")
			var parsedValues []interface{}
			for _, val := range values {
				val = strings.TrimSpace(val)
				parsedValues = append(parsedValues, val)
			}
			params[k] = parsedValues
		} else {
			params[k] = v
		}
	}

	// Convert numeric values to float64
	for k, v := range params {
		switch val := v.(type) {
		case int:
			params[k] = float64(val)
		case string:
			if numValue, err := strconv.ParseFloat(val, 64); err == nil {
				params[k] = numValue
			}
		}
	}
	if dynamicInstance.ParamsNumberated != nil {
		paramCount := 1
		for k, v := range dynamicInstance.ParamsNumberated {
			if strValue, ok := v.(string); ok && strings.Contains(strValue, ",") {
				values := strings.Split(strValue, ",")
				for _, val := range values {
					val = strings.TrimSpace(val)
					params[fmt.Sprintf("%s%d", k, paramCount)] = val
					paramCount++
				}
			}
		}
	}

	return params, globalParamsMap, ignoredKeys
}

func checkInstancesSkip(config *models.Config, countSoFar int) string {
	if config.MaxInstances > 0 && countSoFar >= config.MaxInstances {
		return fmt.Sprintf("Max instances reached: %d", config.MaxInstances)
	}
	return ""
}

func checkRegexPattern(config *models.Config, configuredInstanceConfig models.ConfiguredInstanceConfig, inputPath models.InputPath) string {
	if config.RegexPattern != "" {
		regex, err := regexp.Compile(config.RegexPattern)
		if err != nil {
			return fmt.Sprintf("Error compiling regex pattern: %v", err)
		}
		// Check instance name
		if regex.MatchString(configuredInstanceConfig.Name) {
			if config.Debug {
				logCreation(fmt.Sprintf("Regex Match (configuredInstanceConfig) %s %s", config.RegexPattern, configuredInstanceConfig.Name))
			}
			return ""
		}
		// Check input path
		if regex.MatchString(inputPath.Path) {
			if config.Debug {
				logCreation(fmt.Sprintf("Regex Match (inputPath) %s %s", config.RegexPattern, inputPath.Path))
			}
			return ""
		}
		// Check param set names from instance
		if configuredInstanceConfig.ParamSets != "" {
			for _, paramSetName := range strings.Split(configuredInstanceConfig.ParamSets, ",") {
				paramSetName = strings.TrimSpace(paramSetName)
				if regex.MatchString(paramSetName) {
					if config.Debug {
						logCreation(fmt.Sprintf("Regex Match (param_set name) %s %s", config.RegexPattern, paramSetName))
					}
					return ""
				}
			}
		}
		// Check param set names from input path
		if inputPath.ParamSets != "" {
			for _, paramSetName := range strings.Split(inputPath.ParamSets, ",") {
				paramSetName = strings.TrimSpace(paramSetName)
				if regex.MatchString(paramSetName) {
					if config.Debug {
						logCreation(fmt.Sprintf("Regex Match (inputPath param_set name) %s %s", config.RegexPattern, paramSetName))
					}
					return ""
				}
			}
		}
		// Check param keys from configuredInstanceConfig.Params
		for paramKey := range configuredInstanceConfig.Params {
			if regex.MatchString(paramKey) {
				if config.Debug {
					logCreation(fmt.Sprintf("Regex Match (param key) %s %s", config.RegexPattern, paramKey))
				}
				return ""
			}
		}
		// Check param values from configuredInstanceConfig.Params
		for _, param := range configuredInstanceConfig.Params {
			if strValue, ok := param.(string); ok {
				for _, val := range strings.Split(strValue, ",") {
					val = strings.TrimSpace(val)
					if regex.MatchString(val) {
						if config.Debug {
							logCreation(fmt.Sprintf("Regex Match (configuredInstanceConfig param) %s %s", config.RegexPattern, val))
						}
						return ""
					}
				}
			}
		}
		// Check param sets from config for referenced param sets (check their names and params)
		paramSetNames := append(strings.Split(configuredInstanceConfig.ParamSets, ","), strings.Split(inputPath.ParamSets, ",")...)
		for _, paramSet := range config.Design.ParamSets {
			// Check if this param set is referenced
			for _, refName := range paramSetNames {
				refName = strings.TrimSpace(refName)
				if refName == "" {
					continue
				}
				if paramSet.Name == refName {
					// Check param set name (this is already checked earlier, but keeping for completeness)
					if regex.MatchString(paramSet.Name) {
						if config.Debug {
							logCreation(fmt.Sprintf("Regex Match (referenced param_set name) %s %s", config.RegexPattern, paramSet.Name))
						}
						return ""
					}
					// Check param keys from this param set
					for paramKey := range paramSet.Params {
						if regex.MatchString(paramKey) {
							if config.Debug {
								logCreation(fmt.Sprintf("Regex Match (param_set param key) %s %s", config.RegexPattern, paramKey))
							}
							return ""
						}
					}
					// Check param values from this param set
					for _, paramValue := range paramSet.Params {
						if strValue, ok := paramValue.(string); ok {
							for _, val := range strings.Split(strValue, ",") {
								val = strings.TrimSpace(val)
								if regex.MatchString(val) {
									if config.Debug {
										logCreation(fmt.Sprintf("Regex Match (param_set param value) %s %s", config.RegexPattern, val))
									}
									return ""
								}
							}
						}
					}
				}
			}
		}
		// No match found
		return fmt.Sprintf("Regex pattern (%s) didn't match: (checked: %s & %s)", config.RegexPattern, configuredInstanceConfig.Name, inputPath.Path)
	}
	return ""
}

func checkRegexPatternV2(config *models.Config, instance models.InstanceConfig) string {
	if config.RegexPattern != "" {
		regex, err := regexp.Compile(config.RegexPattern)
		if err != nil {
			return fmt.Sprintf("Error compiling regex pattern: %v", err)
		}
		// Check AutoName first (new check)
		if regex.MatchString(instance.AutoName) {
			if config.Debug {
				logCreation(fmt.Sprintf("Regex Match (AutoName) %s %s", config.RegexPattern, instance.AutoName))
			}
			return ""
		}
		// Check instance name
		if regex.MatchString(instance.Name) {
			if config.Debug {
				logCreation(fmt.Sprintf("Regex Match (instance.Name) %s %s", config.RegexPattern, instance.Name))
			}
			return ""
		}
		// Check input path
		if regex.MatchString(instance.InputPath.Path) {
			if config.Debug {
				logCreation(fmt.Sprintf("Regex Match (inputPath) %s %s", config.RegexPattern, instance.InputPath.Path))
			}
			return ""
		}
		// Check param set names from instance (need to get from config since InstanceConfig doesn't have ParamSets directly)
		// We'll need to check the config for param sets that match the instance
		// For now, we'll check what we can from the instance structure
		// Check param keys from instance.Params
		for paramKey := range instance.Params {
			if regex.MatchString(paramKey) {
				if config.Debug {
					logCreation(fmt.Sprintf("Regex Match (param key) %s %s", config.RegexPattern, paramKey))
				}
				return ""
			}
		}
		// Check param values from instance.Params
		for _, param := range instance.Params {
			if strValue, ok := param.(string); ok {
				for _, val := range strings.Split(strValue, ",") {
					val = strings.TrimSpace(val)
					if regex.MatchString(val) {
						if config.Debug {
							logCreation(fmt.Sprintf("Regex Match (instance param value) %s %s", config.RegexPattern, val))
						}
						return ""
					}
				}
			}
		}
		// Check param set names from input path
		if instance.InputPath.ParamSets != "" {
			for _, paramSetName := range strings.Split(instance.InputPath.ParamSets, ",") {
				paramSetName = strings.TrimSpace(paramSetName)
				if regex.MatchString(paramSetName) {
					if config.Debug {
						logCreation(fmt.Sprintf("Regex Match (inputPath param_set name) %s %s", config.RegexPattern, paramSetName))
					}
					return ""
				}
			}
		}
		// Check param sets from config for referenced param sets (check their names and params)
		paramSetNames := strings.Split(instance.InputPath.ParamSets, ",")
		for _, paramSet := range config.Design.ParamSets {
			// Check if this param set is referenced
			for _, refName := range paramSetNames {
				refName = strings.TrimSpace(refName)
				if refName == "" {
					continue
				}
				if paramSet.Name == refName {
					// Check param set name (this is already checked earlier, but keeping for completeness)
					if regex.MatchString(paramSet.Name) {
						if config.Debug {
							logCreation(fmt.Sprintf("Regex Match (referenced param_set name) %s %s", config.RegexPattern, paramSet.Name))
						}
						return ""
					}
					// Check param keys from this param set
					for paramKey := range paramSet.Params {
						if regex.MatchString(paramKey) {
							if config.Debug {
								logCreation(fmt.Sprintf("Regex Match (param_set param key) %s %s", config.RegexPattern, paramKey))
							}
							return ""
						}
					}
					// Check param values from this param set
					for _, paramValue := range paramSet.Params {
						if strValue, ok := paramValue.(string); ok {
							for _, val := range strings.Split(strValue, ",") {
								val = strings.TrimSpace(val)
								if regex.MatchString(val) {
									if config.Debug {
										logCreation(fmt.Sprintf("Regex Match (param_set param value) %s %s", config.RegexPattern, val))
									}
									return ""
								}
							}
						}
					}
				}
			}
		}
		// No match found
		return fmt.Sprintf("Regex pattern (%s) didn't match: (checked: %s, %s & %s)", config.RegexPattern, instance.AutoName, instance.Name, instance.InputPath.Path)
	}
	return ""
}

func generateAutoName(configuredInstanceConfig models.ConfiguredInstanceConfig, inputPath models.InputPath) string {
	return fmt.Sprintf("%s_%s", configuredInstanceConfig.Name, inputPath.Path)
}

func generateUniqueAutoName(exportPath string, instanceName string, inputPath string) string {
	// Extract the filename from the export path (without extension)
	baseName := filepath.Base(exportPath)
	if strings.Contains(baseName, ".") {
		baseName = baseName[:strings.LastIndex(baseName, ".")]
	}

	// Extract the parent directory and add the name to the baseName, until we hit the export folder
	currentDir := filepath.Dir(exportPath)
	parentNames := []string{}

	// Walk up the directory tree until we hit an "export" folder
	for {
		dirName := filepath.Base(currentDir)
		if dirName == "export" {
			break
		}
		parentNames = append([]string{dirName}, parentNames...) // Prepend to maintain order
		currentDir = filepath.Dir(currentDir)

		// Safety check to prevent infinite loop
		if currentDir == filepath.Dir(currentDir) {
			break // Reached root directory
		}
	}

	// Build the final name with parent directories
	if len(parentNames) > 0 {
		for i := 1; i < len(parentNames); i++ {
			if len(parentNames[i]) > 0 {
				baseName = parentNames[i] + "_" + baseName
			} else {
				baseName = baseName
			}
		}
	}

	// If the baseName is just the instance name, add the input path for uniqueness
	if baseName == instanceName {
		return fmt.Sprintf("%s_%s", instanceName, filepath.Base(inputPath))
	}

	return baseName
}

// sanitizeForHTMLID converts a string to a valid HTML ID by replacing invalid characters
func sanitizeForHTMLID(input string) string {
	// Replace dots with underscores (dots are invalid in CSS selectors)
	result := strings.ReplaceAll(input, ".", "_")

	// Replace other invalid characters for HTML IDs
	invalidChars := []string{":", "*", "?", "\"", "<", ">", "|", " ", "/", "\\", "+", "=", "&", "%", "#", "@", "!", "$", "^", "(", ")", "[", "]", "{", "}", ";", "'", ",", "~", "`"}
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Ensure the ID starts with a letter or underscore (HTML ID requirements)
	if len(result) > 0 && !((result[0] >= 'a' && result[0] <= 'z') || (result[0] >= 'A' && result[0] <= 'Z') || result[0] == '_') {
		result = "id_" + result
	}

	// Remove multiple consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	// Remove leading/trailing underscores
	result = strings.Trim(result, "_")

	// Ensure we have a valid ID
	if result == "" {
		result = "instance"
	}

	return result
}

func GenerateInstances(config *models.Config, configuredInstanceConfig models.ConfiguredInstanceConfig, inputPath models.InputPath) ([]models.InstanceConfig, string, error) {
	if config.Debug {
		logStage("=== Generating Instances === ")
	}

	if inputPath.Path == "" {
		return nil, "", fmt.Errorf("GenerateInstances - input path is empty")
	}

	inputPath.RawOpenSCADFile = "N/A (yet)"

	// Create a map of parameters that should be ignored
	ignoredParams := make(map[string]bool)
	if inputPath.IgnoreParamsWhenProcessing != "" {
		for _, key := range strings.Split(inputPath.IgnoreParamsWhenProcessing, ",") {
			ignoredParams[strings.TrimSpace(key)] = true
		}
	}

	// Get all parameters and filter out ignored ones
	params, globalParamsMap, ignoredParamsKeys := getAllParams(configuredInstanceConfig, config.Design.GlobalParams, config.Design.ParamSets, inputPath)

	for _, key := range ignoredParamsKeys {
		ignoredParams[key] = true
	}

	if config.Debug {
		log.Printf("params from getAllParams: %v", params)
		log.Printf("globalParamsMap: %v", globalParamsMap)
	}

	// Filter out ignored parameters from both params and globalParamsMap
	filteredParams := make(map[string]interface{})
	for k, v := range params {
		if !ignoredParams[k] {
			filteredParams[k] = v
		}
	}

	filteredGlobalParamsMap := make(map[string][]interface{})
	for k, v := range globalParamsMap {
		if !ignoredParams[k] {
			filteredGlobalParamsMap[k] = v
		}
	}

	if config.Debug {
		log.Printf("ignoredParamsKeys: %v", ignoredParamsKeys)
		log.Printf("ignoredParams: %v", ignoredParams)
		log.Printf("filteredParams: %v", filteredParams)
		log.Printf("filteredGlobalParamsMap: %v", filteredGlobalParamsMap)
	}

	// Convert parameters to combinations
	paramCombos, err := convertToParamCombinations(filteredParams, make(map[string]bool), configuredInstanceConfig.DirectArrayParams)
	if err != nil {
		return nil, "", fmt.Errorf("error converting parameters: %v", err)
	}

	// Generate all possible combinations
	var instances []models.InstanceConfig

	// If no parameters have multiple values and no global parameters, create a single instance
	/*if len(filteredParams) == 0 && len(filteredGlobalParamsMap) == 0 {
		instance := models.InstanceConfig{
			Name:               configuredInstanceConfig.Name,
			Params:             make(map[string]interface{}),
			InputPath:          inputPath,
			SkipImages:         configuredInstanceConfig.SkipImages || inputPath.SkipImages,
			SkippedReason:      checkInstancesSkip(config, len(instances)) + checkRegexPattern(config, configuredInstanceConfig, inputPath),
			ID:                 uuid.New().String(),
			ExportImages:       []models.ExportCameraCoordinates{},
			ImageResults:       []models.GenerateImageResult{},
			RunOutputImagePath: "",
			AutoName:           generateAutoName(configuredInstanceConfig, inputPath),
		}

		// Add required parameters
		rawScadFilename := filepath.Base(inputPath.Path)
		instance.Params["designFileName"] = strings.Split(rawScadFilename, ".")[0]

		//LogKeyValuePair("param.designFileName", instance.Params["designFileName"].(string))
		if len(configuredInstanceConfig.Name) > 0 {
			instance.Params["name"] = configuredInstanceConfig.Name
		}
		exportNameFormat := strings.ReplaceAll(config.Design.ExportNameFormat, "{instanceName}", "{name}")
		versionSafe := strings.ReplaceAll(config.Design.Version, " ", "_")
		instance.Params["version"] = versionSafe

		// Format the export name
		if exportNameFormat == "" {
			exportNameFormat = "{designFileName}_{version}_name_{name}"
			if config.Debug {
				log.Printf("[DEBUG] No export_name_format set, using default: %s", exportNameFormat)
			}
		}
		exportName := formatExportName(exportNameFormat, instance.Params, ignoredParams)

		// Normalize versionSafe to use underscores for both spaces and dots
		versionSafe = strings.ReplaceAll(versionSafe, " ", "_")

		configFileDir := filepath.Dir(config.ConfigFile)
		outputPath := filepath.Join(configFileDir, "export", versionSafe, exportName+".stl")
		instance.OutputPathV2 = outputPath
		instance.RunOutputPathV3 = filepath.ToSlash(outputPath)

		for k := range ignoredParams {
			instance.IgnoredParams = append(instance.IgnoredParams, k)
		}

		return []models.InstanceConfig{instance}, "", nil
	}*/

	// Get current working directory
	workingDir, _ := os.Getwd()

	// Generate combinations for parameters
	var parameterCombos []map[string]interface{}
	if len(paramCombos) > 0 {
		parameterCombos = generateCombinations(paramCombos)
	} else {
		parameterCombos = []map[string]interface{}{{}}
	}

	// Generate combinations for global parameters
	var globalCombos []map[string]interface{}
	if len(filteredGlobalParamsMap) > 0 {
		globalCombos = generateCombinations(filteredGlobalParamsMap)
	} else {
		globalCombos = []map[string]interface{}{{}}
	}
	// Combine all parameter combinations
	for _, paramCombo := range parameterCombos {
		for _, globalCombo := range globalCombos {
			instance := models.InstanceConfig{
				Name:                        configuredInstanceConfig.Name,
				Params:                      make(map[string]interface{}),
				InputPath:                   inputPath,
				SkipImages:                  configuredInstanceConfig.SkipImages || inputPath.SkipImages,
				SkippedReason:               "",
				ExportImages:                []models.ExportCameraCoordinates{},
				ImageResults:                []models.GenerateImageResult{},
				RunOutputImagePath:          "",
				AutoName:                    "", // Will be set after parameters are processed
				IsComplete:                  false,
				InputConfigFilePath:         config.ConfigFile,
				InputConfigFilePathRelative: getRelativePath(config.ConfigFile, workingDir),
			}

			// Add non-ignored parameters that don't have multiple values
			for k, v := range filteredParams {
				if !paramHasMultipleValues(v) {
					instance.Params[k] = v
				}
			}

			// Add parameter combination values
			for k, v := range paramCombo {
				instance.Params[k] = v
			}

			// Add global parameter combination values
			for k, v := range globalCombo {
				instance.Params[k] = v
			}

			rawScadFilename := filepath.Base(inputPath.Path)
			instance.Params["designFileName"] = strings.Split(rawScadFilename, ".")[0]
			if len(configuredInstanceConfig.Name) > 0 {
				instance.Params["name"] = configuredInstanceConfig.Name
			}
			exportNameFormat := strings.ReplaceAll(config.Design.ExportNameFormat, "{instanceName}", "{name}")
			versionSafe := strings.ReplaceAll(config.Design.Version, " ", "_")
			instance.Params["version"] = versionSafe

			// Format the export name
			if exportNameFormat == "" || exportNameFormat == "." {
				exportNameFormat = "{designFileName}_name_{name}"
				if config.Debug {
					log.Printf("[DEBUG] No export_name_format set, using default: %s", exportNameFormat)
				}
			}
			//exportName := formatExportName(exportNameFormat, instance.Params, ignoredParams)

			// Normalize versionSafe to use underscores for both spaces and dots
			versionSafe = strings.ReplaceAll(versionSafe, " ", "_")

			configFileDir := filepath.Dir(config.ConfigFile)

			baseExportPath := path.Join(configFileDir, "export", versionSafe, exportNameFormat)

			// Always use the full relative path for output
			//outputPath := filepath.Join(config.ConfigFile, versionSafe, exportName+".stl")

			if config.Debug {
				log.Printf("[DEBUG] MakeFileNameReplacements input: baseExportPath=%s, instanceParams=%v", baseExportPath, instance.Params)
			}
			outputPathReplace := models.MakeFileNameReplacements(config.Design.GlobalParams, instance.Params, instance.IgnoredParams, baseExportPath, config.Design.Version, instance.Params["designFileName"].(string), config.Quality, configuredInstanceConfig.Name, instance.PartIDLetter)
			if config.Debug {
				log.Printf("[DEBUG] MakeFileNameReplacements output: %s", outputPathReplace)
			}

			// Generate unique AutoName based on the export path
			instance.AutoName = generateUniqueAutoName(outputPathReplace, configuredInstanceConfig.Name, inputPath.Path)
			// Use AutoName as the ID
			instance.ID = instance.AutoName
			// Generate a sanitized UniqueID for HTML usage
			instance.UniqueID = sanitizeForHTMLID(instance.AutoName)
			if config.Debug {
				log.Printf("[DEBUG] UniqueID: AutoName=%s, UniqueID=%s", instance.AutoName, instance.UniqueID)
			}

			//	instance.OutputPathV2 = outputPath
			instance.RunOutputPathV3 = filepath.ToSlash(outputPathReplace + ".stl")
			instance.RunOutputPathRelative = getRelativePath(instance.RunOutputPathV3, config.ConfigFile)
			instance.RunOutputImagePath = outputPathReplace

			for k := range ignoredParams {
				instance.IgnoredParams = append(instance.IgnoredParams, k)
			}

			instance.SkippedReason = checkInstancesSkip(config, len(instances)) + checkRegexPatternV2(config, instance)

			instances = append(instances, instance)
		}
	}

	return instances, "NOT USED", nil
}

// Helper function to get relative path
func getRelativePath(fullPath string, basePath string) string {
	// If the full path is already relative, return it as is
	if !strings.HasPrefix(fullPath, "/") {
		return fullPath
	}

	// If base path is empty, just return the filename
	if basePath == "" {
		return filepath.Base(fullPath)
	}

	// Try to make the path relative to the base path
	relPath, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		// If we can't make it relative, return just the filename
		return filepath.Base(fullPath)
	}
	return relPath
}

func commonPrefix(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	i := 0
	for i < minLen && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// Helper function to check if a parameter has multiple values
func paramHasMultipleValues(value interface{}) bool {
	if strValue, ok := value.(string); ok {
		return strings.Contains(strValue, ",")
	}
	// Check if it's a slice/array
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		return true
	}
	return false
}

// formatExportName replaces placeholders in the format string with actual parameter values
func formatExportName(format string, params map[string]interface{}, ignoredParams map[string]bool) string {
	result := format

	if config.Debug {
		logStage("=== formatExportName === ")
		log.Printf("formatExportName:format: %s", format)
		log.Printf("formatExportName:params: %v", params)
		log.Printf("formatExportName:ignoredParams: %v", ignoredParams)
	}

	// Replace version placeholder with sanitized version
	if version, ok := params["version"].(string); ok {
		result = strings.ReplaceAll(result, "{version}", version)
	}

	paramsToProcess := make(map[string]interface{})
	for k, v := range params {
		paramsToProcess[k] = v
	}

	for k, _ := range ignoredParams {
		paramsToProcess[k] = ""
	}

	// Replace other placeholders
	for key, value := range paramsToProcess {
		placeholder := fmt.Sprintf("{%s}", key)
		var strValue string
		if ignoredParams[key] {
			strValue = ""
		} else {
			if key == "version" {
				continue // Already handled above
			}
			if strings.Contains(result, placeholder) {
				// Convert value to string while preserving $ characters
				switch v := value.(type) {
				case string:
					strValue = v
				case float64:
					strValue = fmt.Sprintf("%g", v)
				case bool:
					strValue = fmt.Sprintf("%v", v)
				default:
					strValue = fmt.Sprintf("%v", v)
				}
			}
		}
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	if config.Debug {
		log.Printf("formatExportName:result: %s", result)
	}
	return result
}

// Helper function to generate all possible combinations of parameters
func generateCombinations(paramCombos map[string][]interface{}) []map[string]interface{} {
	if len(paramCombos) == 0 {
		return []map[string]interface{}{{}}
	}

	// Get all keys and sort them for consistent ordering
	keys := make([]string, 0, len(paramCombos))
	for k := range paramCombos {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Get values in the same order as sorted keys
	values := make([][]interface{}, 0, len(paramCombos))
	for _, k := range keys {
		values = append(values, paramCombos[k])
	}

	// Generate all possible combinations
	var combinations []map[string]interface{}
	generateCombinationsHelper(keys, values, 0, make(map[string]interface{}), &combinations)
	return combinations
}

func generateCombinationsHelper(keys []string, values [][]interface{}, index int, current map[string]interface{}, combinations *[]map[string]interface{}) {
	if index == len(keys) {
		// Create a copy of the current combination
		combination := make(map[string]interface{})
		for k, v := range current {
			combination[k] = v
		}
		*combinations = append(*combinations, combination)
		return
	}

	// Try each value for the current key
	for _, value := range values[index] {
		current[keys[index]] = value
		generateCombinationsHelper(keys, values, index+1, current, combinations)
	}
}

// (see GetInputPaths)
func getAllInputPaths(config *models.Config) []models.InputPath {
	if len(config.Design.InputPaths) > 0 {
		inputPaths := make([]models.InputPath, len(config.Design.InputPaths))
		for i, inputPath := range config.Design.InputPaths {
			inputFileName := filepath.Base(inputPath.Path)
			inputPaths[i] = models.InputPath{Path: inputPath.Path, RawOpenSCADFileName: inputFileName}
		}
		return inputPaths
	}
	inputFileName := filepath.Base(config.Design.InputPath)
	return []models.InputPath{{Path: config.Design.InputPath, RawOpenSCADFileName: inputFileName}}
}

func getAbsPath(configFile, inputPath string) string {
	// Get the absolute path of the config file directory
	configDir, err := filepath.Abs(filepath.Dir(configFile))
	if err != nil {
		log.Printf("Could not get absolute path for config directory: %s", err)
		return inputPath
	}

	// Try the path relative to the config file directory first
	absPath := filepath.Join(configDir, inputPath)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// If that doesn't exist, try getting the absolute path of the input path
		absPath, err = filepath.Abs(inputPath)
		if err != nil {
			log.Printf("Could not resolve absolute path: %s", err)
			return inputPath
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		logError(fmt.Sprintf("Input path does not exist: %s", inputPath))
		logError("critical error, existing")
		os.Exit(1)
	}

	// Clean the path to normalize separators and remove any "./" prefix
	absPath = filepath.Clean(absPath)

	return absPath
}

func isNumericRange(val map[string]interface{}) bool {
	_, hasStart := val["start"]
	_, hasEnd := val["end"]
	_, hasStep := val["step"]
	return hasStart && hasEnd && hasStep
}

func getNumericValue(val interface{}) interface{} {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
		return nil
	default:
		return nil
	}
}

func generateNumericRange(start, end, step float64) []float64 {
	var values []float64
	for i := start; i <= end; i += step {
		values = append(values, i)
	}
	return values
}

func interfaceSlice(values []float64) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[i] = v
	}
	return result
}

func copyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func Copy(src string, dst string) error {
	// Read all content of src to data, may cause OOM for a large file.
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// Write data to dst
	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

type OpenSCADVersion struct {
	Version     string
	Path        string
	IsOutOfDate bool
}

// findOpenSCAD finds the OpenSCAD executable and returns its version info
func findOpenSCAD() (OpenSCADVersion, error) {
	// Try to run openscad --version directly
	cmdVer := exec.Command("openscad", "--version")
	outputVer, err := cmdVer.CombinedOutput()
	if err != nil {
		return OpenSCADVersion{}, fmt.Errorf("OpenSCAD not found in PATH (version check). Make sure you can run openscad from the command line: %w", err)
	}

	versionStr := strings.TrimSpace(string(outputVer))

	// Get the path for future use
	path, err := exec.LookPath("openscad")
	if err != nil {
		// If we can run it but can't find the path, just use "openscad"
		path = "openscad"
	}

	return OpenSCADVersion{
		Version:     versionStr,
		Path:        path,
		IsOutOfDate: false,
	}, nil
}

// ProbeOpenSCAD reports whether the openscad CLI is on PATH and returns
// the trimmed combined output of `openscad --version` and the resolved path.
func ProbeOpenSCAD() (OpenSCADVersion, error) {
	return findOpenSCAD()
}

var configTemplate = `[openscadgen]
name = "{{projectName}}"
description = ""

version = "v0.1"

export_name_format = "{designFileName}"

# These params will be passed into all the generated designs
global_params = { }

[[openscadgen.input_paths]]
path = "./{{projectName}}.scad"
params = { }

[[openscadgen.images]]
name = "nice"
`

var configTemplateExtended = `[openscadgen]
name = "{{projectName}}"
description = ""

version = "v0.1"

export_name_format = "{designFileName}_{renderType}"

global_params = { renderType = "obj,vertSlice,horzSlice" }

[[openscadgen.input_paths]]
path = "./{{projectName}}.scad"

[[openscadgen.images]]
name = "nice"
param_filter = { renderType= "obj,vertSlice,horzSlice"}
`

func openScadTemplateExtended(projectNameUnderLined string) string {
	return fmt.Sprintf(`

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";


	module %s(){
		cuboid([100,100,100]);
	}


    sliced(renderType=renderType) {
        %s();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

`, projectNameUnderLined, projectNameUnderLined)
}

var openScadTemplate = func(projectNameUnderLined string) string {
	return fmt.Sprintf(`
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


module %s(){
	cuboid([10,10,10]);
}


%s();
`, projectNameUnderLined, projectNameUnderLined)
}

// InitConfig creates a project from a path-shaped argument: relative or absolute "parent/leaf" or "leaf".
// Only the basename is sanitized; see InitConfigInParent for explicit parent + name.
func InitConfig(projectPathRaw string, extended bool) error {
	parent := filepath.Dir(projectPathRaw)
	leaf := filepath.Base(projectPathRaw)
	if parent == "" || parent == "." {
		parent = "."
	}
	return InitConfigInParent(parent, leaf, extended)
}

func LogKeys(flags models.CmdFlags) {
	LogKeyValuePair("Debug", "true")
	LogKeyValuePair("Quiet", "false")
	LogKeyValuePair("NoProcessing", "false")
	LogKeyValuePair("RegexPattern", "false")

}

func InitLogger(logFilePath string) error {
	if logFilePath == "memory" {
		logToMemory = true
		logger = log.New(io.MultiWriter(os.Stdout, &logBuffer), "", log.Ldate|log.Ltime|log.Lshortfile)
		return nil
	}

	//if config.IncludeExportLog {
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)

	if logToMemory {
		// Flush the buffer to the log file
		_, err := logFile.Write(logBuffer.Bytes())
		if err != nil {
			return err
		}
		logBuffer.Reset()
		logToMemory = false
	}
	/*} else {
		// Ensure we still log to console even when not logging to file
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	}*/

	return nil
}

func LogKeyValuePair(key string, value string) {
	logger.Printf(colorYellow+"%s: "+colorWhite+"\t\t\t\t%s"+colorReset, key, value)
}

func logSkip(message string) {
	logger.Printf(colorYellow+"%s"+colorReset, message)
}

func LogWarn(message string, critical bool) {
	if critical {
		logger.Printf(colorRed+"%s"+colorReset, message)
	} else {
		logger.Printf(colorOrange+"%s"+colorReset, message)
	}
}

func logTip(message string) {
	logger.Printf(colorCyan+"\t%s"+colorReset, message)
}

// Exclude symbols and use multiple letters if the number is greater than 26
func getPartIDLetter(stlIndex int) string {
	if stlIndex < 26 {
		letter := string(rune(65 + stlIndex))
		if letter >= "A" && letter <= "Z" {
			return letter
		}
	}
	// For numbers >= 26, use multiple letters (AA, AB, AC, etc.)
	quotient := stlIndex / 26
	remainder := stlIndex % 26

	var result string
	if quotient > 0 {
		result += string(rune(64 + quotient)) // First letter
	}
	result += string(rune(65 + remainder)) // Second letter or only letter if < 26

	return result
}

func logCreation(message string) {
	logger.Printf(colorGreen+"%s"+colorReset, message)
}

func logError(message string) {
	logger.Printf(colorRed+"%s"+colorReset, message)
}

func logStage(stage string) {
	logger.Printf(colorBlue+"\n\n========== %s =========="+colorReset, stage)
}

func SetMetadata(fileName string, metadata map[string]string, config *models.Config) error {
	if config.Debug {
		logStage("Setting metadata")
	}
	// Check if the file exists
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		LogWarn(fmt.Sprintf("[SetMetadata] warning: file '%s' does not exist", fileName), false)
		return fmt.Errorf("warning: file '%s' does not exist", fileName)
	} else if err != nil {
		LogWarn(fmt.Sprintf("[SetMetadata] warning: error accessing file '%s': %v", fileName, err), false)
		return fmt.Errorf("error accessing file '%s': %v", fileName, err)
	}

	// Get OS details
	currentOS := runtime.GOOS
	if config.Debug {
		fmt.Printf("Running on OS: %s\n", currentOS)
	}
	// Set metadata based on the OS
	switch currentOS {
	case "linux", "darwin":
		// For Linux and macOS, use xattrs
		for key, value := range metadata {
			xattrKey := "user." + key
			if err := xattr.Set(fileName, xattrKey, []byte(value)); err != nil {
				LogWarn(fmt.Sprintf("warning: error setting xattr '%s' on file '%s': %v", key, fileName, err), false)
				return fmt.Errorf("error setting xattr '%s' on file '%s': %v", key, fileName, err)
			} else if config.Debug {
				LogKeyValuePair("Set xattr", xattrKey)
				fmt.Printf("Set xattr '%s' on file '%s' with value: %s\n", xattrKey, fileName, value)
			}
		}
	case "windows":
		// For Windows, use NTFS Alternate Data Streams (ADS)
		for key, value := range metadata {
			adsName := fileName + ":" + key
			file, err := os.OpenFile(adsName, os.O_CREATE|os.O_RDWR, 0600)
			if err != nil {
				LogWarn(fmt.Sprintf("warning: error opening ADS '%s': %v", adsName, err), false)
				return fmt.Errorf("error opening ADS '%s': %v", adsName, err)
			} else if config.Debug {
				log.Printf("📊 Set ADS: %s", adsName)
			}
			defer file.Close()

			_, err = file.Write([]byte(value))
			if err != nil {
				LogWarn(fmt.Sprintf("warning: error writing to ADS '%s': %v", adsName, err), false)
				return fmt.Errorf("error writing to ADS '%s': %v", adsName, err)
			} else if config.Debug {
				log.Printf("📊 Set ADS: %s with value: %s", adsName, value)
			}
			fmt.Printf("Set ADS '%s' on file '%s' with value: %s\n", key, fileName, value)
		}
	default:
		LogWarn(fmt.Sprintf("warning: unsupported operating system: %s", currentOS), false)
		return fmt.Errorf("unsupported operating system: %s", currentOS)
	}

	return nil
}

// GetMetadata retrieves metadata from a file
func GetMetadata(fileName string) (map[string]string, error) {
	metadata := make(map[string]string)

	// Check if the file exists
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return metadata, fmt.Errorf("file '%s' does not exist", fileName)
	} else if err != nil {
		return metadata, fmt.Errorf("error accessing file '%s': %v", fileName, err)
	}

	// Get OS details
	currentOS := runtime.GOOS

	// Get metadata based on the OS
	switch currentOS {
	case "linux", "darwin":
		// For Linux and macOS, use xattrs
		attrs, err := xattr.List(fileName)
		if err != nil {
			return metadata, fmt.Errorf("error listing xattrs for '%s': %v", fileName, err)
		}

		for _, attr := range attrs {
			if strings.HasPrefix(attr, "user.") {
				value, err := xattr.Get(fileName, attr)
				if err != nil {
					continue // Skip this attribute if we can't read it
				}
				key := strings.TrimPrefix(attr, "user.")
				metadata[key] = string(value)
			}
		}
	case "windows":
		// For Windows, use NTFS Alternate Data Streams (ADS)
		// This is a simplified implementation - in practice you might want to use a library
		// that properly handles ADS enumeration
		dir, err := os.ReadDir(filepath.Dir(fileName))
		if err != nil {
			return metadata, fmt.Errorf("error reading directory for '%s': %v", fileName, err)
		}

		baseName := filepath.Base(fileName)
		for _, entry := range dir {
			if strings.HasPrefix(entry.Name(), baseName+":") {
				streamName := strings.TrimPrefix(entry.Name(), baseName+":")
				adsPath := fileName + ":" + streamName
				if data, err := os.ReadFile(adsPath); err == nil {
					metadata[streamName] = string(data)
				}
			}
		}
	default:
		return metadata, fmt.Errorf("unsupported operating system: %s", currentOS)
	}

	return metadata, nil
}

// loadEstimatedTotalTime loads the estimated total processing time from metadata
func loadEstimatedTotalTime(config *models.Config) time.Duration {
	// Try to read from config file metadata first
	if metadata, err := GetMetadata(config.ConfigFile); err == nil {
		if lastTimeStr, exists := metadata["last_process_time_taken"]; exists {
			if lastTime, err := time.ParseDuration(lastTimeStr); err == nil {
				log.Printf("📊 Using metadata for progress bar estimation: %v (from previous run)", lastTime)
				return lastTime
			}
		}
	}

	// Fallback: estimate based on number of instances
	totalInstances := len(config.GetInputPaths()) * len(config.Design.ConfiguredInstanceConfig)
	if totalInstances == 0 {
		totalInstances = 1 // Prevent division by zero
	}
	estimatedTime := time.Duration(totalInstances) * 30 * time.Second // 30s per instance estimate
	log.Printf("📊 Using fallback estimation for progress bar: %v (%d instances × 30s)", estimatedTime, totalInstances)
	return estimatedTime
}

// storeProcessingTime stores the processing time in metadata for future estimates
func storeProcessingTime(config *models.Config, totalTime time.Duration) {
	metadata := map[string]string{
		"last_process_time_taken": totalTime.String(),
		"last_process_timestamp":  time.Now().Format(time.RFC3339),
	}
	SetMetadata(config.ConfigFile, metadata, config)

}

// Set all the attributes against the file in the metadata
func setBuildInfoInFileAttributes(outputPath string, config *models.Config, instance *models.InstanceConfig) {
	metadata := make(map[string]string)
	metadata["openscadgen.version"] = config.Design.Version
	metadata["openscadgen.instance"] = instance.Name
	for name, value := range instance.Params {
		metadata[fmt.Sprintf("openscadgen.params.%s", name)] = fmt.Sprintf("%v", value)
	}
	SetMetadata(outputPath, metadata, config)
}

// isValidPath checks if the path contains any invalid characters from the provided string
func isValidPath(path string, invalidChars string) bool {
	for _, char := range invalidChars {
		if strings.ContainsRune(path, char) {
			return false
		}
	}
	return true
}

func generateSTL(instance *models.InstanceConfig, config *models.Config) (models.GenerateSTLResult, error) {
	result := models.GenerateSTLResult{
		InstanceConfig: *instance,
		OutputPath:     "",
		Command:        "",
		Skipped:        false,
		LowQuality:     false,
		Error:          "",
		AppliedParams:  make(map[string]interface{}),
	}

	startTime := time.Now()

	if config.Debug {
		logStage("Generating STL")
		LogKeyValuePair("inputPath", instance.InputPath.Path)
		LogKeyValuePair("exportFolderPath", instance.RunOutputPathV3)
		log.Printf("[DEBUG] Starting STL generation for instance: %s", instance.Name)
		log.Printf("[DEBUG] Instance params: %+v", instance.Params)
	}

	// Validate output path for invalid characters
	if !isValidPath(instance.OutputPathV2, invalidChars) {
		errMsg := fmt.Sprintf("Output path '%s' contains invalid character(s) from '%s'. Please update your export_name_format in config.", instance.OutputPathV2, invalidChars)
		result.Error = errMsg
		if config.StopOnError {
			return result, nil
		}
		return result, fmt.Errorf(errMsg)
	}

	// Ensure output directory exists before proceeding
	outputDir := filepath.Dir(instance.OutputPathV2)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create output directory %s: %v", outputDir, err)
		if config.StopOnError {
			return result, nil
		}
		return result, fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Find the matching input path to get its IgnoreParamsWhenProcessing
	var ignoredKeys []string
	for _, inputPath := range config.Design.InputPaths {
		if inputPath.Path == instance.InputPath.Path {
			if inputPath.IgnoreParamsWhenProcessing != "" {
				ignoredKeys = strings.Split(inputPath.IgnoreParamsWhenProcessing, ",")
				for i, key := range ignoredKeys {
					ignoredKeys[i] = strings.TrimSpace(key)
				}
			}
			break
		}
	}

	if config.Debug && len(ignoredKeys) > 0 {
		log.Printf("Ignored keys for STL generation: %v", ignoredKeys)
	}

	// Copy parameters, skipping ignored ones
	for k, v := range instance.Params {
		// Skip if this parameter should be ignored
		shouldSkip := false
		for _, ignoredKey := range ignoredKeys {
			if k == ignoredKey {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			if config.Debug {
				log.Printf("Skipping parameter %s for STL generation", k)
			}
			continue
		}
		result.AppliedParams[k] = v
	}

	if config.Debug {
		LogKeyValuePair("Applied parameters for STL generation", "")
		for key, value := range result.AppliedParams {
			LogKeyValuePair(key, fmt.Sprintf("%v", value))
		}
	}

	name := instance.Name
	if name == "" {
		name = filepath.Base(instance.InputPath.Path)
	}

	// Get instance paths
	paths := instance.GetInstancePaths(config)

	// Debug: check if input file exists
	if config.Debug {
		if fileInfo, err := os.Stat(paths.InputPath); err == nil {
			log.Printf("Input file exists: %s (size: %d bytes)", paths.InputPath, fileInfo.Size())
		} else {
			log.Printf("Input file does NOT exist: %s", paths.InputPath)
		}
	}

	// Build OpenSCAD command
	openscadCmd := FindOpenSCAD()
	if openscadCmd == "" {
		return result, fmt.Errorf("openscad command not found")
	}

	if config.Debug {
		log.Printf("Using OpenSCAD command: %s", openscadCmd)
	}

	// Create directory if not exists
	stlOutputDir := filepath.Dir(instance.RunOutputPathV3)
	if config.Debug {
		log.Printf("[DEBUG] Creating output directory: %s", stlOutputDir)
	}
	if err := os.MkdirAll(stlOutputDir, 0755); err != nil {
		if config.Debug {
			log.Printf("[DEBUG] Failed to create output directory: %v", err)
		}
		return result, fmt.Errorf("failed to create output directory: %w", err)
	}
	if config.Debug {
		log.Printf("[DEBUG] Output directory created/verified: %s", stlOutputDir)
	}

	// Build command arguments
	args := []string{
		"-o", fmt.Sprintf("'%s'", instance.RunOutputPathV3),
	}

	//check for ? charcters
	if !isValidPath(instance.RunOutputPathV3, invalidChars) {
		errMsg := fmt.Sprintf("Output path '%s' contains invalid character(s) from '%s'. Please update your export_name_format in config.", instance.OutputPathV2, invalidChars)
		result.Error = errMsg
		if config.StopOnError {
			return result, nil
		}
		return result, fmt.Errorf(errMsg)
	}
	if !config.SkipRender {
		args = append(args, "--render")
	}

	if !config.Design.DontUseManifold {
		//	args = append(args, "--backend=manifold")
	}

	if config.Debug {
		args = append(args, "--debug", "all")
		args = append(args, "--summary", "all")
	}

	// Add custom OpenSCAD arguments if provided
	if config.Design.CustomOpenSCADArgs != "" {
		args = append(args, strings.Split(config.Design.CustomOpenSCADArgs, " ")...)
	}

	for name, value := range result.AppliedParams {
		if name == "version" {
			continue
		}
		if reflect.TypeOf(value).Kind() == reflect.String && value != "true" && value != "false" {
			args = append(args, "-D", fmt.Sprintf("'%s=\"%v\"'", name, value))
		} else {
			args = append(args, "-D", fmt.Sprintf("'%s=%v'", name, value))
		}
	}

	if config.OverrideFN > 0 {
		args = append(args, "-D", fmt.Sprintf("'$fn=%d'", config.OverrideFN))
	}

	// Create command string
	commandStr := fmt.Sprintf("%s %s %s", openscadCmd, strings.Join(args, " "), paths.InputPath)

	if config.Debug {
		log.Printf("[DEBUG] Built OpenSCAD command string")
		log.Printf("[DEBUG] OpenSCAD executable: %s", openscadCmd)
		log.Printf("[DEBUG] Arguments: %v", args)
		log.Printf("[DEBUG] Input path: %s", paths.InputPath)
		log.Printf("[DEBUG] Full command: %s", commandStr)
	}

	if config.Debug {
		logStage(fmt.Sprintf("Running STL generation: %s", instance.Name))
		LogKeyValuePair("Command", commandStr)
	}

	// Run command through shell
	cmd := exec.Command("sh", "-c", commandStr)

	if config.Debug {
		log.Printf("[DEBUG] About to execute OpenSCAD command")
		log.Printf("[DEBUG] Command: %s", commandStr)
		log.Printf("[DEBUG] Working directory: %s", filepath.Dir(paths.InputPath))
	}

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	result.Command = commandStr
	result.CommandOutput = string(output)
	result.TimeTaken = time.Since(startTime)

	if config.Debug {
		log.Printf("[DEBUG] OpenSCAD command completed")
		log.Printf("[DEBUG] Error: %v", err)
		log.Printf("[DEBUG] Output length: %d bytes", len(output))
		if len(output) > 0 {
			log.Printf("[DEBUG] Output: %s", string(output))
			result.OutputLog = string(output)

		}
		log.Printf("[DEBUG] Time taken: %v", result.TimeTaken)
	}

	if err != nil {
		result.Error = fmt.Sprintf("error running openscad: %v\nOutput: %s", err, string(output))
		if config.Debug {
			log.Printf("OpenSCAD command failed with error: %v", err)
			log.Printf("Command output: %s", string(output))
		}
		if config.StopOnError {
			return result, nil
		}
		return result, fmt.Errorf("OpenSCAD command failed: %w\nOutput: %s", err, string(output))
	}

	// Check if file was created and has content
	if config.Debug {
		log.Printf("[DEBUG] Checking if output file exists: %s", instance.RunOutputPathV3)
	}

	fileInfo, statErr := os.Stat(instance.RunOutputPathV3)
	if os.IsNotExist(statErr) {
		result.Error = fmt.Sprintf("output file was not created: %s", instance.RunOutputPathV3)
		if config.Debug {
			log.Printf("[DEBUG] Output file does not exist: %s", instance.RunOutputPathV3)
		}
		if config.StopOnError {
			return result, nil
		}
		return result, fmt.Errorf("output file was not created: %s", instance.RunOutputPathV3)
	} else if statErr != nil {
		result.Error = fmt.Sprintf("error checking output file: %v", statErr)
		if config.Debug {
			log.Printf("[DEBUG] Error checking output file: %v", statErr)
		}
		if config.StopOnError {
			return result, nil
		}
		return result, fmt.Errorf("error checking output file: %w", statErr)
	} else if fileInfo.Size() == 0 {
		result.Error = fmt.Sprintf("output file is empty: %s", instance.RunOutputPathV3)
		if config.Debug {
			log.Printf("[DEBUG] Output file is empty (size: %d): %s", fileInfo.Size(), instance.RunOutputPathV3)
		}
		if config.StopOnError {
			return result, nil
		}
		return result, fmt.Errorf("output file is empty: %s", instance.RunOutputPathV3)
	} else {
		if config.Debug {
			log.Printf("[DEBUG] Output file created successfully (size: %d bytes): %s", fileInfo.Size(), instance.RunOutputPathV3)
		}
		logCreation(fmt.Sprintf("\n📦 STL\t\t%s\t\t%s", result.InstanceConfig.AutoName, result.TimeTaken.String()))
		if config.Debug {
			LogKeyValuePair("STL File", instance.RunOutputPathV3)
		}
	}

	result.OutputPath = instance.OutputPathV2

	if config.Debug {
		log.Printf("[DEBUG] STL generation completed successfully for instance: %s", instance.Name)
		log.Printf("[DEBUG] Total time taken: %v", result.TimeTaken)
	}

	return result, nil
}

func FindOpenSCAD() string {
	// Use exec.LookPath which is cross-platform
	path, err := exec.LookPath("openscad")
	if err != nil {
		log.Fatalf("OpenSCAD not found in PATH. (%+v)", err)
	}
	return path
}

func BuildHomeURL(serverFolder string) string {
	var homeURL string = "/"
	if serverFolder != "" {
		encodedServerFolder := base64.StdEncoding.EncodeToString([]byte(serverFolder))
		homeURL = fmt.Sprintf("/?server_folder=%s", encodedServerFolder)
	}
	return homeURL
}

func BuildConfigFileURL(configFilePath string, serverFolder string) string {
	var configFileURL string = "/"
	if configFilePath != "" {
		encodedConfigFilePath := base64.StdEncoding.EncodeToString([]byte(configFilePath))
		encodedServerFolder := base64.StdEncoding.EncodeToString([]byte(serverFolder))
		configFileURL = fmt.Sprintf("/?config=%s&server_folder=%s", encodedConfigFilePath, encodedServerFolder)
	}
	return configFileURL
}

func BuildPageUrl(configFilePath string, serverFolder string) models.PageUrlInfo {
	var pageURL string = "/?"
	var encodedConfigFilePath string
	if configFilePath != "" {
		encodedConfigFilePath = base64.StdEncoding.EncodeToString([]byte(configFilePath))
		pageURL += fmt.Sprintf("config=%s&", encodedConfigFilePath)
	}

	var encodedServerFolder string
	if serverFolder != "" {
		encodedServerFolder = base64.StdEncoding.EncodeToString([]byte(serverFolder))
		pageURL += fmt.Sprintf("server_folder=%s", encodedServerFolder)
	}

	return models.PageUrlInfo{
		PageURL:               pageURL,
		HomeURL:               BuildHomeURL(serverFolder),
		ConfigFileURL:         BuildConfigFileURL(configFilePath, serverFolder),
		ConfigFilePath:        configFilePath,
		ConfigFilePathEncoded: encodedConfigFilePath,
		ServerFolder:          serverFolder,
		ServerFolderEncoded:   encodedServerFolder,
	}
}

func BuildReportMeta(params models.BuildReportMetaParams, results models.Results) models.ReportMeta {
	pageUrlInfo := BuildPageUrl(params.ConfigFilePath, params.ServerFolder)

	reportMeta := models.ReportMeta{
		HomeURL:               BuildHomeURL(params.ServerFolder),
		IsServerMode:          params.IsServerMode,
		TotalQueuedInstances:  params.TotalQueuedInstances,
		ConfigFilePath:        params.ConfigFilePath,
		ConfigFilePathEncoded: pageUrlInfo.ConfigFilePathEncoded,
		ServerFolder:          params.ServerFolder,
		ServerFolderEncoded:   pageUrlInfo.ServerFolderEncoded,
		Results:               results,
	}

	return reportMeta
}

func GenerateOutputReport(config *models.Config, instances []models.InstanceConfig, stlResults []models.GenerateSTLResult, imageResults []models.GenerateImageResult, outputDir string, toFile bool, totalTimeTaken time.Duration) (templ.Component, string, error) {
	logStage("Generating HTML report")
	if config.Debug && toFile {
		LogKeyValuePair("Generating HTML report at", outputDir)
	}

	//return nil, nil
	// Get all parameter names
	var allParamNames []string
	for _, instance := range instances {
		for paramName := range instance.Params {
			found := false
			for _, existingName := range allParamNames {
				if existingName == paramName {
					found = true
					break
				}
			}
			if !found {
				allParamNames = append(allParamNames, paramName)
			}
		}
	}

	outputFile := filepath.Join(outputDir, "report.html")

	if config.Debug {
		LogKeyValuePair("REPORT OUTPUT", outputFile)

		logStage("Report Params")
		LogKeyValuePair("config", fmt.Sprintf("%+v", config))
		LogKeyValuePair("instances", fmt.Sprintf("%+v", instances))
		LogKeyValuePair("outputFile", fmt.Sprintf("%+v", outputFile))
		LogKeyValuePair("stlResults", fmt.Sprintf("%+v", stlResults))

		LogKeyValuePair("imageResults", fmt.Sprintf("%+v", imageResults))
		LogKeyValuePair("allParamNames", fmt.Sprintf("%+v", allParamNames))
		LogKeyValuePair("false", fmt.Sprintf("%+v", false))
		LogKeyValuePair("serveroutputfile", "")

	}

	reportMeta := BuildReportMeta(models.BuildReportMetaParams{
		IsServerMode:   false,
		ConfigFilePath: "",
		ServerFolder:   "",
	}, models.Results{
		TimeTake: totalTimeTaken,
	})
	htmlContent := templates.Report("complete", config, instances, outputFile, stlResults, imageResults, allParamNames, reportMeta, totalTimeTaken, nil)

	var htmlFile *os.File
	var err error
	if toFile {
		if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
			return nil, outputFile, fmt.Errorf("failed to create output directory: %w", err)
		}
		htmlFile, err = os.Create(outputFile)
		if err != nil {
			return nil, outputFile, fmt.Errorf("GenerateOutputReport - failed to create HTML file: %w", err)
		}
		if config.Debug {
			// Log file handle state
			log.Printf("[DEBUG] About to render HTML: htmlFile.Name()=%v, htmlFile is nil? %v", htmlFile.Name(), htmlFile == nil)
			log.Printf("[DEBUG] htmlContent type: %T, value: %+v", htmlContent, htmlContent)
		}
	}

	// Write the HTML content to the file
	if toFile {
		if config.Debug {
			log.Printf("[DEBUG] Calling htmlContent.Render with file: %v", htmlFile)
		}
		err := htmlContent.Render(context.Background(), htmlFile)
		if err != nil {
			log.Printf("[ERROR] htmlContent.Render failed: %v", err)
			return nil, outputFile, fmt.Errorf("htmlContent.Render failed to RENDER HTML (%s)\n contents: %+v", outputFile, err)
		}
		defer htmlFile.Close()
		if config.Debug {
			log.Printf("[DEBUG] htmlContent.Render succeeded for file: %v", outputFile)
		}
		LogKeyValuePair("Html Report:", outputFile)

	}

	if !config.Quiet && !toFile {
		absPath, err := filepath.Abs(outputFile)
		if err != nil {
			logError(fmt.Sprintf("failed to get absolute path for report: %v", err))
		} else {
			LogKeyValuePair("HTML report", absPath)
			logCreation(fmt.Sprintf("file://%s", absPath))
		}
	} else if config.Debug {
		log.Print("HTML content generated.")
	}

	return htmlContent, outputFile, nil
}

// generateAllCameraCombinations creates all possible camera combinations for a given direction
func generateAllCameraCombinations(direction string) []models.ExportCameraCoordinates {
	var cameras []models.ExportCameraCoordinates

	// Check if the direction is valid
	if _, ok := cameraPresets[direction]; !ok {
		return cameras
	}

	// Generate cameras for each distance
	for distanceKey, _ := range cameraDistances {
		// Skip empty distance key for "all" combinations
		if direction == "all" && distanceKey == "" {
			continue
		}

		// Create camera name
		cameraName := direction
		if distanceKey != "" {
			cameraName = fmt.Sprintf("%s-%s", direction, distanceKey)
		}

		// Generate coordinates
		coordinates := generateCameraCoordinates(direction, distanceKey)

		// Add to result
		cameras = append(cameras, models.ExportCameraCoordinates{
			CameraName:        cameraName,
			CameraCoordinates: coordinates,
		})
	}

	return cameras
}

func executeCommand(cmd string) error {
	command := exec.Command("sh", "-c", cmd)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		// Build detailed error message with command output
		errorDetails := fmt.Sprintf("command failed: %v", err)
		if stdout.Len() > 0 {
			errorDetails += fmt.Sprintf("\nstdout: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			errorDetails += fmt.Sprintf("\nstderr: %s", stderr.String())
		}
		return fmt.Errorf("%s", errorDetails)
	}
	return nil
}

func executeCommandWithOutput(cmd string) (string, error) {
	command := exec.Command("sh", "-c", cmd)

	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		// Build detailed error message with command output
		errorDetails := fmt.Sprintf("command failed: %v", err)
		if stdout.Len() > 0 {
			errorDetails += fmt.Sprintf("\nstdout: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			errorDetails += fmt.Sprintf("\nstderr: %s", stderr.String())
		}
		return stdout.String() + stderr.String(), fmt.Errorf("%s", errorDetails)
	}
	return stdout.String() + stderr.String(), nil
}

func processImage(config *models.Config, instance *models.InstanceConfig, progress ProgressReporter) ([]models.GenerateImageResult, error) {

	var imageResults []models.GenerateImageResult

	if instance.SkipImages {
		return imageResults, nil
	}

	// Create a dummy STL result for image generation
	stlResult := models.GenerateSTLResult{
		InstanceConfig: *instance,
		AppliedParams:  instance.Params,
	}

	//totalImages := len(instance.ExportImages)
	completedImages := 0

	for _, camera := range instance.ExportImages {
		if len(instance.SkippedImageReason) > 0 {
			if config.Debug {
				logSkip(fmt.Sprintf("Skipping image %s", instance.SkippedImageReason))
			}
		} else {
			// Update progress for current image
			//if progress != nil {
			//		progress.ProgressInstance(instance.ID, int(float64(completedImages)/float64(totalImages)*100))
			//}

			result, cmdStr, err := generateImage(instance, config, camera, stlResult)
			if err != nil {
				LogKeyValuePair(
					"Command", cmdStr,
				)
				logError(fmt.Sprintf("Error generating image: '%s' (%s) %v", camera.CameraName, camera.CameraCoordinates, err))
				return imageResults, fmt.Errorf("error generating image: %w", err)
			}
			if config.Debug {
				LogKeyValuePair("Generated Image", result.OutputPath)
			}

			imageResults = append(imageResults, result)
		}

		completedImages++
	}

	return imageResults, nil
}

func GetImagePath(runOutputImagePath string, cameraName string) string {
	fileName := filepath.Base(runOutputImagePath)
	// Remove file extension
	fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	outputPath := filepath.Dir(runOutputImagePath)
	return outputPath + "/" + fileNameWithoutExt + "-" + cameraName + ".png"
}

func generateImage(instance *models.InstanceConfig, config *models.Config, camera models.ExportCameraCoordinates, stlResult models.GenerateSTLResult) (models.GenerateImageResult, string, error) {
	// Initialize result
	imageResult := models.GenerateImageResult{
		InstanceConfig: *instance,
		AppliedParams:  stlResult.AppliedParams,
		CameraName:     camera.CameraName,
		CameraCoords:   camera.CameraCoordinates,
	}

	// Get instance paths
	paths := instance.GetInstancePaths(config)

	// Create output path for PNG
	if instance.RunOutputImagePath == "" {
		log.Panic("RunOutputImagePath is empty")
	}

	outputImgPath := GetImagePath(instance.RunOutputImagePath, camera.CameraName)
	if !config.Quiet && config.Debug {
		LogKeyValuePair("outputPath", outputImgPath)
	}
	/*
		// Ensure output directory exists
		absOutputPath := filepath.Join(filepath.Dir(config.ConfigFile), outputImgPath)
		if err := os.MkdirAll(filepath.Dir(absOutputPath), 0755); err != nil {
			return imageResult, fmt.Errorf("error creating output directory: %w", err)
		}
	*/
	// Find OpenSCAD executable
	openscadCmd := FindOpenSCAD()
	if openscadCmd == "" {
		return imageResult, "", fmt.Errorf("OpenSCAD not found")
	}

	// Start timing
	startTime := time.Now()

	// Set default image size if not specified
	imageSize := "1920,1080"
	if camera.ImageSize != "" {
		imageSize = camera.ImageSize
	}
	if config.Debug {
		LogKeyValuePair("outputImgPath", outputImgPath)
	}

	// Build command arguments
	args := []string{
		"-o", outputImgPath,
		"--imgsize", imageSize,
		"--projection", "perspective", //ortho
		"--camera", camera.CameraCoordinates,
		"--preview",
	}

	// Add custom OpenSCAD arguments if provided
	if config.Design.CustomOpenSCADArgs != "" {
		args = append(args, strings.Split(config.Design.CustomOpenSCADArgs, " ")...)
	}

	args = append(args, "-D", "genType=\"image\"")

	for name, value := range stlResult.AppliedParams {
		if reflect.TypeOf(value).Kind() == reflect.String && value != "true" && value != "false" {

			args = append(args, "-D", fmt.Sprintf("'%s=\"%v\"'", name, value))
		} else {

			//	logError("WOAH!")
			args = append(args, "-D", fmt.Sprintf("'%s=%v'", name, value))
		}
	}

	// Add input file using the correct path
	args = append(args, paths.InputPath)

	imgStartTime := time.Now()
	// Create the command
	cmd := exec.Command(openscadCmd, args...)

	if config.Debug {
		logStage("Running Image generation")
		LogKeyValuePair("Command", cmd.String())
	}

	cmdStr := cmd.String()
	// Run the command and capture output
	output, err := executeCommandWithOutput(cmdStr)
	imageResult.Command = cmdStr
	imageResult.CommandOutput = output

	if err != nil {
		// Log detailed error information
		logError(fmt.Sprintf("OpenSCAD command failed for image generation"))
		logError(fmt.Sprintf("Command: %s", cmd.String()))
		logError(fmt.Sprintf("Output path: %s", outputImgPath))
		logError(fmt.Sprintf("Error details: %v", err))

		return imageResult, cmdStr, fmt.Errorf("error running OpenSCAD: %w", err)
	} else {
		logCreation(fmt.Sprintf("\n🖼️  Image\t%s\t\t%s", instance.AutoName, time.Since(imgStartTime)))
		if config.Debug {
			LogKeyValuePair("Image Created", outputImgPath)
		}
	}

	// Update result with timing information
	imageResult.TimeTaken = time.Since(startTime)
	imageResult.OutputPath = outputImgPath

	return imageResult, cmdStr, nil
}

func printDirectoryContents(dir string) {
	printDirectoryContentsRecursive(dir, 0)
}

func printDirectoryContentsRecursive(dir string, depth int) {
	contents, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Failed to read directory %s: %v", dir, err)
		return
	}

	// Print header only for the root directory
	if depth == 0 {
		log.Printf("\n===== Directory Contents:  =====\n (%s)\n", dir)
	}

	// Create indentation based on depth
	indent := strings.Repeat("  ", depth)

	for _, entry := range contents {
		if entry.IsDir() {
			log.Printf("%s📁 %s/\n", indent, entry.Name())
			// Recursively print subdirectory contents
			subDir := filepath.Join(dir, entry.Name())
			printDirectoryContentsRecursive(subDir, depth+1)
		} else {
			log.Printf("%s📄 %s\n", indent, entry.Name())
		}
	}

	// Print footer only for the root directory
	if depth == 0 {
		log.Printf("=============================\n")
	}
}
