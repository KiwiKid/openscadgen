package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/browser"
	"github.com/kiwikid/openscadgen/cmd"
	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/server"
)

func main() {
	cmdFlags := cmd.ParseFlags()

	// If -c / --config was not provided, fall back to CONFIG_FILE env var
	if cmdFlags.ConfigFile == "" {
		if envCfg := os.Getenv("CONFIG_FILE"); envCfg != "" {
			cmdFlags.ConfigFile = envCfg
		}
	}

	// Initialize logger before loading config
	if err := pkg.InitLogger("memory"); err != nil {
		pkg.LogErrorf("Failed to initialize logger: %v", err)
		os.Exit(1)
	}

	version := pkg.GetVersion()
	if cmdFlags.Debug || cmdFlags.Version {
		pkg.LogInfof("OpenSCADGen Version: %s", version.OpenSCADGen)
		pkg.LogInfof("OpenSCAD Version: %s", version.OpenSCAD)
	}
	if cmdFlags.Version {
		return
	}

	if cmdFlags.InitProjectName != "" || cmdFlags.InitProjectNameExtended != "" || cmdFlags.InitProjectNameExplainer != "" {
		parent := pkg.ResolveInitParentDir(cmdFlags)
		template := "basic"
		name := strings.TrimSpace(cmdFlags.InitProjectName)
		stageLabel := "project"

		switch {
		case cmdFlags.InitProjectNameExplainer != "":
			name = strings.TrimSpace(cmdFlags.InitProjectNameExplainer)
			template = "explainer"
			stageLabel = "explainer project"
		case cmdFlags.InitProjectNameExtended != "":
			name = strings.TrimSpace(cmdFlags.InitProjectNameExtended)
			template = "extended"
			stageLabel = "extended project"
		default:
			if chosen, err := pkg.ChooseInitTemplate(cmdFlags, parent); err != nil {
				pkg.LogErrorf("init project: %v", err)
				os.Exit(1)
			} else if chosen {
				template = "explainer"
				stageLabel = "explainer project"
			}
		}

		pkg.LogStagef("init", "Initializing %s: %s", stageLabel, name)
		if filepath.IsAbs(name) {
			// Preserve prior behaviour: absolute init names are treated as a leaf under -init-dir / -sf.
			name = filepath.Base(name)
		}
		if err := pkg.InitConfigInParentWithTemplate(parent, name, template); err != nil {
			pkg.LogErrorf("init project: %v", err)
			os.Exit(1)
		}
		if !cmdFlags.Server && cmdFlags.ServerFolder == "" {
			return
		}
	}

	if cmdFlags.Debug {
		pkg.LogKeys(cmdFlags)
	}

	if !cmdFlags.Quiet {
		pkg.LogInit()
	}

	if cmdFlags.ShowMan {
		pkg.ShowMan()
		return
	}
	if cmdFlags.ShowConfigOptions {
		fmt.Print(pkg.RenderConfigOptionsCLI(cmdFlags.ConfigOptionsTopic))
		return
	}

	if cmdFlags.DeleteExportSTLsDir != "" {
		files, err := pkg.FindExportSTLFiles(cmdFlags.DeleteExportSTLsDir)
		if err != nil {
			pkg.LogErrorf("FindExportSTLFiles Error: %v", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			pkg.LogInfof("No export/*.stl files found under: %s", cmdFlags.DeleteExportSTLsDir)
			return
		}

		fmt.Printf("Found %d .stl files under export/ folders in: %s\n\n", len(files), cmdFlags.DeleteExportSTLsDir)
		for _, f := range files {
			fmt.Printf("- %s\n", f)
		}
		fmt.Printf("\nDelete these files? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if strings.ToLower(line) != "y" {
			pkg.LogWarnf("Aborted (no delete performed).")
			return
		}

		res := pkg.DeleteFiles(files)
		if len(res.Failed) > 0 {
			pkg.LogWarnf("Deleted %d files; %d failed:", len(res.Deleted), len(res.Failed))
			for p, e := range res.Failed {
				pkg.LogWarnf("  - %s: %v", p, e)
			}
			os.Exit(1)
			return
		}
		pkg.LogInfof("Deleted %d files.", len(res.Deleted))
		return
	}

	if cmdFlags.Server || cmdFlags.ServerFolder != "" {
		cmdFlags.Server = true
		server.StartServer(cmdFlags.ServerFolder, cmdFlags, nil)
		return
	}

	if cmdFlags.ProcessFolder != "" {
		cmdFlags.OverwriteExisting = true
		processResults, err := pkg.ProcessFolder(cmdFlags.ProcessFolder, cmdFlags)
		if err != nil {
			pkg.LogErrorf("ProcessFolder Error: %v", err)
			os.Exit(1)
		}
		pkg.LogInfof("Processed %d config files", len(processResults))
		return
	}

	if cmdFlags.ConfigFile != "" {

		config, _, err := pkg.LoadConfigFromFile(cmdFlags)
		if err != nil {
			pkg.LogErrorf("LoadConfig Error: %v", err)
			os.Exit(1)
		}

		// Use terminal progress reporter instead of NoopProgress
		var progress pkg.ProgressReporter
		if config.Quiet {
			progress = &pkg.NoopProgress{}
		} else {
			progress = pkg.NewTerminalProgressReporter(config)
		}

		processResult, err := pkg.Process(config, progress, nil, pkg.Operations{
			GenerateReport: true,
		}, false)
		if err != nil {
			pkg.LogErrorf("Process Error: %v", err)
			os.Exit(1)
		}

		//if !config.Quiet {
		pkg.LogInfof("Total processing time: %v", processResult.TotalTimeTaken)
		//}

		if len(processResult.STLResults) == 0 && len(processResult.ImageResults) == 0 {
			pkg.LogInfof("Match options:")
			for _, instance := range processResult.Instances {
				instanceStr := fmt.Sprintf("  - %s", instance.Name)
				/*for _, param := range instance.Params {
					instanceStr += fmt.Sprintf(" %v", param)
				}*/
				pkg.LogInfof("%s", instanceStr)
			}
			pkg.LogWarnWithCritical("No STLs or images generated", true)
			pkg.LogWarnWithCritical(fmt.Sprintf("Regex pattern didn't match any instances: %s\n\n(match options listed above)", config.RegexPattern), false)

			os.Exit(1)
			return
		}
	} else {
		onStart := func(port string) error {
			// launch browser
			browser.OpenURL(fmt.Sprintf("http://localhost%s", port))
			return nil
		}
		server.StartServer("", cmdFlags, onStart)
	}

}
