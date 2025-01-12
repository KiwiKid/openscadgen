package main

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
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/xattr"
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

func getFileName(path string) string {
	fileName := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
	if strings.Contains(fileName, ".") {
		fileName = fileName[:strings.LastIndex(fileName, ".")]
	}
	return fileName
}

// InstanceConfig represents a single instance configuration
type InstanceConfig struct {
	Name         string                 `toml:"name"`
	AutoName     string                 `toml:"auto_name"`
	Description  string                 `toml:"description,omitempty"`
	InputPath    string                 `toml:"input_path"`
	Params       map[string]interface{} `toml:"params"`
	PartIDLetter string                 `toml:"part_id_letter"`
	isDynamic    bool
}

type DynamicInstanceConfig struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description,omitempty"`
	Params      map[string]string `toml:"params"`
}

func getInstanceConfigSaveLocation(config *Config, instance *InstanceConfig, inputPath string) string {

	fileName := getFileName(inputPath)
	if fileName == "" {
		log.Panicf("designFileName is invalid for dynamic instance %s", instance.Name)
	}

	formatToUse := instance.Name
	if formatToUse == "" {
		formatToUse = config.Design.ExportNameFormat
	}

	if formatToUse != "" {
		for key, value := range instance.Params {
			formatToUse = strings.ReplaceAll(formatToUse, "{"+key+"}", fmt.Sprintf("%v", value))
		}

		if strings.Contains(formatToUse, "{designFileName}") {
			formatToUse = strings.ReplaceAll(formatToUse, "{designFileName}", fileName)
		}
	} else {
		formatToUse = fileName
	}

	logKeyValuePair("exportNameFormat in getInstanceConfigSaveLocation", formatToUse)
	res := path.Join(config.Design.OutputPath, config.Design.Version, formatToUse+".stl")
	return res
}

func generateDynamicInstances(config *Config) []InstanceConfig {
	instances := []InstanceConfig{}
	// for each key in params, split the value by comma, parse the value (1,2,3 or 1-5, or 1,2,5-10) into a range of specific values and then generate all combinations of those values

	if len(config.Design.DynamicInstanceConfig) == 0 {
		return []InstanceConfig{}
	}

	inputPaths := config.getInputPaths()

	for _, inputPath := range inputPaths {
		// Reset PartIDLetter for each new filepath
		for diIndex, dynamicInstance := range config.Design.DynamicInstanceConfig {
			paramCombinations := map[string][]interface{}{}

			for key, value := range dynamicInstance.Params {
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

					if strings.Contains(fileName, ".") {
						fileName = fileName[:strings.LastIndex(fileName, ".")]
					}
					if fileName == "" || fileName == "." {
						log.Panicf("designFileName is invalid for dynamic instance %s", dynamicInstance.Name)
					} else {
						current["designFileName"] = fileName
					}

					instances = append(instances, InstanceConfig{
						Name:        dynamicInstance.Name,
						AutoName:    instanceName,
						isDynamic:   true,
						Description: dynamicInstance.Description,
						InputPath:   inputPath.Path,
						Params:      copyMap(current),
					})
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
			if config.Verbose {
				logKeyValuePair(fmt.Sprintf("Generated PartIDLetter from index (%d) for instance %s", index, instance.Name), letter)
			}
		}
	}

	if config.Verbose {
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

type InputPath struct {
	Path             string `toml:"path" validate:"required"`
	ExportNameFormat string `toml:"export_name_format"`
}

type DesignConfig struct {
	Name           string      `toml:"name"`
	Description    string      `toml:"description"`
	InputPath      string      `toml:"input_path"`
	InputPaths     []InputPath `toml:"input_paths"`
	OutputPath     string      `toml:"output_path" validate:"required"`
	Version        string      `toml:"version"`
	NoPartIDLetter bool        `toml:"no_part_id_letter"`
	// @@ deprecated
	ExportNameFormat      string                  `toml:"export_name_format"`
	CustomOpenSCADArgs    string                  `toml:"custom_openscad_args"`
	Instances             []InstanceConfig        `toml:"instances"`
	DynamicInstanceConfig []DynamicInstanceConfig `toml:"dynamic_instances"`
}

// Config holds the overall configuration structure
type Config struct {
	Design                       DesignConfig `toml:"openscadgen"`
	ConfigFile                   string       `flag:"c,config"`
	Quiet                        bool         `flag:"q"`
	Verbose                      bool         `flag:"v"`
	RegexPattern                 string       `flag:"f"`
	MaxInstances                 int          `flag:"n"`
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

func (config *Config) getInputPaths() []InputPath {
	if len(config.Design.InputPaths) > 0 {
		return config.Design.InputPaths
	}
	return []InputPath{
		{Path: config.Design.InputPath, ExportNameFormat: config.Design.ExportNameFormat},
	}
}

var config Config

var logger *log.Logger

// ANSI escape codes for colored output
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
func initLogger(logFilePath string) error {
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
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

func getOrMakeExportFolder(config *Config) (string, error) {
	if config.Verbose {
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
	} else {
		// Remove /export from outputPath
		outputPath = strings.TrimSuffix(outputPath, "/export")
	}

	if config.Design.Version == "" {
		config.Design.Version = "v0.1"
	}

	fullExportFolderPath := path.Join(exportFolderPath, config.Design.Version)

	if config.IncludePartIDLetter {
		fullExportFolderPath = path.Join(fullExportFolderPath, "with_embedded_part_letter")
	}

	// Check if exportFolderPath has any files or directories
	if files, err := os.ReadDir(fullExportFolderPath); err == nil && len(files) > 0 {
		filesStr := ""
		for i, file := range files {
			if i < 5 {
				filesStr += fmt.Sprintf("\t- %s\n", file.Name())
			} else {
				filesStr += fmt.Sprintf("\tand %d other files ...\n", len(files)-5)
				break
			}
		}

		if !config.Quiet {
			log.Printf(colorYellow + "(tip: if you want to keep the existing stl export files, cancel this run and update the 'version' in the config file, this will generate a new folder and keep the existing files)")
			log.Printf(colorYellow + "(the '-ow' flag will skip this check)")
		}
		logWarn(fmt.Sprintf("\nThe export folder (%s) has %d existing files: \n%s\n\n"+colorRed+" %d files will be deleted from: \n\t%s\n", fullExportFolderPath, len(files), filesStr, len(files), fullExportFolderPath), false)
		logWarn(colorOrange+"\n\nDo you want to continue? (y/n):\n"+colorReset, false)
		if !config.OverwriteExisting {

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if response != "y\n" && response != "Y\n" {
				fmt.Println("Aborting operation.")
				os.Exit(1)
			}
		} else if config.Verbose {
			logKeyValuePair("OverwriteExisting set, skipping check", fullExportFolderPath)
		}

		if !config.Quiet {
			logStage(fmt.Sprintf("Clearing %d files from export folder", len(files)))
		}
		err := os.RemoveAll(fullExportFolderPath)
		if config.Verbose {
			logKeyValuePair("Removed files from export folder", exportFolderPath)
		}
		if err != nil {
			log.Printf(colorRed+"Failed to remove file %s: %v", exportFolderPath, err)
		}

	}

	if config.Verbose {
		logStage("Creating export folder")
		logKeyValuePair("Export folder", fullExportFolderPath)
	}
	os.MkdirAll(fullExportFolderPath, 0755)

	return fullExportFolderPath, nil
}

func SetMetadata(fileName string, metadata map[string]string, config *Config) error {
	if !config.Quiet {
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
	if config.Verbose {
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
			if config.Verbose {
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
			if config.Verbose {
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
	log.Printf("\n\nGENERATE STL FOR inputPath: %s", instance.InputPath)

	name := instance.Name
	if name == "" {
		name = filepath.Base(instance.InputPath)
	}
	outputPath := path.Join(exportFolderPath, name)
	if !config.Quiet && config.Verbose {
		logKeyValuePair("outputPath", outputPath)
		logKeyValuePair("[1]exportFolderPath", exportFolderPath)
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
			if !config.Quiet {
				logCreation(fmt.Sprintf("Created export folder: %s", exportFolderPath))
			}
		}
	}

	// copy design file to export folder
	//designFilePath := config.Design.InputPath

	if config.Verbose {
		logStage("Create export folder")
		logKeyValuePair("Design file", instance.InputPath)
		logKeyValuePair("Export folder", exportFolderPath)
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

	if config.Verbose {
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
		} else if !config.Quiet && config.Verbose {
			log.Printf(colorBlue + "Copied .scad file to export folder" + colorReset)
		}
	} else {
		if !config.Quiet {
			log.Printf(colorBlue + "Design file already exists in export folder, skipping copy" + colorReset)
		}
		if config.Verbose {
			logKeyValuePair("Design file", designFileCopyPath)
		}
	}
	designFileName := strings.Split(filepath.Base(instance.InputPath), ".")[0]

	//if outputFileName == "" {

	outputFileName := fmt.Sprintf("%s", designFileName)
	for name, value := range instance.Params {
		outputFileName += fmt.Sprintf("_%s%v", name, value)
	}
	if config.Verbose {
		logKeyValuePair("outputFileName[1]", outputFileName)
		logKeyValuePair("Design file name", designFileName)
	}
	//}

	outputFileName += ".stl"

	logKeyValuePair("exportFolderPath", exportFolderPath)
	logKeyValuePair("outputFileName", outputFileName)

	outputPath = getInstanceConfigSaveLocation(config, &instance, instance.InputPath)
	logKeyValuePair("outputPath from getInstanceConfigSaveLocation", outputPath)
	args := []string{"-o", outputPath}

	if config.Verbose {
		logKeyValuePair("creating output folder", outputPath)
	}
	outputFolder := filepath.Dir(outputPath)
	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		os.MkdirAll(outputFolder, 0755)
		if config.Verbose {
			logKeyValuePair("created output folder", outputPath)
		}
	}

	if config.Verbose {
		logKeyValuePair("output folder confirmed", outputPath)
	}

	if config.Quiet {
		args = append(args, "-q")
	}

	for name, value := range instance.Params {
		if config.Verbose {
			logKeyValuePair(fmt.Sprintf("CustomParameter [%s]", name), fmt.Sprintf("%v", value))
		}
		args = append(args, "-D", fmt.Sprintf("%s=%v", name, value))
	}

	if config.IncludePartIDLetter {
		args = append(args, "-D", fmt.Sprintf("'optional_part_id_letter=\"%s\"'", instance.PartIDLetter))
		if config.Verbose {
			logKeyValuePair("OptionalPartIDLetter set on model", instance.PartIDLetter)
		}
	} else {
		if config.Verbose {
			logKeyValuePair("OptionalPartIDLetter NOT set on model", "false")
		}
	}

	if config.OverrideFN > 0 {
		args = append(args, "-D", fmt.Sprintf("'$fn=%d'", config.OverrideFN))
		if config.Verbose {
			logKeyValuePair("OverrideFN", fmt.Sprintf("%d", config.OverrideFN))
		}
	}

	if config.Verbose || true {
		logKeyValuePair("InputPath", instance.InputPath)
	}
	args = append(args, instance.InputPath)

	args = append(args, "--export-format", "binstl")

	args = append(args, "--backend=manifold")

	if !config.SkipRender {
		args = append(args, "--render")
	} else if !config.Quiet {
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
		if !config.Quiet {
			logKeyValuePair("Custom OpenSCAD command", command)
		}
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	commandStr := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	//if !config.Quiet {
	log.Printf("Running command: %s", commandStr)
	//	}
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
		if config.Verbose {
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

	if !config.Quiet {
		logStage("Finished generating STL")
	}

	return outputPath, nil
}

// Define a struct to hold the command-line flags
type CmdFlags struct {
	Quiet                        bool
	Verbose                      bool
	RegexPattern                 string
	MaxInstances                 int
	OverwriteExisting            bool
	ShowMan                      bool
	ConfigFile                   string
	SkipRender                   bool
	SkipReadme                   bool
	CustomOpenSCADCommand        string
	Concurrent                   bool
	MaxConcurrentRequests        int
	IncludePartIDLetter          bool
	SetBuildInfoInFileAttributes bool
	OverrideFN                   int
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
	conf.SkipReadme = flags.SkipReadme
	conf.OverwriteExisting = flags.OverwriteExisting
	conf.CustomOpenSCADCommand = flags.CustomOpenSCADCommand
	conf.Concurrent = flags.Concurrent
	conf.MaxConcurrentRequests = flags.MaxConcurrentRequests
	conf.IncludePartIDLetter = flags.IncludePartIDLetter
	conf.SetBuildInfoInFileAttributes = flags.SetBuildInfoInFileAttributes

	conf.OverrideFN = flags.OverrideFN

	if conf.Design.OutputPath == "" {
		conf.Design.OutputPath = getOutputPath(conf)
	}

	if conf.Design.DynamicInstanceConfig != nil {
		if len(conf.Design.DynamicInstanceConfig) > 0 {
			for dynamicInstanceIndex, dynamicInstance := range conf.Design.DynamicInstanceConfig {
				for paramName, _ := range dynamicInstance.Params {

					if len(conf.Design.InputPaths) > 1 {
						nameHasDesignFileName := strings.Contains(dynamicInstance.Name, "{designFileName}")
						if !nameHasDesignFileName {
							log.Panicf("designFileName is required when a build has multiple input paths for dynamic instances %s. (add {designFileName} to the name)", dynamicInstance.Name)
						}
					}

					nameHasParams := strings.Contains(dynamicInstance.Name, fmt.Sprintf("{%s}", paramName))
					if !nameHasParams {
						logKeyValuePair("Dynamic instance index", fmt.Sprintf("%d", dynamicInstanceIndex))
						logKeyValuePair("Dynamic instance name", dynamicInstance.Name)
						logKeyValuePair("Missing Param name", paramName)
						logKeyValuePair("Config file", configFile)
						logWarn(fmt.Sprintf(`Dynamic instance name: 
%s 

does not contain param:
 %s

Include every param in the export name (in the format '{param_name}') to ensure all instances are generated.`, dynamicInstance.Name, paramName), true)
						os.Exit(1)
					}
				}
			}

			if conf.Design.ExportNameFormat != "" {
				for _, paramName := range conf.Design.DynamicInstanceConfig {
					if !strings.Contains(conf.Design.ExportNameFormat, fmt.Sprintf("{%s}", paramName)) {
						logKeyValuePair("Export name format", conf.Design.ExportNameFormat)
						logKeyValuePair("Missing Param name", paramName.Name)
						logKeyValuePair("Config file", configFile)
						logWarn(fmt.Sprintf(`Export name format: 
	%s 
	
does not contain param 

%s

	Include every param in the export name (in the format '{param_name}') to ensure all instances are generated.					 
`, conf.Design.ExportNameFormat, paramName.Name), true)
						os.Exit(1)
					}
				}
			}

		}
	}

	return &conf, nil
}

func generateReadme(config *Config, dynamicInstances []InstanceConfig, designName string, version string, openscadVersion string) {
	if config.SkipReadme {
		log.Printf(colorYellow + "Skipping readme generation" + colorReset)
		return
	}

	if !config.Quiet {
		logStage("Generating README.md")
	}

	contents := fmt.Sprintf("# %s\n\n%s\n\n", designName, config.Design.Description)
	contents += "## Table of Contents\n"
	for _, instance := range config.Design.Instances {
		contents += fmt.Sprintf("- [%s](#%s)\n", instance.Name, strings.ToLower(strings.ReplaceAll(instance.Name, " ", "-")))
	}
	for _, instance := range dynamicInstances {
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

	for _, instance := range dynamicInstances {
		contents += fmt.Sprintf("## %s\n", instance.Name)
		for name, value := range instance.Params {
			contents += fmt.Sprintf("- **%s**: %v\n", name, value)
		}
		contents += "\n"
	}

	// Optionally add a footer or additional information
	contents += "## Additional Information\n"
	contents += fmt.Sprintf("This README was generated by [openscadgen](https://github.com/KiwiKid/openscadgen) %s %s.\n", version, openscadVersion)

	readmePath := path.Join(config.Design.OutputPath, config.Design.Version, "README.md")

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

func findOpenSCAD() string {
	// Try to find openscad using `which` command
	cmd := exec.Command("which", "openscad")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("OpenSCAD not found in PATH.")
	}
	return string(output)
}

func main() {
	// Set the PATH environment variable
	VERSION := "v1.1.5-ALPHA"

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

	flag.BoolVar(&cmdFlags.SkipReadme, "skip-readme", false, "Skip generating a README.md file")
	flag.BoolVar(&cmdFlags.SkipReadme, "sr", false, "Alias for -skip-readme")

	flag.IntVar(&cmdFlags.MaxInstances, "n", 0, "Maximum number of instances to process")

	flag.BoolVar(&cmdFlags.OverwriteExisting, "ow", false, "Overrwite existing files")
	flag.BoolVar(&cmdFlags.OverwriteExisting, "overwrite", false, "Alias for -ow")

	flag.BoolVar(&cmdFlags.IncludePartIDLetter, "pid", false, "Include optional_part_id_letter variable in the call the openscad")

	flag.BoolVar(&cmdFlags.Concurrent, "concurrent", false, "Run instances concurrently")
	flag.BoolVar(&cmdFlags.Concurrent, "con", false, "Run instances concurrently")

	flag.IntVar(&cmdFlags.MaxConcurrentRequests, "max-concurrent-requests", 10, "Maximum number of concurrent requests")

	flag.StringVar(&cmdFlags.CustomOpenSCADCommand, "custom-openscad-command", "", "Custom OpenSCAD command to use")

	flag.IntVar(&cmdFlags.OverrideFN, "fn", 0, "Override the default fn value (default none)")

	flag.BoolVar(&cmdFlags.SetBuildInfoInFileAttributes, "fi", true, "Set build info in file attributes (default true)")
	// Create a logger that writes to both the file and stdout (console)

	// Load configuration
	flag.Parse()

	initLogger("openscadgen.log")

	if cmdFlags.Concurrent {
		logWarn("Concurrent mode is not (yet) supported", true)

	}

	if cmdFlags.ShowMan {
		flag.PrintDefaults()
		return
	}

	if cmdFlags.ConfigFile == "" {
		flag.PrintDefaults()

		logWarn("No config file provided, use -c or -config to specify a config file", true)
		os.Exit(1)
	}

	if cmdFlags.Verbose && cmdFlags.Quiet {
		log.Print("**whispers**WHAT DO YOU WANT FROM ME? Being quiet and verbose is better suited for when writing passive aggressive post-it notes **whispers**")
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

	openscadPath := findOpenSCAD()
	//	os.Setenv("PATH", openscadPath+":"+os.Getenv("PATH"))

	if config.Verbose {
		log.Printf("Openscad path: %s", openscadPath)
	}
	cmd := exec.Command("openscad", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Panicf(colorRed+"Failed to get openscad version: %s"+colorReset, err)
	}

	openscadVersion := strings.TrimSuffix(out.String(), "\n")
	openscadVersionNumberStr := strings.Replace(openscadVersion, "OpenSCAD version ", "", 1)

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
		if config.Verbose {
			logKeyValuePair("OpenSCAD version", openscadVersion)
			logKeyValuePair("OpenSCAD year", fmt.Sprintf("%d", openscadYear))
			logKeyValuePair("OpenSCAD version to", fmt.Sprintf("%d", openscadYear))
		}
		if openscadYear < OPENSCAD_VERSION_WARN_IF_OLDER_THAN {
			logWarn("OpenSCAD version is older than the latest available (2024), consider updating to the latest version of OpenSCAD as it has more features and improved rendering time", true)
		}

	}

	if !cmdFlags.Quiet {
		logStage("Loading config file")
		if cmdFlags.Verbose {
			log.Printf("Config file %s", cmdFlags.ConfigFile)
		}
	}
	config, err := loadConfig(cmdFlags.ConfigFile, cmdFlags)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Failed to load config: %v\n"+colorReset, err)
		os.Exit(1)
	}

	if config.Verbose {
		log.Printf("Loaded config %+v", config)
	}

	design := config.Design
	dynamicInstances := generateDynamicInstances(config)

	for _, instance := range dynamicInstances {
		if instance.PartIDLetter == "" {
			log.Panicf(colorRed + "PartIDLetter is required for dynamic instances, please set the PartIDLetter field for each instance")
		}
	}

	if !config.Quiet {

		log.Printf(colorBlue+"Config provided %d possible instances "+colorYellow+"(%d specific, %d dynamic)"+colorBlue+" to generate from scad file '%s'"+colorReset, len(design.Instances)+len(dynamicInstances), len(design.Instances), len(dynamicInstances), design.Name)
		logKeyValuePair("Input File", design.InputPath)
		logKeyValuePair("Design Name", design.Name)
		logKeyValuePair("Design Version", design.Version)
		if config.MaxInstances > 0 {
			logWarn(fmt.Sprintf("Max Limit of %d instances", config.MaxInstances), false)
		}
		if config.RegexPattern != "" {
			logWarn(fmt.Sprintf("Filter to: %s", config.RegexPattern), false)
		}
		if config.Verbose {
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

	pathsToProcess := config.getInputPaths()

	for _, path := range pathsToProcess {

		designFileExists, err := os.Stat(path.Path)
		if err != nil {

			logWarn(fmt.Sprintf("Could not find scad file: %s", err), true)
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

	exportFolderPath, err := getOrMakeExportFolder(config)
	if err != nil {
		log.Printf(colorRed+"Failed to get or make export folder: %s", err)
		os.Exit(1)
	}

	initLogger(path.Join(exportFolderPath, "export_log.log"))

	// Create a WaitGroup to manage concurrency
	var wg sync.WaitGroup
	errorChannel := make(chan error, len(dynamicInstances)+len(design.Instances))

	// Generate STL files for dynamic instances
	if len(dynamicInstances) > 0 {
		logStage(fmt.Sprintf("Starting Dynamic %d Instances", len(dynamicInstances)))
	}

	//if config.Verbose {
	logStage("\n\n\nGot %d paths to process")

	for _, path := range pathsToProcess {
		logKeyValuePair("Path", path.Path)
	}
	//}

	processedCount := 0
	skippedCount := 0
	for _, path := range pathsToProcess {
		logKeyValuePair("=============== path", path.Path)
		stlIndex := 0

		for diIndex, instance := range dynamicInstances {
			if regex != nil && !regex.MatchString(instance.Name) {
				if !config.Quiet && config.Verbose {
					log.Printf(colorYellow+"Skipping instance %s as it does not match the regex pattern", instance.Name)
				}
				skippedCount++
				continue
			}

			if !config.Quiet {
				logStage(fmt.Sprintf("\nDynamic Model - (%d/%d) [%d processed] - '%s' ", diIndex, len(dynamicInstances), processedCount, instance.Name))
				if config.Verbose {
					logKeyValuePair("Params", fmt.Sprintf("%+v", instance.Params))
				}
			}

			if config.MaxInstances > 0 && processedCount >= config.MaxInstances {
				log.Printf(colorBlue+"Max instance of %d processed, stopping", config.MaxInstances)
				break
			} else if config.Verbose {
				logStage("Max instance check passed")
				logKeyValuePair("Max instances", fmt.Sprintf("%d", config.MaxInstances))
				logKeyValuePair("Processed instances", fmt.Sprintf("%d", processedCount))
			}

			if config.Concurrent {
				wg.Add(1)

				// Run generateSTL concurrently if Concurrent flag is set
				go func(instance InstanceConfig, inputPath string) {
					defer wg.Done()

					exportSTLPath, err := generateSTL(instance, config, exportFolderPath)
					if err != nil {
						errorChannel <- fmt.Errorf("(concurrent) Error generating STL for instance %s: %v", instance.Name, err)
					} else if !config.Quiet {
						logKeyValuePair("Outputting to", exportSTLPath)
					}
					processedCount++

				}(instance, path.Path)
			} else {
				_, err := generateSTL(instance, config, exportFolderPath)
				if err != nil {
					logWarn(fmt.Sprintf("Error generating STL for instance '%s': %v", instance.Name, err), false)
				} else if !config.Quiet {
					processedCount++
				}

			}
			stlIndex++
		}

		if config.Concurrent {

			logStage("Waiting for dynamic instances to finish")

			// Wait for all goroutines to finish
			wg.Wait()
			close(errorChannel)

			// Handle any errors that occurred during concurrent execution
			for err := range errorChannel {
				fmt.Fprintf(os.Stderr, colorRed+"%v\n"+colorReset, err)
			}
		}

		// Generate STL files for specific instances
		if len(design.Instances) > 0 {
			logStage("Starting Specific Instances")
			for index, instance := range design.Instances {

				name := instance.Name
				instance.PartIDLetter = getPartIDLetter(index + len(dynamicInstances))
				if config.Verbose {
					logKeyValuePair(fmt.Sprintf("Generated PartIDLetter from index (%d)", index), instance.PartIDLetter)
				}

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

				prefix := fmt.Sprintf("Generating (%d/%d) - '%s' ", processedCount, len(design.Instances), name)

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

				if config.Concurrent {
					wg.Add(1)

					// Run generateSTL concurrently if Concurrent flag is set
					go func(instance InstanceConfig) {
						defer wg.Done()

						exportSTLPath, err := generateSTL(instance, config, exportFolderPath)
						if err != nil {
							errorChannel <- fmt.Errorf("Error generating STL for instance %s: %v", instance.Name, err)
						} else if !config.Quiet {
							logKeyValuePair("Outputting to", exportSTLPath)
							processedCount++
						}
						wg.Done()

					}(instance)
				} else {
					exportSTLPath, err := generateSTL(instance, config, exportFolderPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, colorRed+"Error generating STL for instance %s: %+v\n"+colorReset, instance.Name, err)
					} else {
						processedCount++

						if !config.Quiet {
							log.Printf(colorGreen+"%s successfully in %s\n"+colorReset, prefix, time.Since(startTime))
							logKeyValuePair("Outputting to", exportSTLPath)
						}
					}
				}

				if config.MaxInstances > 0 && processedCount >= config.MaxInstances {
					log.Printf(colorBlue+"Max instance of %d processed, stopping", config.MaxInstances)
					break
				}
			}
		} else if config.Verbose {
			logStage("No Static Instances")
		}
	}

	generateReadme(config, dynamicInstances, design.Name, VERSION, openscadVersion)
	if !config.Quiet {
		logStage("STL generation completed")

		msg := ""
		if skippedCount > 0 {
			msg += fmt.Sprintf(colorYellow+"%d instances were skipped as they did not match the regex pattern\n\n"+colorReset, skippedCount)
		}
		msg += fmt.Sprintf(colorGreen+"openscadgen completed! \n\n%d stl files generated in %s\n\n\nthanks for using openscadgen! "+colorReset, processedCount, time.Since(startTime))

		log.Print(msg)
	}
}
