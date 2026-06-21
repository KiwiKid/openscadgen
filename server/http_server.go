package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/a-h/templ"
	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

var configFiles []models.ConfigFile
var fileWatcher *FileWatcher
var fileWatcherEnabled bool
var globalServerFolder string

type ServerInfo struct {
	Port    string
	Address string
}

// resolveConfigPath resolves a config path using the server folder structure
// If the path is absolute, it's returned as-is
// If the path is relative and serverFolder is set, they are joined
// Otherwise, the path is resolved relative to the current working directory
func resolveConfigPath(configPath string, serverFolder string) string {
	if filepath.IsAbs(configPath) {
		return configPath
	}
	if serverFolder != "" {
		return filepath.Join(serverFolder, configPath)
	}
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return configPath // fallback to original if abs fails
	}
	return absPath
}

// tryListenOnPort attempts to listen on the specified port.
// If portWasSpecified is false and the port is in use, it tries random ports.
// Returns the listener, the actual port being used, and any error.
func tryListenOnPort(port string, portWasSpecified bool) (net.Listener, string, error) {
	// Try to listen on the requested port
	listener, err := net.Listen("tcp", port)
	if err == nil {
		// Success! Get the actual address
		addr := listener.Addr().(*net.TCPAddr)
		actualPort := fmt.Sprintf(":%d", addr.Port)
		return listener, actualPort, nil
	}

	// Port is in use
	if portWasSpecified {
		// Port was explicitly specified, so fail
		return nil, "", fmt.Errorf("port %s is already in use", port)
	}

	// Port wasn't specified, try random port
	log.Printf("Port %s is in use, trying random port...", port)
	// Use port 0, which lets the OS assign a random available port
	listener, err = net.Listen("tcp", ":0")
	if err != nil {
		return nil, "", fmt.Errorf("could not find an available port: %v", err)
	}
	addr := listener.Addr().(*net.TCPAddr)
	actualPort := fmt.Sprintf(":%d", addr.Port)
	return listener, actualPort, nil
}

// StartServer starts the HTTP server with the specified folder for config files
func StartServer(serverFolder string, cmdFlags models.CmdFlags, onStart func(port string) error) ServerInfo {
	port := ":8080"
	if cmdFlags.ServerPort != 0 {
		port = fmt.Sprintf(":%d", cmdFlags.ServerPort)
	}

	var err error
	if cmdFlags.Debug {
		log.Printf("[Debug]")
	}
	var msg = fmt.Sprintf("Starting server on port %s", port)

	// Initialize file watcher only if enabled
	fileWatcherEnabled = cmdFlags.EnableFileWatcher
	if cmdFlags.EnableFileWatcher {
		fileWatcher, err = NewFileWatcher()
		if err != nil {
			log.Fatalf("Error creating file watcher: %v", err)
		}
	}

	if serverFolder != "" {
		globalServerFolder = serverFolder
		configFiles, err = pkg.ScanFolderForConfigFiles(serverFolder)
		if err != nil {
			log.Fatalf("ScanFolderForConfigFiles Error: %v", err)
		}
		log.Printf("Found %d config files in %s", len(configFiles), serverFolder)
		msg += fmt.Sprintf(" and %d config files in %s", len(configFiles), serverFolder)

		// Start file watching only if enabled
		if cmdFlags.EnableFileWatcher && fileWatcher != nil {
			if err := fileWatcher.StartWatching(serverFolder); err != nil {
				log.Printf("Warning: Could not start file watching: %v", err)
			} else {
				msg += " with file watching enabled"
			}
		}
	}

	// Setup handlers
	http.HandleFunc("/", handleMainRequest)

	http.HandleFunc("/start", StartHandler)
	http.HandleFunc("/progress", ProgressHandler)
	http.HandleFunc("/cancel", CancelHandler)
	http.HandleFunc("/chat", handleChatRequest)
	http.HandleFunc("/api/config", handleConfigRequest)
	http.HandleFunc("/api/open", handleOpenFile)
	http.HandleFunc("/api/edit", handleEditFile)

	http.HandleFunc("/static/", handleStaticFiles)
	http.HandleFunc("/images", handleImageRequest)

	http.HandleFunc("/api/openscad/status", handleOpenSCADStatus)
	http.HandleFunc("/api/watcher/status", handleWatcherStatus)
	http.HandleFunc("/api/watcher/pause", handleWatcherPause)
	http.HandleFunc("/api/watcher/resume", handleWatcherResume)
	http.HandleFunc("/api/watcher/ui", handleWatcherUI)
	http.HandleFunc("/api/preview", handlePreviewRequest)
	http.HandleFunc("/api/stl", handleSTLRequest)
	http.HandleFunc("/delete-export-stls", handleDeleteExportSTLs)

	// Check if port is available, and if not specified, try random port
	portWasSpecified := cmdFlags.ServerPort != 0
	originalPort := port

	listener, actualPort, err := tryListenOnPort(port, portWasSpecified)
	if err != nil {
		log.Printf("Is Openscadgen already running? %v", err)
		log.Fatalf("Error: Could not bind to port %s: %v\n End the existing process or use -p to specify a different port", port, err)
	}

	// Update port if we got a different one
	if actualPort != port {
		log.Printf("Port %s was in use, using %s instead", originalPort, actualPort)
		port = actualPort
		// Update message with actual port
		msg = fmt.Sprintf("Starting server on port %s", port)
		if serverFolder != "" {
			msg += fmt.Sprintf(" and %d config files in %s", len(configFiles), serverFolder)
			if cmdFlags.EnableFileWatcher && fileWatcher != nil && fileWatcher.IsWatching() {
				msg += " with file watching enabled"
			}
		}
	}

	msg += fmt.Sprintf("\n\n Running. Will soon navigate to http://localhost%s", port)
	log.Print(msg)

	// Call onStart callback before starting server
	if onStart != nil {
		log.Print("onStart starting")
		err = onStart(port)
		if err != nil {
			listener.Close()
			log.Fatalf("Error on start: %v", err)
		}
	}

	log.Printf("Server started on port %s", port)

	// Start server on the listener (non-blocking for now, but we'll block here)
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalf("Error on serve: %v", err)
	}

	return ServerInfo{Port: port, Address: fmt.Sprintf("http://localhost%s", port)}
}

func handleDeleteExportSTLs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		rootDir := r.URL.Query().Get("dir")
		if rootDir == "" {
			if globalServerFolder != "" {
				rootDir = globalServerFolder
			} else {
				rootDir = "."
			}
		}
		files, err := pkg.FindExportSTLFiles(rootDir)
		data := templates.DeleteExportSTLsPageData{
			RootDir:  rootDir,
			STLFiles: files,
		}
		if err != nil {
			data.Error = err.Error()
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.DeleteExportSTLsPage(data).Render(r.Context(), w)
		return
	case "POST":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		rootDir := r.FormValue("dir")
		if rootDir == "" {
			if globalServerFolder != "" {
				rootDir = globalServerFolder
			} else {
				rootDir = "."
			}
		}
		performDelete := r.FormValue("performDelete") == "true"

		files, err := pkg.FindExportSTLFiles(rootDir)
		data := templates.DeleteExportSTLsPageData{
			RootDir:       rootDir,
			STLFiles:      files,
			PerformDelete: performDelete,
		}
		if err != nil {
			data.Error = err.Error()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			templates.DeleteExportSTLsPage(data).Render(r.Context(), w)
			return
		}
		if performDelete && len(files) > 0 {
			delRes := pkg.DeleteFiles(files)
			data.Deleted = delRes.Deleted
			if len(delRes.Failed) > 0 {
				data.Failed = map[string]string{}
				for p, e := range delRes.Failed {
					data.Failed[p] = e.Error()
				}
			}
			// Refresh list after delete attempt
			remaining, _ := pkg.FindExportSTLFiles(rootDir)
			data.STLFiles = remaining
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.DeleteExportSTLsPage(data).Render(r.Context(), w)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// handleMainRequest handles the main HTTP requests (GET, POST, PUT)
func handleMainRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGETRequest(w, r)
	case "POST":
		handlePOSTRequest(w, r)
	case "PUT":
		handlePUTRequest(w, r)
	case "DELETE":
		handleDeleteRequest(w, r)
	default:
		http.Error(w, "handleMainRequest Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGETRequest handles GET requests for displaying configs
func handleGETRequest(w http.ResponseWriter, r *http.Request) {
	log.Print("GET (Display) request" + r.Method)
	configEntryPathEncoded := r.URL.Query().Get("config")
	configEntryPath, err := url.QueryUnescape(configEntryPathEncoded)
	if err != nil {
		log.Printf("QueryUnescape Error: %v", err)
	}

	serverFolderEncoded := r.URL.Query().Get("server_folder")
	if serverFolderEncoded != "" {
		log.Printf("(url) server_folder query: %s", serverFolderEncoded)
		serverFolderBytes, err := base64.StdEncoding.DecodeString(serverFolderEncoded)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("Error decoding server folder: %v", err))
			projectFolderForm := templates.ProjectFolderForm(warning)
			projectFolderForm.Render(context.Background(), w)
			return
		}
		decoded := string(serverFolderBytes)
		scanned, err := pkg.ScanFolderForConfigFiles(decoded)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("Could not find any config.toml files in the scanned folder %s: %v", decoded, err))
			projectFolderForm := templates.ProjectFolderForm(warning)
			projectFolderForm.Render(context.Background(), w)
			return
		}
		prev := globalServerFolder
		globalServerFolder = decoded
		configFiles = scanned
		log.Printf("Found %d config files in %s (server_folder query overrides default)", len(configFiles), decoded)

		if fileWatcherEnabled && fileWatcher != nil && prev != decoded {
			if fileWatcher.IsWatching() {
				fileWatcher.StopWatching()
				nw, werr := NewFileWatcher()
				if werr != nil {
					log.Printf("Warning: could not recreate file watcher: %v", werr)
					fileWatcher = nil
				} else {
					fileWatcher = nw
				}
			}
			if fileWatcher != nil && !fileWatcher.IsWatching() {
				if err := fileWatcher.StartWatching(decoded); err != nil {
					log.Printf("Warning: could not start file watching on %s: %v", decoded, err)
				}
			}
		}
	} else if globalServerFolder != "" {
		log.Printf("globalServerFolder: %s", globalServerFolder)
	}

	// If no server_folder is provided and no config is specified, show the server folder selector
	if serverFolderEncoded == "" && configEntryPath == "" {
		projectFolderForm := templates.ProjectFolderForm(nil)
		projectFolderForm.Render(context.Background(), w)
		return
	}

	if configEntryPath == "" {
		if len(configFiles) == 0 {
			projectFolderForm := templates.ProjectFolderForm(nil)
			projectFolderForm.Render(context.Background(), w)
			return
		}

		configEntry := templates.EnterConfigPage(configFiles, globalServerFolder)
		configEntry.Render(context.Background(), w)
	} else {

		var decodedConfigEntryPath string
		decodedConfigEntryPathBytes, err := base64.StdEncoding.DecodeString(configEntryPath)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("DecodeString Error: %v", err))
			warning.Render(context.Background(), w)
			return
		} else {
			decodedConfigEntryPath = string(decodedConfigEntryPathBytes)
		}

		// Construct full path if we have a server folder and the path is relative
		fullConfigPath := decodedConfigEntryPath
		if globalServerFolder != "" && !filepath.IsAbs(decodedConfigEntryPath) {
			fullConfigPath = filepath.Join(globalServerFolder, decodedConfigEntryPath)
		}

		config, warn, err := pkg.LoadConfigFromFile(models.CmdFlags{ConfigFile: fullConfigPath, ServerFolder: globalServerFolder, Server: true})
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("LoadConfigFromFile Error: %v (%s)\n\nglobalServerFolder: %s", err, fullConfigPath, globalServerFolder))
			warning.Render(context.Background(), w)
			return
		}

		instances, err := pkg.GenerateInstanceConfigs(config)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("GenerateInstanceConfigsError: %v", err))
			warning.Render(context.Background(), w)
			return
		}

		for _, instance := range instances {
			if len(instance.SkippedReason) == 0 {
				config.TotalQueuedInstances++
			} else {
				log.Printf("Skipped instance: %s - %s", instance.AutoName, instance.SkippedReason)
			}
		}

		log.Printf("Generating report for %s with %d instances (total queued instances: %d)", decodedConfigEntryPath, len(instances), config.TotalQueuedInstances)

		var warning templ.Component
		if warn != nil {
			warning = templates.Warning(fmt.Sprintf("Warning: %v", warn))
		}

		reportMeta := pkg.BuildReportMeta(models.BuildReportMetaParams{
			IsServerMode:         true,
			ConfigFilePath:       decodedConfigEntryPath,
			ServerFolder:         globalServerFolder,
			Config:               config,
			Instances:            instances,
			TotalQueuedInstances: config.TotalQueuedInstances,
		}, models.Results{
			TimeTake: 0,
		})
		report := templates.Report("view", config, instances, "", []models.GenerateSTLResult{}, []models.GenerateImageResult{}, []string{}, reportMeta, 0, warning)
		report.Render(context.Background(), w)
	}
}

// handlePOSTRequest handles POST requests for adding configs
func handlePOSTRequest(w http.ResponseWriter, r *http.Request) {
	log.Print("POST (Add Config) request" + r.Method)

	if r.FormValue("create_project") == "1" {
		name := strings.TrimSpace(r.FormValue("project_name"))
		parent := strings.TrimSpace(r.FormValue("project_parent"))
		if parent == "" && globalServerFolder != "" {
			parent = globalServerFolder
		}
		extended := r.FormValue("project_extended") == "1" || r.FormValue("project_extended") == "on"
		if name == "" {
			warning := templates.Warning("project_name is required")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			warning.Render(context.Background(), w)
			return
		}
		if parent == "" {
			warning := templates.Warning("project parent path is required (or open the config list with a server folder so it can default)")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			warning.Render(context.Background(), w)
			return
		}
		if err := pkg.InitConfigInParent(parent, name, extended); err != nil {
			warning := templates.Warning(fmt.Sprintf("Create project: %v", err))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			warning.Render(context.Background(), w)
			return
		}
		if globalServerFolder != "" {
			scanned, err := pkg.ScanFolderForConfigFiles(globalServerFolder)
			if err != nil {
				log.Printf("rescan after create project: %v", err)
			} else {
				configFiles = scanned
			}
		}
		http.Redirect(w, r, pkg.BuildHomeURL(globalServerFolder), http.StatusSeeOther)
		return
	}

	projectFolder := r.FormValue("project_folder")
	if projectFolder != "" {
		configFiles, err := pkg.ScanFolderForConfigFiles(projectFolder)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("ScanFolderForConfigFiles Error: %v \n(projectFolder: %s)", err, projectFolder))
			warning.Render(context.Background(), w)
			return
		}
		log.Printf("Found %d config files in %s", len(configFiles), projectFolder)
		configFilesComponent := templates.ListConfig(configFiles)
		configFilesComponent.Render(context.Background(), w)
		return
	}

	configEntry := r.FormValue("config_path")
	if configEntry == "" {
		warning := templates.Warning("No config_path file provided")
		warning.Render(context.Background(), w)
		return
	}

	pageUrlInfo := pkg.BuildPageUrl(configEntry, globalServerFolder)

	http.Redirect(w, r, pageUrlInfo.PageURL, http.StatusSeeOther)
}

type StartProcessingForm struct {
	Path                   string `json:"config_path" form:"config_path"`
	PageInstancesSignature string `json:"page_instances_signature" form:"page_instances_signature"`
	ServerFolder           string `json:"server_folder" form:"server_folder"`
	Regex                  string `json:"regex" form:"regex"`
	ServerModeConfigFile   string `json:"server_mode_config_file" form:"server_mode_config_file"`
	Quiet                  bool   `json:"quiet" form:"quiet"`
	Debug                  bool   `json:"debug" form:"debug"`
	NoProcessing           bool   `json:"no_processing" form:"no_processing"`
	Version                bool   `json:"version" form:"version"`
	RegexPattern           string `json:"regex_pattern" form:"regex_pattern"`
	MaxInstances           int    `json:"max_instances" form:"max_instances"`
	StopOnError            bool   `json:"stop_on_error" form:"stop_on_error"`
	IncludeExportLog       bool   `json:"include_export_log" form:"include_export_log"`
	ConfigFile             string `json:"config_file" form:"config_file"`
	SkipRender             bool   `json:"skip_render" form:"skip_render"`
	SkipReadme             bool   `json:"skip_readme" form:"skip_readme"`
	LowQuality             bool   `json:"low_quality" form:"low_quality"`
}

// formToCmdFlags maps StartProcessingForm to models.CmdFlags with explicit field mapping
func formToCmdFlags(form StartProcessingForm) models.CmdFlags {
	return models.CmdFlags{
		ConfigFile:        form.Path,
		RegexPattern:      form.Regex,
		Quiet:             form.Quiet,
		Debug:             form.Debug,
		NoProcessing:      form.NoProcessing,
		Version:           form.Version,
		MaxInstances:      form.MaxInstances,
		StopOnError:       form.StopOnError,
		IncludeExportLog:  form.IncludeExportLog,
		ShowMan:           false,
		Server:            true, // Always true for server requests
		ServerFolder:      form.ServerFolder,
		ServerPort:        0,
		SkipRender:        form.SkipRender,
		SkipReadme:        form.SkipReadme,
		EnableFileWatcher: false,
	}
}

// handlePUTRequest handles PUT requests for processing
func handlePUTRequest(w http.ResponseWriter, r *http.Request) {
	log.Print("PUT (Processing) request" + r.Method)

	var form StartProcessingForm

	// Parse request based on content type
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("PUT - Processing - Invalid JSON body: " + err.Error()))
			return
		}
	} else {
		// Parse form data manually
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("PUT - Processing - Failed to parse form: " + err.Error()))
			return
		}

		pathEncoded := r.FormValue("config_path")
		configFilePathBytes, err := base64.StdEncoding.DecodeString(pathEncoded)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("PUT - Processing - Invalid path encoding for form value config_path %v\n\n(pathEncoded): %s", err, pathEncoded))
			warning.Render(context.Background(), w)
			return
		}
		form.Path = string(configFilePathBytes)
		form.PageInstancesSignature = r.FormValue("page_instances_signature")

		serverFolderEncoded := r.FormValue("server_folder")
		serverFolderBytes, err := base64.StdEncoding.DecodeString(serverFolderEncoded)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("Error decoding server folder: %v", err))
			warning.Render(context.Background(), w)
			return
		}
		form.ServerFolder = string(serverFolderBytes)
		// Map form values to struct fields
		form.Regex = r.FormValue("regex")
		form.ServerModeConfigFile = r.FormValue("server_mode_config_file")
		form.Quiet = r.FormValue("quiet") == "true" || r.FormValue("quiet") == "1" || r.FormValue("quiet") == "on"
		form.Debug = r.FormValue("debug") == "true" || r.FormValue("debug") == "1" || r.FormValue("debug") == "on"
		form.NoProcessing = r.FormValue("no_processing") == "true" || r.FormValue("no_processing") == "1" || r.FormValue("no_processing") == "on"
		form.RegexPattern = r.FormValue("regex_pattern")
		if maxInstances := r.FormValue("max_instances"); maxInstances != "" {
			if val, err := strconv.Atoi(maxInstances); err == nil {
				form.MaxInstances = val
			}
		}
		form.StopOnError = r.FormValue("stop_on_error") == "true" || r.FormValue("stop_on_error") == "1" || r.FormValue("stop_on_error") == "on"
		rawConfigFile := r.FormValue("raw_config_file")
		rawConfigFileBytes, err := base64.StdEncoding.DecodeString(rawConfigFile)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("DecodeString Error: %v", err))
			warning.Render(context.Background(), w)
			return
		}
		form.ConfigFile = string(rawConfigFileBytes)
		form.SkipRender = r.FormValue("skip_render") == "true" || r.FormValue("skip_render") == "1" || r.FormValue("skip_render") == "on"
		form.SkipReadme = r.FormValue("skip_readme") == "true" || r.FormValue("skip_readme") == "1" || r.FormValue("skip_readme") == "on"
		form.LowQuality = r.FormValue("low_quality") == "true" || r.FormValue("low_quality") == "1" || r.FormValue("low_quality") == "on"
	}

	// Map form to cmdFlags
	cmdFlags := formToCmdFlags(form)

	if cmdFlags.ConfigFile == "" {
		warning := templates.Warning("No config_path file provided")
		warning.Render(context.Background(), w)
		return
	}

	var useOOBUpdates bool
	if form.Regex != "" {
		useOOBUpdates = true
	}

	cmdFlags.OverwriteExisting = true
	cmdFlags.IncludeExportLog = true

	config, _, err := pkg.LoadConfigFromFile(cmdFlags)
	if err != nil {
		warning := templates.Warning(fmt.Sprintf("LoadConfigFromFileError: %v \n\n (ConfigFile: %s, ServerFolder: %s)", err, cmdFlags.ConfigFile, cmdFlags.ServerFolder))
		warning.Render(context.Background(), w)
		return
	}

	instancesForPage, err := pkg.GenerateInstanceConfigs(config)
	if err != nil {
		warning := templates.Warning(fmt.Sprintf("GenerateInstanceConfigsError: %v", err))
		warning.Render(context.Background(), w)
		return
	}
	config.TotalQueuedInstances = 0
	for _, instance := range instancesForPage {
		if instance.SkippedReason == "" {
			config.TotalQueuedInstances++
		}
	}
	loadedInstanceSignature := pkg.BuildInstanceSetSignature(instancesForPage)
	if form.PageInstancesSignature != "" && loadedInstanceSignature != "" && form.PageInstancesSignature != loadedInstanceSignature {
		pageUrlInfo := pkg.BuildPageUrl(form.Path, form.ServerFolder)
		w.Header().Set("HX-Redirect", pageUrlInfo.PageURL)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("Reloading updated instances before processing..."))
		return
	}

	id := StartProcessingJob(config)

	if useOOBUpdates {
		w.Header().Set("Content-Type", "text/html")
		progressContainer := templates.ProgressContainer(id)
		progressContainer.Render(context.Background(), w)
	} else {
		w.Header().Set("Content-Type", "text/html")

		pageUrlInfo := pkg.BuildPageUrl(form.Path, globalServerFolder)
		// set the url to the full config file path with htmx and no reload
		w.Header().Set("HX-Push-Url", pageUrlInfo.PageURL)

		progressHTML := templates.GetProgressHTML(id)
		progressHTML.Render(context.Background(), w)
		//fmt.Fprintf(w, progressHTML)
	}
}

func handleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	configFile := r.URL.Query().Get("config_file")
	dryRun := r.URL.Query().Get("dryRun") == "true"
	cleanOldVersions := r.URL.Query().Get("cleanOldVersions") == "true"

	if len(configFile) > 0 {
		// Clean specific config
		folder := filepath.Dir(configFile)
		_, err := pkg.CleanConfig(dryRun, cleanOldVersions, folder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Clean entire directory
		folder := r.URL.Query().Get("server_folder")
		if folder == "" {
			folder = "."
		}
		_, err := pkg.CleanDirectory(dryRun, cleanOldVersions, folder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// handleOpenSCADStatus returns HTML: full page by default, or HTMX fragment when ?fragment=1.
func handleOpenSCADStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx := r.Context()
	info, err := pkg.ProbeOpenSCAD()
	var v templates.OpenSCADNavView
	if err != nil {
		v = templates.OpenSCADNavView{Available: false, Error: err.Error()}
	} else {
		v = templates.OpenSCADNavView{
			Available: true,
			Path:      info.Path,
			Version:   info.Version,
			OutOfDate: info.IsOutOfDate,
		}
	}
	v.DetailsOpen = r.URL.Query().Get("isExpanded") == "1"
	if r.URL.Query().Get("fragment") == "1" {
		templates.OpenSCADNavWidget(v).Render(ctx, w)
		return
	}
	templates.OpenSCADStatusFullPage(v).Render(ctx, w)
}

// handleWatcherStatus returns the current file watching status
func handleWatcherStatus(w http.ResponseWriter, r *http.Request) {
	if !fileWatcherEnabled {
		w.Header().Set("Content-Type", "application/json")
		status := map[string]interface{}{
			"watching": false,
			"configs":  0,
			"enabled":  false,
		}
		json.NewEncoder(w).Encode(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	status := map[string]interface{}{
		"watching": fileWatcher.IsWatching(),
		"configs":  len(fileWatcher.GetWatchedConfigs()),
		"enabled":  true,
	}
	json.NewEncoder(w).Encode(status)
}

// handleWatcherPause pauses file watching
func handleWatcherPause(w http.ResponseWriter, r *http.Request) {
	if fileWatcher == nil {
		http.Error(w, "File watcher not enabled", http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "handleWatcherPause Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileWatcher.StopWatching()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File watching paused"))
}

// handleWatcherResume resumes file watching
func handleWatcherResume(w http.ResponseWriter, r *http.Request) {
	if fileWatcher == nil {
		http.Error(w, "File watcher not enabled", http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "handleWatcherResume Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, just restart watching the server folder
	// In a real implementation, you'd want to store the folder path
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File watching resumed"))
}

// handleWatcherUI renders the watcher status UI
func handleWatcherUI(w http.ResponseWriter, r *http.Request) {
	if !fileWatcherEnabled {
		status := models.WatcherStatusUI{
			Watching:    false,
			ConfigPaths: []string{},
			Enabled:     false,
		}
		templates.WatcherStatusComponent(status).Render(r.Context(), w)
		return
	}
	configs := fileWatcher.GetWatchedConfigs()
	paths := make([]string, 0, len(configs))
	for path := range configs {
		paths = append(paths, path)
	}
	status := models.WatcherStatusUI{
		Watching:    fileWatcher.IsWatching(),
		ConfigPaths: paths,
		Enabled:     true,
	}
	templates.WatcherStatusComponent(status).Render(r.Context(), w)
}

func buildEditConfigParams(configPath string, serverFolder string, content string, errorMsg templ.Component) models.EditConfigParams {
	configPathEncoded := base64.StdEncoding.EncodeToString([]byte(configPath))
	serverFolderEncoded := base64.StdEncoding.EncodeToString([]byte(serverFolder))
	return models.EditConfigParams{
		ConfigFilePath:        configPath,
		ConfigFilePathEncoded: configPathEncoded,
		FilePathEncoded:       base64.StdEncoding.EncodeToString([]byte(content)),
		ServerFolder:          serverFolder,
		ServerFolderEncoded:   serverFolderEncoded,
		Content:               content,
		ErrorMsg:              errorMsg,
	}
}

func renderTOMLEditPage(w http.ResponseWriter, r *http.Request, statusCode int, configPath string, serverFolder string, content string, message templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if statusCode != 0 {
		w.WriteHeader(statusCode)
	}
	configParams := buildEditConfigParams(configPath, serverFolder, content, message)
	templates.TOMLEditForm(&configParams).Render(r.Context(), w)
}

// handleConfigRequest handles GET and POST requests for config file operations
func handleConfigRequest(w http.ResponseWriter, r *http.Request) {
	configPathEncoded := r.URL.Query().Get("config_path")
	if configPathEncoded == "" {
		http.Error(w, "Missing 'config_path' query parameter", http.StatusBadRequest)
		return
	}

	configPathBytes, err := base64.StdEncoding.DecodeString(configPathEncoded)
	if err != nil {
		http.Error(w, "Invalid config path encoding", http.StatusBadRequest)
		return
	}
	configPath := string(configPathBytes)

	serverFolderEncoded := r.URL.Query().Get("server_folder")
	if serverFolderEncoded == "" {
		http.Error(w, "Missing 'server_folder' query parameter", http.StatusBadRequest)
		return
	}
	serverFolderBytes, err := base64.StdEncoding.DecodeString(serverFolderEncoded)
	if err != nil {
		http.Error(w, "Invalid server folder encoding", http.StatusBadRequest)
		return
	}
	serverFolder := string(serverFolderBytes)

	// Resolve path using server folder structure
	//configPath = resolveConfigPath(configPath, serverFolder)

	switch r.Method {
	case "GET":
		handleConfigGet(w, r, configPath, serverFolder)
	case "POST":
		handleConfigPost(w, r, configPath, serverFolder)
	default:
		http.Error(w, "handleConfigRequest - Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleConfigGet reads and returns the config file content
func handleConfigGet(w http.ResponseWriter, r *http.Request, configPath string, serverFolder string) {
	configPath = resolveConfigPath(configPath, serverFolder)
	content, err := os.ReadFile(configPath)
	if err != nil {
		renderTOMLEditPage(w, r, http.StatusInternalServerError, configPath, serverFolder, "", templates.Warning(fmt.Sprintf("Failed to read config file:\n%s", err)))
		return
	}

	renderTOMLEditPage(w, r, http.StatusOK, configPath, serverFolder, string(content), nil)
}

// handleConfigPost validates TOML and updates the config file
func handleConfigPost(w http.ResponseWriter, r *http.Request, configPath string, serverFolder string) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		renderTOMLEditPage(w, r, http.StatusBadRequest, configPath, serverFolder, "", templates.Warning(fmt.Sprintf("Failed to parse form:\n%s", err)))
		return
	}

	path := resolveConfigPath(configPath, serverFolder)

	configContent := r.FormValue("content")
	if configContent == "" {
		renderTOMLEditPage(w, r, http.StatusBadRequest, configPath, serverFolder, "", templates.Warning("Missing 'content' form field"))
		return
	}

	// Validate TOML by attempting to decode it
	var testConfig models.Config
	_, err := toml.Decode(configContent, &testConfig)
	if err != nil {
		renderTOMLEditPage(w, r, http.StatusBadRequest, configPath, serverFolder, configContent, templates.Warning("Invalid TOML:\n"+pkg.FormatTOMLDecodeError(configContent, err)))
		return
	}

	// Write the content to the file
	err = os.WriteFile(path, []byte(configContent), 0644)
	if err != nil {
		renderTOMLEditPage(w, r, http.StatusInternalServerError, configPath, serverFolder, configContent, templates.Warning(fmt.Sprintf("Failed to write config file:\n%s", err)))
		return
	}

	success := templates.Success("Config saved successfully")
	renderTOMLEditPage(w, r, http.StatusOK, configPath, serverFolder, configContent, success)

}

// handleImageRequest serves image files for the web interface
func handleImageRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "handleImageRequest - Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imagePath := r.URL.Query().Get("config_path")
	if imagePath == "" {
		http.Error(w, "Missing 'config_path' query parameter", http.StatusBadRequest)
		return
	}

	// Resolve path using server folder structure
	imagePath = resolveConfigPath(imagePath, globalServerFolder)

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "Image file not found at: "+imagePath, http.StatusNotFound)
		return
	}

	// Set appropriate content type based on file extension
	ext := filepath.Ext(imagePath)
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	default:
		warning := templates.Warning("Unsupported image file type: " + ext)
		warning.Render(r.Context(), w)
	}

	// Prevent stale same-path renders from sticking around after HTMX swaps.
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve the file
	http.ServeFile(w, r, imagePath)
}

// handleStaticFiles serves static files from the static directory
func handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "handleStaticFiles - Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Remove /static/ prefix and construct file path
	filePath := r.URL.Path[len("/static/"):]
	if filePath == "" {
		http.Error(w, "Missing file path", http.StatusBadRequest)
		return
	}

	// Construct absolute path to static file
	absPath := filepath.Join("static", filePath)

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set appropriate content type based on file extension
	ext := filepath.Ext(absPath)
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".stl":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "inline")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Serve the file
	http.ServeFile(w, r, absPath)
}

// handleOpenFile opens a file in the system's default editor
func handleOpenFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "handleOpenFile - Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("config_path")
	if filePath == "" {
		http.Error(w, "Missing 'config_path' query parameter", http.StatusBadRequest)
		return
	}

	// Resolve path using server folder structure
	filePath = resolveConfigPath(filePath, globalServerFolder)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Open file with system default editor
	// This will work on macOS, Linux, and Windows
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filePath)
	default:
		http.Error(w, "Unsupported operating system", http.StatusBadRequest)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "File opened in default editor",
		"path":    filePath,
	})
}

// handleEditFile handles GET and POST requests for editing TOML files
func handleEditFile(w http.ResponseWriter, r *http.Request) {
	filePathEncoded := r.URL.Query().Get("config_path")
	if filePathEncoded == "" {
		http.Error(w, "Missing 'config_path' query parameter", http.StatusBadRequest)
		return
	}
	filePathBytes, err := base64.StdEncoding.DecodeString(filePathEncoded)
	if err != nil {
		http.Error(w, "Invalid config path encoding", http.StatusBadRequest)
		return
	}
	filePath := string(filePathBytes)

	serverFolderEncoded := r.URL.Query().Get("server_folder")
	if serverFolderEncoded == "" {
		http.Error(w, "Missing 'server_folder' query parameter", http.StatusBadRequest)
		return
	}
	serverFolderBytes, err := base64.StdEncoding.DecodeString(serverFolderEncoded)
	if err != nil {
		http.Error(w, "Invalid server folder encoding", http.StatusBadRequest)
		return
	}
	serverFolder := string(serverFolderBytes)

	// Resolve path using server folder structure
	filePath = resolveConfigPath(filePath, serverFolder)

	// Validate file extension
	if filepath.Ext(filePath) != ".toml" && filepath.Base(filePath) != "config.toml" {
		http.Error(w, "Only config.toml files are allowed", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleEditGet(w, r, serverFolder, filePath)
	case "POST":
		handleEditPost(w, r, serverFolder, filePath)
	default:
		http.Error(w, "HandleEditFile - Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleEditGet loads and displays the TOML file in an editable form
func handleEditGet(w http.ResponseWriter, r *http.Request, serverFolder string, filePath string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		renderTOMLEditPage(w, r, http.StatusNotFound, filePath, serverFolder, "", templates.Warning("File not found"))
		return
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		renderTOMLEditPage(w, r, http.StatusInternalServerError, filePath, serverFolder, "", templates.Warning(fmt.Sprintf("Failed to read file:\n%s", err)))
		return
	}

	// Render the edit form
	renderTOMLEditPage(w, r, http.StatusOK, filePath, serverFolder, string(content), nil)
}

// handleEditPost validates and saves the TOML file
func handleEditPost(w http.ResponseWriter, r *http.Request, filePath string, serverFolder string) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		renderTOMLEditPage(w, r, http.StatusBadRequest, filePath, serverFolder, "", templates.Warning(fmt.Sprintf("Failed to parse form:\n%s", err)))
		return
	}

	content := r.FormValue("content")
	if content == "" {
		error := templates.Warning("Content cannot be empty")
		// Show form with error
		configParams := buildEditConfigParams(filePath, serverFolder, content, error)
		editForm := templates.TOMLEditForm(&configParams)
		editForm.Render(r.Context(), w)
		return
	}

	// Validate TOML by attempting to decode it
	var testConfig models.Config
	metadata, err := toml.Decode(content, &testConfig)
	if err != nil {
		errorMsg := templates.Warning("Invalid TOML:\n" + pkg.FormatTOMLDecodeError(content, err))
		// Show form with validation error
		configParams := buildEditConfigParams(filePath, serverFolder, content, errorMsg)
		editForm := templates.TOMLEditForm(&configParams)
		editForm.Render(r.Context(), w)
		return
	}

	// Check for undecoded keys (invalid fields)
	undecoded := metadata.Undecoded()
	if len(undecoded) > 0 {
		var invalidFields []string
		for _, key := range undecoded {
			invalidFields = append(invalidFields, key.String())
		}
		errorMsg := templates.Warning(fmt.Sprintf("Invalid fields in config: %v", invalidFields))
		// Show form with validation error
		configParams := buildEditConfigParams(filePath, serverFolder, content, errorMsg)
		editForm := templates.TOMLEditForm(&configParams)
		editForm.Render(r.Context(), w)
		return
	}

	// Write the content to the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		errorMsg := templates.Warning(fmt.Sprintf("Failed to save file: %v", err))
		// Show form with save error
		configParams := buildEditConfigParams(filePath, serverFolder, content, errorMsg)
		editForm := templates.TOMLEditForm(&configParams)
		editForm.Render(r.Context(), w)
		return
	}
	success := templates.Success("File saved successfully")
	configParams := buildEditConfigParams(filePath, serverFolder, content, success)
	editForm := templates.TOMLEditForm(&configParams)
	editForm.Render(r.Context(), w)
}

// handlePreviewRequest handles STL preview requests with three.js viewer
func handlePreviewRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("handlePreviewRequest Method not allowed: %s", r.Method), http.StatusMethodNotAllowed)
		return
	}

	configFilePath := r.URL.Query().Get("config_path")
	if configFilePath == "" {
		http.Error(w, "Missing 'config_path' query parameter", http.StatusBadRequest)
		return
	}

	serverFolder := r.URL.Query().Get("server_folder")
	if serverFolder == "" {
		http.Error(w, "Missing 'server_folder' query parameter", http.StatusBadRequest)
		return
	}

	instanceID := r.URL.Query().Get("instance")
	if instanceID == "" {
		http.Error(w, "Missing 'instance' query parameter", http.StatusBadRequest)
		return
	}

	filePathEncoded := r.URL.Query().Get("file_path")
	if filePathEncoded == "" {
		http.Error(w, "Missing 'file_path' query parameter", http.StatusBadRequest)
		return
	}

	filePathBytes, err := base64.StdEncoding.DecodeString(filePathEncoded)
	if err != nil {
		http.Error(w, "Invalid config path encoding", http.StatusBadRequest)
		return
	}
	filePath := string(filePathBytes)

	log.Printf("Preview request for instance: %s", instanceID)

	// Load config and find the instance to get the STL file path
	log.Printf("Checking config file: %s (encoded as: %s)", filePath, filePathEncoded)
	config, _, err := pkg.LoadConfigFromFile(models.CmdFlags{ConfigFile: filePath, Server: true, ServerFolder: globalServerFolder})
	if err != nil {
		log.Printf("handlePreviewRequest Error loading config %s: %v", filePath, err)
		warning := templates.Warning(fmt.Sprintf("handlePreviewRequest Error loading config %s: %v", filePath, err))
		warning.Render(r.Context(), w)
		return
	}

	instances, err := pkg.GenerateInstanceConfigs(config)
	if err != nil {
		log.Printf("Error generating instances for %s: %v", filePath, err)
		http.Error(w, "Error generating instances", http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d instances in %s", len(instances), filePath)
	var targetInstance *models.InstanceConfig
	for _, instance := range instances {
		log.Printf("Checking instance: %s", instance.UniqueID)
		if instance.UniqueID == instanceID {
			targetInstance = &instance
			log.Printf("Found matching instance: %s, STL path: %s", instance.UniqueID, instance.RunOutputPathV3)
			break
		}
	}

	if targetInstance == nil {
		log.Printf("Instance not found: %s", instanceID)
		http.Error(w, fmt.Sprintf("Instance not found: %s", instanceID), http.StatusNotFound)
		return
	}

	// Check if STL file exists
	if _, err := os.Stat(targetInstance.RunOutputPathV3); os.IsNotExist(err) {
		log.Printf("STL file not found: %s", targetInstance.RunOutputPathV3)
		http.Error(w, "STL file not found", http.StatusNotFound)
		return
	}

	if len(targetInstance.RunOutputPathV3) == 0 {
		http.Error(w, "STL file not found", http.StatusNotFound)
		return
	}

	// Create API path for serving the STL file
	// Base64 encode the STL file path
	encodedSTLPath := base64.StdEncoding.EncodeToString([]byte(targetInstance.RunOutputPathV3))
	stlPath := fmt.Sprintf("/api/stl?file_path=%s", encodedSTLPath)

	log.Printf("Using STL file path: %s", targetInstance.RunOutputPathV3)
	log.Printf("Encoded STL path: %s", encodedSTLPath)

	pageUrlInfo := pkg.BuildPageUrl(filePath, serverFolder)

	// Render the three.js viewer template
	viewer := templates.STLViewerHeaderBody(models.STLViewerParams{
		InstanceID:  instanceID,
		STLPath:     stlPath,
		PageUrlInfo: pageUrlInfo,
	})
	viewer.Render(r.Context(), w)
	return
}

// handleSTLRequest serves STL files for the 3D viewer
func handleSTLRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "handleSTLRequest Method not allowed: "+r.Method, http.StatusMethodNotAllowed)
		return
	}

	// Get the configPath from the query parameters
	filePathEncoded := r.URL.Query().Get("file_path")
	if filePathEncoded == "" {
		http.Error(w, "Missing file_path path", http.StatusBadRequest)
		return
	} else {
		log.Printf("handleSTLRequest - filePathEncoded: %s", filePathEncoded)
	}

	// Decode the base64 encoded config path
	/*configPathBytes, err := base64.StdEncoding.DecodeString(configPathEncoded)
	if err != nil {
		http.Error(w, "Invalid config path encoding", http.StatusBadRequest)
		return
	}*/
	//configPath := string(configPathBytes)

	// decode base64 path
	var filePath string
	filePathBytes, err := base64.StdEncoding.DecodeString(filePathEncoded)
	if err != nil {
		http.Error(w, "Invalid file_path encoding", http.StatusBadRequest)
		return
	} else {
		filePath = string(filePathBytes)
	}

	log.Printf("STL handler - filePath: %s", filePath)

	// Check if STL file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("STL file not found: %s", filePath)
		http.Error(w, "STL file not found", http.StatusNotFound)
		return
	}

	// Check file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		http.Error(w, "Error accessing file", http.StatusInternalServerError)
		return
	}

	log.Printf("STL file size: %d bytes", fileInfo.Size())

	// Set headers for STL file
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Serve the STL file
	log.Printf("Serving STL file: %s", filePath)
	http.ServeFile(w, r, filePath)
}
