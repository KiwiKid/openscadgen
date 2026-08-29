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
	flag.BoolVar(&cmdFlags.ShowConfigOptions, "config-options", false, "Display config.toml options")
	flag.BoolVar(&cmdFlags.ShowConfigOptions, "co", false, "Alias for -config-options")
	flag.StringVar(&cmdFlags.ConfigOptionsTopic, "topic", "", "Filter config options by topic")

	initDesc := "Create a new project folder (name or relative path, e.g. examples/my-part). Optional location: -init-dir, or -sf when starting the server in the same command."
	flag.StringVar(&cmdFlags.InitProjectName, "init", "", initDesc)
	flag.StringVar(&cmdFlags.InitProjectName, "i", "", "Alias for -init")
	flag.StringVar(&cmdFlags.InitProjectName, "new", "", "Alias for -init (create new project)")

	initExtDesc := "Create a new project with BOSL2 / renderSlicing template (same as -ie). Location flags same as -init."
	flag.StringVar(&cmdFlags.InitProjectNameExtended, "init-extended", "", initExtDesc)
	flag.StringVar(&cmdFlags.InitProjectNameExtended, "ie", "", "Alias for -init-extended")
	flag.StringVar(&cmdFlags.InitProjectNameExtended, "new-extended", "", "Alias for -init-extended")
	flag.StringVar(&cmdFlags.InitProjectNameExtended, "newe", "", "Alias for -new-extended")
	flag.StringVar(&cmdFlags.InitProjectTemplate, "t", "", "Template file name under openscad-init-templates/ to use with -ie")

	initExplainerDesc := "Create a new project with the explainer starter template (same as -ix). Location flags same as -init."
	flag.StringVar(&cmdFlags.InitProjectNameExplainer, "init-explainer", "", initExplainerDesc)
	flag.StringVar(&cmdFlags.InitProjectNameExplainer, "ix", "", "Alias for -init-explainer")

	flag.BoolVar(&cmdFlags.NoInput, "no-input", false, "Bypass interactive prompts and assume the default action")

	initParentDesc := "Parent directory for a new project from -init/-i/-new or -ie/-newe (default: . or -sf folder when set)"
	flag.StringVar(&cmdFlags.InitProjectParentDir, "init-dir", "", initParentDesc)
	flag.StringVar(&cmdFlags.InitProjectParentDir, "id", "", "Alias for -init-dir")

	flag.StringVar(&cmdFlags.RegexPattern, "regex", "", "Regex pattern to only run a specific instances when generating files")
	flag.StringVar(&cmdFlags.RegexPattern, "r", "", "Alias for -regex")

	flag.BoolVar(&cmdFlags.Quiet, "quiet", false, "quiet mode, no log output")
	flag.BoolVar(&cmdFlags.Quiet, "q", false, "Alias for -quiet")

	flag.BoolVar(&cmdFlags.NoProcessing, "no-processing", false, "'dry-run' mode - will check config and provide instances that will be processed, but not do any processing")
	flag.BoolVar(&cmdFlags.NoProcessing, "np", false, "Alias for -no-processing")

	flag.BoolVar(&cmdFlags.Debug, "debug", false, "debug mode, more output")
	flag.BoolVar(&cmdFlags.Debug, "D", false, "Alias for -debug (note: -d is used for delete-export-stls-dir)")

	deleteDirDesc := "Directory to scan for export/ folders and list .stl files for deletion (confirm required)"
	flag.StringVar(&cmdFlags.DeleteExportSTLsDir, "delete-export-stls-dir", "", deleteDirDesc)
	flag.StringVar(&cmdFlags.DeleteExportSTLsDir, "d", "", "Alias for -delete-export-stls-dir")
	flag.BoolVar(&cmdFlags.BuildPages, "build-pages", false, "Build a GitHub Pages-ready copy of examples/ and bols2otropolis/ into the pages/ folder")
	flag.StringVar(&cmdFlags.PagesOutputDir, "pages-output-dir", "pages", "Destination directory for -build-pages output")
	flag.BoolVar(&cmdFlags.BuildSite, "build-site", false, "Build all example and bols2otropolis reports in GitHub Pages mode, then assemble the pages/ folder")
	flag.StringVar(&cmdFlags.BuildSiteIgnoreSections, "build-site-ignore-sections", "", "Comma-separated path sections to skip when building the site")

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
	flag.BoolVar(&cmdFlags.DangerouslySkipPermissions, "dangerously-skip-permissions", false, "Allow unsafe OpenSCAD execution settings such as custom command paths and extra command arguments")

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

	flag.BoolVar(&cmdFlags.EnableFileWatcher, "enable-file-watcher", false, "[NOT YET IMPLEMENTED] sEnable file watchers to automatically regenerate on changes (default false)")
	flag.BoolVar(&cmdFlags.EnableFileWatcher, "efw", false, "Alias for -enable-file-watcher")
	flag.BoolVar(&cmdFlags.ShowAI, "ai", false, "Show AI tools in server mode")

	flag.Parse()

	return cmdFlags
}
