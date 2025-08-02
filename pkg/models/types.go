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
	CameraName        string                 `toml:"name"`
	CameraCoordinates string                 `toml:"coord"`
	ImageSize         string                 `toml:"image_size"`
	ParamFilter       map[string]interface{} `toml:"param_filter"`
}

type ParamSet struct {
	Name   string                 `toml:"name"`
	Params map[string]interface{} `toml:"params"`
}

type DesignConfig struct {
	Name                       string                    `toml:"name" validate:"required,min=1"`
	Description                string                    `toml:"description"`
	InputPath                  string                    `toml:"input_path"`
	InputPaths                 []InputPath               `toml:"input_paths"`
	OutputPath                 string                    `toml:"output_path"`
	Version                    string                    `toml:"version"`
	NoPartIDLetter             bool                      `toml:"no_part_id_letter"`
	RunType                    string                    `toml:"run_type"` // 'clearAndCreate', 'appendOrOverwrite'
	ExportNameFormat           string                    `toml:"export_name_format" validate:"required,min=1"`
	GlobalParams               map[string]interface{}    `toml:"global_params"`
	ParamSets                  []ParamSet                `toml:"param_sets"`
	CustomOpenSCADOutputFormat string                    `toml:"custom_openscad_output_format"`
	CustomOpenSCADArgs         string                    `toml:"custom_openscad_args"`
	ExportImageQuality         string                    `toml:"export_image_quality"`
	ExportImages               []ExportCameraCoordinates `toml:"export_images"` // 'all', 'front' 'back', 'front,back' etc
	// Instances             []InstanceConfig        `toml:"instances"`
	ConfiguredInstanceConfig []ConfiguredInstanceConfig `toml:"instances"`
	DontUseManifold          bool                       `toml:"dont_use_manifold"`
}

func (d *DesignConfig) ClearVersion(version string) string {
	// Replace forward slashes with underscores
	version = strings.ReplaceAll(version, "/", "_")
	return version
}

type ConfiguredInstanceConfig struct {
	Name             string                    `toml:"name"`
	Description      string                    `toml:"description,omitempty"`
	Params           map[string]interface{}    `toml:"params"`
	ParamSets        string                    `toml:"param_sets"`      // comma separated list of param sets to use
	ParamsNumberated map[string]interface{}    `toml:"params_numbered"` // comma separated list of keys to number
	ExportImages     []ExportCameraCoordinates `toml:"export_images"`
	SkipImages       bool                      `toml:"skip_images"`
}

// Define a struct to hold the command-line flags
type CmdFlags struct {
	Quiet                        bool   `json:"quiet"`
	Debug                        bool   `json:"debug"`
	NoProcessing                 bool   `json:"no_processing"`
	Version                      bool   `json:"version"`
	RegexPattern                 string `json:"regex_pattern"`
	MaxInstances                 int    `json:"max_instances"`
	ContinueOnError              bool   `json:"continue_on_error"`
	IncludeExportLog             bool   `json:"include_export_log"`
	OverwriteExisting            bool   `json:"overwrite_existing"`
	ShowMan                      bool   `json:"show_man"`
	Server                       bool   `json:"server"`
	ServerFolder                 string `json:"server_folder"`
	InitProjectName              string `json:"init_project_name"`
	InitProjectNameExtended      string `json:"init_project_name_extended"`
	ConfigFile                   string `json:"config_file"`
	SkipRender                   bool   `json:"skip_render"`
	SkipReadme                   bool   `json:"skip_readme"`
	CustomOpenSCADCommand        string `json:"custom_openscad_command"`
	CustomOpenSCADOutputFormat   string `json:"custom_openscad_output_format"`
	Concurrent                   bool   `json:"concurrent"`
	MaxConcurrentRequests        int    `json:"max_concurrent_requests"`
	OnlyImages                   bool   `json:"only_images"`
	OnlyExport                   bool   `json:"only_export"`
	IncludePartIDLetter          bool   `json:"include_part_id_letter"`
	SetBuildInfoInFileAttributes bool   `json:"set_build_info_in_file_attributes"`
	OverrideFN                   int    `json:"override_fn"`
	HighQuality                  bool   `json:"high_quality"`
	LowQuality                   bool   `json:"low_quality"`
	EnableFileWatcher            bool   `json:"enable_file_watcher"`
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
	RawConfigFile                string
	ConfigFile                   string //`flag:"c,config"`
	Quiet                        bool   //`flag:"q"`
	Debug                        bool   //`flag:"d"`
	NoProcessing                 bool   //`flag:"np"`
	Quality                      string //`flag:"quality"`
	Version                      bool   ///`flag:"v"`
	RegexPattern                 string //`flag:"f"`
	MaxInstances                 int    //`flag:"n"`
	ContinueOnError              bool   //`flag:"coe"`
	IncludeExportLog             bool   // `flag:"el"`
	Overwrite                    bool   //`flag:"r"`
	SkipRender                   bool   //`flag:"sr"`
	OverwriteExisting            bool   //`flag:"ow"`
	SkipReadme                   bool   //`flag:"skip-readme"`
	CustomOpenSCADCommand        string //`flag:"cmd"`
	Concurrent                   bool   //`flag:"p"`
	MaxConcurrentRequests        int    //`flag:"pn"`
	IncludePartIDLetter          bool   //`flag:"pid"`
	OverrideFN                   int    //`flag:"fn"`
	OnlyImages                   bool   //`flag:"oi"`
	OnlyExport                   bool   //`flag:"oe"`
	SetBuildInfoInFileAttributes bool   //`flag:"fi"`
	Server                       bool   //`flag:"s"`
	ServerFolder                 string //`flag:"sf"`
	EnableFileWatcher            bool   //`flag:"efw"`
	OpenSCADVersion              string
	OpenScadGenVersion           string
	InitProjectName              string
	InitProjectNameExtended      string
}

type InputPath struct {
	Path                string                 `toml:"path" validate:"required"`
	RawOpenSCADFile     string                 `toml:"raw_openscad_file"`
	RawOpenSCADFileName string                 `toml:"raw_openscad_file_name"`
	ExportNameFormat    string                 `toml:"export_name_format"`
	ParamSets           string                 `toml:"param_sets"`
	Params              map[string]interface{} `toml:"params"`
	SkipImages          bool                   `toml:"skip_images"`
	/*
		IgnoreParamsWhenProcessing are params that will be ignored when processing this input path
	*/
	IgnoreParamsWhenProcessing string `toml:"ignore_param_when_processing"`
}

// InstanceConfig represents a single instance configuration
type InstanceConfig struct {
	ID                 string
	Name               string
	AutoName           string
	Description        string
	ExportNameFormat   string
	InputPath          InputPath
	Params             map[string]interface{}
	PartIDLetter       string
	isDynamic          bool
	UniqueID           string
	ConfigError        string
	OutputPathV2       string
	RunOutputPathV3    string // Path used for OpenSCAD -o command, relative to config.toml location
	IgnoredParams      []string
	ImageResults       []GenerateImageResult
	ExportImages       []ExportCameraCoordinates
	RunOutputImagePath string
	SkipImages         bool
	SkippedReason      string
	SkippedImageReason string
}

type GenerateSTLResult struct {
	InstanceConfig      InstanceConfig
	OutputPath          string
	Command             string
	Error               string
	AppliedParams       map[string]interface{}
	TimeTaken           time.Duration
	Skipped             bool
	LowQuality          bool
	SkippedReason       string
	GenerateImageResult []GenerateImageResult
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
	InputPath         string
	OutputFolderPath  string
	OutputPath        string
	FullOutputPath    string
	PartIDOutputPath  string
	InputPathRelative string
}

func (instance *InstanceConfig) GetInstancePaths(config *Config) *InstancePaths {
	configDir := filepath.Dir(config.ConfigFile)
	absPath := filepath.Join(configDir, instance.InputPath.Path)

	if config.Debug {
		log.Printf("[DEBUG] GetInstancePaths: configDir=%s, inputPath=%s, absPath=%s", configDir, instance.InputPath.Path, absPath)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		if config.Debug {
			log.Printf("[DEBUG] GetInstancePaths: file not found at config-relative path, trying absolute: %s", instance.InputPath.Path)
		}
		absPath, err = filepath.Abs(instance.InputPath.Path)
		if err != nil {
			log.Panicf("Could not resolve absolute path: %s", err)
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Panicf("Input path does not exist: %s", instance.InputPath.Path)
	}

	versionPathSafe := config.Design.ClearVersion(config.Design.Version)
	outputFolderPath := path.Join(config.Design.OutputPath, versionPathSafe)
	relativeOutputPath := strings.TrimPrefix(instance.OutputPathV2, outputFolderPath)
	relativeOutputPath = strings.TrimPrefix(relativeOutputPath, string(filepath.Separator))

	relInputPath, err := filepath.Rel(configDir, absPath)
	if err != nil {
		relInputPath = absPath
	}

	if config.Debug {
		log.Printf("[DEBUG] GetInstancePaths: relInputPath=%s, outputFolderPath=%s, relativeOutputPath=%s", relInputPath, outputFolderPath, relativeOutputPath)
	}

	return &InstancePaths{
		InputPath:         absPath,
		InputPathRelative: relInputPath,
		OutputFolderPath:  outputFolderPath,
		OutputPath:        relativeOutputPath,
		FullOutputPath:    instance.OutputPathV2,
		PartIDOutputPath:  path.Join(config.Design.OutputPath, versionPathSafe, "with_embedded_part_letter"),
	}
}

type ConfigFile struct {
	Path string
}

type ProcessResult struct {
	Instances      []InstanceConfig
	STLResults     []GenerateSTLResult
	ImageResults   []GenerateImageResult
	TotalTimeTaken time.Duration
	ExportLocation string
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

func MakeFileNameReplacements(globalParams map[string]interface{}, instanceParams map[string]interface{}, ignoredParams []string, formatToUse string, version string, fileName string, quality string, instanceName string, partIdLetter string) string {
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

	// Replace special placeholders
	formatToUse = strings.ReplaceAll(formatToUse, "{designFileName}", fileName)
	formatToUse = strings.ReplaceAll(formatToUse, "${designFileName}", fileName)
	formatToUse = strings.ReplaceAll(formatToUse, "{part_id_letter}", partIdLetter)
	formatToUse = strings.ReplaceAll(formatToUse, "${part_id_letter}", partIdLetter)

	if quality != "" {
		formatToUse = strings.ReplaceAll(formatToUse, "{quality}", quality)
		formatToUse = strings.ReplaceAll(formatToUse, "${quality}", quality)
	}

	if instanceName != "" {
		formatToUse = strings.ReplaceAll(formatToUse, "{name}", instanceName)
		formatToUse = strings.ReplaceAll(formatToUse, "${name}", instanceName)
		formatToUse = strings.ReplaceAll(formatToUse, "{instanceName}", instanceName)
		formatToUse = strings.ReplaceAll(formatToUse, "${instanceName}", instanceName)
	}

	for _, ignoredParam := range ignoredParams {
		formatToUse = strings.ReplaceAll(formatToUse, "{"+ignoredParam+"}", "")
		formatToUse = strings.ReplaceAll(formatToUse, "${"+ignoredParam+"}", "")
	}

	nonPathValidChars := []string{":", "*", "?", "\"", "<", ">", "|", " "}
	for _, char := range nonPathValidChars {
		formatToUse = strings.ReplaceAll(formatToUse, char, "")
	}

	return formatToUse
}

func decorateInputPath(inputPath InputPath) InputPath {
	inputFileName := filepath.Base(inputPath.Path)
	inputPath.RawOpenSCADFileName = strings.Split(inputFileName, ".")[0]
	return inputPath
}

func (config *Config) GetInputPaths() []InputPath {
	if config.Debug {
		log.Printf("[DEBUG] GetInputPaths: InputPaths len=%d, InputPath=%s", len(config.Design.InputPaths), config.Design.InputPath)
	}
	if len(config.Design.InputPaths) > 0 {
		var inputPathResult []InputPath
		for i, ip := range config.Design.InputPaths {
			if config.Debug {
				log.Printf("[DEBUG] GetInputPaths: InputPaths[%d]=%+v", i, ip)
			}
			inputPathResult = append(inputPathResult, decorateInputPath(ip))
		}

		return inputPathResult
	}
	if config.Debug {
		log.Printf("[DEBUG] GetInputPaths: Using fallback InputPath: %s", config.Design.InputPath)
	}
	return []InputPath{
		decorateInputPath(InputPath{Path: config.Design.InputPath}),
	}
}

// WatcherStatusUI holds data for the watcher status UI
// (for use in templ components, to avoid import cycles)
type WatcherStatusUI struct {
	Watching    bool
	ConfigPaths []string
	Enabled     bool
}
