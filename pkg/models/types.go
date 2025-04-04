package models

import (
	"os"
	"path/filepath"
	"log"
	"strings"
	"path"
	"fmt"
	"time"
)


type ExportCameraCoordinates struct {
	CameraName string `toml:"camera_name"`
	CameraCoordinates string `toml:"camera_coordinates"`
}

type ParamSet struct {
	Name string `toml:"name"`
	Params map[string]string `toml:"params"`
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
	ExportNameFormat           string            `toml:"export_name_format"`
	GlobalParams               map[string]interface{} `toml:"global_params"`
	ParamSets                  []ParamSet        `toml:"param_sets"`
	CustomOpenSCADOutputFormat string            `toml:"custom_openscad_output_format"`
	CustomOpenSCADArgs         string            `toml:"custom_openscad_args"`
	ExportImageQuality         string            `toml:"export_image_quality"`
	ExportImages               []ExportCameraCoordinates              `toml:"export_images"` // 'all', 'front' 'back', 'front,back' etc
	// Instances             []InstanceConfig        `toml:"instances"`
	DynamicInstanceConfig []DynamicInstanceConfig `toml:"instances"`
}

type DynamicInstanceConfig struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description,omitempty"`
	Params      map[string]interface{} `toml:"params"`
	ParamSets   string          `toml:"param_sets"`  // comma separated list of param sets to use
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
}

type InputPath struct {
	Path             string `toml:"path" validate:"required"`
	ExportNameFormat string `toml:"export_name_format"`
	/*
		IgnoreParamsWhenProcessing are params that will be ignored when processing this input path
	*/
	IgnoreParamsWhenProcessing string `toml:"ignore_param_when_processing"`
}



// InstanceConfig represents a single instance configuration
type InstanceConfig struct {
	Name             string                 `toml:"name"`
	AutoName         string                 `toml:"auto_name"`
	Description      string                 `toml:"description,omitempty"`
	ExportNameFormat string                 `toml:"export_name_format"`
	InputPath        string                 `toml:"input_path"`
	Params           map[string]interface{} `toml:"params"`
	PartIDLetter     string                 `toml:"part_id_letter"`
	isDynamic        bool
	UniqueID         string 
	ConfigError      string
}


type GenerateSTLResult struct {
	InstanceConfig InstanceConfig
	OutputPath     string
	Command        string
	Error          string
	TimeTaken      time.Duration
	Skipped        bool
	LowQuality     bool
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
	absPath := filepath.Join(configDir, instance.InputPath)

	// If file doesn't exist at config-relative path, try absolute/working dir
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		absPath, err = filepath.Abs(instance.InputPath)
		if err != nil {
			log.Panicf("Could not resolve absolute path: %s", err)
		}
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Panicf("Input path does not exist: %s", instance.InputPath)
	}

	saveLocation := GetInstanceConfigSaveLocation(config, instance)
	nameFormat := instance.ExportNameFormat
	if nameFormat == "" {
		nameFormat = config.Design.ExportNameFormat
	}

	if nameFormat == "" {
		nameFormat = "{designFileName}-{name}"
	}

	outputFolderPath := path.Join(config.Design.OutputPath, config.Design.Version)
	return &InstancePaths{
		InputPath:        instance.InputPath,
		OutputFolderPath: outputFolderPath,
		OutputPath:       strings.Replace(saveLocation, outputFolderPath, "", 1),
		FullOutputPath:   path.Join(config.Design.OutputPath, config.Design.Version, nameFormat),
		PartIDOutputPath: path.Join(config.Design.OutputPath, config.Design.Version, "with_embedded_part_letter"),
	}
}


func GetInstanceConfigSaveLocation(config *Config, instance *InstanceConfig) string {

	fileName := GetFileName(instance.InputPath)
	if fileName == "" {
		log.Panicf("inputPath: '%s' is invalid, could not get fileName", instance.InputPath)
	}

	formatToUse := instance.ExportNameFormat
	if formatToUse == "" {
		formatToUse = "{designFileName}"
	}else {
		formatToUse = makeFileNameReplacements(config.Design.GlobalParams, instance.Params, formatToUse, config.Design.Version, fileName, config.Quality, instance.Name)
	}

	outputFormat := "stl"
	if config.Design.CustomOpenSCADOutputFormat != "" {
		outputFormat = config.Design.CustomOpenSCADOutputFormat
	}

	configDir := filepath.Dir(config.ConfigFile)

	res := path.Join(configDir, config.Design.OutputPath, config.Design.Version, formatToUse+"."+outputFormat)
	return res
}



func GetFileName(path string) string {
	fileName := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
	if strings.Contains(fileName, ".") {
		fileName = fileName[:strings.LastIndex(fileName, ".")]
	}
	return fileName
}


func makeFileNameReplacements(globalParams map[string]interface{}, instanceParams map[string]interface{}, formatToUse string, version string, name string, quality string, instanceName string) string {
	for key, value := range instanceParams {
		formatToUse = strings.ReplaceAll(formatToUse, "{"+key+"}", fmt.Sprintf("%v", value))
	}

	if strings.Contains(formatToUse, "{designFileName}") {
		formatToUse = strings.ReplaceAll(formatToUse, "{designFileName}", name)
	}

	if strings.Contains(formatToUse, "{version}") {
		formatToUse = strings.ReplaceAll(formatToUse, "{version}", version)
	}

	if strings.Contains(formatToUse, "{quality}") {
		formatToUse = strings.ReplaceAll(formatToUse, "{quality}", quality)
	}
	
	if strings.Contains(formatToUse, "{name}") {
		formatToUse = strings.ReplaceAll(formatToUse, "{name}", instanceName)
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