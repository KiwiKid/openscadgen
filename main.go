package main

import (
	"bufio"
	"fmt"
	"log"
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

	// Initialize logger before loading config
	if err := pkg.InitLogger("memory"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	version := pkg.GetVersion()
	if cmdFlags.Debug || cmdFlags.Version {
		log.Printf("OpenSCADGen Version: %s", version.OpenSCADGen)
		log.Printf("OpenSCAD Version: %s", version.OpenSCAD)
	}
	if cmdFlags.Version {
		return
	}

	if cmdFlags.InitProjectName != "" {
		log.Printf("Initializing project: %s", cmdFlags.InitProjectName)
		name := strings.TrimSpace(cmdFlags.InitProjectName)
		parent := pkg.ResolveInitParentDir(cmdFlags)
		var err error
		if filepath.IsAbs(name) {
			// Preserve prior behaviour: absolute -i is treated as a leaf under -init-dir / -sf.
			err = pkg.InitConfigInParent(parent, filepath.Base(name), false)
		} else {
			err = pkg.InitConfigInParent(parent, name, false)
		}
		if err != nil {
			log.Fatalf("init project: %v", err)
		}
		if !cmdFlags.Server && cmdFlags.ServerFolder == "" {
			return
		}
	} else if cmdFlags.InitProjectNameExtended != "" {
		log.Printf("Initializing extended project: %s", cmdFlags.InitProjectNameExtended)
		name := strings.TrimSpace(cmdFlags.InitProjectNameExtended)
		parent := pkg.ResolveInitParentDir(cmdFlags)
		var err error
		if filepath.IsAbs(name) {
			err = pkg.InitConfigInParent(parent, filepath.Base(name), true)
		} else {
			err = pkg.InitConfigInParent(parent, name, true)
		}
		if err != nil {
			log.Fatalf("init project: %v", err)
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

	if cmdFlags.DeleteExportSTLsDir != "" {
		files, err := pkg.FindExportSTLFiles(cmdFlags.DeleteExportSTLsDir)
		if err != nil {
			log.Fatalf("FindExportSTLFiles Error: %v", err)
		}
		if len(files) == 0 {
			log.Printf("No export/*.stl files found under: %s", cmdFlags.DeleteExportSTLsDir)
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
			log.Printf("Aborted (no delete performed).")
			return
		}

		res := pkg.DeleteFiles(files)
		if len(res.Failed) > 0 {
			log.Printf("Deleted %d files; %d failed:", len(res.Deleted), len(res.Failed))
			for p, e := range res.Failed {
				log.Printf("  - %s: %v", p, e)
			}
			os.Exit(1)
			return
		}
		log.Printf("Deleted %d files.", len(res.Deleted))
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
			log.Fatalf("ProcessFolder Error: %v", err)
		}
		log.Printf("Processed %d config files", len(processResults))
		return
	}

	if cmdFlags.ConfigFile != "" {

		config, _, err := pkg.LoadConfigFromFile(cmdFlags)
		if err != nil {
			log.Fatalf("LoadConfig Error: %v", err)
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
			log.Fatalf("Process Error: %v", err)
		}

		//if !config.Quiet {
		log.Printf("Total processing time: %v", processResult.TotalTimeTaken)
		//}

		if len(processResult.STLResults) == 0 && len(processResult.ImageResults) == 0 {
			log.Printf("Match options:")
			for _, instance := range processResult.Instances {
				instanceStr := fmt.Sprintf("  - %s", instance.Name)
				/*for _, param := range instance.Params {
					instanceStr += fmt.Sprintf(" %v", param)
				}*/
				log.Println(instanceStr)
			}
			pkg.LogWarn("No STLs or images generated", true)
			pkg.LogWarn(fmt.Sprintf("Regex pattern didn't match any instances: %s\n\n(match options listed above)", config.RegexPattern), false)

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
