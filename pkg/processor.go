package pkg

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/xattr"
)

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
	GlobalParams               map[string]string `toml:"global_params"`
	CustomOpenSCADOutputFormat string            `toml:"custom_openscad_output_format"`
	CustomOpenSCADArgs         string            `toml:"custom_openscad_args"`
	// Instances             []InstanceConfig        `toml:"instances"`
	DynamicInstanceConfig []DynamicInstanceConfig `toml:"instances"`
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
	OverwriteExisting            bool
	ShowMan                      bool
	InitProjectName              string
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

func (config *Config) getInputPaths() []InputPath {
	if len(config.Design.InputPaths) > 0 {
		return config.Design.InputPaths
	}
	return []InputPath{
		{Path: config.Design.InputPath},
	}
}

func getOutputPaths(config Config) OutputPaths {
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

	if config.Design.OutputPath != "" {
		// When output path is explicitly specified, use it relative to config dir
		absOutputPath := filepath.Join(configDir, config.Design.OutputPath)

		if config.Debug {
			log.Printf("Output path specified in config: %s", absOutputPath)
		}

		logKeyValuePair("Version", config.Design.Version)
		logKeyValuePair("getOutputPaths:Output path", filepath.Join(absOutputPath, config.Design.Version))

		return OutputPaths{
			OutputPath:            filepath.Join(absOutputPath, config.Design.Version),
			ExportFolderPath:      filepath.Join(absOutputPath, config.Design.Version),
			LowQualityWarningPath: filepath.Join(absOutputPath, config.Design.Version, "LOW_QUALITY_WARNING.md"),
			ReadmePath:            filepath.Join(absOutputPath, config.Design.Version, "README.md"),
			LogOutputPath:         filepath.Join(absOutputPath, config.Design.Version, fmt.Sprintf("export_log_%s.log", time.Now().Format("2006-01-02_15-04-05"))),
		}
	}

	// If output_path is not specified, derive it based on the input path, but still relative to config dir
	// Get design name from input path
	designFilename := filepath.Base(absInputPath)
	designName := strings.TrimSuffix(designFilename, filepath.Ext(designFilename))

	// Match the path structure expected by the tests
	if filepath.IsAbs(config.ConfigFile) {
		// For absolute config file paths, maintain the original structure
		return OutputPaths{
			OutputPath:            filepath.Join(".", filepath.Dir(absInputPath), "export", config.Design.Version, designName, "export", config.Design.Version),
			ExportFolderPath:      filepath.Join(filepath.Dir(absInputPath), "export", config.Design.Version),
			LowQualityWarningPath: filepath.Join(filepath.Dir(absInputPath), "export", config.Design.Version, designName, "LOW_QUALITY_WARNING.md"),
			ReadmePath:            filepath.Join(filepath.Dir(absInputPath), "export", config.Design.Version, designName, "README.md"),
			LogOutputPath:         filepath.Join(filepath.Dir(absInputPath), "export", config.Design.Version, designName, "export_log.log"),
		}
	} else {
		// For relative config file paths, use paths relative to the config file directory
		return OutputPaths{
			OutputPath:            filepath.Join(configDir, "export", config.Design.Version),
			ExportFolderPath:      filepath.Join(configDir, "export", config.Design.Version),
			LowQualityWarningPath: filepath.Join(configDir, "export", config.Design.Version, "LOW_QUALITY_WARNING.md"),
			ReadmePath:            filepath.Join(configDir, "export", config.Design.Version, "README.md"),
			LogOutputPath:         filepath.Join(configDir, "export", config.Design.Version, "export_log.log"),
		}
	}
}

const OPENSCAD_VERSION_WARN_IF_OLDER_THAN = 2024

var config Config

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
git tag "v2.0.0"
git push && git push --tags
```

To create a new version:

```sh
git commit -m "New and improved version"
git tag "v[NEW_VERSION_HERE]-alpha"
*/
const VERSION = "v2.0.6-BETA"

func Process(cmdFlags CmdFlags) {

	startTime := time.Now()

	if cmdFlags.ShowMan {
		flag.PrintDefaults()
		return
	}

	if cmdFlags.InitProjectName != "" {
		initLogger("memory")
		initConfig(cmdFlags.InitProjectName)
		return
	}

	initLogger("memory")

	openscadPath := findOpenSCAD()

	cmd := exec.Command("openscad", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	findScadErr := cmd.Run()
	if findScadErr != nil {
		log.Panicf(colorRed+"Failed to get openscad version: %s"+colorReset, findScadErr)
	}

	openscadVersion := strings.TrimSuffix(out.String(), "\n")
	openscadVersionNumberStr := strings.Replace(openscadVersion, "OpenSCAD version ", "", 1)

	if cmdFlags.Version {
		log.Printf("Openscadgen version %s", VERSION)
		log.Printf("Openscad version %s", openscadVersion)
		os.Exit(0)
	}

	// Load configuration

	config, err := LoadConfig(cmdFlags)
	if err != nil {
		msg := fmt.Sprintf("Failed to load config: %v", err)
		logWarn(msg, true)
		fmt.Fprintf(os.Stderr, colorRed+"Failed to load config: %v\n"+colorReset, err)
		os.Exit(1)
	}

	if config.Debug {
		log.Printf("Loaded config %+v", config)
	}

	if cmdFlags.Debug {
		logStage("Loading config file")
		if cmdFlags.Debug {
			log.Printf("Config file %s", cmdFlags.ConfigFile)
			logKeyValuePair("Config file", cmdFlags.ConfigFile)
		}
	}

	//	os.Setenv("PATH", openscadPath+":"+os.Getenv("PATH"))

	if config.Debug {
		log.Printf("Openscad path: %s", openscadPath)
	}

	// Split the version string to parse the year
	versionParts := strings.Split(openscadVersionNumberStr, ".")
	if len(versionParts) < 1 {
		log.Panic(colorRed + "Invalid OpenSCAD version format. Please check the version output." + colorReset)
	}

	openscadYear, err := strconv.Atoi(versionParts[0])
	if err != nil {
		log.Printf(colorRed+"Failed to parse OpenSCAD year from version %s: %s"+colorReset, openscadVersion, err)
	}

	if len(openscadVersion) == 0 {
		log.Panic(colorRed + "OpenSCAD version output is empty. Please check if OpenSCAD is installed and accessible." + colorReset)
	} else if !cmdFlags.Quiet {
		if config.Debug {
			logKeyValuePair("OpenSCAD version", openscadVersion)
			logKeyValuePair("OpenSCAD year", fmt.Sprintf("%d", openscadYear))
			logKeyValuePair("OpenSCAD version to", fmt.Sprintf("%d", openscadYear))
		}
		if openscadYear < OPENSCAD_VERSION_WARN_IF_OLDER_THAN {
			logWarn("OpenSCAD version is older than the latest available (2024), consider updating to the latest version of OpenSCAD as it has more features and improved rendering time", true)
		}

	}

	design := config.Design

	if design.Name == "" {
		log.Panic(colorRed + "Design name is required, please set the Name field in the config file")
	}

	dynamicInstances := generateDynamicInstances(config)

	for _, instance := range dynamicInstances {
		if instance.PartIDLetter == "" {
			log.Panicf(colorRed + "PartIDLetter is required for dynamic instances, please set the PartIDLetter field for each instance")
		}
	}

	if !config.Quiet {

		log.Printf(colorBlue+"Config provided %d possible instances "+colorYellow+"(%d dynamic)"+colorBlue+" to generate from scad file '%s'"+colorReset, len(dynamicInstances), len(dynamicInstances), design.Name)

		if design.InputPath != "" {
			logKeyValuePair("Input File", design.InputPath)
		} else {
			for i, inputPath := range design.InputPaths {
				logKeyValuePair(fmt.Sprintf("Input File [%d/%d]", i+1, len(design.InputPaths)), inputPath.Path)
			}
		}
		logKeyValuePair("Design Version", design.Version)
		if config.MaxInstances > 0 {
			logWarn(fmt.Sprintf("Max Limit of %d instances", config.MaxInstances), false)
		}
		if config.RegexPattern != "" {
			logWarn(fmt.Sprintf("Filter to: %s", config.RegexPattern), false)
		}
		if config.Debug {
			logKeyValuePair("Input Flags", fmt.Sprintf("%+v", cmdFlags))
			logKeyValuePair("Config File", cmdFlags.ConfigFile)
			logKeyValuePair("Export Location", design.OutputPath)
		}
	}
	// Compile regex if provided
	var regex *regexp.Regexp
	if config.RegexPattern != "" {
		regex, err = regexp.Compile(config.RegexPattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, colorRed+"Invalid regex pattern: %v\n"+colorReset, err)
			os.Exit(1)
		}
	}

	if cmdFlags.Version {
		log.Printf("Openscadgen version %s", VERSION)
		log.Printf("Openscad version %s", openscadVersion)
		os.Exit(0)
	}

	if cmdFlags.ConfigFile == "" {
		flag.PrintDefaults()

		logWarn("No config file provided, use -c or -config to specify a config file", true)
		os.Exit(1)
	}

	if cmdFlags.Debug && cmdFlags.Quiet {
		log.Print("**whispers**WHAT DO YOU WANT FROM ME? Being quiet (-q) and debug (-v) is not supported (defaulting to debug)")
		cmdFlags.Quiet = false
	}

	// New message indicating config file location and number of instances
	if !cmdFlags.Quiet {

		log.Println(`   ___                                     _                      `)
		log.Println(`  / _ \ _ __   ___ _ __  ___  ___ __ _  __| | __ _  ___ _ __      `)
		log.Println(" | | | | '_ \\ / _ \\ '_ \\/ __|/ __/ _` |/ _` |/ _` |/ _ \\ '_ \\     ")
		log.Println(` | |_| | |_) |  __/ | | \__ \ (_| (_| | (_| | (_| |  __/ | | |    `)
		log.Println(`  \___/| .__/ \___|_| |_|___/\___\__,_|\__,_|\__, |\___|_| |_|    `)
		log.Println(` 	    |_|                                   |___/                `)

		log.Printf(colorGreen + "Welcome to openscadgen" + colorReset)
		logWarn("You are running an ALPHA version, this software is being worked on and not yet stable, please report any bugs to https://github.com/KiwiKid/openscadgen/issues", false)
		log.Printf("Openscadgen version %s", VERSION)
	}

	pathsToProcess := config.getInputPaths()

	for _, path := range pathsToProcess {

		// First try relative to config file location
		configDir := filepath.Dir(cmdFlags.ConfigFile)
		absPath := filepath.Join(configDir, path.Path)

		// If file doesn't exist at config-relative path, try absolute/working dir
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			absPath, err = filepath.Abs(path.Path)
			if err != nil {
				logWarn(fmt.Sprintf("Could not resolve absolute path: %s", err), true)
				os.Exit(1)
			}
		}

		designFileExists, err := os.Stat(absPath)
		if err != nil {

			logWarn(fmt.Sprintf("Could not find scad file: at '%s' %s", absPath, err), true)
			logKeyValuePair("Config File", cmdFlags.ConfigFile)
			logKeyValuePair("Design File Path", path.Path)
			os.Exit(1)
		} else if designFileExists == nil {
			logWarn(fmt.Sprintf("Design file %s does not exist", path.Path), true)
			os.Exit(1)
		}
	}

	if !config.Quiet {
		logStage("Starting STL generation")
		if config.RegexPattern != "" {
			logWarn(fmt.Sprintf("Filter: Only generating file matching pattern %s", config.RegexPattern), false)
		}
		if config.MaxInstances > 0 {
			logWarn(fmt.Sprintf("Limit: Only generating first %d instances", config.MaxInstances), false)
		}
	}

	outputPaths := getOutputPaths(*config)

	getOrMakeExportFolder(config, &outputPaths)

	initLogger(outputPaths.LogOutputPath)

	// Generate STL files for dynamic instances
	if len(dynamicInstances) > 0 && !config.Quiet {
		logStage(fmt.Sprintf("Found Dynamic %d Instances", len(dynamicInstances)))
		for _, instance := range dynamicInstances {
			logKeyValuePair("Dynamic Instance", instance.Name)
		}
	}

	if config.Debug {
		logStage(fmt.Sprintf("Got %d paths to process", len(pathsToProcess)))
	}

	processedCount := 0
	skippedCount := 0
	for pathIndex, path := range pathsToProcess {
		stlIndex := 0

		if !config.Quiet {
			logStage(fmt.Sprintf("Starting Dynamic %d Instances - %s", len(dynamicInstances), path.Path))
		}
		for diIndex, instance := range dynamicInstances {
			if config.Debug {
				logStage(fmt.Sprintf("\t\tDynamic Model - (path:[%d/%d]-instance:[%d/%d]) [%d processed] - '%s' %s ", pathIndex+1, len(pathsToProcess), diIndex+1, len(dynamicInstances), processedCount, instance.PartIDLetter, instance.AutoName))
			}
			if regex != nil && !regex.MatchString(instance.Name) {
				if !config.Quiet && config.Debug {
					log.Printf(colorYellow+"Skipping instance %s as it does not match the regex pattern", instance.Name)
				}
				skippedCount++
				continue
			}

			if !config.Quiet {
				logStage(fmt.Sprintf("Dynamic Model - (path:[%d/%d]-instance:[%d/%d]) [%d processed] - '%s' %s ", pathIndex, len(pathsToProcess), diIndex, len(dynamicInstances), processedCount, instance.PartIDLetter, instance.AutoName))
				logKeyValuePair("InputPath", instance.InputPath)
				logKeyValuePair("AutoName", instance.AutoName)
				if config.IncludePartIDLetter {
					logKeyValuePair("PartIDLetter", instance.PartIDLetter)
				}
				if config.Debug {
					logKeyValuePair("Params", fmt.Sprintf("%+v", instance.Params))
				}
			}

			if config.MaxInstances > 0 && processedCount >= config.MaxInstances {
				log.Printf(colorBlue+"Max instance of %d processed, stopping", config.MaxInstances)
				break
			} else if config.Debug {
				logStage("Max instance check passed")
				logKeyValuePair("Max instances", fmt.Sprintf("%d", config.MaxInstances))
				logKeyValuePair("Processed instances", fmt.Sprintf("%d", processedCount))
			}

			if config.NoProcessing {
				log.Printf(colorBlue+"No processing requested, skipping instance '%s'", instance.Name)
				continue
			}

			_, err := generateSTL(instance, config, outputPaths.ExportFolderPath)
			if err != nil {
				logWarn(fmt.Sprintf("Error generating STL for instance '%s': %v", instance.Name, err), false)
			} else if !config.Quiet {
				processedCount++
			}

			stlIndex++
		}

	}

	generateReadme(config, dynamicInstances, VERSION, openscadVersion, outputPaths.ReadmePath)

	generateLowQualityWarningFile(config, outputPaths.LowQualityWarningPath)

	if !config.Quiet {

		msg := ""
		if skippedCount > 0 {
			msg += fmt.Sprintf(colorYellow+"%d instances were skipped as they did not match the regex pattern\n\n"+colorReset, skippedCount)
		}
		fileFormat := "stl"
		if config.Design.CustomOpenSCADOutputFormat != "" {
			fileFormat = config.Design.CustomOpenSCADOutputFormat
		}

		if processedCount > 0 {
			msg += fmt.Sprintf(colorGreen+"openscadgen completed! \n\n%d %s files generated in %s\n\nthanks for using openscadgen! "+colorReset, processedCount, fileFormat, time.Since(startTime))

		} else if config.NoProcessing {
			log.Printf(colorBlue + "(as requested) No instances were processed" + colorReset)
		} else {
			logWarn("No instances were processed, please check the log, config and scad file and try again", true)
		}

		log.Print(msg)
	}

}

func generateConfFieldMap(conf interface{}) map[string]struct{} {
	fieldMap := make(map[string]struct{})
	val := reflect.ValueOf(conf)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fieldMap
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tomlTag := field.Tag.Get("toml")
		if tomlTag != "" {
			fieldMap[tomlTag] = struct{}{}
		}
	}
	return fieldMap
}

// loadConfig reads the configuration file and populates the Config struct
func LoadConfig(flags CmdFlags) (*Config, error) {
	var conf Config
	data, err := ioutil.ReadFile(flags.ConfigFile)
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

	log.Printf("Loaded config")

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
	conf.ConfigFile = flags.ConfigFile

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

	if conf.Design.DynamicInstanceConfig != nil {
		if len(conf.Design.DynamicInstanceConfig) > 0 {
			for dynamicInstanceIndex, dynamicInstance := range conf.Design.DynamicInstanceConfig {
				for paramName, paramValue := range dynamicInstance.Params {
					if config.Debug {
						logKeyValuePair("LoadConfig:Param name", paramName)
						logKeyValuePair("LoadConfig:Param value", paramValue)
					}

					paramHasMoreThanOneValue := strings.Contains(paramValue, ",")
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

					nameHasParams := strings.Contains(exportNameFormat, fmt.Sprintf("{%s}", paramName))
					if !nameHasParams && paramHasMoreThanOneValue {
						logKeyValuePair("Dynamic instance index 1", fmt.Sprintf("%d", dynamicInstanceIndex))
						logKeyValuePair("Export Name Format", exportNameFormat)
						logKeyValuePair("Missing Param name", paramName)
						logKeyValuePair("Param value", paramValue)
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

			//	if exportNameFormat != "" {
			/*for _, dynamicInstanceConfig := range conf.Design.DynamicInstanceConfig {
							exportNameFormat := getExportNameFormat(&conf, &dynamicInstanceConfig)
							for _, paramName := range dynamicInstanceConfig.Params {
								paramHasMoreThanOneValue := strings.Contains(paramName, ",")

								if config.Debug {
									logKeyValuePair("Param name", paramName)
									logKeyValuePair("Param has more than one value", fmt.Sprintf("%t", paramHasMoreThanOneValue))
								}

								if paramHasMoreThanOneValue && !strings.Contains(exportNameFormat, fmt.Sprintf("{%s}", paramName)) {
									logKeyValuePair("Export name format", exportNameFormat)
									logKeyValuePair("Missing Param name", dynamicInstanceConfig.Name)
									logKeyValuePair("Config file", configFile)
									logWarn(fmt.Sprintf(`Export name format:
			%s

			does not contain param

			%s

				Include every param in the export name (in the format '{param_name}') to ensure all instances are generated.
			`, conf.Design.ExportNameFormat, paramName), true)
									os.Exit(1)
								}

							}
						}
						//	}
			*/
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

	return &conf, nil
}

func generateLowQualityWarningFile(config *Config, outputPath string) {

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

func generateReadme(config *Config, dynamicInstances []InstanceConfig, version string, openscadVersion string, readmePath string) {
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
		paths := instance.getInstancePaths(config)
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
}

type InstancePaths struct {
	InputPath        string
	OutputFolderPath string
	OutputPath       string
	FullOutputPath   string
	PartIDOutputPath string
}

func (instance *InstanceConfig) getInstancePaths(config *Config) *InstancePaths {
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

	saveLocation := getInstanceConfigSaveLocation(config, instance)
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

type DynamicInstanceConfig struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description,omitempty"`
	Params      map[string]string `toml:"params"`
}

func getExportNameFormat(config *Config) string {

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

func getFileName(path string) string {
	fileName := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
	if strings.Contains(fileName, ".") {
		fileName = fileName[:strings.LastIndex(fileName, ".")]
	}
	return fileName
}

func getInstanceConfigSaveLocation(config *Config, instance *InstanceConfig) string {

	fileName := getFileName(instance.InputPath)
	if fileName == "" {
		log.Panicf("inputPath: '%s' is invalid, could not get fileName", instance.InputPath)
	}

	formatToUse := instance.ExportNameFormat
	if formatToUse == "" {
		formatToUse = "{designFileName}"
	}

	if formatToUse != "" {
		for key, value := range instance.Params {
			formatToUse = strings.ReplaceAll(formatToUse, "{"+key+"}", fmt.Sprintf("%v", value))
		}

		if strings.Contains(formatToUse, "{designFileName}") {
			formatToUse = strings.ReplaceAll(formatToUse, "{designFileName}", fileName)
		}

		if strings.Contains(formatToUse, "{version}") {
			formatToUse = strings.ReplaceAll(formatToUse, "{version}", config.Design.Version)
		}

		if config.Quality != "" {
			formatToUse = fmt.Sprintf("%s_%s", config.Quality, formatToUse)
		}
	}

	outputFormat := "stl"
	if config.Design.CustomOpenSCADOutputFormat != "" {
		outputFormat = config.Design.CustomOpenSCADOutputFormat
	}

	configDir := filepath.Dir(config.ConfigFile)

	res := path.Join(configDir, config.Design.OutputPath, config.Design.Version, formatToUse+"."+outputFormat)
	return res
}

func getAllParams(dynamicInstance DynamicInstanceConfig, ignoreParams string, globalParams map[string]string) (map[string]string, []string) {
	params := make(map[string]string)
	for key, value := range dynamicInstance.Params {
		params[key] = value
	}
	if globalParams != nil {
		for key, value := range globalParams {
			params[key] = value
		}
	}
	ignoredKeys := []string{}
	paramsToIgnore := strings.Split(ignoreParams, ",")
	for _, ignoreParam := range paramsToIgnore {
		ignoredKeys = append(ignoredKeys, ignoreParam)
		delete(params, ignoreParam)
	}

	return params, ignoredKeys
}

func generateDynamicInstances(config *Config) []InstanceConfig {
	instances := []InstanceConfig{}
	// for each key in params, split the value by comma, parse the value (1,2,3 or 1-5, or 1,2,5-10) into a range of specific values and then generate all combinations of those values

	if len(config.Design.DynamicInstanceConfig) == 0 {
		config.Design.DynamicInstanceConfig = []DynamicInstanceConfig{
			{
				Name:        "default",
				Description: "Default dynamic instance",
				Params:      make(map[string]string),
			},
		}
	}

	inputPaths := config.getInputPaths()

	for _, inputPath := range inputPaths {
		// Check if the input path exists
		configDir := filepath.Dir(config.ConfigFile)
		absPath := filepath.Join(configDir, inputPath.Path)

		// If file doesn't exist at config-relative path, try absolute/working dir
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			absPath, err = filepath.Abs(inputPath.Path)
			if err != nil {
				logWarn(fmt.Sprintf("Could not resolve absolute path: %s", err), true)
				continue
			}
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			logWarn(fmt.Sprintf("Input path does not exist: %s", inputPath.Path), true)
			os.Exit(1)
			continue
		}

		// Reset PartIDLetter for each new filepath
		for diIndex, dynamicInstance := range config.Design.DynamicInstanceConfig {
			paramCombinations := map[string][]interface{}{}

			params, ignoredKeys := getAllParams(dynamicInstance, inputPath.IgnoreParamsWhenProcessing, config.Design.GlobalParams)

			if len(ignoredKeys) > 0 {
				for _, param := range ignoredKeys {
					delete(params, param)
					splitOptions := []string{"-", "_", ""}
					for _, splitOption := range splitOptions {
						inputPath.ExportNameFormat = strings.ReplaceAll(inputPath.ExportNameFormat, fmt.Sprintf("{%s}%s", param, splitOption), "")
					}
				}
			}

			for key, value := range params {
				values := strings.Split(value, ",")
				var parsedValues []interface{}

				for _, val := range values {
					val = strings.TrimSpace(val)
					if strings.Contains(val, "-") && isNumericRange(val) {
						// parse the value (1-5) into a range of specific values
						rangeValues := strings.Split(val, "-")
						start, err := strconv.Atoi(rangeValues[0])
						if err != nil {
							log.Printf(colorRed+"Failed to parse start value for dynamic instance %d: %s", diIndex, dynamicInstance.Name)
							continue
						}
						end, err := strconv.Atoi(rangeValues[1])
						if err != nil {
							log.Printf(colorRed+"Failed to parse end value for dynamic instance %d: %s", diIndex, dynamicInstance.Name)
							continue
						}
						for i := start; i <= end; i++ {
							parsedValues = append(parsedValues, i)
						}
					} else if val == "true" || val == "false" {
						// handle boolean values
						parsedValues = append(parsedValues, val == "true")
					} else if num, err := strconv.ParseFloat(val, 64); err == nil {
						// handle integer values
						parsedValues = append(parsedValues, num)
					} else {
						// handle string values
						parsedValues = append(parsedValues, val)
					}
				}
				paramCombinations[key] = parsedValues
			}

			// Generate all combinations of parameter values
			keys := make([]string, 0, len(paramCombinations))
			for k := range paramCombinations {
				keys = append(keys, k)
			}

			var generateCombinations func(map[string]interface{}, int)
			generateCombinations = func(current map[string]interface{}, index int) {
				if index == len(keys) {

					instanceName := dynamicInstance.Name
					if instanceName == "" {
						instanceName = config.Design.ExportNameFormat
					}
					for k, v := range current {
						placeholder := fmt.Sprintf("{%s}", k)
						instanceName = strings.ReplaceAll(instanceName, placeholder, fmt.Sprintf("%v", v))
					}

					fileName := getFileName(inputPath.Path)

					if strings.Contains(instanceName, "{designFileName}") {
						instanceName = strings.ReplaceAll(instanceName, "{designFileName}", fileName)
					}

					if strings.Contains(instanceName, "{version}") {
						instanceName = strings.ReplaceAll(instanceName, "{version}", config.Design.Version)
					}

					if strings.Contains(fileName, ".") {
						fileName = fileName[:strings.LastIndex(fileName, ".")]
					}
					if fileName == "" || fileName == "." {
						log.Panicf("designFileName is invalid for dynamic instance %s", dynamicInstance.Name)
					} else {
						current["designFileName"] = fileName
					}

					exportNameFormat := inputPath.ExportNameFormat
					if exportNameFormat == "" {
						exportNameFormat = config.Design.ExportNameFormat
					}

					filteredParams := copyMap(current)

					// Old way of ignoring params
					if inputPath.IgnoreParamsWhenProcessing != "" {
						for _, param := range strings.Split(inputPath.IgnoreParamsWhenProcessing, ",") {
							delete(filteredParams, param)
							splitOptions := []string{"-", "_", ""}
							for _, splitOption := range splitOptions {
								exportNameFormat = strings.ReplaceAll(exportNameFormat, fmt.Sprintf("{%s}%s", param, splitOption), "")
							}
						}
					}

					newInstance := InstanceConfig{
						Name:             dynamicInstance.Name,
						AutoName:         instanceName,
						isDynamic:        true,
						Description:      dynamicInstance.Description,
						InputPath:        absPath,
						Params:           filteredParams,
						ExportNameFormat: exportNameFormat,
					}
					if config.Debug {
						logKeyValuePair("Filtered Params", fmt.Sprintf("%v", filteredParams))
						logKeyValuePair("newInstance", fmt.Sprintf("%+v", newInstance))
					}

					instances = append(instances, newInstance)
					return
				}

				key := keys[index]
				for _, value := range paramCombinations[key] {
					current[key] = value
					generateCombinations(current, index+1)
				}
			}

			generateCombinations(make(map[string]interface{}), 0)
		}

		// Each new filepath should reset the pathId lettering
		for index, instance := range instances {
			letter := getPartIDLetter(index)
			instances[index].PartIDLetter = letter
			if !config.Design.NoPartIDLetter {
				instances[index].Name = fmt.Sprintf("%s_%s", instances[index].Name, letter)
			}
			if config.Debug {
				logKeyValuePair(fmt.Sprintf("Generated PartIDLetter from index (%d) for instance %s", index, instance.Name), letter)
			}
		}
	}

	if config.Debug {
		logStage("Making Part ID Letter")
	}

	return instances
}

func isNumericRange(val string) bool {
	parts := strings.Split(val, "-")
	if len(parts) != 2 {
		return false
	}
	_, err1 := strconv.Atoi(parts[0])
	_, err2 := strconv.Atoi(parts[1])
	return err1 == nil && err2 == nil
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
	return string(output)
}

var configTemplate = `[openscadgen]
name = "{{projectName}}"
description = "{{projectDescription}}"

output_path = "./export/"
version = "v0.1"

[[openscadgen.input_paths]]
path = "./{{projectName}}.scad"

# [[openscadgen.input_paths]]
# path = "./a-second-related-openscad-file.scad"
# ignore_param_when_processing = "param3" # dont process param3 when generating instances for this file 
# export_name_format = "custom-export-name-for-this-file/{designFileName}-{name}"


[[openscadgen.instances]]
params = { name= "firstInstanceName", param1 = "firstValue", param2 = "secondValue", param3 = "thirdValue" }

# [[openscadgen.instances]]
# params = { name= "secondInstanceName", param1 = "firstValue", param2 = "secondValue", param3 = "thirdValue" }
`

var openScadTemplate = `
$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;
`

func initConfig(projectName string) error {
	// make a directory at current working directory with the name of the initPath
	os.Mkdir(projectName, 0755)

	configTemplate = strings.ReplaceAll(configTemplate, "{{projectName}}", projectName)

	os.Create(filepath.Join(projectName, "config.toml"))
	os.WriteFile(filepath.Join(projectName, "config.toml"), []byte(configTemplate), 0644)

	os.Create(filepath.Join(projectName, projectName+".scad"))
	os.WriteFile(filepath.Join(projectName, projectName+".scad"), []byte(openScadTemplate), 0644)

	return nil
}

func initLogger(logFilePath string) error {
	if logFilePath == "memory" {
		logToMemory = true
		logger = log.New(io.MultiWriter(os.Stdout, &logBuffer), "", log.Ldate|log.Ltime|log.Lshortfile)
		return nil
	}

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

func getOrMakeExportFolder(config *Config, outputPaths *OutputPaths) {
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
		} else {
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

func SetMetadata(fileName string, metadata map[string]string, config *Config) error {
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
func setBuildInfoInFileAttributes(outputPath string, config *Config, instance InstanceConfig) {
	metadata := make(map[string]string)
	metadata["openscadgen.version"] = config.Design.Version
	metadata["openscadgen.instance"] = instance.Name
	for name, value := range instance.Params {
		metadata[fmt.Sprintf("openscadgen.params.%s", name)] = fmt.Sprintf("%v", value)
	}
	SetMetadata(outputPath, metadata, config)
}

func generateSTL(instance InstanceConfig, config *Config, exportFolderPath string) (string, error) {
	if !config.Quiet {
		logStage("Generating STL")
		logKeyValuePair("inputPath", instance.InputPath)
		logKeyValuePair("exportFolderPath", exportFolderPath)
	}

	name := instance.Name
	if name == "" {
		name = filepath.Base(instance.InputPath)
	}
	outputPath := path.Join(exportFolderPath, name)
	if !config.Quiet && config.Debug {
		logKeyValuePair("outputPath", outputPath)
	}

	if config.Debug {
		logKeyValuePair("exportFolderPath to check:", exportFolderPath)
	}
	if _, exportFolderExists := os.Stat(exportFolderPath); os.IsNotExist(exportFolderExists) {

		if _, outputPathExists := os.Stat(outputPath); os.IsNotExist(outputPathExists) {
			log.Panicf(colorRed+"Failed to create instance output path as it does not exists, \n check the folder exists at: \n\n\t%s \n%+v ", outputPath, outputPathExists)
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
	fileName := path.Base(instance.InputPath)

	designFileCopyPath := path.Join(exportFolderPath, fileName)

	// Ensure the directory structure for the design file copy path exists
	designFileCopyFolder := filepath.Dir(designFileCopyPath)
	if _, err := os.Stat(designFileCopyFolder); os.IsNotExist(err) {
		os.MkdirAll(designFileCopyFolder, 0755)
		if !config.Quiet {
			logCreation(fmt.Sprintf("Created design folder: %s", designFileCopyPath))
		}
	}
	/*
		if config.Debug {
			logCreation("Copying design file to export folder")
			logKeyValuePair("Design file", instance.InputPath)
			logKeyValuePair("Export folder", designFileCopyPath)
		}
		// Check if the design file already exists in the export folder
		if _, err := os.Stat(designFileCopyPath); os.IsNotExist(err) {

			err := Copy(instance.InputPath, designFileCopyPath)
			if !config.Quiet {
				logCreation(fmt.Sprintf("Copied design file to export folder: %s", designFileCopyPath))
			}
			if err != nil {
				log.Panicf(colorRed+"Failed to copy design file to export folder: %s", err)
			} else if !config.Quiet && config.Debug {
				log.Printf(colorBlue + "Copied .scad file to export folder" + colorReset)
			}

		} else {
			if config.Debug {
				log.Printf(colorBlue + "Design file already exists in export folder, skipping copy" + colorReset)

				logKeyValuePair("Design file", designFileCopyPath)
			}
		}
	*/
	configFileName := strings.Split(config.ConfigFile, "/")[len(strings.Split(config.ConfigFile, "/"))-1]

	configCopyPath := path.Join(path.Dir(exportFolderPath), config.Design.Version, configFileName)

	if _, err := os.Stat(configCopyPath); os.IsNotExist(err) {

		if config.Debug {
			log.Printf("Copying config file from \n\nconfig.ConfigFile: \t%s\n\n to \n\nconfigCopyPath:\t%s", config.ConfigFile, configCopyPath)
		}
		configErr := Copy(config.ConfigFile, configCopyPath)
		if configErr != nil {
			log.Panicf(colorRed+"Failed to copy config file to export folder: %s", configErr)
		} else { //if !config.Quiet && config.Debug
			log.Printf(colorBlue + "Copied config file to export folder" + colorReset)
		}
	} else if config.Debug {
		logKeyValuePair("config exists in export", configCopyPath)
	}

	designFileName := strings.Split(filepath.Base(instance.InputPath), ".")[0]

	outputFileName := fmt.Sprintf("%s", designFileName)
	for name, value := range instance.Params {
		outputFileName += fmt.Sprintf("_%s%v", name, value)
	}
	if config.Debug {
		logKeyValuePair("outputFileName[1]", outputFileName)
		logKeyValuePair("Design file name", designFileName)
	}

	outputFileName += ".stl"

	paths := instance.getInstancePaths(config)

	if config.Debug {
		logStage("(in-progress) Instance paths")
		logKeyValuePair("Instance paths", fmt.Sprintf("%+v", paths))
	}

	if paths.InputPath == "" {
		logWarn("InputPath is empty, please set at least one the input path in the config file", true)
		os.Exit(1)
	}

	outputPath = getInstanceConfigSaveLocation(config, &instance)
	args := []string{"-o", fmt.Sprintf("'./%s'", outputPath)}

	if config.Debug {
		logKeyValuePair("creating output folder", outputPath)
	}
	outputFolder := filepath.Dir(outputPath)
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

	for name, value := range instance.Params {
		if config.Debug {
			logKeyValuePair(fmt.Sprintf("CustomParameter [%s]", name), fmt.Sprintf("%v", value))
		}
		if reflect.TypeOf(value).Kind() == reflect.String {
			args = append(args, "-D", fmt.Sprintf("%s='\"%v\"'", name, value))
		} else {
			args = append(args, "-D", fmt.Sprintf("'%s=%v'", name, value))
		}
	}

	if config.IncludePartIDLetter {
		args = append(args, "-D", fmt.Sprintf("'optional_part_id_letter=\"%s\"'", instance.PartIDLetter))
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
		if config.Debug {
			logKeyValuePair("OverrideFN", fmt.Sprintf("%d", config.OverrideFN))
		}
	}

	if config.Debug {
		logKeyValuePair("InputPath", instance.InputPath)
	}
	args = append(args, instance.InputPath)

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
		//	args = append(args, "--export-format", "binstl")
		args = append(args, "--backend=manifold")
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	commandStr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	if !config.Quiet {
		log.Printf("Running command: %s", commandStr)
	}
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

	// Check if the context was canceled due to timeout
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out")
	}

	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			log.Printf("Command failed with exit code: %d", exitError.ExitCode())
		}
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	if config.SetBuildInfoInFileAttributes {
		setBuildInfoInFileAttributes(outputPath, config, instance)
		if config.Debug {
			logKeyValuePair("Set build info in file attributes", outputPath)
		}
	}

	_, fileErr := os.Stat(outputPath)
	if os.IsNotExist(fileErr) {
		logWarn(fmt.Sprintf("warning: file '%s' does not exist", outputPath), false)
		return outputPath, fmt.Errorf("warning: file '%s' does not exist", outputPath)
	} else if err != nil {
		logWarn(fmt.Sprintf("warning: error accessing file '%s': %v", outputPath, err), false)
		return outputPath, fmt.Errorf("error accessing file '%s': %v", outputPath, err)
	}

	if config.Debug {
		logStage("Finished generating STL in ")
		logKeyValuePair("MetaData set on Path", outputPath)
	}

	return outputPath, nil
}

/*
func saveReleaseFile(filePath string, data []byte) error {
	// Check if the file exists and is not empty
	fileInfo, err := os.Stat(filePath)
	if err == nil && fileInfo.Size() > 0 {
		filesStr := ""
		fileNames := fileInfo.Name()
		for i, file := range fileNames {
			if i < 5 {
				filesStr += fmt.Sprintf("%s\n", file)
			} else {
				filesStr += "...\n"
				break
			}
		}
		// File is not empty, prompt the user for confirmation
		logWarn("The export folder is not empty", false)
		logKeyValuePair("(non-empty) Export Folder", filePath)
		logKeyValuePair("Files", filesStr)
		fmt.Printf("These files here will be deleted: %s \n %s \n Do you want to continue? (y/n): ", filePath, filesStr)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if response != "y\n" && response != "Y\n" {
			fmt.Println("Aborting save operation.")
			return nil
		}
	}

	// Proceed with saving the file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	fmt.Println("File saved successfully.")
	return nil
}*/

func FindOpenSCAD() string {
	// Try to find openscad using `which` command
	cmd := exec.Command("which", "openscad")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("OpenSCAD not found in PATH.")
	}
	return string(output)
}
