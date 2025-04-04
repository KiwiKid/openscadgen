package main

import (
	"flag"
	"log"
	"github.com/kiwikid/openscadgen/pkg/models"

	"github.com/kiwikid/openscadgen/pkg"
)

// https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file
// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. If that fail, copy the file contents from src to dst.

func main() {

	cmdFlags := models.CmdFlags{}

	// Parse command-line flags into the struct
	flag.StringVar(&cmdFlags.ConfigFile, "config", "", "Path to config file")
	flag.StringVar(&cmdFlags.ConfigFile, "c", "", "Alias for -config")

	flag.BoolVar(&cmdFlags.ShowMan, "man", false, "Display help message")
	flag.BoolVar(&cmdFlags.ShowMan, "m", false, "Alias for -man")
	flag.BoolVar(&cmdFlags.ShowMan, "h", false, "Alias for -man")

	flag.StringVar(&cmdFlags.InitProjectName, "init", "", "Initialize a new project at the current directory with the given name")
	flag.StringVar(&cmdFlags.InitProjectName, "i", "", "Alias for -init")

	flag.StringVar(&cmdFlags.InitProjectNameExtended, "init-extended", "", "Initialize a new project at the current directory with the given name - with bosl2 and renderSlicing support")
	flag.StringVar(&cmdFlags.InitProjectNameExtended, "ie", "", "Alias for -init")

	flag.StringVar(&cmdFlags.RegexPattern, "regex", "", "Regex pattern to filter instances by name")
	flag.StringVar(&cmdFlags.RegexPattern, "r", "", "Alias for -regex")

	flag.BoolVar(&cmdFlags.Quiet, "quiet", false, "quiet mode, no log output")
	flag.BoolVar(&cmdFlags.Quiet, "q", false, "Alias for -quiet")

	flag.BoolVar(&cmdFlags.NoProcessing, "no-processing", false, "'dry-run' mode - will check config and provide instances that will be processed, but not do any processing")
	flag.BoolVar(&cmdFlags.NoProcessing, "np", false, "Alias for -no-processing")

	flag.BoolVar(&cmdFlags.Debug, "debug", false, "debug mode, more output")
	flag.BoolVar(&cmdFlags.Debug, "d", false, "Alias for -debug")

	flag.BoolVar(&cmdFlags.Version, "version", false, "just output the openscadgen and openscad version number")
	flag.BoolVar(&cmdFlags.Version, "v", false, "Alias for -version")

	flag.BoolVar(&cmdFlags.SkipRender, "manifold", false, "Dont run a render before export")
	flag.BoolVar(&cmdFlags.SkipRender, "rm", false, "Alias for -manifold")

	flag.BoolVar(&cmdFlags.SkipReadme, "skip-readme", false, "Skip generating a README.md file")
	flag.BoolVar(&cmdFlags.SkipReadme, "sr", false, "Alias for -skip-readme")

	flag.IntVar(&cmdFlags.MaxInstances, "n", 0, "Maximum number of instances to process")

	flag.BoolVar(&cmdFlags.ContinueOnError, "co", false, "Continue on error")
	flag.BoolVar(&cmdFlags.ContinueOnError, "continue-on-error", false, "Alias for -co")

	flag.BoolVar(&cmdFlags.IncludeExportLog, "include-export-log-file", false, "Include the export log in the README.md file")
	flag.BoolVar(&cmdFlags.IncludeExportLog, "el", false, "Alias for -include-export-log-file")

	flag.BoolVar(&cmdFlags.FullExport, "full-export", false, "Include the the config file and extended output ")
	flag.BoolVar(&cmdFlags.FullExport, "fe", false, "Alias for -full-export")

	flag.BoolVar(&cmdFlags.OverwriteExisting, "ow", false, "Overrwite existing files")
	flag.BoolVar(&cmdFlags.OverwriteExisting, "overwrite", false, "Alias for -ow")

	flag.BoolVar(&cmdFlags.IncludePartIDLetter, "pid", false, "Include optional_part_id_letter variable in the call the openscad")

	flag.StringVar(&cmdFlags.CustomOpenSCADCommand, "custom-openscad-command", "", "Custom OpenSCAD command to use")

	flag.IntVar(&cmdFlags.OverrideFN, "fn", 0, "Override the default fn value (default none)")

	flag.BoolVar(&cmdFlags.HighQuality, "hq", false, "Set high quality (fn = 200)")

	flag.BoolVar(&cmdFlags.LowQuality, "lq", false, "Set low quality (fn = 20)")

	flag.BoolVar(&cmdFlags.SetBuildInfoInFileAttributes, "fi", true, "Set build info in file attributes (default true)")

	flag.Parse()

	err := pkg.Process(cmdFlags)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Use the openscadgen package to generate STL files
	// ...

}
