package pkg

import (
	"bufio"
	"bytes"
	"context"
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

	// Convert configDir to relative path if possible
	workingDir, err := os.Getwd()
	if err != nil {
		logError(fmt.Sprintf("Warning: Could not get working directory: %s", err))
		workingDir = configDir
	}

	// Try to make configDir relative to working directory
	relConfigDir, err := filepath.Rel(workingDir, configDir)
	if err != nil {
		logError(fmt.Sprintf("Warning: Could not make config dir relative: %s", err))
		relConfigDir = configDir
	}

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

	/*if config.Design.OutputPath != "" {
		// When output path is explicitly specified, use it relative to config dir
		outputPath := filepath.Join(relConfigDir, config.Design.OutputPath)

		if config.Debug {
			log.Printf("Output path specified in config: %s", outputPath)
		}

		if !config.Quiet {
			LogKeyValuePair("v", config.Design.Version)
			LogKeyValuePair("getOutputPaths:Output path", filepath.Join(outputPath, config.Design.Version))
		}

		// Construct paths with the correct directory structure
		var baseExportPath, exportFolderPath string
		var output models.OutputPaths
		if hasExportPrefix {
			output = models.OutputPaths{
				OutputPath:            filepath.Join(baseExportPath),
				ExportFolderPath:      exportFolderPath,
				LowQualityWarningPath: filepath.Join(baseExportPath, "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(baseExportPath, "README.md"),
				LogOutputPath:         filepath.Join(baseExportPath, "export_log.log"),
				ReportPath:            filepath.Join(baseExportPath, "report.html"),
			}

				exportFolderPath = relConfigDir
				baseExportPath = relConfigDir
		} else {
			exportFolderPath = filepath.Join("export", outputPath)
			baseExportPath = filepath.Join("export", outputPath, designName)
			output = models.OutputPaths{
				OutputPath:            filepath.Join(baseExportPath, "export", versionPath),
				ExportFolderPath:      exportFolderPath,
				LowQualityWarningPath: filepath.Join(baseExportPath, "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(baseExportPath, "README.md"),
				LogOutputPath:         filepath.Join(baseExportPath, "export_log.log"),
				ReportPath:            filepath.Join(relConfigDir, "export", versionPath, "report.html"),
			}
		}

		if exportFolderPath == "" {
			log.Panicf("hmm empty export folder path ")
		}

		if filepath.Join(baseExportPath, "export", versionPath) == "" {
			log.Panicf("hmm empty export folder path ")
		}

		return output

	}*/

	// For relative config file paths, use paths relative to the config file directory
	var baseExportPath, exportFolderPath string
	if hasExportPrefix {
		exportFolderPath = relConfigDir
		baseExportPath = relConfigDir
	} else {
		exportFolderPath = filepath.Join(relConfigDir)
		baseExportPath = filepath.Join(relConfigDir, designName)
	}

	// Get the directory of the first input path for prefixing
	var inputDir string
	if len(config.GetInputPaths()) > 0 {
		inputDir = filepath.Dir(config.GetInputPaths()[0].Path)
	} else {
		inputDir = "."
	}

	outputPath := filepath.Join(baseExportPath, "export", versionPath)

	output := models.OutputPaths{
		OutputPath:            outputPath,
		ExportFolderPath:      filepath.Join(baseExportPath, exportFolderPath),
		LowQualityWarningPath: filepath.Join(baseExportPath, "LOW_QUALITY_WARNING.md"),
		ReadmePath:            filepath.Join(baseExportPath, "README.md"),
		LogOutputPath:         filepath.Join(baseExportPath, "export_log.log"),
		ReportPath:            filepath.Join(baseExportPath, inputDir, "report.html"),
	}

	log.Printf("ExportNameFormat: %+v", config.Design.ExportNameFormat)
	log.Printf("relConfigDir: %+v", relConfigDir)
	log.Printf("outputPath: %+v", outputPath)
	log.Printf("exportFolderPath: %+v", exportFolderPath)
	log.Printf("baseExportPath: %+v", baseExportPath)

	log.Printf("OutputPaths:\n\n %+v", output)

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
git tag "v2.0.11-BETA"
git push && git push --tags
```

To create a new version:

```sh
git commit -m "New and improved version"
git tag "v[NEW_VERSION_HERE]-alpha"
*/
const VERSION = "v2.5.0__2025.06.06-BETA"

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
		log.Printf("Error: %v", err)
	}
	return Version{
		OpenSCADGen: VERSION,
		OpenSCAD:    openSCADVersion.Version,
		IsOutOfDate: openSCADVersion.IsOutOfDate,
	}
}

func Process(config *models.Config, progress ProgressReporter, cancel <-chan struct{}) (models.ProcessResult, error) {
	start := time.Now()
	if config.Debug {
		logStage("=== Processing === ")
	}

	// Get output paths
	/*outputPaths := getOutputPaths(config)

	// Create export folder if it doesn't exist
	if err := os.MkdirAll(outputPaths.ExportFolderPath, 0755); err != nil {
		return models.ProcessResult{}, fmt.Errorf("Process: failed to create export folder '%s': %w", outputPaths.ExportFolderPath, err)
	}*/

	// Check if export folder has existing files
	//clearExportFolder(config, outputPaths)

	// Generate instances
	var instances []models.InstanceConfig
	var stlResults []models.GenerateSTLResult
	var allImageResults []models.GenerateImageResult
	if config.Debug {
		log.Printf("[DEBUG] Generating instances for %d configured instances and %d input paths", len(config.Design.ConfiguredInstanceConfig), len(config.GetInputPaths()))
	}

	if len(config.Design.ConfiguredInstanceConfig) == 0 {
		config.Design.ConfiguredInstanceConfig = []models.ConfiguredInstanceConfig{
			{
				Name:   "default",
				Params: map[string]interface{}{},
			},
		}
	}

	for _, dynamicInstance := range config.Design.ConfiguredInstanceConfig {
		for _, inputPath := range config.GetInputPaths() {
			if config.Debug {
				logStage(fmt.Sprintf("Generating instance %s", dynamicInstance.Name))
			}
			var err error
			var newInstances []models.InstanceConfig
			newInstances, _, err = generateInstances(config, dynamicInstance, inputPath)
			if err != nil {
				if config.ContinueOnError {
					logError(fmt.Sprintf("Warning: failed to generate instances: %v", err))
					continue
				}
				return models.ProcessResult{}, fmt.Errorf("failed to generate instances: %w", err)
			}

			newInstancesWithImages, err := populateExportImages(config, newInstances)
			if err != nil {
				if config.ContinueOnError {
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
	}

	if config.Debug {
		log.Printf("[DEBUG] Total instances generated: %d", len(instances))
	}

	errors := validateInstances(instances)
	if len(errors) > 0 {
		logError("Validation of generated instances failed:")
		for _, error := range errors {
			logError(runErrorCodeName[error.ErrorCode] + "\n" + error.Message)
			for k, v := range error.KVPs {
				LogKeyValuePair(k, v)
			}
		}
		if !config.ContinueOnError {
			if !config.Server {
				os.Exit(1)
			}
		}
	}

	for i := range instances {
		// Set PartIDLetter
		instances[i].PartIDLetter = getPartIDLetter(i)

		if instances[i].SkippedReason == "" {
			if config.Debug {
				logCreation(fmt.Sprintf("Generated instance %s", instances[i].AutoName))
				log.Printf("[DEBUG] Generating STL for instance %d: OutputPathV2=%s", i, instances[i].OutputPathV2)
			}
			var result models.GenerateSTLResult
			var err error
			if !config.OnlyImages {
				result, err = generateSTL(&instances[i], config)
				if err != nil {
					if config.ContinueOnError && !config.Server {
						logError(fmt.Sprintf("Warning: failed to generate STL for instance %s:\n Error:\n%+v", instances[i].Name, err))
						stlResults = append(stlResults, result)
						continue
					} else {
						return models.ProcessResult{}, fmt.Errorf("failed to generate STL: %w", err)
					}
				} else {
					stlResults = append(stlResults, result)
				}
			}

			if !config.OnlyExport {
				genImageResult, err := processImage(config, &instances[i])
				if err != nil {
					if config.ContinueOnError && !config.Server {
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

		} else {
			if config.Debug {
				logSkip(fmt.Sprintf("STL skipped %s %s", instances[i].AutoName, instances[i].SkippedReason))
			}
		}

		if config.Debug {
			LogKeyValuePair("ImageResults Count", fmt.Sprintf("%d", len(instances[i].ImageResults)))
		}
	}

	configDirectory := filepath.Dir(config.ConfigFile)
	exportLoc := filepath.Join(configDirectory, "export", config.Design.Version)

	if config.Debug {
		LogKeyValuePair("Process complete - ExportLocation:", exportLoc)
	}
	return models.ProcessResult{
		ExportLocation: exportLoc,
		Instances:      instances,
		STLResults:     stlResults,
		ImageResults:   allImageResults,
		TotalTimeTaken: time.Since(start),
	}, nil
}

type RunError struct {
	ErrorCode RunErrorCode
	Message   string
	KVPs      map[string]string
}

func validateInstances(instances []models.InstanceConfig) []RunError {
	if config.Debug {
		logStage("Validating instances")
	}

	instanceParamCount := make(map[string]int)
	errors := []RunError{}
	exportPaths := make(map[string]bool)
	for _, instance := range instances {
		if _, exists := exportPaths[instance.RunOutputPathV3]; exists {

			var paramStr string
			for k, v := range instance.Params {
				paramStr += fmt.Sprintf("%s=%v\n", k, v)
				instanceParamCount[k]++
			}

			LogKeyValuePair("Validation", fmt.Sprintf("%s: %+v", instance.RunOutputPathV3, exists))

			errors = append(errors, RunError{
				ErrorCode: RunErrorCode_DuplicateExportPath,
				Message:   fmt.Sprintf("Duplicate export path: \n\n\t%s. \n\n Ensure the export_name_format includes all parameters (in {curlyBrackets}) that are different between instances.", instance.OutputPathV2),
				KVPs: map[string]string{
					"exportNameFormat": config.Design.ExportNameFormat,
					"exportPaths":      paramStr,
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
			LogKeyValuePair("OverwriteExisting set, skipping check", outputPaths.ExportFolderPath)
		}
		if config.Server {
			LogKeyValuePair("Server mode, skipping check", outputPaths.ExportFolderPath)
		} else if !config.OverwriteExisting {
			logWarn(fmt.Sprintf("\nThe export folder (%s) has %d existing files: \n%s\n\n(the '-ow' flag will skip this check)\n\n(tip: if you want to keep the existing stl export files, cancel this run and update the 'version' in the config file, this will generate a new folder and keep the existing files)", outputPaths.ExportFolderPath, len(files), filesStr), false)

			logWarn(fmt.Sprintf("\n\n %d files will be deleted from: \n\n\t%s\n\nDo you want to continue? (y/n):", len(files), outputPaths.ExportFolderPath), true)

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if response != "y\n" && response != "Y\n" {
				fmt.Println("Aborting operation.")
				os.Exit(1)
			}
		} else if !strings.HasPrefix(outputPaths.ExportFolderPath, "export") {
			log.Printf("Export folder path does not start with export, skipping deletion")
			return
		}

		err := os.RemoveAll(outputPaths.ExportFolderPath)
		if err != nil {
			log.Panicf(colorRed+"Clear export folder Failed to delete export folder (outputPaths.ExportFolderPath): '%s' %s", outputPaths.ExportFolderPath, err)
		}
	}
}

func ScanFolderForConfigFiles(folder string) ([]models.ConfigFile, error) {
	var configFiles []models.ConfigFile
	maxSize := int64(2 * 1024 * 1024) // 2MB

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
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

		f, err := os.Open(path)
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
		configFiles = append(configFiles, models.ConfigFile{Path: path})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return configFiles, nil
}

var PRESET_EXPORT_IMAGES = []models.ExportCameraCoordinates{
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
		CameraName:        "nice",
		CameraCoordinates: "0,0,0,45,0,45,300",
	},
	{
		CameraName:        "nice-near",
		CameraCoordinates: "0,0,0,45,0,45,150",
	},
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

		if instances[i].SkipImages {
			if config.Debug {
				logSkip(fmt.Sprintf("Skipping images for instance %s", instances[i].AutoName))
			}
			instances[i].SkippedImageReason = "Skipping images because of skip_images flag"
			continue
		}

		for _, instance := range config.Design.ExportImages {
			if instance.ParamFilter != nil {
				for k, v := range instance.ParamFilter {
					if v != instances[i].Params[k] {
						if config.Debug {
							logSkip(fmt.Sprintf("Skipping image for instance %s because param %s does not match %s", instances[i].AutoName, k, v))
						}
						instances[i].SkippedImageReason = fmt.Sprintf("Skipping image for instance %s because param %s does not match %s", instances[i].AutoName, k, v)
						continue
					}
				}

			}
			exportImages := makePresetReplacement(instance)
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
		return parts[0], ""
	}

	if len(parts) == 2 {
		// Check if the second part is a distance keyword
		if _, ok := cameraDistances[parts[1]]; ok {
			return parts[0], parts[1]
		}
	}

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

func makePresetReplacement(exportImage models.ExportCameraCoordinates) []models.ExportCameraCoordinates {
	if exportImage.CameraName == "all" {
		return PRESET_EXPORT_IMAGES
	}

	// If custom coordinates are provided, use them directly
	if exportImage.CameraCoordinates != "" {
		return []models.ExportCameraCoordinates{
			{
				CameraName:        exportImage.CameraName,
				CameraCoordinates: exportImage.CameraCoordinates,
				ImageSize:         exportImage.ImageSize,
				ParamFilter:       exportImage.ParamFilter,
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
				CameraName:        exportImage.CameraName,
				CameraCoordinates: coordinates,
				ImageSize:         exportImage.ImageSize,
				ParamFilter:       exportImage.ParamFilter,
			},
		}
	}

	// Look for matching preset camera in the static list
	for _, preset := range PRESET_EXPORT_IMAGES {
		if preset.CameraName == exportImage.CameraName {
			return []models.ExportCameraCoordinates{
				{
					CameraName:        exportImage.CameraName,
					CameraCoordinates: preset.CameraCoordinates,
					ImageSize:         preset.ImageSize,
					ParamFilter:       exportImage.ParamFilter,
				},
			}
		}
	}

	log.Printf("Preset Export Camera Names:")
	for _, preset := range PRESET_EXPORT_IMAGES {
		log.Printf(preset.CameraName)
	}

	// If no preset found and no coordinates provided, log an error
	log.Panicf(`Camera '%s' is not a preset and has no coordinates specified.
	
	Options are listed above
	`, exportImage.CameraName)

	return []models.ExportCameraCoordinates{}
}

/*
func getPresetExportImages(config *models.Config) []models.ExportCameraCoordinates {
	allExportImages := []models.ExportCameraCoordinates{}

	// First, handle any "all" camera requests
	for _, exportImage := range config.Design.ExportImages {
		if exportImage.CameraName == "all" {
			// If "all" is specified, add all preset cameras
			allExportImages = append(allExportImages, PRESET_EXPORT_IMAGES...)
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
		for _, preset := range PRESET_EXPORT_IMAGES {
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

// loadConfig reads the configuration file and populates the Config struct
func LoadConfig(flags models.CmdFlags) (*models.Config, error) {
	var conf models.Config
	if flags.ConfigFile == "" {
		log.Printf(colorRed + "Run directly with a config file use '-c' like '-c you-project/config.toml'\n\nUse -s to run in Server mode\n\n Or Server Folder Scan mode:  -sf parent-project-folder " + colorReset)
		return nil, fmt.Errorf("no config file provided")
	}

	// Resolve config file path relative to current working directory
	configPath := flags.ConfigFile
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			log.Printf(colorRed+"Failed to resolve config file path '%s': %v", configPath, err)
			return nil, err
		}
		configPath = absPath
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf(colorRed+"Failed to read config file at path '%s'\n\n Error: %v", configPath, err)
		return nil, err
	}

	// First decode into a map to check for unmapped fields
	var metadata toml.MetaData
	metadata, err = toml.Decode(string(data), &conf)
	if err != nil {
		LogKeyValuePair("Config file", string(data))
		log.Printf(colorRed+"Config file is not valid toml: %v", err)
		return nil, err
	}

	// Check for undecoded keys
	undecoded := metadata.Undecoded()
	if len(undecoded) > 0 {
		LogKeyValuePair("Config file", flags.ConfigFile)
		for _, key := range undecoded {
			logError(fmt.Sprintf("Invalid field: %s", key.String()))
		}
		if flags.ContinueOnError {
			log.Printf(colorYellow + "Continuing on error" + colorReset)
		} else {
			return nil, fmt.Errorf("invalid fields in config: %v", undecoded)
		}
	}

	// Validate the config
	validate := validator.New()
	err = validate.Struct(conf)
	if err != nil {
		log.Printf(colorRed+"Failed to validate config: %v", err)
		return nil, err
	}

	if flags.Debug {
		log.Printf("Loaded config")
		LogKeyValuePair("Config", conf.ConfigFile)
	}

	conf.RawConfigFile = string(data)

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

	conf.ContinueOnError = flags.ContinueOnError
	conf.ConfigFile = configPath
	conf.IncludeExportLog = flags.IncludeExportLog
	conf.OverwriteExisting = flags.OverwriteExisting
	conf.Server = flags.Server
	conf.ServerFolder = flags.ServerFolder
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
			logWarn(fmt.Sprintf("ExportNameFormat contains param: \n\n -\t(%s)\n\n that is not in the params. Include every param in the export_name_format (in the format '{param_name}') to ensure all instances are generated to unique files.", paramName), true)
		}
	}

	openSCADVersion, err := findOpenSCAD()
	if err != nil {
		return nil, fmt.Errorf("failed to find OpenSCAD version: %w", err)
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
								logWarn("If more than one input is specified, the export_name_format need to include designFileName (add {designFileName} to the export_name_format)", true)
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
							logWarn(fmt.Sprintf(`Export instance name:
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
			logWarn(fmt.Sprintf("ExportNameFormat contains param (%s) that is not in the params", name), true)
		}
	}

	// Add validation for export_name_format
	if err := validateExportNameFormat(&conf); err != nil {
		return nil, fmt.Errorf("export name format validation failed: %w", err)
	}	*/

	return &conf, nil
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

		contents += fmt.Sprintf("\t- **%s**: %v\n", "InputPath", paths.InputPath)

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

// Helper function to convert a map of parameters to a map of parameter combinations
func convertToParamCombinations(params map[string]interface{}, ignoredParams map[string]bool) (map[string][]interface{}, error) {
	result := make(map[string][]interface{})
	for k, v := range params {
		// Skip ignored parameters
		if ignoredParams[k] {
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

		if strValue, ok := value.(string); ok && strings.Contains(strValue, ",") {
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

	for _, paramSet := range config.Design.ParamSets {
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

	for k, v := range inputPath.Params {
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
		// Check params
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
		// No match found
		return fmt.Sprintf("Regex pattern (%s) didn't match: (checked: %s & %s)", config.RegexPattern, configuredInstanceConfig.Name, inputPath.Path)
	}
	return ""
}

func generateAutoName(configuredInstanceConfig models.ConfiguredInstanceConfig, inputPath models.InputPath) string {
	return fmt.Sprintf("%s_%s", configuredInstanceConfig.Name, inputPath.Path)
}

func generateInstances(config *models.Config, configuredInstanceConfig models.ConfiguredInstanceConfig, inputPath models.InputPath) ([]models.InstanceConfig, string, error) {
	if config.Debug {
		logStage("=== Generating Instances === ")
	}

	if inputPath.Path == "" {
		return nil, "", fmt.Errorf("generateInstances - input path is empty")
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
	paramCombos, err := convertToParamCombinations(filteredParams, make(map[string]bool))
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
				Name:               configuredInstanceConfig.Name,
				Params:             make(map[string]interface{}),
				InputPath:          inputPath,
				SkipImages:         configuredInstanceConfig.SkipImages || inputPath.SkipImages,
				SkippedReason:      checkInstancesSkip(config, len(instances)) + checkRegexPattern(config, configuredInstanceConfig, inputPath),
				ExportImages:       []models.ExportCameraCoordinates{},
				ImageResults:       []models.GenerateImageResult{},
				RunOutputImagePath: "",
				AutoName:           generateAutoName(configuredInstanceConfig, inputPath),
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

			outputPathReplace := models.MakeFileNameReplacements(config.Design.GlobalParams, instance.Params, instance.IgnoredParams, baseExportPath, config.Design.Version, filepath.Dir(instance.InputPath.Path), config.Quality, instance.AutoName, instance.PartIDLetter)

			//	instance.OutputPathV2 = outputPath
			instance.RunOutputPathV3 = filepath.ToSlash(outputPathReplace + ".stl")
			instance.RunOutputImagePath = outputPathReplace

			for k := range ignoredParams {
				instance.IgnoredParams = append(instance.IgnoredParams, k)
			}

			instances = append(instances, instance)
		}
	}

	return instances, "NOT USED", nil
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

// hmm windows support?
func findOpenSCAD() (OpenSCADVersion, error) {
	// Try to find openscad using `which` command
	cmd := exec.Command("which", "openscad")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("OpenSCAD not found in PATH. Make sure you can run openscad from the command line")
	}

	cmdVer := exec.Command("openscad", "--version")
	outputVer, err := cmdVer.CombinedOutput()
	if err != nil {
		return OpenSCADVersion{}, fmt.Errorf("OpenSCAD not found in PATH (version check). Make sure you can run openscad from the command line: %w", err)
	}

	versionStr := strings.TrimSpace(string(outputVer))
	var versionDate time.Time
	var isOutOfDate bool

	// Try to extract date from version string - handle both YYYY.MM and YYYY.MM.DD formats
	dateRegex := regexp.MustCompile(`(\d{4})\.(\d{2})(?:\.(\d{2}))?`)
	matches := dateRegex.FindStringSubmatch(versionStr)

	if len(matches) >= 2 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])

		// Check if we have a day component
		day := 1 // Default to 1st day of month if no day specified
		if len(matches) >= 4 && matches[3] != "" {
			day, _ = strconv.Atoi(matches[3])
		}

		// Create date from year, month, and day
		versionDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

		// Check if version is more than 6 months old
		sixMonthsAgo := time.Now().AddDate(0, -6, 0)
		isOutOfDate = versionDate.Before(sixMonthsAgo)

		if isOutOfDate {
			logWarn(fmt.Sprintf("Warning: OpenSCAD version %s is more than 6 months old. Consider updating to the latest version.", versionStr), false)
		}
	} else {
		// If we can't parse the date, assume it's not dated
		isOutOfDate = false

		logCreation(fmt.Sprintf("OpenSCAD version %s is unknown", versionStr))

	}

	return OpenSCADVersion{
		Version:     versionStr,
		Path:        strings.TrimSpace(string(output)),
		IsOutOfDate: isOutOfDate,
	}, nil
}

var configTemplate = `[openscadgen]
name = "{{projectName}}"
description = ""

version = "v0.1"

export_name_format = "{designFileName}"

[[openscadgen.input_paths]]
path = "./{{projectName}}.scad"

`

func openScadTemplateExtended(projectNameUnderLined string) string {
	return fmt.Sprintf(`

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	renderType = "all"


	module %s(){
		cuboid([100,100,100]);
	}


    sliced(renderType="") {
        %s();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.3,
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
	cuboid([100,100,100]);
}
`, projectNameUnderLined)
}

func InitConfig(projectPathRaw string, extended bool) error {

	replacements := map[string]string{
		" ":  "_",
		"-":  "_",
		".":  "_",
		"__": "_",
	}

	projectPath := projectPathRaw
	for old, new := range replacements {
		projectPath = strings.ReplaceAll(projectPath, old, new)
	}
	projectPath = filepath.Clean(projectPath)

	// check if the project name is already taken
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		os.Mkdir(projectPath, 0755)
	} else {
		logError(fmt.Sprintf("Project name already taken: %s", projectPath))
		return fmt.Errorf("project name already taken")
	}

	projectName := filepath.Base(projectPath)
	projectNameLocation := filepath.Dir(projectPath)
	projectNameUnderLined := strings.NewReplacer(
		" ", "_",
	).Replace(projectName)

	configTemplate = strings.ReplaceAll(configTemplate, "{{projectName}}", projectNameUnderLined)

	configPath := filepath.Join(projectNameLocation, projectNameUnderLined, "config.toml")
	configFile, err := os.Create(configPath)
	if err != nil {
		logError(fmt.Sprintf("Failed to create config file: %s", err))
	} else {
		logCreation(fmt.Sprintf("Created config file: %s", configPath))
	}
	defer configFile.Close()
	_, err = configFile.WriteString(configTemplate)
	if err != nil {
		logError(fmt.Sprintf("Failed to write template to config file: %s", err))
	} else {
		logCreation(fmt.Sprintf("Wrote template to config file: %s", configPath))
	}

	scadPath := filepath.Join(projectNameLocation, projectNameUnderLined, projectNameUnderLined+".scad")
	scadFile, err := os.Create(scadPath)
	if err != nil {
		logError(fmt.Sprintf("Failed to create scad file: %s", err))
	}
	defer scadFile.Close()

	if extended {
		_, err = scadFile.WriteString(openScadTemplateExtended(projectNameUnderLined))
		if err != nil {
			logError(fmt.Sprintf("Failed to write template to scad file: %s", err))
		} else {
			logCreation(fmt.Sprintf("Wrote template to scad file: %s", scadPath))
		}

	} else {
		_, err = scadFile.WriteString(openScadTemplate(projectNameUnderLined))
		if err != nil {
			logError(fmt.Sprintf("Failed to write template to scad file: %s", err))
		} else {
			logCreation(fmt.Sprintf("Wrote template to scad file: %s", scadPath))
		}

	}

	logCreation(fmt.Sprintf("Project Successfully Initialized: %s", projectName))
	LogKeyValuePair("Project Path", projectPath)
	return nil
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

	if config.IncludeExportLog {
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
	} else {
		// Ensure we still log to console even when not logging to file
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	}

	return nil
}

func LogKeyValuePair(key string, value string) {
	logger.Printf(colorYellow+"%s: "+colorWhite+"\t\t\t%s"+colorReset, key, value)
}

func logSkip(message string) {
	logger.Printf(colorYellow+"%s"+colorReset, message)
}

func logWarn(message string, critical bool) {
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
		logWarn(fmt.Sprintf("[SetMetadata] warning: file '%s' does not exist", fileName), false)
		return fmt.Errorf("warning: file '%s' does not exist", fileName)
	} else if err != nil {
		logWarn(fmt.Sprintf("[SetMetadata] warning: error accessing file '%s': %v", fileName, err), false)
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
				logWarn(fmt.Sprintf("warning: error setting xattr '%s' on file '%s': %v", key, fileName, err), false)
				return fmt.Errorf("error setting xattr '%s' on file '%s': %v", key, fileName, err)
			}
			if config.Debug {
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
				logWarn(fmt.Sprintf("warning: error opening ADS '%s': %v", adsName, err), false)
				return fmt.Errorf("error opening ADS '%s': %v", adsName, err)
			}
			defer file.Close()

			_, err = file.Write([]byte(value))
			if err != nil {
				logWarn(fmt.Sprintf("warning: error writing to ADS '%s': %v", adsName, err), false)
				return fmt.Errorf("error writing to ADS '%s': %v", adsName, err)
			}
			if config.Debug {
				LogKeyValuePair("Set ADS", adsName)
			}
			fmt.Printf("Set ADS '%s' on file '%s' with value: %s\n", key, fileName, value)
		}
	default:
		logWarn(fmt.Sprintf("warning: unsupported operating system: %s", currentOS), false)
		return fmt.Errorf("unsupported operating system: %s", currentOS)
	}

	return nil
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
	}

	// Validate output path for invalid characters
	if !isValidPath(instance.OutputPathV2, invalidChars) {
		errMsg := fmt.Sprintf("Output path '%s' contains invalid character(s) from '%s'. Please update your export_name_format in config.", instance.OutputPathV2, invalidChars)
		result.Error = errMsg
		if config.ContinueOnError {
			return result, nil
		}
		return result, fmt.Errorf(errMsg)
	}

	// Ensure output directory exists before proceeding
	outputDir := filepath.Dir(instance.OutputPathV2)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create output directory %s: %v", outputDir, err)
		if config.ContinueOnError {
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
		LogKeyValuePair("Applied parameters for STL generation:", "")
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
	os.MkdirAll(filepath.Dir(instance.RunOutputPathV3), 0755)

	// Build command arguments
	args := []string{
		"-o", fmt.Sprintf("'%s'", instance.RunOutputPathV3),
	}

	//check for ? charcters
	if !isValidPath(instance.RunOutputPathV3, invalidChars) {
		errMsg := fmt.Sprintf("Output path '%s' contains invalid character(s) from '%s'. Please update your export_name_format in config.", instance.OutputPathV2, invalidChars)
		result.Error = errMsg
		if config.ContinueOnError {
			return result, nil
		}
		return result, fmt.Errorf(errMsg)
	}
	if !config.SkipRender {
		args = append(args, "--render")
	}

	if !config.Design.DontUseManifold {
		args = append(args, "--backend=manifold")
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
	if !config.Quiet {
		logStage(fmt.Sprintf("Running STL generation: %s", instance.Name))
		LogKeyValuePair("Command", commandStr)
	}

	// Run command through shell
	cmd := exec.Command("sh", "-c", commandStr)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	result.Command = commandStr
	result.TimeTaken = time.Since(startTime)

	if err != nil {
		result.Error = fmt.Sprintf("error running openscad: %v\nOutput: %s", err, string(output))
		if config.Debug {
			log.Printf("OpenSCAD command failed with error: %v", err)
			log.Printf("Command output: %s", string(output))
		}
		if config.ContinueOnError {
			return result, nil
		}
		return result, fmt.Errorf("OpenSCAD command failed: %w\nOutput: %s", err, string(output))
	}

	// Check if file was created and has content
	if fileInfo, err := os.Stat(instance.RunOutputPathV3); os.IsNotExist(err) || fileInfo.Size() == 0 {
		result.Error = fmt.Sprintf("output file was not created or is empty: %s", instance.OutputPathV2)
		if config.Debug {
			log.Printf("Output file was not created or is empty at: %s", instance.OutputPathV2)
		}
		if config.ContinueOnError {
			return result, nil
		}
		return result, fmt.Errorf("output file was not created or is empty: %s", instance.OutputPathV2)
	} else {
		logCreation(fmt.Sprintf("STL created in %s at %s", result.TimeTaken, instance.RunOutputPathV3))
	}

	result.OutputPath = instance.OutputPathV2
	return result, nil
}

func FindOpenSCAD() string {
	// Try to find openscad using `which` command
	cmd := exec.Command("which", "openscad")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("OpenSCAD not found in PATH.")
	}
	return strings.TrimSpace(string(output))
}

func GenerateOutputReport(config *models.Config, instances []models.InstanceConfig, stlResults []models.GenerateSTLResult, imageResults []models.GenerateImageResult, outputDir string, toFile bool) (templ.Component, string, error) {
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
	htmlContent := templates.Report(config, instances, outputFile, stlResults, imageResults, allParamNames, false, "")

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

/*
func processInstance(config *models.Config, instance models.InstanceConfig) error {
	if config.Debug {
		logStage("=== Processing Instance === ")
		log.Printf("Processing instance: %s", instance.Name)
	}

	// Skip if instance is marked to be skipped
	if instance.SkippedReason != "" {
		log.Printf("Skipping instance %s: %s", instance.Name, instance.SkippedReason)
		return nil
	}

	// Generate OpenSCAD command
	cmd := generateOpenSCADCommand(config, &instance)
	if config.Debug {
		log.Printf("OpenSCAD command: %s", cmd)
	}

	// Execute OpenSCAD command
	if err := executeCommand(cmd); err != nil {
		return fmt.Errorf("error executing OpenSCAD command: %v", err)
	}
	/*
		// Process images if not skipped
		if !instance.SkipImages {
			_, err := processImage(config, &instance)
			if err != nil {
				return fmt.Errorf("error processing images: %v", err)
			}
		}

	return nil
}

func generateOpenSCADCommand(config *models.Config, instance *models.InstanceConfig) string {
	args := []string{"-o", instance.RunOutputPathV3}

	if config.Quiet {
		args = append(args, "-q")
	}

	// Add parameters to command
	for name, value := range instance.Params {
		if reflect.TypeOf(value).Kind() == reflect.String && value != "true" && value != "false" {
			args = append(args, "-D", fmt.Sprintf("'%s=\"%v\"'", name, value))
		} else {
			args = append(args, "-D", fmt.Sprintf("'%s=%v'", name, value))
		}
	}

	if config.IncludePartIDLetter || !config.Design.NoPartIDLetter {
		args = append(args, "-D", fmt.Sprintf("'part_id_letter=\"%s\"'", instance.PartIDLetter))
	}

	if config.OverrideFN > 0 {
		args = append(args, "-D", fmt.Sprintf("'$fn=%d'", config.OverrideFN))
	}

	// Get the absolute path of the input file
	absPath := getAbsPath(config.ConfigFile, instance.InputPath.Path)
	args = append(args, fmt.Sprintf("\"%s\"", absPath))

	if !config.SkipRender {
		args = append(args, "--render")
	}

	if !config.Design.DontUseManifold {
		args = append(args, "--backend=manifold")
	}

	command := "openscad"
	if config.CustomOpenSCADCommand != "" {
		command = config.CustomOpenSCADCommand
	}

	return fmt.Sprintf("%s %s", command, strings.Join(args, " "))
}*/

func executeCommand(cmd string) error {
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		return fmt.Errorf("error executing command: %w", err)
	}
	return nil
}

func processImage(config *models.Config, instance *models.InstanceConfig) ([]models.GenerateImageResult, error) {

	var imageResults []models.GenerateImageResult

	if instance.SkipImages {
		return imageResults, nil
	}

	// Create a dummy STL result for image generation
	stlResult := models.GenerateSTLResult{
		InstanceConfig: *instance,
		AppliedParams:  instance.Params,
	}

	for _, camera := range instance.ExportImages {
		if len(instance.SkippedImageReason) > 0 {
			logSkip(fmt.Sprintf("Skipping image %s", instance.SkippedImageReason))

		} else {
			result, err := generateImage(instance, config, camera, stlResult)
			if err != nil {
				logError(fmt.Sprintf("Error generating image: %v", err))
				return imageResults, fmt.Errorf("error generating image: %w", err)
			}
			if config.Debug {
				LogKeyValuePair("Generated Image", result.OutputPath)
			}

			imageResults = append(imageResults, result)
		}

	}

	return imageResults, nil
}

func generateImage(instance *models.InstanceConfig, config *models.Config, camera models.ExportCameraCoordinates, stlResult models.GenerateSTLResult) (models.GenerateImageResult, error) {
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

	outputImgPath := instance.RunOutputImagePath + "-" + camera.CameraName + ".png"
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
		return imageResult, fmt.Errorf("OpenSCAD not found")
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
		"--projection", "perspective",
		"--camera", camera.CameraCoordinates,
		"--preview",
		"--backend=manifold",
		"--enable=fast-csg",
	}

	// Add custom OpenSCAD arguments if provided
	if config.Design.CustomOpenSCADArgs != "" {
		args = append(args, strings.Split(config.Design.CustomOpenSCADArgs, " ")...)
	}

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

	// Create the command
	cmd := exec.Command(openscadCmd, args...)

	if !config.Quiet {
		logStage("Running Image generation")
		LogKeyValuePair("Command", cmd.String())
	}

	// Run the command
	if err := executeCommand(cmd.String()); err != nil {
		return imageResult, fmt.Errorf("error running OpenSCAD: %w", err)
	} else {
		logCreation("Image generation completed")
		if config.Debug {
			LogKeyValuePair("Image Created", outputImgPath)
		}
	}

	// Update result with timing information
	imageResult.TimeTaken = time.Since(startTime)
	imageResult.OutputPath = outputImgPath

	return imageResult, nil
}

type ProgressReporter interface {
	Update(msg string)
	Done()
	Error(err error)
}

type NoopProgress struct{}

func (n *NoopProgress) Update(msg string) {}
func (n *NoopProgress) Done()             {}
func (n *NoopProgress) Error(err error)   {}

type ChanProgress struct {
	Updates chan<- string
}

func (c *ChanProgress) Update(msg string) { c.Updates <- msg }
func (c *ChanProgress) Done()             { c.Updates <- "done" }
func (c *ChanProgress) Error(err error)   { c.Updates <- "error: " + err.Error() }
