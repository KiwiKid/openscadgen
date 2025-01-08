package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

// https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file
// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. If that fail, copy the file contents from src to dst.

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

// InstanceConfig represents a single instance configuration
type InstanceConfig struct {
	Name        string                 `toml:"name"`
	Description string                 `toml:"description,omitempty"`
	Params      map[string]interface{} `toml:"params"`
}

type DesignConfig struct {
	Name               string           `toml:"name"`
	Description        string           `toml:"description"`
	InputPath          string           `toml:"input_path" validate:"required"`
	OutputPath         string           `toml:"output_path" validate:"required"`
	Version            string           `toml:"version"`
	CustomOpenSCADArgs string           `toml:"custom_openscad_args"`
	Instances          []InstanceConfig `toml:"instances" validate:"required"`
}

// Config holds the overall configuration structure
type Config struct {
	Design                DesignConfig `toml:"design"`
	ConfigFile            string       `flag:"c"`
	Quiet                 bool         `flag:"q"`
	Verbose               bool         `flag:"v"`
	RegexPattern          string       `flag:"f"`
	MaxInstances          int          `flag:"n"`
	Overwrite             bool         `flag:"r"`
	SkipRender            bool         `flag:"sr"`
	NoReadmeGeneration    bool         `flag:"nr"`
	SkipReadme            bool         `flag:"skip-readme"`
	CustomOpenSCADCommand string       `flag:"cmd"`
}

// ANSI escape codes for colored output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

/*
// generateCombinations recursively generates all combinations of parameter values
func generateCombinations(params map[string][]int, keys []string, index int, current map[string]int, result *[]map[string]int) {
	if index == len(keys) {
		// Add a copy of the current combination to the result
		combination := make(map[string]int)
		for k, v := range current {
			combination[k] = v
		}
		*result = append(*result, combination)
		return
	}

	key := keys[index]
	for _, value := range params[key] {
		current[key] = value
		generateCombinations(params, keys, index+1, current, result)
	}
}
*/

func logKeyValuePair(key string, value string) {
	log.Printf(colorYellow+"%s: "+colorWhite+"\t\t%s"+colorReset, key, value)
}

func logStage(stage string) {
	log.Printf(colorBlue+"%s"+colorReset, stage)
}

func generateSTL(instance InstanceConfig, config *Config) (string, error) {
	outputFileName := instance.Name
	designFileName := strings.Split(config.Design.InputPath, "/")[len(strings.Split(config.Design.InputPath, "/"))-1]

	outputPath := config.Design.OutputPath
	if outputPath == "" {
		outputPath = path.Join("./", designFileName)
	}

	exportFolderPath := strings.Clone(outputPath)
	if !strings.HasSuffix(outputPath, "/export") && !strings.HasSuffix(outputPath, "/export/") {
		exportFolderPath = path.Join(exportFolderPath, "export")
	} else {
		// Remove /export from outputPath
		outputPath = strings.TrimSuffix(outputPath, "/export")
	}

	if config.Design.Version == "" {
		config.Design.Version = "v0.1"
	}

	exportFolderPath = path.Join(exportFolderPath, config.Design.Version)

	if !config.Quiet && config.Verbose {
		logKeyValuePair("outputPath", outputPath)
		logKeyValuePair("exportFolderPath", exportFolderPath)
	}

	if config.Verbose {
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
		}
	}

	// copy design file to export folder
	designFilePath := config.Design.InputPath

	designFileCopyPath := path.Join(exportFolderPath, designFileName)

	fileExists, err := os.Stat(designFilePath)

	if err != nil {
		log.Panicf(colorRed+"Failed to check if design file exists: %s", err)
	}

	if fileExists.IsDir() {
		log.Panicf(colorRed+"Design file is a directory, not a file: %s", designFilePath)
	}

	if fileExists == nil {
		log.Printf(colorBlue + "Design file already exists in export folder, skipping copy" + colorReset)
	} else {

		err := Copy(designFilePath, designFileCopyPath)
		if err != nil {
			log.Panicf(colorRed+"Failed to copy design file to export folder: %s", err)
		} else if !config.Quiet && config.Verbose {
			log.Printf(colorBlue + "Copied .scad file to export folder" + colorReset)
		}
	}

	if outputFileName == "" {
		outputFileName = designFileName
		for name, value := range instance.Params {
			outputFileName += fmt.Sprintf("_%s%d", name, value)
		}
	}
	outputFileName += ".stl"

	outputPath = path.Join(exportFolderPath, outputFileName)
	args := []string{"-o", outputPath}
	if config.Quiet {
		args = append(args, "-q")
	}

	for name, value := range instance.Params {
		args = append(args, "-D", fmt.Sprintf("%s=%v", name, value))
	}
	args = append(args, config.Design.InputPath)

	args = append(args, "--export-format", "binstl")

	if !config.SkipRender {
		args = append(args, "--render")
	} else {
		log.Printf(colorYellow + "Skipping render" + colorReset)
	}

	if !config.Quiet && config.Verbose {
		logStage("Running openscad")
		if config.Verbose {
			logKeyValuePair("Command", fmt.Sprintf("openscad %v", strings.Join(args, " ")))
		}
	}

	command := "openscad"
	if config.CustomOpenSCADCommand != "" {
		command = config.CustomOpenSCADCommand
		logKeyValuePair("Custom OpenSCAD command", command)
	}
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return outputPath, cmd.Run()
}

// Define a struct to hold the command-line flags
type CmdFlags struct {
	Quiet                 bool
	Verbose               bool
	NoReadmeGeneration    bool
	RegexPattern          string
	MaxInstances          int
	ShowMan               bool
	ConfigFile            string
	SkipRender            bool
	SkipReadme            bool
	CustomOpenSCADCommand string
}

// loadConfig reads the configuration file and populates the Config struct
func loadConfig(configFile string, flags CmdFlags) (*Config, error) {
	var conf Config
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf(colorRed+"Failed to read config file: %v", err)
		return nil, err
	}

	_, err = toml.Decode(string(data), &conf)
	if err != nil {
		log.Printf(colorRed+"Failed to unmarshal config: %v", err)
		return nil, err
	}

	// Validate the config
	validate := validator.New()
	err = validate.Struct(conf)
	if err != nil {
		log.Printf(colorRed+"Config Validation failed: %v", err)
		return nil, err
	}

	// Merge command-line flags into the config
	conf.Quiet = flags.Quiet
	conf.Verbose = flags.Verbose
	conf.RegexPattern = flags.RegexPattern
	conf.MaxInstances = flags.MaxInstances
	conf.SkipRender = flags.SkipRender
	conf.NoReadmeGeneration = flags.NoReadmeGeneration
	conf.SkipReadme = flags.SkipReadme
	conf.CustomOpenSCADCommand = flags.CustomOpenSCADCommand

	if conf.Design.OutputPath == "" {
		conf.Design.OutputPath = getOutputPath(conf)
	}

	return &conf, nil
}

func printUsage() {
	log.Println("Usage of openscadgen:")
	log.Println("  -config, -c string")
	log.Println("        Path to config file")
	log.Println("  -h, -help, -man")
	log.Println("        Display this help message")
	log.Println("  -regex, -r string")
	log.Println("        Regex pattern to filter instances by name")
	log.Println("  -n, -max-instances int")
	log.Println("        Maximum number of instances to process")
	log.Println("  -q, -quiet")
	log.Println("        Quiet mode, no output")
	log.Println("  -v, -verbose")
	log.Println("        Verbose mode, more debug output")

	// Add more usage information as needed
}

func generateReadme(config *Config, designName string, version string, openscadVersion string) {
	if config.NoReadmeGeneration {
		log.Printf(colorYellow + "Skipping readme generation" + colorReset)
		return
	}

	contents := fmt.Sprintf("# %s\n\n%s\n\n", designName, config.Design.Description)
	contents += "## Table of Contents\n"
	for _, instance := range config.Design.Instances {
		contents += fmt.Sprintf("- [%s](#%s)\n", instance.Name, strings.ToLower(strings.ReplaceAll(instance.Name, " ", "-")))
	}
	contents += "\n"

	for _, instance := range config.Design.Instances {
		contents += fmt.Sprintf("## %s\n", instance.Name)
		if instance.Description != "" {
			contents += fmt.Sprintf("### Description\n%s\n", instance.Description)
		}
		contents += "### Parameters:\n"
		for name, value := range instance.Params {
			contents += fmt.Sprintf("- **%s**: %v\n", name, value)
		}
		contents += "\n"
	}

	// Optionally add a footer or additional information
	contents += "## Additional Information\n"
	contents += fmt.Sprintf("This README was generated by [openscadgen](https://github.com/KiwiKid/openscadgen) %s.\n and openscad version %s.\n", version, openscadVersion)

	readmePath := path.Join(config.Design.OutputPath, config.Design.Version, "README.md")

	readmeFile, err := os.Create(readmePath)
	if err != nil {
		log.Panicf(colorRed+"Failed to create README.md file: %s", err)
	}
	defer readmeFile.Close()

	_, err = readmeFile.WriteString(contents)
	if err != nil {
		log.Panicf(colorRed+"Failed to write to README.md file: %s", err)
	}
}

func getOutputPath(config Config) string {
	if config.Design.OutputPath != "" {
		return config.Design.OutputPath
	}

	// If output_path is not specified, derive it from input_path
	inputPath := config.Design.InputPath
	designFilename := filepath.Base(inputPath)
	designName := strings.TrimSuffix(designFilename, filepath.Ext(designFilename))

	// Construct the new output path
	return filepath.Join(filepath.Dir(inputPath), designName)
}

const OPENSCAD_VERSION_WARN_IF_OLDER_THAN = 2024

func main() {
	VERSION := "v0.9.5-alpha"

	startTime := time.Now()

	// Create an instance of the CmdFlags struct
	cmdFlags := CmdFlags{}

	// Parse command-line flags into the struct
	flag.StringVar(&cmdFlags.ConfigFile, "config", "", "Path to config file")
	flag.StringVar(&cmdFlags.ConfigFile, "c", "", "Alias for -config")

	flag.BoolVar(&cmdFlags.ShowMan, "man", false, "Display help message")
	flag.BoolVar(&cmdFlags.ShowMan, "m", false, "Alias for -man")
	flag.BoolVar(&cmdFlags.ShowMan, "h", false, "Alias for -man")

	flag.StringVar(&cmdFlags.RegexPattern, "regex", "", "Regex pattern to filter instances by name")
	flag.StringVar(&cmdFlags.RegexPattern, "r", "", "Alias for -regex")

	flag.BoolVar(&cmdFlags.Quiet, "quiet", false, "quiet mode, no output")
	flag.BoolVar(&cmdFlags.Quiet, "q", false, "Alias for -quiet")

	flag.BoolVar(&cmdFlags.Verbose, "verbose", false, "verbose mode, more output")
	flag.BoolVar(&cmdFlags.Verbose, "v", false, "Alias for -verbose")

	flag.BoolVar(&cmdFlags.SkipRender, "manifold", false, "Dont run a render before export")
	flag.BoolVar(&cmdFlags.SkipRender, "rm", false, "Alias for -manifold")

	flag.BoolVar(&cmdFlags.NoReadmeGeneration, "nr", false, "Skip readme generation")
	flag.BoolVar(&cmdFlags.NoReadmeGeneration, "no-readme", false, "Alias for -nr")

	flag.IntVar(&cmdFlags.MaxInstances, "n", 0, "Maximum number of instances to process")

	// Load configuration
	flag.Parse()

	if cmdFlags.ShowMan {
		printUsage()
		return
	}

	if cmdFlags.ConfigFile == "" {
		log.Println(colorRed + "No config file provided, use -c or -config to specify a config file" + colorReset)
		printUsage()
		os.Exit(1)
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
		logKeyValuePair("Openscadgen version", VERSION)
	}

	cmd := exec.Command("openscad", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Printf(colorRed+"Failed to get openscad version: %s"+colorReset, err)
	}

	openscadVersion := strings.TrimSuffix(out.String(), "\n")
	openscadVersionNumberStr := strings.Replace(openscadVersion, "OpenSCAD version ", "", 1)
	openscadVersionNumber, err := strconv.ParseFloat(openscadVersionNumberStr, 64)
	if err != nil {
		log.Printf(colorRed+"Failed to get openscad version number from %s: %s"+colorReset, openscadVersion, err)
	}

	if err != nil {
		log.Printf(colorRed+"Failed to get openscad version: %s"+colorReset, err)
	} else if len(openscadVersion) == 0 {
		log.Printf(colorRed + "Openscad version output is empty. Please check if OpenSCAD is installed and accessible." + colorReset)
	} else {
		logKeyValuePair("Openscad version found", openscadVersion)
		if openscadVersionNumber < OPENSCAD_VERSION_WARN_IF_OLDER_THAN {
			log.Printf(colorRed+"Openscad version is older than the latest available (%d), consider updating to the latest version of openscad as has more features and improved rendering time"+colorReset, OPENSCAD_VERSION_WARN_IF_OLDER_THAN)
		}
	}

	logStage("Loading config file")
	if cmdFlags.Verbose {
		logKeyValuePair("Config file", cmdFlags.ConfigFile)
	}
	config, err := loadConfig(cmdFlags.ConfigFile, cmdFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Failed to load config: %v\n"+colorReset, err)
		os.Exit(1)
	}

	if config.Verbose {
		logKeyValuePair("Loaded config", fmt.Sprintf("%+v", config))
	}

	design := config.Design

	if !config.Quiet {
		log.Printf(colorBlue+"Config provided %d possible instances to generate from scad file at: "+colorReset+"%s", len(design.Instances), design.InputPath)
		if config.Verbose {
			logKeyValuePair("Input File", design.InputPath)
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

	if !config.Quiet {
		logStage("Starting STL generation")
		if config.RegexPattern != "" {
			logKeyValuePair("Filter: Only generating file matching pattern", config.RegexPattern)
		}
	}
	// Generate STL files for each instance
	processedCount := 0
	skippedCount := 0
	for _, instance := range design.Instances {

		name := instance.Name
		if name == "" {
			// Create a default name using up to the first 3 parameters
			paramCount := 0
			for paramName, paramValue := range instance.Params {
				if paramCount >= 3 {
					break
				}
				if paramCount > 0 {
					name += "_"
				}
				name += fmt.Sprintf("%s%d", paramName, paramValue)
				paramCount++
			}
		}

		prefix := fmt.Sprintf("Generating: '%s' (%d/%d) ", name, processedCount, len(design.Instances))

		if regex != nil && !regex.MatchString(name) {
			if !config.Quiet && config.Verbose {
				log.Printf(colorYellow+"Skipping instance %s as it does not match the regex pattern", name)
			}
			skippedCount++
			continue
		}

		if !config.Quiet {
			logStage(fmt.Sprintf("========== %s ==========", prefix))
		}

		exportSTLPath, err := generateSTL(instance, config)

		if err != nil {
			fmt.Fprintf(os.Stderr, colorRed+"Error generating STL for instance %s: %v\n"+colorReset, instance.Name, err)
		} else if !config.Quiet {
			log.Printf(colorGreen+"%s was successful\n"+colorReset, prefix)
			logKeyValuePair("Outputting to", exportSTLPath)
			processedCount++

		}

		if config.MaxInstances > 0 && processedCount >= config.MaxInstances {
			log.Printf(colorBlue + "Max instances processed, stopping")
			break
		}
	}

	generateReadme(config, design.Name, VERSION, openscadVersion)
	if !config.Quiet {
		logStage("STL generation completed")

		msg := ""
		if skippedCount > 0 {
			msg += fmt.Sprintf(colorYellow+"%d instances were skipped as they did not match the regex pattern\n\n"+colorReset, skippedCount)
		}
		msg += fmt.Sprintf(colorGreen+"STL generation completed. %d .stl instances generated in %s\n\n"+colorReset, processedCount, time.Since(startTime))

		log.Print(msg)
	}
}
