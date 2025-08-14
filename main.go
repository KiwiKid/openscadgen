package main

import (
	"fmt"
	"log"
	"os"

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
		pkg.InitConfig(cmdFlags.InitProjectName, false)
		return
	}

	if cmdFlags.InitProjectNameExtended != "" {
		log.Printf("Initializing extended project: %s", cmdFlags.InitProjectNameExtended)
		pkg.InitConfig(cmdFlags.InitProjectNameExtended, true)
		return
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

	if cmdFlags.Server || cmdFlags.ServerFolder != "" {
		cmdFlags.Server = true
		server.StartServer(cmdFlags.ServerFolder, cmdFlags)
		return
	}

	config, err := pkg.LoadConfig(cmdFlags)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Use terminal progress reporter instead of NoopProgress
	var progress pkg.ProgressReporter
	if config.Quiet {
		progress = &pkg.NoopProgress{}
	} else {
		progress = pkg.NewTerminalProgressReporter(config)
	}

	processResult, err := pkg.Process(config, progress, nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
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

	_, location, genReportErr := pkg.GenerateOutputReport(config, processResult.Instances, processResult.STLResults, processResult.ImageResults, processResult.ExportLocation, true, processResult.TotalTimeTaken)
	if genReportErr != nil {
		if config.ContinueOnError {
			log.Printf("Warning: failed to generate output report: %v", err)
		} else {
			log.Fatalf("failed to generate output report: %v", err)
		}
	} else if config.Debug {
		pkg.LogKeyValuePair("Output report generated at", location)
	}
}
