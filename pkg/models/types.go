package models

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type ExportCameraCoordinates struct {
	CameraName        string `toml:"camera_name"`
	CameraCoordinates string `toml:"camera_coordinates"`
}

type ParamSet struct {
	Name   string                 `toml:"name"`
	Params map[string]interface{} `toml:"params"`
}

type DesignConfig struct {
	Name           string      `toml:"name" validate:"required,min=1"`
	Description    string      `toml:"description"`
	InputPath      string      `toml:"input_path"`
	InputPaths     []InputPath `toml:"input_paths"`
	OutputPath     string      `toml:"output_path"`
	Version        string      `toml:"version"`
	NoPartIDLetter bool        `toml:"no_part_id_letter"`
	// @@ deprecated
	ExportNameFormat           string                    `toml:"export_name_format"`
	GlobalParams               map[string]interface{}    `toml:"global_params"`
	ParamSets                  []ParamSet                `toml:"param_sets"`
	CustomOpenSCADOutputFormat string                    `toml:"custom_openscad_output_format"`
	CustomOpenSCADArgs         string                    `toml:"custom_openscad_args"`
	ExportImageQuality         string                    `toml:"export_image_quality"`
	ExportImages               []ExportCameraCoordinates `toml:"export_images"` // 'all', 'front' 'back', 'front,back' etc
	// Instances             []InstanceConfig        `toml:"instances"`
	ConfiguredInstanceConfig []ConfiguredInstanceConfig `toml:"instances"`
}

type ConfiguredInstanceConfig struct {
	Name                 string                 `toml:"name"`
	Description          string                 `toml:"description,omitempty"`
	Params               map[string]interface{} `toml:"params"`
	ParamSets            string                 `toml:"param_sets"`             // comma separated list of param sets to use
	ParamNumberationKeys string                 `toml:"param_numberation_keys"` // comma separated list of keys to number
}

// Define a struct to hold the command-line flags
type CmdFlags struct {
	Quiet                        bool
	Debug                        bool
	NoProcessing                 bool
	Version                      bool
	RegexPattern                 string
	MaxInstances                 int
	ContinueOnError              bool
	IncludeExportLog             bool
	FullExport                   bool
	OverwriteExisting            bool
	ShowMan                      bool
	InitProjectName              string
	InitProjectNameExtended      string
	ConfigFile                   string
	SkipRender                   bool
	SkipReadme                   bool
	CustomOpenSCADCommand        string
	CustomOpenSCADOutputFormat   string
	Concurrent                   bool
	MaxConcurrentRequests        int
	IncludePartIDLetter          bool
	SetBuildInfoInFileAttributes bool
	OverrideFN                   int
	HighQuality                  bool
	LowQuality                   bool
}

type OutputPaths struct {
	OutputPath            string
	ExportFolderPath      string
	LowQualityWarningPath string
	ReadmePath            string
	LogOutputPath         string
	ReportPath            string
}

// Config holds the overall configuration structure
type Config struct {
	Design                       DesignConfig `toml:"openscadgen"`
	ConfigFile                   string       `flag:"c,config"`
	Quiet                        bool         `flag:"q"`
	Debug                        bool         `flag:"d"`
	NoProcessing                 bool         `flag:"np"`
	Quality                      string       `flag:"quality"`
	Version                      bool         `flag:"v"`
	RegexPattern                 string       `flag:"f"`
	MaxInstances                 int          `flag:"n"`
	ContinueOnError              bool         `flag:"co"`
	IncludeExportLog             bool         `flag:"el"`
	FullExport                   bool         `flag:"fe"`
	Overwrite                    bool         `flag:"r"`
	SkipRender                   bool         `flag:"sr"`
	OverwriteExisting            bool         `flag:"ow"`
	SkipReadme                   bool         `flag:"skip-readme"`
	CustomOpenSCADCommand        string       `flag:"cmd"`
	Concurrent                   bool         `flag:"p"`
	MaxConcurrentRequests        int          `flag:"pn"`
	IncludePartIDLetter          bool         `flag:"pid"`
	OverrideFN                   int          `flag:"fn"`
	SetBuildInfoInFileAttributes bool         `flag:"fi"`
	OpenSCADVersion              string
	OpenScadGenVersion           string
	InitProjectName              string
	InitProjectNameExtended      string
}

type InputPath struct {
	Path             string                 `toml:"path" validate:"required"`
	ExportNameFormat string                 `toml:"export_name_format"`
	ParamSets        string                 `toml:"param_sets"`
	Params           map[string]interface{} `toml:"params"`
	/*
		IgnoreParamsWhenProcessing are params that will be ignored when processing this input path
	*/
	IgnoreParamsWhenProcessing string `toml:"ignore_param_when_processing"`
}

// InstanceConfig represents a single instance configuration
type InstanceConfig struct {
	Name             string
	AutoName         string
	Description      string
	ExportNameFormat string
	InputPath        InputPath
	Params           map[string]interface{}
	PartIDLetter     string
	isDynamic        bool
	UniqueID         string
	ConfigError      string
	OutputPathV2     string
	IgnoredParams    []string
}

type GenerateSTLResult struct {
	InstanceConfig InstanceConfig
	OutputPath     string
	Command        string
	Error          string
	AppliedParams  map[string]interface{}
	TimeTaken      time.Duration
	Skipped        bool
	LowQuality     bool
}

type GenerateImageResult struct {
	InstanceConfig InstanceConfig
	OutputPath     string
	Command        string
	Error          string
	AppliedParams  map[string]interface{}
	TimeTaken      time.Duration
	Skipped        bool
	CameraName     string
	CameraCoords   string
}

type InstancePaths struct {
	InputPath        string
	OutputFolderPath string
	OutputPath       string
	FullOutputPath   string
	PartIDOutputPath string
}

func (instance *InstanceConfig) GetInstancePaths(config *Config) *InstancePaths {
	// Check if the input path exists
	configDir := filepath.Dir(config.ConfigFile)
	absPath := filepath.Join(configDir, instance.InputPath.Path)

	// If file doesn't exist at config-relative path, try absolute/working dir
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		absPath, err = filepath.Abs(instance.InputPath.Path)
		if err != nil {
			log.Panicf("Could not resolve absolute path: %s", err)
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Panicf("Input path does not exist: %s", instance.InputPath)
	}

	versionPathSafe := strings.ReplaceAll(config.Design.Version, ".", "_")

	outputFolderPath := path.Join(config.Design.OutputPath, versionPathSafe)

	// Get the relative output path by removing the output folder path prefix
	relativeOutputPath := strings.TrimPrefix(instance.OutputPathV2, outputFolderPath)
	relativeOutputPath = strings.TrimPrefix(relativeOutputPath, string(filepath.Separator))

	versionPathSafe = strings.ReplaceAll(config.Design.Version, ".", "_")
	return &InstancePaths{
		InputPath:        "./" + absPath,
		OutputFolderPath: outputFolderPath,
		OutputPath:       relativeOutputPath,
		FullOutputPath:   instance.OutputPathV2,
		PartIDOutputPath: path.Join(config.Design.OutputPath, versionPathSafe, "with_embedded_part_letter"),
	}
}

/*
func GetInstanceConfigSaveLocation(config *Config, inputPath string, instanceName string, exportNameFormat string, params map[string]interface{}, partIdLetter string) string {
	fileName := GetFileName(inputPath)
	if fileName == "" {
		log.Panicf("inputPath: '%s' is invalid, could not get fileName", inputPath)
	}
	if partIdLetter == "" {
		log.Panicf("partIdLetter: '%s' is invalid, could not get partIdLetter", partIdLetter)
	}

	// Determine which export name format to use
	formatToUse := exportNameFormat
	if formatToUse == "" {
		log.Panicf("exportNameFormat: '%s' is invalid, could not get formatToUse", exportNameFormat)
	}

	// Apply all replacements
	formatToUse = makeFileNameReplacements(config.Design.GlobalParams, params, formatToUse, config.Design.Version, fileName, config.Quality, instanceName, partIdLetter)

	// Determine output format
	outputFormat := "stl"
	if config.Design.CustomOpenSCADOutputFormat != "" {
		outputFormat = config.Design.CustomOpenSCADOutputFormat
	}

	// Get config directory and build final path
	configDir := filepath.Dir(config.ConfigFile)
	outputPath := config.Design.OutputPath

	// Ensure the formatToUse doesn't already have an extension
	/* strings.Contains(formatToUse, ".") {
		formatToUse = formatToUse[:strings.LastIndex(formatToUse, ".")]
	}*/

// If output path is absolute, use it directly
/*	if filepath.IsAbs(outputPath) {
		return filepath.Join(outputPath, config.Design.Version, formatToUse+"."+outputFormat)
	}
	versionPathSafe := strings.ReplaceAll(config.Design.Version, ".", "_")
	// Otherwise, join with config directory
	return filepath.Join(configDir, outputPath, versionPathSafe, formatToUse+"."+outputFormat)
}*/

func GetFileName(path string) string {
	fileName := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
	if strings.Contains(fileName, ".") {
		fileName = fileName[:strings.LastIndex(fileName, ".")]
	}
	return fileName
}

func makeFileNameReplacements(globalParams map[string]interface{}, instanceParams map[string]interface{}, ignoredParams []string, formatToUse string, version string, fileName string, quality string, instanceName string, partIdLetter string) string {
	// First replace instance parameters and global parameters from instance params
	for key, value := range instanceParams {
		formatToUse = strings.ReplaceAll(formatToUse, "{"+key+"}", fmt.Sprintf("%v", value))
		formatToUse = strings.ReplaceAll(formatToUse, "${"+key+"}", fmt.Sprintf("%v", value))
	}

	// Then replace any remaining global parameters that weren't in instance params
	for key, value := range globalParams {
		if _, exists := instanceParams[key]; !exists {
			formatToUse = strings.ReplaceAll(formatToUse, "{"+key+"}", fmt.Sprintf("%v", value))
			formatToUse = strings.ReplaceAll(formatToUse, "${"+key+"}", fmt.Sprintf("%v", value))
		} else {
			formatToUse = strings.ReplaceAll(formatToUse, "{"+key+"}", "")
			formatToUse = strings.ReplaceAll(formatToUse, "${"+key+"}", "")
		}
	}
	versionPathSafe := strings.ReplaceAll(version, ".", "_")

	// Replace special placeholders
	formatToUse = strings.ReplaceAll(formatToUse, "{designFileName}", fileName)
	formatToUse = strings.ReplaceAll(formatToUse, "${designFileName}", fileName)
	formatToUse = strings.ReplaceAll(formatToUse, "{version}", versionPathSafe)
	formatToUse = strings.ReplaceAll(formatToUse, "${version}", versionPathSafe)
	formatToUse = strings.ReplaceAll(formatToUse, "{part_id_letter}", partIdLetter)
	formatToUse = strings.ReplaceAll(formatToUse, "${part_id_letter}", partIdLetter)

	if quality != "" {
		formatToUse = strings.ReplaceAll(formatToUse, "{quality}", quality)
		formatToUse = strings.ReplaceAll(formatToUse, "${quality}", quality)
	}

	if instanceName != "" {
		formatToUse = strings.ReplaceAll(formatToUse, "{name}", instanceName)
		formatToUse = strings.ReplaceAll(formatToUse, "${name}", instanceName)
	}

	for _, ignoredParam := range ignoredParams {
		formatToUse = strings.ReplaceAll(formatToUse, "{"+ignoredParam+"}", "")
		formatToUse = strings.ReplaceAll(formatToUse, "${"+ignoredParam+"}", "")
	}

	return formatToUse
}

func (config *Config) GetInputPaths() []InputPath {
	if len(config.Design.InputPaths) > 0 {
		return config.Design.InputPaths
	}
	return []InputPath{
		{Path: config.Design.InputPath},
	}
}
