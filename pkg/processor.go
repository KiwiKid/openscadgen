package pkg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
	"github.com/pkg/xattr"
)

func getOutputPaths(config *models.Config) models.OutputPaths {
	// Get the directory containing the config file - this is our anchor point
	configDir := filepath.Dir(config.ConfigFile)

	// Determine the input path, checking relative to the config file first
	absInputPath := filepath.Join(configDir, config.Design.InputPath)

	// If file doesn't exist at config-relative path, try absolute path
	if _, err := os.Stat(absInputPath); os.IsNotExist(err) {
		absInputPath, err = filepath.Abs(config.Design.InputPath)
		if err != nil {
			log.Panicf("Could not resolve absolute path: %s", err)
		}
	}

	versionPathSafe := strings.ReplaceAll(config.Design.Version, ".", "_")

	if config.Design.OutputPath != "" {
		// When output path is explicitly specified, use it relative to config dir
		absOutputPath := filepath.Join(configDir, config.Design.OutputPath)

		if config.Debug {
			log.Printf("Output path specified in config: %s", absOutputPath)
		}

		if !config.Quiet {
			logKeyValuePair("Version", config.Design.Version)
			logKeyValuePair("getOutputPaths:Output path", filepath.Join(absOutputPath, config.Design.Version))
		}

		return models.OutputPaths{
			OutputPath:            filepath.Join(absOutputPath, "export", versionPathSafe),
			ExportFolderPath:      filepath.Join(absOutputPath, "export", versionPathSafe),
			LowQualityWarningPath: filepath.Join(absOutputPath, "export", versionPathSafe, "LOW_QUALITY_WARNING.md"),
			ReadmePath:            filepath.Join(absOutputPath, "export", versionPathSafe, "README.md"),
			LogOutputPath:         filepath.Join(absOutputPath, "export", versionPathSafe, "export_log.log"),
			ReportPath:            filepath.Join(absOutputPath, "export", versionPathSafe, "report.html"),
		}
	}

	// If output_path is not specified, derive it based on the input path, but still relative to config dir
	// Get design name from input path
	designFilename := filepath.Base(absInputPath)
	designName := strings.TrimSuffix(designFilename, filepath.Ext(designFilename))

	// Match the path structure expected by the tests
	if filepath.IsAbs(config.ConfigFile) {
		// For absolute config file paths, maintain the original structure
		return models.OutputPaths{
			OutputPath:            filepath.Join(filepath.Dir(absInputPath), "export", versionPathSafe, designName),
			ExportFolderPath:      filepath.Join(filepath.Dir(absInputPath), "export", versionPathSafe),
			LowQualityWarningPath: filepath.Join(filepath.Dir(absInputPath), "export", versionPathSafe, designName, "LOW_QUALITY_WARNING.md"),
			ReadmePath:            filepath.Join(filepath.Dir(absInputPath), "export", versionPathSafe, designName, "README.md"),
			LogOutputPath:         filepath.Join(filepath.Dir(absInputPath), "export", versionPathSafe, designName, "export_log.log"),
			ReportPath:            filepath.Join(filepath.Dir(absInputPath), "export", versionPathSafe, designName, "report.html"),
		}
	} else {
		// For relative config file paths, use paths relative to the config file directory
		return models.OutputPaths{
			OutputPath:            filepath.Join(configDir, "export", versionPathSafe),
			ExportFolderPath:      filepath.Join(configDir, "export", versionPathSafe),
			LowQualityWarningPath: filepath.Join(configDir, "export", versionPathSafe, "LOW_QUALITY_WARNING.md"),
			ReadmePath:            filepath.Join(configDir, "export", versionPathSafe, "README.md"),
			LogOutputPath:         filepath.Join(configDir, "export", versionPathSafe, "export_log.log"),
			ReportPath:            filepath.Join(configDir, "export", versionPathSafe, "report.html"),
		}
	}
}

const OPENSCAD_VERSION_WARN_IF_OLDER_THAN = 2024

var config models.Config

var logger *log.Logger

var logBuffer bytes.Buffer
var logToMemory bool

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
const VERSION = "v2.2.4-BETA"

type Version struct {
	OpenSCADGen string
	OpenSCAD    string
}

func GetVersion() Version {
	return Version{
		OpenSCADGen: VERSION,
		OpenSCAD:    findOpenSCAD(),
	}
}

func Process(config *models.Config) error {

	// Get output paths
	outputPaths := getOutputPaths(config)
	if config.Debug {
		logStage("Output paths")
		logKeyValuePair("Output path", outputPaths.OutputPath)
		logKeyValuePair("Export folder path", outputPaths.ExportFolderPath)
		logKeyValuePair("Low quality warning path", outputPaths.LowQualityWarningPath)
		logKeyValuePair("Readme path", outputPaths.ReadmePath)
		logKeyValuePair("Log output path", outputPaths.LogOutputPath)
		logKeyValuePair("Report path", outputPaths.ReportPath)
	}

	clearExportFolder(config, outputPaths)

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
		for _, inputPath := range config.Design.InputPaths {
			newInstances, err := generateInstances(config, dynamicInstance, inputPath, outputPaths.ExportFolderPath)
			if err != nil {
				return fmt.Errorf("failed to generate instances: %w", err)
			}
			instances = append(instances, newInstances...)
		}
	}

	// Generate STL files
	var stlResults []models.GenerateSTLResult
	for i, instance := range instances {
		// Set PartIDLetter
		instance.PartIDLetter = getPartIDLetter(i)

		// Set OutputPathV2 using the instance's parameters
		exportNameFormat := config.Design.ExportNameFormat
		if exportNameFormat == "" {
			log.Panicf("exportNameFormat: '%s' is invalid, could not get formatToUse", exportNameFormat)
		} else if config.Debug {
			logKeyValuePair("generate-instance:exportNameFormat", exportNameFormat)
		}

		result, err := generateSTL(&instance, config, outputPaths.ExportFolderPath)
		if err != nil {
			return fmt.Errorf("failed to generate STL: %w", err)
		}

		if result.Error != "" {
			logWarn(fmt.Sprintf("Warning: %s", result.Error), false)
		}

		stlResults = append(stlResults, result)
	}

	config.Design.ExportImages = getPresetExportImages(config.Design.ExportImages)

	// Generate images if configured
	var imageResults []models.GenerateImageResult
	if config.Design.ExportImages != nil && len(config.Design.ExportImages) > 0 {
		for _, instance := range instances {
			if instance.SkipImages {
				continue
			}
			for _, camera := range config.Design.ExportImages {
				result, err := generateImage(&instance, config, outputPaths.ExportFolderPath, camera)
				if err != nil {
					if config.ContinueOnError {
						logWarn(fmt.Sprintf("Warning: failed to generate image: %v", err), false)
						continue
					}
					return fmt.Errorf("failed to generate image: %w", err)
				}

				if result.Error != "" {
					logWarn(fmt.Sprintf("Warning: %s", result.Error), false)
				}

				instance.ImageResults = append(instance.ImageResults, result)

				imageResults = append(imageResults, result)
			}
		}
	}

	// Generate output report
	if err := GenerateOutputReport(config, instances, outputPaths, stlResults, imageResults); err != nil {
		return fmt.Errorf("failed to generate output report: %w", err)
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

		/*		// get the absolute path
				absPath, err := filepath.Abs(outputPaths.ExportFolderPath)
				if err != nil {
					logWarn(fmt.Sprintf("Could not get absolute path for export folder: %s", err), true)
					os.Exit(1)
				}
		*/
		if config.NoProcessing {
			log.Printf(colorBlue + "No processing requested, skipping export folder actions" + colorReset)
			return
		} else if !config.Quiet {
			logKeyValuePair("[processing] Export folder", outputPaths.ExportFolderPath)
		}

		if !config.OverwriteExisting {
			logWarn(fmt.Sprintf("\nThe export folder (%s) has %d existing files: \n%s\n\n(the '-ow' flag will skip this check)\n\n(tip: if you want to keep the existing stl export files, cancel this run and update the 'version' in the config file, this will generate a new folder and keep the existing files)", outputPaths.ExportFolderPath, len(files), filesStr), false)

			logWarn(fmt.Sprintf(" %d files will be deleted from: \n\n\t%s\n\nDo you want to continue? (y/n):", len(files), outputPaths.ExportFolderPath), true)

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if response != "y\n" && response != "Y\n" {
				fmt.Println("Aborting operation.")
				os.Exit(1)
			}
		} else if config.Debug {
			logKeyValuePair("OverwriteExisting set, skipping check", outputPaths.ExportFolderPath)
		}

		err := os.RemoveAll(outputPaths.ExportFolderPath)
		if err != nil {
			log.Panicf(colorRed+"Failed to delete export folder: %s", err)
		}
	}
}

var PRESET_EXPORT_IMAGES = []models.ExportCameraCoordinates{
	{
		CameraName:        "top",
		CameraCoordinates: "0,0,0,0,0,0,300",
	},
	{
		CameraName:        "bottom",
		CameraCoordinates: "0,0,0,180,0,0,300",
	},
	{
		CameraName:        "front",
		CameraCoordinates: "0,0,0,90,0,0,300",
	},
	{
		CameraName:        "back",
		CameraCoordinates: "0,0,0,270,0,0,300",
	},
	{
		CameraName:        "left",
		CameraCoordinates: "0,0,0,0,90,0,300",
	},
	{
		CameraName:        "right",
		CameraCoordinates: "0,0,0,0,0,0,300",
	},
}

func getPresetExportImages(exportImages []models.ExportCameraCoordinates) []models.ExportCameraCoordinates {
	allExportImages := []models.ExportCameraCoordinates{}

	// First, handle any "all" camera requests
	for _, exportImage := range exportImages {
		if exportImage.CameraName == "all" {
			// If "all" is specified, add all preset cameras
			allExportImages = append(allExportImages, PRESET_EXPORT_IMAGES...)
			break
		}
	}

	// Then process each export image request
	for _, exportImage := range exportImages {
		// Skip "all" as it's already handled
		if exportImage.CameraName == "all" {
			continue
		}

		// If custom coordinates are provided, use them directly
		if exportImage.CameraCoordinates != "" {
			allExportImages = append(allExportImages, exportImage)
			continue
		}

		// Look for matching preset camera
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
	}

	return allExportImages
}

// loadConfig reads the configuration file and populates the Config struct
func LoadConfig(flags models.CmdFlags) (*models.Config, error) {
	var conf models.Config
	if flags.ConfigFile == "" {
		log.Printf(colorRed + "No config file provided - use '-c' like '-c you-project/config.toml' to specify a config file" + colorReset)
		return nil, fmt.Errorf("no config file provided")
	}
	data, err := os.ReadFile(flags.ConfigFile)
	if err != nil {
		log.Printf(colorRed+"Failed to read config file at path '%s': %v", flags.ConfigFile, err)
		return nil, err
	}

	// Decode into a map to check for unmapped fields
	var rawData map[string]interface{}
	_, err = toml.Decode(string(data), &rawData)
	if err != nil {
		log.Printf(colorRed+"Failed to unmarshal config: %v", err)
		return nil, err
	}

	_, err = toml.Decode(string(data), &conf)
	if err != nil {
		log.Printf(colorRed+"Failed to unmarshal config: %v", err)
		return nil, err
	}

	// toml validation
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(conf)
	if err != nil {
		log.Printf(colorRed+"Config Validation failed: %v", err)
		return nil, err
	}

	if !flags.Debug {
		log.Printf("Loaded config")
	}

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
	conf.FullExport = flags.FullExport
	conf.ConfigFile = flags.ConfigFile
	conf.IncludeExportLog = flags.IncludeExportLog
	conf.OverwriteExisting = flags.OverwriteExisting

	if flags.CustomOpenSCADOutputFormat != "" {
		conf.Design.CustomOpenSCADOutputFormat = flags.CustomOpenSCADOutputFormat
	}

	config.Debug = flags.Debug

	config.Quality = ""
	if flags.OverrideFN > 0 {
		conf.OverrideFN = flags.OverrideFN
		config.Quality = fmt.Sprintf("fn-%d", flags.OverrideFN)
	} else if flags.HighQuality {
		conf.OverrideFN = 200
		config.Quality = "high"
	} else if flags.LowQuality {
		conf.OverrideFN = 20
		config.Quality = "low"
	}

	if config.Design.Version == "" {
		config.Design.Version = "v0.1"
	}

	exportNameFormat := getExportNameFormat(&conf)

	exportNameFormatParams := getExportNameFormatParams(exportNameFormat)

	if flags.Debug {
		logStage("DEBUG Validating: export_name_format params")
	}
	for _, paramName := range exportNameFormatParams {
		if config.Debug {
			logKeyValuePair("Param name to confirm", paramName)
			logKeyValuePair("ExportNameFormat", exportNameFormat)
		}
		if !strings.Contains(conf.Design.ExportNameFormat, paramName) {
			logWarn(fmt.Sprintf("ExportNameFormat contains param (%s) that is not in the params", paramName), true)
		}
	}

	config.OpenSCADVersion = findOpenSCAD()
	config.OpenScadGenVersion = VERSION
	/* (temp disabled instance validation for now)
		if conf.Design.ConfiguredInstanceConfig != nil {
			if len(conf.Design.ConfiguredInstanceConfig) > 0 {
				for dynamicInstanceIndex, dynamicInstance := range conf.Design.ConfiguredInstanceConfig {
					for paramName, paramValue := range dynamicInstance.Params {
						if config.Debug {
							logKeyValuePair("LoadConfig:Param name", paramName)
							logKeyValuePair("LoadConfig:Param value", fmt.Sprintf("%v", paramValue))
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
								logKeyValuePair("ExportNameFormat missing {designFileName}", exportNameFormat)
								logKeyValuePair("from config file:", flags.ConfigFile)
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
							logKeyValuePair("Dynamic instance index 1", fmt.Sprintf("%d", dynamicInstanceIndex))
							logKeyValuePair("Export Name Format", exportNameFormat)
							logKeyValuePair("Missing Param name", paramName)
							logKeyValuePair("Param value", fmt.Sprintf("%v", paramValue))
							logKeyValuePair("Config file", flags.ConfigFile)
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
			logKeyValuePair("Param name to confirm", paramName)
			logKeyValuePair("ExportNameFormat", conf.Design.ExportNameFormat)
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
			} else if strings.Contains(fmt.Sprintf("%v", paramValue), ",") || strings.Contains(fmt.Sprintf("%v", paramValue), "-") {
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
		logKeyValuePair("LOW_QUALITY_WARNING.md written to", outputPath)
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
	contents += fmt.Sprintf("This README was generated by [openscadgen](https://github.com/KiwiKid/openscadgen) %s %s. The free, local, open source openscad stl release generator.\n", version, openscadVersion)

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
			logKeyValuePair("README.md written to", readmePath)
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
		log.Printf("Export name format is not set, defaulting to '{designFileName}'")
		exportNameFormat = "{designFileName}"
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
			if strings.Contains(val, "-") {
				// Handle numeric ranges
				rangeValues := strings.Split(val, "-")
				if len(rangeValues) != 2 {
					return nil, fmt.Errorf("invalid range format: %s", val)
				}
				start, err := strconv.ParseFloat(rangeValues[0], 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse start value: %s", err)
				}
				end, err := strconv.ParseFloat(rangeValues[1], 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse end value: %s", err)
				}
				for i := start; i <= end; i++ {
					parsedValues = append(parsedValues, float64(i))
				}
			} else if val == "true" || val == "false" {
				// Handle boolean values
				parsedValues = append(parsedValues, val == "true")
			} else if num, err := strconv.ParseFloat(val, 64); err == nil {
				// Handle numeric values
				parsedValues = append(parsedValues, num)
			} else {
				// Handle string values
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
				parsedValues = append(parsedValues, val)
			}
			globalParamsMap[key] = parsedValues
		} else {
			globalParamsMap[key] = []interface{}{value}
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
		params[k] = v
	}

	for k, v := range inputPath.Params {
		params[k] = v
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

	return params, globalParamsMap, ignoredKeys
}

func checkInstancesSkip(config *models.Config, countSoFar int) string {
	if config.MaxInstances > 0 && countSoFar >= config.MaxInstances {
		return fmt.Sprintf("Max instances reached: %d", config.MaxInstances)
	}
	return ""
}

func generateInstances(config *models.Config, configuredInstanceConfig models.ConfiguredInstanceConfig, inputPath models.InputPath, exportFolderPath string) ([]models.InstanceConfig, error) {
	if config.Debug {
		logStage("=== Generating Instances === ")
	}

	if inputPath.Path == "" {
		return nil, fmt.Errorf("generateInstances - input path is empty")
	}

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
		return nil, fmt.Errorf("error converting parameters: %v", err)
	}

	// Generate all possible combinations
	var instances []models.InstanceConfig

	// If no parameters have multiple values, create a single instance
	if len(paramCombos) == 0 && len(filteredGlobalParamsMap) == 0 {
		instance := models.InstanceConfig{
			Name:          configuredInstanceConfig.Name,
			Params:        make(map[string]interface{}),
			InputPath:     inputPath,
			SkipImages:    configuredInstanceConfig.SkipImages,
			SkippedReason: checkInstancesSkip(config, len(instances)),
		}

		// Add all non-ignored parameters
		for k, v := range filteredParams {
			instance.Params[k] = v
		}

		// Add required parameters
		instance.Params["designFileName"] = strings.TrimSuffix(filepath.Base(inputPath.Path), filepath.Ext(inputPath.Path))
		instance.Params["instanceName"] = configuredInstanceConfig.Name

		instance.Params["version"] = config.Design.Version

		// Create version subdirectory
		//		versionDir := strings.ReplaceAll(config.Design.Version, ".", "_")

		// Set output path
		instance.OutputPathV2 = filepath.Join(exportFolderPath, formatExportName(config.Design.ExportNameFormat, instance.Params, ignoredParams))

		for k, _ := range ignoredParams {
			instance.IgnoredParams = append(instance.IgnoredParams, k)
		}

		return []models.InstanceConfig{
			instance,
		}, nil
	}

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
				Name:          configuredInstanceConfig.Name,
				Params:        make(map[string]interface{}),
				InputPath:     inputPath,
				SkipImages:    configuredInstanceConfig.SkipImages,
				SkippedReason: checkInstancesSkip(config, len(instances)),
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

			// Add required parameters
			instance.Params["designFileName"] = strings.TrimSuffix(filepath.Base(inputPath.Path), filepath.Ext(inputPath.Path))
			instance.Params["name"] = configuredInstanceConfig.Name
			instance.Params["version"] = config.Design.Version

			// Create version subdirectory
			//versionDir := strings.ReplaceAll(config.Design.Version, ".", "_")

			if config.Debug {
				logStage("=== Generate OutputPathV2 === ")
				logKeyValuePair("exportFolderPath", exportFolderPath)
				logKeyValuePair("ExportNameFormat", config.Design.ExportNameFormat)
				logKeyValuePair("ExportNameFormat", config.Design.OutputPath)
				logKeyValuePair("formatExportName", formatExportName(config.Design.ExportNameFormat, instance.Params, ignoredParams))
				logKeyValuePair("instance.Params", fmt.Sprintf("%v", instance.Params))
				logKeyValuePair("instance.IgnoredParams", fmt.Sprintf("%v", instance.IgnoredParams))
			}
			instance.OutputPathV2 = filepath.Join(exportFolderPath, formatExportName(config.Design.ExportNameFormat, instance.Params, ignoredParams))

			instances = append(instances, instance)
		}
	}

	if config.Debug {
		logStage(fmt.Sprintf("=== Finished Generating %d Instances === ", len(instances)))
	}

	return instances, nil
}

// Helper function to check if a parameter has multiple values
func paramHasMultipleValues(value interface{}) bool {
	if strValue, ok := value.(string); ok {
		return strings.Contains(strValue, ",") || strings.Contains(strValue, "-")
	}
	return false
}

// formatExportName replaces placeholders in the format string with actual parameter values
func formatExportName(format string, params map[string]interface{}, ignoredParams map[string]bool) string {
	result := format

	logStage("=== formatExportName === ")
	log.Printf("formatExportName:format: %s", format)
	log.Printf("formatExportName:params: %v", params)
	log.Printf("formatExportName:ignoredParams: %v", ignoredParams)

	// Replace version placeholder with sanitized version
	if version, ok := params["version"].(string); ok {
		result = strings.ReplaceAll(result, "{version}", strings.ReplaceAll(version, ".", "_"))
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

	log.Printf("formatExportName:result: %s", result)
	return result + ".stl"
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

func getAllInputPaths(config *models.Config) []models.InputPath {
	if len(config.Design.InputPaths) > 0 {
		return config.Design.InputPaths
	}
	return []models.InputPath{{Path: config.Design.InputPath}}
}

func getAbsPath(configFile, inputPath string) string {
	// Check if the input path exists relative to config file
	configDir := filepath.Dir(configFile)
	absPath := filepath.Join(configDir, inputPath)

	// If file doesn't exist at config-relative path, try absolute/working dir
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		absPath, err = filepath.Abs(inputPath)
		if err != nil {
			log.Printf("Could not resolve absolute path: %s", err)
			return inputPath
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Printf("Input path does not exist: %s", inputPath)
		os.Exit(1)
	}

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

// hmm windows support?
func findOpenSCAD() string {
	// Try to find openscad using `which` command
	cmd := exec.Command("which", "openscad")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("OpenSCAD not found in PATH.")
	}
	return strings.TrimSpace(string(output))
}

var configTemplate = `[openscadgen]
name = "{{projectName}}"
description = ""

output_path = "./export/"
version = "v0.1"

export_name_format = "{designFileName}"

# export_name_format = "custom-export-name-for-this-file/{designFileName}-{name}"

[[openscadgen.input_paths]]
path = "./{{projectName}}.scad"

# [[openscadgen.input_paths]]
# path = "./a-second-related-openscad-file.scad"
# ignore_param_when_processing = "param3" # dont process param3 when generating instances for this file 

# [[openscadgen.instances]]
# params = {renderType = "horzSlice,vertSlice,all", param1 = "a-custom-value" }

# [[openscadgen.instances]]
# params = { renderType = "horzSlice,vertSlice,all", param1 = "a-diff-custom-value" }
`

func openScadTemplateExtended(projectNameUnderLined string) string {
	return fmt.Sprintf(`

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;


	module test_create(){
		cuboid([100,100,100]);
	}


    sliced(renderType="horzSlice") {
        test_create();
    }
        
       sliced(renderType=renderType){
        hand_press_plate();
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

var openScadTemplate = `
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;
`

func InitConfig(projectName string, extended bool) error {

	// check if the project name is already taken
	if _, err := os.Stat(projectName); os.IsNotExist(err) {
		return fmt.Errorf("project name already taken")
	}

	os.Mkdir(projectName, 0755)

	configTemplate = strings.ReplaceAll(configTemplate, "{{projectName}}", projectName)

	os.Create(filepath.Join(projectName, "config.toml"))
	os.WriteFile(filepath.Join(projectName, "config.toml"), []byte(configTemplate), 0644)

	os.Create(filepath.Join(projectName, projectName+".scad"))

	if extended {
		projectNameUnderLined := strings.NewReplacer(
			"-", "_",
			" ", "_",
			".", "_",
		).Replace(projectName)
		os.WriteFile(filepath.Join(projectName, projectName+".scad"), []byte(openScadTemplateExtended(projectNameUnderLined)), 0644)
	} else {
		os.WriteFile(filepath.Join(projectName, projectName+".scad"), []byte(openScadTemplate), 0644)
	}

	return nil
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

func logKeyValuePair(key string, value string) {
	logger.Printf(colorYellow+"%s: "+colorWhite+"\t\t\t%s"+colorReset, key, value)
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

func logStage(stage string) {
	logger.Printf(colorBlue+"%s"+colorReset, stage)
}

func getOrMakeExportFolder(config *models.Config, outputPaths models.OutputPaths) {
	if config.Debug {
		logStage("Getting or making export folder")
	}
	designFileName := strings.Split(config.Design.InputPath, "/")[len(strings.Split(config.Design.InputPath, "/"))-1]

	outputPath := config.Design.OutputPath
	if outputPath == "" {
		outputPath = path.Join("./", designFileName)
	}

	exportFolderPath := strings.Clone(outputPath)
	if !strings.HasSuffix(outputPath, "/export") && !strings.HasSuffix(outputPath, "/export/") {
		exportFolderPath = path.Join(exportFolderPath, "export")
	}

	// Check if exportFolderPath has any files or directories
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

		// get the absolute path
		absPath, err := filepath.Abs(outputPaths.ExportFolderPath)
		if err != nil {
			logWarn(fmt.Sprintf("Could not get absolute path for export folder: %s", err), true)
			os.Exit(1)
		}

		if config.NoProcessing {
			log.Printf(colorBlue + "No processing requested, skipping export folder actions" + colorReset)
			return
		} else if !config.Quiet {
			logKeyValuePair("[processing] Export folder", absPath)
		}

		if !config.OverwriteExisting {
			logWarn(fmt.Sprintf("\nThe export folder (%s) has %d existing files: \n%s\n\n(the '-ow' flag will skip this check)\n\n(tip: if you want to keep the existing stl export files, cancel this run and update the 'version' in the config file, this will generate a new folder and keep the existing files)", outputPaths.ExportFolderPath, len(files), filesStr), false)

			logWarn(fmt.Sprintf(" %d files will be deleted from: \n\n\t%s\n\nDo you want to continue? (y/n):", len(files), absPath), true)

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if response != "y\n" && response != "Y\n" {
				fmt.Println("Aborting operation.")
				os.Exit(1)
			}
		} else if config.Debug {
			logKeyValuePair("OverwriteExisting set, skipping check", absPath)
		}

		if !config.Quiet {
			logStage(fmt.Sprintf("Clearing %d files from export folder", len(files)))
		}
		remRrr := os.RemoveAll(absPath)
		if config.Debug {
			logKeyValuePair("Removed files from export folder", absPath)
		}
		if remRrr != nil {
			log.Printf(colorRed+"Failed to remove file %s: %v", absPath, remRrr)
		}

	}

	if config.Debug {
		logStage("Creating export folder")
		logKeyValuePair("Export folder", outputPaths.ExportFolderPath)
	}
	os.MkdirAll(outputPaths.ExportFolderPath, 0755)

}

func SetMetadata(fileName string, metadata map[string]string, config *models.Config) error {
	if config.Debug {
		logStage("Setting metadata")
	}
	// Check if the file exists
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		logWarn(fmt.Sprintf("warning: file '%s' does not exist", fileName), false)
		return fmt.Errorf("warning: file '%s' does not exist", fileName)
	} else if err != nil {
		logWarn(fmt.Sprintf("warning: error accessing file '%s': %v", fileName, err), false)
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
				logKeyValuePair("Set xattr", xattrKey)
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
				logKeyValuePair("Set ADS", adsName)
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

func generateSTL(instance *models.InstanceConfig, config *models.Config, exportFolderPath string) (models.GenerateSTLResult, error) {
	result := models.GenerateSTLResult{
		InstanceConfig: *instance,
		OutputPath:     "",
		Command:        "",
		Skipped:        false,
		LowQuality:     false,
		Error:          "",
		AppliedParams:  make(map[string]interface{}),
	}

	if !config.Quiet {
		logStage("Generating STL")
		logKeyValuePair("inputPath", instance.InputPath.Path)
		logKeyValuePair("exportFolderPath", exportFolderPath)
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
			continue
		}
		result.AppliedParams[k] = v
	}

	name := instance.Name
	if name == "" {
		name = filepath.Base(instance.InputPath.Path)
	}
	outputPath := path.Join(exportFolderPath)
	if !config.Quiet && config.Debug {
		logKeyValuePair("outputPath", outputPath)
	}

	if config.Debug {
		logKeyValuePair("exportFolderPath to check:", exportFolderPath)
	}
	if _, exportFolderExists := os.Stat(exportFolderPath); os.IsNotExist(exportFolderExists) {
		if _, outputPathExists := os.Stat(outputPath); os.IsNotExist(outputPathExists) {
			//log.Panicf(colorRed+"Failed to create instance output path (%s) as it does not exists, \n check the folder exists at: \n\n\t%s \n%+v ", outputPath, outputPath, outputPathExists)
			err := os.MkdirAll(outputPath, 0755)
			if err != nil {
				log.Panicf(colorRed+"Failed to create instance output path (%s) as it does not exists, \n check the folder exists at: \n\n\t%s \n%+v ", outputPath, outputPath, outputPathExists)
			}
		} else if !config.Quiet {
			logStage("Creating export folder")
			if !config.Quiet {
				logKeyValuePair("Export folder", exportFolderPath)
			}
			os.MkdirAll(exportFolderPath, 0755)
			if !config.Quiet {
				logCreation(fmt.Sprintf("Created export folder: %s", exportFolderPath))
			}
		}
	}

	// get file name from input path
	fileName := path.Base(instance.InputPath.Path)
	designFileCopyPath := path.Join(exportFolderPath, fileName)

	// Ensure the directory structure for the design file copy path exists
	designFileCopyFolder := filepath.Dir(designFileCopyPath)
	if _, err := os.Stat(designFileCopyFolder); os.IsNotExist(err) {
		os.MkdirAll(designFileCopyFolder, 0755)
		if !config.Quiet {
			logCreation(fmt.Sprintf("Created design folder: %s", designFileCopyPath))
		}
	}

	configFileName := strings.Split(config.ConfigFile, "/")[len(strings.Split(config.ConfigFile, "/"))-1]
	versionPathSafe := strings.ReplaceAll(config.Design.Version, ".", "_")
	configCopyPath := path.Join(path.Dir(exportFolderPath), versionPathSafe, configFileName)

	if _, err := os.Stat(configCopyPath); os.IsNotExist(err) {
		if config.Debug {
			log.Printf("Copying config file from \n\nconfig.ConfigFile: \t%s\n\n to \n\nconfigCopyPath:\t%s", config.ConfigFile, configCopyPath)
		}
		if config.FullExport {
			configErr := Copy(config.ConfigFile, configCopyPath)
			if configErr != nil {
				log.Panicf(colorRed+"Failed to copy config file to export folder: %s", configErr)
			} else if !config.Quiet {
				log.Printf(colorBlue + "Copied config file to export folder" + colorReset)
			}
		}
	} else if config.Debug {
		logKeyValuePair("config exists in export", configCopyPath)
	}

	designFileName := strings.Split(filepath.Base(instance.InputPath.Path), ".")[0]
	if config.Debug {
		logKeyValuePair("Design file name", designFileName)
	}

	paths := instance.GetInstancePaths(config)
	if config.Debug {
		logStage("(in-progress) Instance paths")
		logKeyValuePair("Instance paths", fmt.Sprintf("%+v", paths))
	}

	if paths.InputPath == "" {
		logWarn("InputPath is empty, please set at least one the input path in the config file", true)
		os.Exit(1)
	}

	args := []string{"-o", instance.OutputPathV2}

	if config.Debug {
		logKeyValuePair("creating output folder", instance.OutputPathV2)
	}
	outputFolder := filepath.Dir(instance.OutputPathV2)
	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		os.MkdirAll(outputFolder, 0755)
		if config.Debug {
			logKeyValuePair("created output folder", outputPath)
		}
	}

	if config.Debug {
		logKeyValuePair("output folder confirmed", outputPath)
	}

	if config.Quiet {
		args = append(args, "-q")
	}

	// Add parameters to command, skipping ignored ones
	for name, value := range result.AppliedParams {
		if config.Debug {
			logKeyValuePair(fmt.Sprintf("CustomParameter [%s]", name), fmt.Sprintf("%v", value))
		}
		if reflect.TypeOf(value).Kind() == reflect.String && value != "true" && value != "false" {
			if config.Debug {
				logKeyValuePair(fmt.Sprintf("[String] CustomParameter [%s]", name), fmt.Sprintf("%v", value))
			}
			args = append(args, "-D", fmt.Sprintf("'%s=\"%v\"'", name, value))
		} else {
			if config.Debug {
				logKeyValuePair(fmt.Sprintf("[Number] CustomParameter [%s]", name), fmt.Sprintf("%v", value))
			}
			args = append(args, "-D", fmt.Sprintf("'%s=%v'", name, value))
		}
	}

	if config.IncludePartIDLetter || !config.Design.NoPartIDLetter {
		args = append(args, "-D", fmt.Sprintf("'part_id_letter=\"%s\"'", instance.PartIDLetter))
		result.AppliedParams["part_id_letter"] = instance.PartIDLetter
		if config.Debug {
			logKeyValuePair("OptionalPartIDLetter set on model", instance.PartIDLetter)
		}
	} else {
		if config.Debug {
			logKeyValuePair("OptionalPartIDLetter NOT set on model", "false")
		}
	}

	if config.OverrideFN > 0 {
		args = append(args, "-D", fmt.Sprintf("'$fn=%d'", config.OverrideFN))
		result.AppliedParams["$fn"] = config.OverrideFN
		if config.Debug {
			logKeyValuePair("OverrideFN", fmt.Sprintf("%d", config.OverrideFN))
		}
	}

	if config.Debug {
		logKeyValuePair("InputPath", instance.InputPath.Path)
	}
	args = append(args, fmt.Sprintf("\"%s\"", paths.InputPath))

	if config.Design.CustomOpenSCADOutputFormat != "" {
		if !config.Quiet {
			logKeyValuePair("Custom OpenSCAD export format", config.Design.CustomOpenSCADOutputFormat)
		}
		args = append(args, "--export-format", config.Design.CustomOpenSCADOutputFormat)
	}

	if !config.SkipRender {
		args = append(args, "--render")
	} else if !config.Quiet {
		log.Printf(colorYellow + "Skipping render" + colorReset)
	}

	if !config.Quiet && config.Debug {
		logStage("Running openscad")
		if config.Debug {
			logKeyValuePair("Command", fmt.Sprintf("openscad --enable=experimental textmetrics %v", strings.Join(args, " ")))
		}
	}

	command := "openscad"
	if config.CustomOpenSCADCommand != "" {
		command = config.CustomOpenSCADCommand
		if !config.Quiet {
			logKeyValuePair("Custom OpenSCAD command", command)
			logKeyValuePair("Command", fmt.Sprintf("openscad %v", strings.Join(args, " ")))
			if !strings.Contains(command, "--backend=manifold") {
				logWarn("Warning: The custom OpenSCAD command does not include the --backend=manifold flag, this may result in unexpected behavior", false)
			}
		}
	} else {
		args = append(args, "--backend=manifold")
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	commandStr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	if !config.Quiet {
		log.Printf("Running command: %s", commandStr)
	}

	result.Command = commandStr
	startTime := time.Now()
	// Run the command through a shell
	cmd := exec.Command("sh", "-c", commandStr)

	// Redirect output to the logger
	cmd.Stdout = logger.Writer()
	cmd.Stderr = logger.Writer()

	if !config.Quiet {
		log.Printf("Running command: %s", strings.Join(cmd.Args, " "))
	}
	err := cmd.Run()

	if err != nil {
		log.Printf("Command failed with error: %v", err)
	}

	result.TimeTaken = time.Since(startTime)

	// Check if the context was canceled due to timeout
	if ctx.Err() == context.DeadlineExceeded {
		result.Error = "command timed out"
	}

	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			log.Printf("Command failed with exit code: %d", exitError.ExitCode())
		}
		result.Error = fmt.Sprintf("command execution failed: %v", err)
	}

	if config.SetBuildInfoInFileAttributes {
		setBuildInfoInFileAttributes(instance.OutputPathV2, config, instance)
		if config.Debug {
			logKeyValuePair("Set build info in file attributes", instance.OutputPathV2)
		}
	}

	_, fileErr := os.Stat(instance.OutputPathV2)
	if os.IsNotExist(fileErr) {
		logWarn(fmt.Sprintf("warning: file '%s' does not exist", instance.OutputPathV2), false)
		result.Error = fmt.Sprintf("warning: file '%s' does not exist", instance.OutputPathV2)
	} else if err != nil {
		logWarn(fmt.Sprintf("warning: error accessing file '%s': %v", instance.OutputPathV2, err), false)
		result.Error = fmt.Sprintf("error accessing file '%s': %v", instance.OutputPathV2, err)
	}

	if config.Debug {
		logStage("Finished generating STL in ")
		logKeyValuePair("MetaData set on Path", instance.OutputPathV2)
	}

	result.OutputPath = instance.OutputPathV2
	return result, nil
}

func generateImage(instance *models.InstanceConfig, config *models.Config, exportFolderPath string, camera models.ExportCameraCoordinates) (models.GenerateImageResult, error) {

	result := models.GenerateImageResult{
		InstanceConfig: *instance,
		OutputPath:     "",
		Command:        "",
		Skipped:        false,
		Error:          "",
		AppliedParams:  make(map[string]interface{}),
		CameraName:     camera.CameraName,
		CameraCoords:   camera.CameraCoordinates,
	}

	if !config.Quiet {
		logStage("Generating Image")
		logKeyValuePair("inputPath", instance.InputPath.Path)
		logKeyValuePair("exportFolderPath", exportFolderPath)
		logKeyValuePair("cameraName", camera.CameraName)
		logKeyValuePair("cameraCoordinates", camera.CameraCoordinates)
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

	if config.Debug {
		log.Printf("Ignored keys for image generation: %v", ignoredKeys)
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
				log.Printf("Skipping parameter %s for image generation", k)
			}
			continue
		}
		result.AppliedParams[k] = v
	}

	if config.Debug {
		log.Printf("Applied parameters for image generation: %v", result.AppliedParams)
	}

	name := instance.Name
	if name == "" {
		name = filepath.Base(instance.InputPath.Path)
	}

	// Get instance paths
	paths := instance.GetInstancePaths(config)

	// Create output path for PNG
	outputImgPath := strings.TrimSuffix(instance.OutputPathV2, ".stl") + "-" + camera.CameraName + ".png"
	if !config.Quiet && config.Debug {
		logKeyValuePair("outputPath", outputImgPath)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputImgPath), 0755); err != nil {
		return result, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build OpenSCAD command
	openscadCmd := FindOpenSCAD()
	if openscadCmd == "" {
		return result, fmt.Errorf("openscad command not found")
	}

	if config.Debug {
		log.Printf("Using OpenSCAD command: %s", openscadCmd)
	}

	// Start timing
	startTime := time.Now()

	imageSize := "1920,1080"
	if camera.ImageSize != "" {
		imageSize = camera.ImageSize
	}

	// Build command arguments
	args := []string{
		"-o", outputImgPath,
		"--imgsize", imageSize, // Use a default image size
		"--projection", "perspective",
		"--camera", camera.CameraCoordinates,
		"--preview",
	}

	// Add custom OpenSCAD arguments if provided
	if config.Design.CustomOpenSCADArgs != "" {
		args = append(args, strings.Split(config.Design.CustomOpenSCADArgs, " ")...)
	}

	for name, value := range result.AppliedParams {
		if reflect.TypeOf(value).Kind() == reflect.String && value != "true" && value != "false" {
			args = append(args, "-D", fmt.Sprintf("'%s=\"%v\"'", name, value))
		} else {
			args = append(args, "-D", fmt.Sprintf("'%s=%v'", name, value))
		}
	}

	// Add input file using the correct path
	args = append(args, paths.InputPath)

	// Create command string
	commandStr := fmt.Sprintf("%s %s", openscadCmd, strings.Join(args, " "))
	if !config.Quiet {
		log.Printf("Running image generation command: %s", commandStr)
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
		return result, err
	}

	// Check if file was created
	if _, err := os.Stat(outputImgPath); os.IsNotExist(err) {
		result.Error = fmt.Sprintf("output image file was not created: %s", outputImgPath)
		if config.Debug {
			log.Printf("Output file was not created at: %s", outputImgPath)
		}
		return result, err
	}

	result.OutputPath = outputImgPath
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

func GenerateOutputReport(config *models.Config, instances []models.InstanceConfig, outputPaths models.OutputPaths, stlResults []models.GenerateSTLResult, imageResults []models.GenerateImageResult) error {
	if !config.Quiet {
		logKeyValuePair("Generating HTML report at", outputPaths.ReportPath)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPaths.ReportPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the HTML file
	htmlFile, err := os.Create(outputPaths.ReportPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer htmlFile.Close()

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

	// Generate HTML content for both STL and image results
	htmlContent := templates.Report(config, instances, outputPaths, stlResults, imageResults, allParamNames)

	// Write the HTML content to the file
	if err := htmlContent.Render(context.Background(), htmlFile); err != nil {
		return fmt.Errorf("failed to write HTML content: %w", err)
	}

	if !config.Quiet {
		absPath, err := filepath.Abs(outputPaths.ReportPath)
		if err != nil {
			log.Printf("HTML report generated at: file://%s", outputPaths.ReportPath)
		} else {
			log.Printf("HTML report generated at: file://%s", absPath)
		}
	}

	return nil
}
