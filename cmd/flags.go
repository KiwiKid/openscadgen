package cmd

import (
	"flag"

	"github.com/kiwikid/openscadgen/pkg/models"
)

// ParseFlags parses command-line flags and returns a CmdFlags struct
func ParseFlags() models.CmdFlags {
	cmdFlags := models.CmdFlags{}

	// Parse command-line flags into the struct
	configDesc := "Path to config file \n the absolute or relative path to the .toml config file"
	flag.StringVar(&cmdFlags.ConfigFile, "config", "", configDesc)
	flag.StringVar(&cmdFlags.ConfigFile, "c", "", configDesc)

	flag.BoolVar(&cmdFlags.ShowMan, "man", false, "Display help message")
	flag.BoolVar(&cmdFlags.ShowMan, "m", false, "Alias for -man")
	flag.BoolVar(&cmdFlags.ShowMan, "h", false, "Alias for -man")

	flag.StringVar(&cmdFlags.InitProjectName, "init", "", "Initialize a new project at the current directory with the given name")
	flag.StringVar(&cmdFlags.InitProjectName, "i", "", "Alias for -init")

	flag.StringVar(&cmdFlags.InitProjectNameExtended, "init-extended", "", "Initialize a new project at the current directory with the given name - with bosl2 and renderSlicing support")
	flag.StringVar(&cmdFlags.InitProjectNameExtended, "ie", "", "Alias for -init")

	flag.StringVar(&cmdFlags.RegexPattern, "regex", "", "Regex pattern to only run a specific instances when generating files")
	flag.StringVar(&cmdFlags.RegexPattern, "r", "", "Alias for -regex")

	flag.BoolVar(&cmdFlags.Quiet, "quiet", false, "quiet mode, no log output")
	flag.BoolVar(&cmdFlags.Quiet, "q", false, "Alias for -quiet")

	flag.BoolVar(&cmdFlags.NoProcessing, "no-processing", false, "'dry-run' mode - will check config and provide instances that will be processed, but not do any processing")
	flag.BoolVar(&cmdFlags.NoProcessing, "np", false, "Alias for -no-processing")

	flag.BoolVar(&cmdFlags.Debug, "debug", false, "debug mode, more output")
	flag.BoolVar(&cmdFlags.Debug, "d", false, "Alias for -debug")

	flag.BoolVar(&cmdFlags.Version, "version", false, "just output the openscadgen and openscad version number")
	flag.BoolVar(&cmdFlags.Version, "v", false, "Alias for -version")

	flag.BoolVar(&cmdFlags.SkipRender, "skip-render", false, "Dont run a render before export")
	flag.BoolVar(&cmdFlags.SkipRender, "sr", false, "Alias for -skip-render")

	flag.BoolVar(&cmdFlags.SkipReadme, "skip-docs", false, "Skip generating a README.md file")
	flag.BoolVar(&cmdFlags.SkipReadme, "sd", false, "Alias for -skip-readme")

	flag.IntVar(&cmdFlags.MaxInstances, "n", 0, "Maximum number of instances to process")

	flag.BoolVar(&cmdFlags.StopOnError, "soe", false, "Stop if an error occurs when loading or generating files")
	flag.BoolVar(&cmdFlags.StopOnError, "stop-on-error", false, "Alias for -soe")

	flag.BoolVar(&cmdFlags.IncludeExportLog, "include-export-log-file", false, "Include the export log in the README.md file")
	flag.BoolVar(&cmdFlags.IncludeExportLog, "el", false, "Alias for -include-export-log-file")

	flag.BoolVar(&cmdFlags.OverwriteExisting, "ow", false, "Overrwite existing files")
	flag.BoolVar(&cmdFlags.OverwriteExisting, "overwrite", false, "Alias for -ow")

	flag.BoolVar(&cmdFlags.IncludePartIDLetter, "pid", false, "Include optional_part_id_letter variable in the call the openscad")

	flag.StringVar(&cmdFlags.CustomOpenSCADCommand, "custom-openscad-command", "", "Custom OpenSCAD command to use")

	flag.IntVar(&cmdFlags.OverrideFN, "fn", 0, "Override the default fn value (default none)")

	flag.BoolVar(&cmdFlags.HighQuality, "hq", false, "Set high quality (fn = 200)")

	flag.BoolVar(&cmdFlags.LowQuality, "lq", false, "Set low quality (fn = 20)")

	flag.BoolVar(&cmdFlags.OnlyImages, "oi", false, "Only generate images (default is images and stl)")
	flag.BoolVar(&cmdFlags.OnlyExport, "oe", false, "Only export STLs (default is images and stl)")

	flag.BoolVar(&cmdFlags.SetBuildInfoInFileAttributes, "fi", true, "Set build info in file attributes (default true)")

	flag.BoolVar(&cmdFlags.Server, "s", false, "Start in server mode")

	flag.StringVar(&cmdFlags.ServerFolder, "sf", "", "Start in server mode, optionally specify a folder to scan for config files (use -p to set the port)")
	flag.IntVar(&cmdFlags.ServerPort, "p", 7203, "Set the port for the server")

	flag.StringVar(&cmdFlags.ProcessFolder, "cf", "", "Process a folder of config files. Will search for config.toml files and process them all")

	flag.BoolVar(&cmdFlags.EnableFileWatcher, "enable-file-watcher", false, "Enable file watchers to automatically regenerate on changes (default false)")
	flag.BoolVar(&cmdFlags.EnableFileWatcher, "efw", false, "Alias for -enable-file-watcher")

	flag.Parse()

	return cmdFlags
}
