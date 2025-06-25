package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"

	"github.com/kiwikid/openscadgen/pkg"
)

// https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file
// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. If that fail, copy the file contents from src to dst.

var (
	progressMap = make(map[string]chan string)
	cancelMap   = make(map[string]chan struct{})
	mu          sync.Mutex
)

func main() {

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

	flag.BoolVar(&cmdFlags.ContinueOnError, "coe", false, "Continue if an error occurs when loading or generating files - not recommended as the checks can be handy (experimental)")
	flag.BoolVar(&cmdFlags.ContinueOnError, "continue-on-error", false, "Alias for -co")

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

	flag.StringVar(&cmdFlags.ServerFolder, "sf", "", "Start in server mode, optionally specify a folder to scan for config files")

	flag.Parse()

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
		StartServer(cmdFlags.ServerFolder)
		return
	}

	config, err := pkg.LoadConfig(cmdFlags)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	processResult, err := pkg.Process(config, &pkg.NoopProgress{}, nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(processResult.STLResults) == 0 && len(processResult.ImageResults) == 0 {
		log.Printf("\033[31m" + "No STLs or images generated" + "\033[0m")
		os.Exit(1)
		return
	}

	_, location, genReportErr := pkg.GenerateOutputReport(config, processResult.Instances, processResult.STLResults, processResult.ImageResults, processResult.ExportLocation, true)
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

var configFiles []models.ConfigFile

func StartServer(serverFolder string) {

	var err error
	var msg = "Starting server on port 8080"
	if serverFolder != "" {
		configFiles, err = pkg.ScanFolderForConfigFiles(serverFolder)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		log.Printf("Found %d config files in %s", len(configFiles), serverFolder)
		msg += fmt.Sprintf(" and %d config files in %s", len(configFiles), serverFolder)

	}
	log.Print(msg)

	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/progress", progressHandler)
	http.HandleFunc("/cancel", cancelHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "GET":
			log.Printf("GET (Display) request" + r.Method)
			configEntryPathEncoded := r.URL.Query().Get("config")
			configEntryPath, err := url.QueryUnescape(configEntryPathEncoded)
			if err != nil {
				log.Printf("Error: %v", err)
			}

			if configEntryPath == "" {
				if len(configFiles) > 0 {
					log.Printf("Listing %d config files", len(configFiles))
					listConfig := templates.ListConfig(configFiles)
					listConfig.Render(context.Background(), w)
				}
				configEntry := templates.EnterConfigPage()
				configEntry.Render(context.Background(), w)
			} else {
				config, err := pkg.LoadConfig(models.CmdFlags{ConfigFile: configEntryPath})
				if err != nil {
					warning := templates.Warning(fmt.Sprintf("Error: %v", err))
					warning.Render(context.Background(), w)
				}

				report := templates.Report(config, []models.InstanceConfig{}, "", []models.GenerateSTLResult{}, []models.GenerateImageResult{}, []string{}, true, configEntryPath)
				report.Render(context.Background(), w)
			}
		case "POST":
			log.Printf("POST (Add Config) request" + r.Method)

			configEntry := r.FormValue("path")
			if configEntry == "" {
				warning := templates.Warning("No config file provided")
				warning.Render(context.Background(), w)
			}

			encodedConfigEntry := url.QueryEscape(configEntry)

			http.Redirect(w, r, "/?config="+encodedConfigEntry, http.StatusSeeOther)
		case "PUT":
			log.Printf("PUT (Processing) request" + r.Method)

			var cmdFlags models.CmdFlags
			if r.Header.Get("Content-Type") == "application/json" {
				if err := json.NewDecoder(r.Body).Decode(&cmdFlags); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("PUT - Processing - Invalid JSON body: " + err.Error()))
					return
				}
			} else {
				cmdFlags.ConfigFile = r.FormValue("path")
			}

			if cmdFlags.ConfigFile == "" {
				warning := templates.Warning("No config file provided")
				warning.Render(context.Background(), w)
				return
			}

			config, err := pkg.LoadConfig(cmdFlags)
			if err != nil {
				warning := templates.Warning(fmt.Sprintf("Error: %v", err))
				warning.Render(context.Background(), w)
				return
			}

			id := uuid.New().String()
			updates := make(chan string, 10)
			cancel := make(chan struct{})
			mu.Lock()
			progressMap[id] = updates
			cancelMap[id] = cancel
			mu.Unlock()

			go func() {
				pkg.Process(config, &pkg.ChanProgress{Updates: updates}, cancel)
				close(updates)
			}()

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
			<div id="progress"></div>
			<button onclick="fetch('/cancel?id=%s')">Cancel</button>
			<script>
			function poll() {
				fetch('/progress?id=%s').then(r => r.text()).then(msg => {
					document.getElementById('progress').innerText = msg;
					if(msg !== 'done' && msg !== 'cancelled') setTimeout(poll, 1000);
				});
			}
			poll();
			</script>
			`, id, id)
			return

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// MCP endpoints
	http.HandleFunc("/v1/metadata", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":         "openscadgen",
			"version":      pkg.GetVersion().OpenSCADGen,
			"capabilities": []string{"resources", "tools"},
		})
	})

	http.HandleFunc("/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "cube-config", "type": "openscadgen-config", "description": "Cube config for openscadgen"},
		})
	})

	http.HandleFunc("/v1/resource/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := r.URL.Path[len("/v1/resource/"):]
		configPath := r.URL.Query().Get("config")
		if configPath == "" {
			configPath = "bols2otropolis/cube/config.toml"
		}
		config, err := pkg.LoadConfig(models.CmdFlags{ConfigFile: configPath, Server: true})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var instanceSummaries []map[string]interface{}
		for _, inst := range config.Design.ConfiguredInstanceConfig {
			instanceSummaries = append(instanceSummaries, map[string]interface{}{
				"name":        inst.Name,
				"params":      inst.Params,
				"description": inst.Description,
			})
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          id,
			"config_path": config.ConfigFile,
			"instances":   instanceSummaries,
			"raw_config":  config.RawConfigFile,
		})
	})

	http.HandleFunc("/v1/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "list_instances", "description": "List all configured instances"},
			{"id": "process_instance", "description": "Process a single instance (POST: {\"name\":...})"},
			{"id": "process_all", "description": "Process all instances"},
		})
	})

	http.HandleFunc("/v1/tool/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !((r.Method == "POST") && (len(r.URL.Path) > len("/v1/tool/") && r.URL.Path[len(r.URL.Path)-len("/invoke"):] == "/invoke")) {
			http.NotFound(w, r)
			return
		}
		id := r.URL.Path[len("/v1/tool/") : len(r.URL.Path)-len("/invoke")]
		var req struct {
			Config string `json:"config"`
			Name   string `json:"name"`
		}
		body, _ := ioutil.ReadAll(r.Body)
		_ = json.Unmarshal(body, &req)
		configPath := req.Config
		if configPath == "" {
			configPath = "bols2otropolis/cube/config.toml"
		}
		config, err := pkg.LoadConfig(models.CmdFlags{ConfigFile: configPath, Server: true})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		switch id {
		case "list_instances":
			var instanceNames []string
			for _, inst := range config.Design.ConfiguredInstanceConfig {
				instanceNames = append(instanceNames, inst.Name)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"instances": instanceNames,
			})
		case "process_instance":
			if req.Name == "" {
				http.Error(w, "Missing or invalid instance name", 400)
				return
			}
			config.RegexPattern = req.Name
			result, err := pkg.Process(config, &pkg.NoopProgress{}, nil)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "Processed instance",
				"output": result,
			})
		case "process_all":
			result, err := pkg.Process(config, &pkg.NoopProgress{}, nil)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"result": "Processed all instances",
				"output": result,
			})
		default:
			http.NotFound(w, r)
		}
	})

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	updates := make(chan string, 10)
	cancel := make(chan struct{})
	mu.Lock()
	progressMap[id] = updates
	cancelMap[id] = cancel
	mu.Unlock()

	var flags models.CmdFlags
	err := json.NewDecoder(r.Body).Decode(&flags)
	if err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	config, err := pkg.LoadConfig(flags)
	if err != nil {
		http.Error(w, "Error loading config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		pkg.Process(config, &pkg.ChanProgress{Updates: updates}, cancel)
		close(updates)
	}()

	// Return the ID to the client, which will poll /progress?id=...
	w.Write([]byte(id))
}

func progressHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	mu.Lock()
	updates, ok := progressMap[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	select {
	case msg, ok := <-updates:
		if !ok {
			w.Write([]byte("done"))
			return
		}
		w.Write([]byte(msg))
	case <-time.After(2 * time.Second):
		w.Write([]byte("waiting"))
	}
}

func cancelHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	mu.Lock()
	cancel, ok := cancelMap[id]
	if ok {
		close(cancel)
		delete(cancelMap, id)
		delete(progressMap, id)
	}
	mu.Unlock()
	w.Write([]byte("cancelled"))
}
