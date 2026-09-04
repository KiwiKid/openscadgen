package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
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
var aiToolsEnabled bool
var globalServerFolder string
var startOpenSCADFile = openOpenSCADFile
var installOpenSCAD = installOpenSCADWithHomebrew
var installBOSL2 = pkg.InstallOrUpgradeBOSL2

const bosl2ManualInstallCommand = `set -e
mkdir -p "$HOME/.local/share/OpenSCAD/libraries"
cd "$HOME/.local/share/OpenSCAD/libraries"
if [ -d BOSL2 ]; then mv BOSL2 "BOSL2.backup-$(date +%s)"; fi
git clone --depth 1 https://github.com/BelfrySCAD/BOSL2.git`

func openSCADManualInstallCommand(installNightly bool) string {
	formula := "openscad"
	if installNightly {
		formula = "openscad@snapshot"
	}
	return "brew install " + formula
}

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
	pkg.LogWarnf("Port %s is in use, trying random port...", port)
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
	aiToolsEnabled = cmdFlags.ShowAI
	port := ":8080"
	if cmdFlags.ServerPort != 0 {
		port = fmt.Sprintf(":%d", cmdFlags.ServerPort)
	}

	var err error
	if cmdFlags.Debug {
		pkg.LogDebugf("Debug mode enabled")
	}
	var msg = fmt.Sprintf("Starting server on port %s", port)

	// Initialize file watcher only if enabled
	fileWatcherEnabled = cmdFlags.EnableFileWatcher
	if cmdFlags.EnableFileWatcher {
		fileWatcher, err = NewFileWatcher()
		if err != nil {
			pkg.LogErrorf("Error creating file watcher: %v", err)
			os.Exit(1)
		}
	}

	if serverFolder != "" {
		globalServerFolder = serverFolder
		configFiles, err = pkg.ScanFolderForConfigFiles(serverFolder)
		if err != nil {
			pkg.LogErrorf("ScanFolderForConfigFiles Error: %v", err)
			os.Exit(1)
		}
		pkg.LogInfof("Found %d config files in %s", len(configFiles), serverFolder)
		msg += fmt.Sprintf(" and %d config files in %s", len(configFiles), serverFolder)

		// Start file watching only if enabled
		if cmdFlags.EnableFileWatcher && fileWatcher != nil {
			if err := fileWatcher.StartWatching(serverFolder); err != nil {
				pkg.LogWarnf("Warning: Could not start file watching: %v", err)
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
	http.HandleFunc("/tools", handleToolsRequest)
	http.HandleFunc("/api/chat/run", handleChatRun)
	http.HandleFunc("/api/chat/status", handleChatStatus)
	http.HandleFunc("/api/config", handleConfigRequest)
	http.HandleFunc("/api/config/options", handleConfigOptionsRequest)
	http.HandleFunc("/api/open", handleOpenFile)
	http.HandleFunc("/api/openscad/open", handleOpenSCADFile)
	http.HandleFunc("/api/edit", handleEditFile)
	http.HandleFunc("/health", handleHealth)

	http.HandleFunc("/static/", handleStaticFiles)
	http.HandleFunc("/images", handleImageRequest)

	http.HandleFunc("/api/openscad/status", handleOpenSCADStatus)
	http.HandleFunc("/api/openscad/install", handleOpenSCADInstall)
	http.HandleFunc("/api/openscad/libraries/", handleOpenSCADLibraryAction)
	http.HandleFunc("/api/openscad/health/badge", handleOpenSCADHealthBadge)
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
		pkg.LogErrorf("Is Openscadgen already running? %v", err)
		pkg.LogErrorf("Could not bind to port %s: %v\nEnd the existing process or use -p to specify a different port", port, err)
		os.Exit(1)
	}

	// Update port if we got a different one
	if actualPort != port {
		pkg.LogWarnf("Port %s was in use, using %s instead", originalPort, actualPort)
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
	pkg.LogStagef("server", "%s", msg)

	// Call onStart callback before starting server
	if onStart != nil {
		pkg.LogInfof("onStart starting")
		err = onStart(port)
		if err != nil {
			listener.Close()
			pkg.LogErrorf("Error on start: %v", err)
			os.Exit(1)
		}
	}

	pkg.LogInfof("Server started on port %s", port)

	// Start server on the listener (non-blocking for now, but we'll block here)
	err = http.Serve(listener, nil)
	if err != nil {
		pkg.LogErrorf("Error on serve: %v", err)
		os.Exit(1)
	}

	return ServerInfo{Port: port, Address: fmt.Sprintf("http://localhost%s", port)}
}

func handleDeleteExportSTLs(w http.ResponseWriter, r *http.Request) {
	resolveDeleteScope := func(configPath string, serverFolder string, version string, includeOtherVersions bool) (templates.DeleteExportSTLsPageData, error) {
		flags := models.CmdFlags{ConfigFile: configPath, ServerFolder: serverFolder}
		config, _, err := pkg.LoadConfigFromFile(flags)
		if err != nil {
			return templates.DeleteExportSTLsPageData{}, err
		}
		activeVersion := strings.TrimSpace(version)
		if activeVersion == "" {
			activeVersion = config.Design.Version
		}
		outputRoot := filepath.Join(filepath.Dir(config.ConfigFile), config.Design.OutputPath)
		currentFiles, otherFiles, err := pkg.PreviewVersionFiles(outputRoot, config.Design.ClearVersion(activeVersion), includeOtherVersions)
		if err != nil {
			return templates.DeleteExportSTLsPageData{}, err
		}
		return templates.DeleteExportSTLsPageData{
			RootDir:              outputRoot,
			ConfigFilePath:       config.ConfigFile,
			ConfigVersion:        activeVersion,
			CurrentVersionFiles:  currentFiles,
			OtherVersionFiles:    otherFiles,
			SelectedOtherVersion: includeOtherVersions,
		}, nil
	}

	switch r.Method {
	case "GET":
		configPath := r.URL.Query().Get("config_path")
		if configPath == "" {
			http.Error(w, "Missing config_path", http.StatusBadRequest)
			return
		}
		decodedConfigPath, err := base64.StdEncoding.DecodeString(configPath)
		if err != nil {
			http.Error(w, "Invalid config_path", http.StatusBadRequest)
			return
		}
		serverFolder := r.URL.Query().Get("server_folder")
		if serverFolder != "" {
			decodedServerFolder, err := base64.StdEncoding.DecodeString(serverFolder)
			if err != nil {
				http.Error(w, "Invalid server_folder", http.StatusBadRequest)
				return
			}
			serverFolder = string(decodedServerFolder)
		}
		version := r.URL.Query().Get("version")
		includeOtherVersions := r.URL.Query().Get("include_other_versions") == "true"
		data, err := resolveDeleteScope(string(decodedConfigPath), serverFolder, version, includeOtherVersions)
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
		configPath := r.FormValue("config_path")
		if configPath == "" {
			http.Error(w, "Missing config_path", http.StatusBadRequest)
			return
		}
		version := r.FormValue("version")
		serverFolder := r.FormValue("server_folder")
		includeOtherVersions := r.FormValue("include_other_versions") == "true"
		confirmed := r.FormValue("confirmed") == "true"
		data, err := resolveDeleteScope(configPath, serverFolder, version, includeOtherVersions)
		if err != nil {
			data.Error = err.Error()
		}
		if !confirmed {
			data.PreviewOnly = true
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			templates.DeleteExportSTLsPage(data).Render(r.Context(), w)
			return
		}
		deletePaths := append([]string{}, data.CurrentVersionFiles...)
		for _, files := range data.OtherVersionFiles {
			deletePaths = append(deletePaths, files...)
		}
		if len(deletePaths) > 0 {
			delRes := pkg.DeleteFiles(deletePaths)
			data.Deleted = delRes.Deleted
			if len(delRes.Failed) > 0 {
				data.Failed = map[string]string{}
				for p, e := range delRes.Failed {
					data.Failed[p] = e.Error()
				}
			}
		}
		data.Confirmed = true
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
	pkg.LogInfof("GET (Display) request %s", r.Method)
	configEntryPathEncoded := r.URL.Query().Get("config")
	configEntryPath, err := url.QueryUnescape(configEntryPathEncoded)
	if err != nil {
		pkg.LogWarnf("QueryUnescape Error: %v", err)
	}

	serverFolderEncoded := r.URL.Query().Get("server_folder")
	if serverFolderEncoded != "" {
		pkg.LogInfof("(url) server_folder query: %s", serverFolderEncoded)
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
		pkg.LogInfof("Found %d config files in %s (server_folder query overrides default)", len(configFiles), decoded)

		if fileWatcherEnabled && fileWatcher != nil && prev != decoded {
			if fileWatcher.IsWatching() {
				fileWatcher.StopWatching()
				nw, werr := NewFileWatcher()
				if werr != nil {
					pkg.LogWarnf("Warning: could not recreate file watcher: %v", werr)
					fileWatcher = nil
				} else {
					fileWatcher = nw
				}
			}
			if fileWatcher != nil && !fileWatcher.IsWatching() {
				if err := fileWatcher.StartWatching(decoded); err != nil {
					pkg.LogWarnf("Warning: could not start file watching on %s: %v", decoded, err)
				}
			}
		}
	} else if globalServerFolder != "" {
		pkg.LogInfof("globalServerFolder: %s", globalServerFolder)
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

		configEntry := templates.EnterConfigPage(configFiles, globalServerFolder, aiToolsEnabled)
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
				pkg.LogInfof("Skipped instance: %s - %s", instance.AutoName, instance.SkippedReason)
			}
		}

		pkg.LogStagef("report", "Generating report for %s with %d instances (total queued instances: %d)", decodedConfigEntryPath, len(instances), config.TotalQueuedInstances)

		var warning templ.Component
		if warn != nil {
			warning = templates.Warning(fmt.Sprintf("Warning: %v", warn))
		}

		reportMeta := pkg.BuildReportMeta(models.BuildReportMetaParams{
			IsServerMode:         true,
			ShowAI:               aiToolsEnabled,
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
	pkg.LogInfof("POST (Add Config) request %s", r.Method)

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
				pkg.LogWarnf("rescan after create project: %v", err)
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
		pkg.LogInfof("Found %d config files in %s", len(configFiles), projectFolder)
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
	pkg.LogInfof("PUT (Processing) request %s", r.Method)

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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		templates.ReportAlert(fmt.Sprintf("LoadConfigFromFileError:\n%s", err)).Render(context.Background(), w)
		templates.Warning("Could not start processing because the config could not be loaded. See the alert above.").Render(context.Background(), w)
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func handleOpenSCADHealthBadge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.OpenSCADHealthBadge(true, true, false).Render(r.Context(), w)
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
		v = templates.OpenSCADNavView{
			Available:             false,
			Error:                 err.Error(),
			InstallSupported:      runtime.GOOS == "darwin",
			InstallSupportLabel:   "darwin",
			BOSL2ActionLabel:      "Install",
			BOSL2InstallSupported: runtime.GOOS == "darwin",
		}
	} else {
		bosl2Available, bosl2Err := pkg.ProbeBOSL2(info.Path)
		bosl2ActionLabel := "Install"
		if bosl2Available {
			bosl2ActionLabel = "Upgrade"
		}
		v = templates.OpenSCADNavView{
			Available:             true,
			Path:                  info.Path,
			Version:               info.Version,
			OutOfDate:             info.IsOutOfDate,
			InstallSupported:      runtime.GOOS == "darwin",
			InstallSupportLabel:   "darwin",
			BOSL2Available:        bosl2Available,
			BOSL2ActionLabel:      bosl2ActionLabel,
			BOSL2InstallSupported: runtime.GOOS == "darwin",
		}
		if bosl2Err != nil {
			v.BOSL2Error = bosl2Err.Error()
		}
	}
	v.DetailsOpen = r.URL.Query().Get("isExpanded") == "1"
	if !v.Available {
		v.InstallSupported = runtime.GOOS == "darwin"
		v.InstallSupportLabel = "darwin"
		v.BOSL2InstallSupported = runtime.GOOS == "darwin"
		v.BOSL2ActionLabel = "Install"
	}
	if r.URL.Query().Get("fragment") == "1" {
		templates.OpenSCADNavWidget(v).Render(ctx, w)
		return
	}
	templates.OpenSCADStatusFullPage(v).Render(ctx, w)
}

func installOpenSCADWithHomebrew(ctx context.Context, installNightly bool) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("OpenSCAD installation is currently supported on macOS only")
	}

	brewPath, err := exec.LookPath("brew")
	if err != nil {
		return fmt.Errorf("Homebrew is required to install OpenSCAD: %w", err)
	}

	formula := "openscad"
	if installNightly {
		formula = "openscad@snapshot"
	}
	args := []string{"install", formula}
	if exec.CommandContext(ctx, brewPath, "list", "--versions", formula).Run() == nil {
		args[0] = "upgrade"
	}
	output, err := exec.CommandContext(ctx, brewPath, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Homebrew %s %s failed: %w\n%s", args[0], formula, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func handleOpenSCADInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid install request", http.StatusBadRequest)
		return
	}

	installNightly := r.Form.Get("install_nightly") == "true"
	if err := installOpenSCAD(r.Context(), installNightly); err != nil {
		pkg.LogWarnf("OpenSCAD installation failed: %v", err)
		writeToolActionResult(w, r, "is-danger", "OpenSCAD installation failed", err.Error(), openSCADManualInstallCommand(installNightly))
		return
	}

	installedVersion := "stable OpenSCAD"
	if installNightly {
		installedVersion = "the latest OpenSCAD nightly"
	}
	writeToolActionResult(w, r, "is-success", "OpenSCAD installed", fmt.Sprintf("Installed %s. Reload the page to refresh its status.", installedVersion), "")
}

func handleToolsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	info, err := pkg.ProbeOpenSCAD()
	var v templates.OpenSCADNavView
	if err != nil {
		v = templates.OpenSCADNavView{
			Available:             false,
			Error:                 err.Error(),
			InstallSupported:      runtime.GOOS == "darwin",
			InstallSupportLabel:   "darwin",
			BOSL2ActionLabel:      "Install",
			BOSL2InstallSupported: runtime.GOOS == "darwin",
		}
	} else {
		bosl2Available, bosl2Err := pkg.ProbeBOSL2(info.Path)
		actionLabel := "Install"
		if bosl2Available {
			actionLabel = "Upgrade"
		}
		v = templates.OpenSCADNavView{
			Available:             true,
			Path:                  info.Path,
			Version:               info.Version,
			OutOfDate:             info.IsOutOfDate,
			InstallSupported:      runtime.GOOS == "darwin",
			InstallSupportLabel:   "darwin",
			BOSL2Available:        bosl2Available,
			BOSL2ActionLabel:      actionLabel,
			BOSL2InstallSupported: runtime.GOOS == "darwin",
		}
		if bosl2Err != nil {
			v.BOSL2Error = bosl2Err.Error()
		}
	}
	templates.ToolsPage(v, pkg.OpenSCADToolRegistry(), pkg.BuildHomeURL(globalServerFolder), aiToolsEnabled).Render(r.Context(), w)
}

func handleConfigOptionsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	topic := r.URL.Query().Get("topic")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(pkg.RenderConfigOptionsCLI(topic)))
}

func handleOpenSCADLibraryAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "invalid library action path", http.StatusBadRequest)
		return
	}
	libName := parts[3]
	action := parts[4]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	switch action {
	case "check":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pkg.LogStagef("update", "check for updates requested for %s", libName)
		writeToolActionResult(w, r, "is-info", fmt.Sprintf("%s availability checked", strings.ToUpper(libName)), "OpenSCADGen checked the configured local library support.", "")
	case "update":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pkg.LogStagef("update", "update requested for %s", libName)
		if libName != "bosl2" {
			http.Error(w, "library installation is not implemented for this library", http.StatusNotImplemented)
			return
		}
		if runtime.GOOS != "darwin" {
			writeToolActionResult(w, r, "is-warning", "BOSL2 installation is not supported on this platform", "BOSL2 installation is currently supported on macOS only.", "")
			return
		}
		if err := installBOSL2(r.Context()); err != nil {
			pkg.LogWarnf("BOSL2 installation failed: %v", err)
			writeToolActionResult(w, r, "is-danger", "BOSL2 installation failed", err.Error(), bosl2ManualInstallCommand)
			return
		}
		if err := pkg.RecordUpdateJournal(globalServerFolder, libName, "installed library", "BOSL2 installed", "downloaded official BOSL2 source archive", true); err != nil {
			pkg.LogWarnf("BOSL2 update journal write failed: %v", err)
		}
		writeToolActionResult(w, r, "is-success", "BOSL2 installed", "Downloaded and validated the official BOSL2 source archive. Reload OpenSCAD to use the updated library.", "")
	default:
		http.Error(w, "unsupported library action", http.StatusNotImplemented)
	}
}

func writeToolActionResult(w http.ResponseWriter, r *http.Request, colour, summary, details, manualCommand string) {
	summary = html.EscapeString(summary)
	details = html.EscapeString(details)
	if r.Header.Get("HX-Target") != "tools-results-content" {
		_, _ = fmt.Fprintf(w, `<span class="has-text-%s">%s</span>`, colour, summary)
		return
	}
	manualFix := ""
	if manualCommand != "" {
		manualFix = fmt.Sprintf(`<hr><h3 class="title is-5">Manual fix</h3><p class="mb-2">Run this in Terminal, then reload OpenSCAD:</p><div class="code-block"><button class="button is-small is-info" onclick="navigator.clipboard.writeText(this.parentElement.querySelector('pre code').textContent)">Copy command</button><pre><code>%s</code></pre></div>`, html.EscapeString(manualCommand))
	}
	_, _ = fmt.Fprintf(w, `<article class="message %s"><div class="message-header"><p>%s</p></div><div class="message-body"><pre style="white-space:pre-wrap;word-break:break-word;background:transparent;padding:0;margin:0">%s</pre>%s</div></article>`, colour, summary, details, manualFix)
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
		ReportURL:             pkg.BuildPageUrl(configPath, serverFolder).PageURL,
		Content:               content,
		ErrorMsg:              errorMsg,
	}
}

func buildConfigEditPreview(configPath string, serverFolder string, content string) (*models.Config, []models.InstanceConfig, []models.EditValidationFeedback, templ.Component) {
	config, warn, err := pkg.LoadConfig(content, models.CmdFlags{
		ConfigFile:   configPath,
		ServerFolder: serverFolder,
		Server:       true,
	}, configPath)
	if err != nil {
		return nil, nil, nil, templates.Warning(fmt.Sprintf("Invalid config:\n%s", err))
	}

	instances, instErr := pkg.GenerateInstanceConfigs(config)
	if instErr != nil {
		return config, nil, nil, templates.Warning(fmt.Sprintf("Failed to generate instances:\n%s", instErr))
	}

	var feedback []models.EditValidationFeedback
	if warn != nil {
		feedback = append(feedback, models.EditValidationFeedback{Message: warn.Error()})
	}

	if runErrors := pkg.ValidateInstances(instances, config); len(runErrors) > 0 {
		for _, runErr := range runErrors {
			feedback = append(feedback, models.EditValidationFeedback{
				Message:        runErr.Message,
				IsSaveBlocking: runErr.IsSaveBlocking,
			})
		}
	}

	return config, instances, feedback, nil
}

func previewConfigEditParams(configPath string, serverFolder string, content string) models.EditConfigParams {
	params := buildEditConfigParams(configPath, serverFolder, content, nil)
	config, instances, feedback, errorMsg := buildConfigEditPreview(configPath, serverFolder, content)
	params.ErrorMsg = errorMsg
	params.PreviewInstances = instances
	params.PreviewFeedback = feedback
	params.IsPreview = errorMsg == nil
	params.PreviewSections, params.PreviewSummary = buildEditPreviewSections(config, instances)
	if errorMsg != nil {
		params.SaveBlocked = true
	}
	for _, item := range feedback {
		if item.IsSaveBlocking {
			params.SaveBlocked = true
			break
		}
	}
	return params
}

func buildEditPreviewSections(config *models.Config, instances []models.InstanceConfig) ([]models.EditPreviewSection, []models.EditPreviewSummary) {
	if config == nil {
		return nil, nil
	}

	summary := []models.EditPreviewSummary{
		{Label: "Instances", Value: fmt.Sprintf("%d", len(config.Design.ConfiguredInstanceConfig))},
		{Label: "Param sets", Value: fmt.Sprintf("%d", len(config.Design.ParamSets))},
		{Label: "Images", Value: fmt.Sprintf("%d", len(config.Design.ExportImages))},
		{Label: "Input paths", Value: fmt.Sprintf("%d", len(config.GetInputPaths()))},
	}

	sections := []models.EditPreviewSection{
		{
			Title:       "export_name_format",
			Description: "Template used to name generated exports and detect duplicates.",
			Items: []models.EditPreviewSectionItem{{
				Title: func() string {
					if config.Design.ExportNameFormat != "" {
						return config.Design.ExportNameFormat
					}
					return "No export_name_format set"
				}(),
				Subtitle:    "current value",
				Description: "This drives output naming and should cover every varying parameter.",
			}},
		},
		{
			Title:       "instances",
			Description: "Configured instance blocks that expand into generated outputs.",
			CountLabel:  fmt.Sprintf("%d configured block(s)", len(config.Design.ConfiguredInstanceConfig)),
			HasItems:    len(config.Design.ConfiguredInstanceConfig) > 0,
		},
		{
			Title:       "param_sets",
			Description: "Reusable parameter bundles that instances and input paths can reference.",
			CountLabel:  fmt.Sprintf("%d defined", len(config.Design.ParamSets)),
			HasItems:    len(config.Design.ParamSets) > 0,
		},
		{
			Title:       "images",
			Description: "Top-level preview cameras applied to the generated instances.",
			CountLabel:  fmt.Sprintf("%d defined", len(config.Design.ExportImages)),
			HasItems:    len(config.Design.ExportImages) > 0,
		},
		{
			Title:       "global_params",
			Description: "Parameters merged into every generated instance by default.",
			CountLabel:  fmt.Sprintf("%d defined", len(config.Design.GlobalParams)),
			HasItems:    len(config.Design.GlobalParams) > 0,
		},
		{
			Title:       "input_paths",
			Description: "Source OpenSCAD files or path groups used to generate instances.",
			CountLabel:  fmt.Sprintf("%d defined", len(config.GetInputPaths())),
			HasItems:    len(config.GetInputPaths()) > 0,
		},
	}

	for i := range sections {
		switch sections[i].Title {
		case "export_name_format":
			sections[i].Controls = []models.EditPreviewControl{{
				Label:       "export_name_format",
				Path:        "[openscadgen].export_name_format",
				Example:     `export_name_format = "{designFileName}_{paramSet}"`,
				Description: "Template used to name generated exports and detect duplicates.",
			}}
		case "instances":
			sections[i].Controls = []models.EditPreviewControl{
				{Label: "instances", Path: "[openscadgen].instances", Example: `[[openscadgen.instances]]\nname = "default"\nparams = { width = 100 }`, Description: "Add another instance block."},
				{Label: "instance.name", Path: "[openscadgen.instances].name", Example: `name = "left"`, Description: "Human-readable instance name."},
				{Label: "instance.params", Path: "[openscadgen.instances].params", Example: `params = { side = "left" }`, Description: "Per-instance parameters."},
				{Label: "instance.param_sets", Path: "[openscadgen.instances].param_sets", Example: `param_sets = "wide,tall"`, Description: "Reuse named param sets for an instance."},
				{Label: "instance.images", Path: "[openscadgen.instances].images", Example: `[[openscadgen.instances.images]]\nname = "front"`, Description: "Attach instance-specific image exports."},
			}
		case "param_sets":
			sections[i].Controls = []models.EditPreviewControl{
				{Label: "param_sets", Path: "[openscadgen].param_sets", Example: `[[openscadgen.param_sets]]\nname = "wide"\nparams = { width = 120 }`, Description: "Add another reusable parameter set."},
				{Label: "param_set.name", Path: "[openscadgen.param_sets].name", Example: `name = "wide"`, Description: "Name of the reusable bundle."},
				{Label: "param_set.params", Path: "[openscadgen.param_sets].params", Example: `params = { width = 120 }`, Description: "Values included in the bundle."},
			}
		case "images":
			sections[i].Controls = []models.EditPreviewControl{
				{Label: "images", Path: "[openscadgen].images", Example: `[[openscadgen.images]]\nname = "front"\ncoord = "0,0,0"`, Description: "Add a top-level preview image camera."},
				{Label: "image.name", Path: "[openscadgen.images].name", Example: `name = "front"`, Description: "Name of the image preset."},
				{Label: "image.coord", Path: "[openscadgen.images].coord", Example: `coord = "0,0,0"`, Description: "Camera coordinates or preset direction."},
				{Label: "image.image_size", Path: "[openscadgen.images].image_size", Example: `image_size = "1200x1200"`, Description: "Optional per-camera image size."},
				{Label: "image.param_filter", Path: "[openscadgen.images].param_filter", Example: `param_filter = { finish = "matte" }`, Description: "Filter images to matching params only."},
				{Label: "preset directions", Path: "[openscadgen].images", Example: `top, front, back, left, right, nice`, Description: "Use presets like top / front / back / side / nice, or a custom 7-value camera string."},
			}
		case "global_params":
			sections[i].Controls = []models.EditPreviewControl{
				{Label: "global_params", Path: "[openscadgen].global_params", Example: `global_params = { wall = 2.4 }`, Description: "Add global parameters shared by every instance."},
			}
		case "input_paths":
			sections[i].Controls = []models.EditPreviewControl{
				{Label: "input_paths", Path: "[openscadgen].input_paths", Example: `[[openscadgen.input_paths]]\npath = "./part.scad"`, Description: "Add another input source."},
				{Label: "input_path.path", Path: "[openscadgen.input_paths].path", Example: `path = "./part.scad"`, Description: "Path to an OpenSCAD source file."},
				{Label: "input_path.export_name_format", Path: "[openscadgen.input_paths].export_name_format", Example: `export_name_format = "{designFileName}_{paramSet}"`, Description: "Override naming for just this input path."},
			}
		}

		switch sections[i].Title {
		case "instances":
			items := make([]models.EditPreviewSectionItem, 0, len(config.Design.ConfiguredInstanceConfig))
			for _, inst := range config.Design.ConfiguredInstanceConfig {
				items = append(items, models.EditPreviewSectionItem{
					Title:       inst.Name,
					Subtitle:    fmt.Sprintf("param_sets=%s skip_images=%v", inst.ParamSets, inst.SkipImages),
					Description: inst.Description,
				})
			}
			sections[i].Items = items
		case "param_sets":
			items := make([]models.EditPreviewSectionItem, 0, len(config.Design.ParamSets))
			for _, set := range config.Design.ParamSets {
				items = append(items, models.EditPreviewSectionItem{
					Title:       set.Name,
					Subtitle:    fmt.Sprintf("%d param(s)", len(set.Params)),
					Description: "",
				})
			}
			sections[i].Items = items
		case "images":
			items := make([]models.EditPreviewSectionItem, 0, len(config.Design.ExportImages))
			for _, img := range config.Design.ExportImages {
				items = append(items, models.EditPreviewSectionItem{
					Title:       img.CameraName,
					Subtitle:    img.CameraCoordinates,
					Description: img.ImageSize,
				})
			}
			sections[i].Items = items
		case "global_params":
			items := make([]models.EditPreviewSectionItem, 0, len(config.Design.GlobalParams))
			for name, value := range config.Design.GlobalParams {
				items = append(items, models.EditPreviewSectionItem{
					Title:       name,
					Subtitle:    fmt.Sprintf("%T", value),
					Description: fmt.Sprintf("%v", value),
				})
			}
			sections[i].Items = items
		case "input_paths":
			items := make([]models.EditPreviewSectionItem, 0, len(config.GetInputPaths()))
			for _, inputPath := range config.GetInputPaths() {
				items = append(items, models.EditPreviewSectionItem{
					Title:       inputPath.Path,
					Subtitle:    inputPath.ExportNameFormat,
					Description: inputPath.ParamSets,
				})
			}
			sections[i].Items = items
		case "export_name_format":
			sections[i].Items = []models.EditPreviewSectionItem{{
				Title:       config.Design.ExportNameFormat,
				Subtitle:    "format string",
				Description: "Used for duplicate-path detection and export naming.",
			}}
		}
		if len(sections[i].Items) == 0 {
			sections[i].Items = []models.EditPreviewSectionItem{{
				Title:       "No entries yet",
				Description: "This section can grow into a configurable block with a + control.",
			}}
		}
	}

	if len(instances) > 0 {
		sections = append(sections, models.EditPreviewSection{
			Title:       "generated_instances",
			Description: "The concrete instances produced by the current config.",
			CountLabel:  fmt.Sprintf("%d generated", len(instances)),
			HasItems:    true,
			Items: func() []models.EditPreviewSectionItem {
				items := make([]models.EditPreviewSectionItem, 0, len(instances))
				for _, instance := range instances {
					items = append(items, models.EditPreviewSectionItem{
						Title:       instance.AutoName,
						Subtitle:    instance.Name,
						Description: instance.Description,
					})
				}
				return items
			}(),
		})
	}

	return sections, summary
}

func renderTOMLEditPage(w http.ResponseWriter, r *http.Request, statusCode int, configPath string, serverFolder string, content string, message templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if statusCode != 0 {
		w.WriteHeader(statusCode)
	}
	configParams := buildEditConfigParams(configPath, serverFolder, content, message)
	if r.Header.Get("HX-Request") == "true" {
		templates.EditConfigPanel(&configParams).Render(r.Context(), w)
		return
	}
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

	params := previewConfigEditParams(configPath, serverFolder, string(content))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Header.Get("HX-Request") == "true" {
		templates.EditConfigPanel(&params).Render(r.Context(), w)
		return
	}
	templates.TOMLEditForm(&params).Render(r.Context(), w)
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
	if journalErr := pkg.RecordUpdateJournal(
		filepath.Dir(path),
		filepath.Base(path),
		"config updated in editor",
		"config saved",
		fmt.Sprintf("path=%s server_folder=%s", path, serverFolder),
		true,
	); journalErr != nil {
		pkg.LogWarnf("update journal write failed: %v", journalErr)
	}

	success := templates.Success("Config saved successfully")
	configParams := buildEditConfigParams(configPath, serverFolder, configContent, success)
	instances, instErr := pkg.GenerateInstanceConfigs(&testConfig)
	if instErr == nil {
		countLabel := fmt.Sprintf("(%d)", len(instances))
		configParams.InstanceCountLabel = countLabel
	}
	if r.Header.Get("HX-Request") == "true" {
		if configParams.InstanceCountLabel != "" {
			_, _ = w.Write([]byte(fmt.Sprintf(`<span id="instances-count-top" hx-swap-oob="true">%s</span><span id="instances-count-bottom" hx-swap-oob="true">%s</span>`, configParams.InstanceCountLabel, configParams.InstanceCountLabel)))
		}
		templates.EditConfigPanel(&configParams).Render(r.Context(), w)
		return
	}
	templates.TOMLEditForm(&configParams).Render(r.Context(), w)

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

func decodeBase64QueryValue(r *http.Request, name string) (string, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return "", fmt.Errorf("missing %q query parameter", name)
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("invalid %q query parameter", name)
	}
	return string(decoded), nil
}

func pathWithinDir(path, dir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		absPath = resolvedPath
	}
	resolvedDir, err := filepath.EvalSymlinks(absDir)
	if err == nil {
		absDir = resolvedDir
	}
	rel, err := filepath.Rel(absDir, absPath)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func openOpenSCADFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-a", "OpenSCAD", path)
	case "linux":
		cmd = exec.Command("openscad", path)
	case "windows":
		cmd = exec.Command("OpenSCAD", path)
	default:
		return fmt.Errorf("unsupported operating system")
	}
	return cmd.Start()
}

// handleOpenSCADFile opens a configured SCAD source in OpenSCAD. Both the config and
// source file must resolve within the active server folder, and the source must be
// explicitly referenced by that config.
func handleOpenSCADFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	configRef, err := decodeBase64QueryValue(r, "config_path")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	serverFolder, err := decodeBase64QueryValue(r, "server_folder")
	if err != nil || strings.TrimSpace(serverFolder) == "" {
		http.Error(w, "a server folder is required to open a SCAD file", http.StatusBadRequest)
		return
	}
	sourceRef, err := decodeBase64QueryValue(r, "source_path")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configPath := resolveConfigPath(configRef, serverFolder)
	if !pathWithinDir(configPath, serverFolder) {
		http.Error(w, "config file must be inside the server folder", http.StatusForbidden)
		return
	}
	config, _, err := pkg.LoadConfigFromFile(models.CmdFlags{ConfigFile: configPath, ServerFolder: serverFolder, Server: true})
	if err != nil {
		http.Error(w, fmt.Sprintf("load config: %v", err), http.StatusBadRequest)
		return
	}

	configured := false
	for _, input := range config.GetInputPaths() {
		if strings.EqualFold(filepath.Ext(input.Path), ".scad") && filepath.Clean(input.Path) == filepath.Clean(sourceRef) {
			configured = true
			break
		}
	}
	if !configured {
		http.Error(w, "source file is not a configured .scad input", http.StatusForbidden)
		return
	}

	sourcePath := sourceRef
	if !filepath.IsAbs(sourcePath) {
		sourcePath = filepath.Join(filepath.Dir(configPath), sourcePath)
	}
	if !pathWithinDir(sourcePath, serverFolder) {
		http.Error(w, "source file must be inside the server folder", http.StatusForbidden)
		return
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		http.Error(w, "source file not found", http.StatusNotFound)
		return
	}
	if info.IsDir() {
		http.Error(w, "source path must be a file", http.StatusBadRequest)
		return
	}
	if err := startOpenSCADFile(sourcePath); err != nil {
		http.Error(w, fmt.Sprintf("open in OpenSCAD: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.Success(fmt.Sprintf("Opened %s in OpenSCAD", filepath.Base(sourcePath))).Render(r.Context(), w)
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
	action := strings.ToLower(strings.TrimSpace(r.FormValue("action")))
	if content == "" {
		error := templates.Warning("Content cannot be empty")
		// Show form with error
		configParams := buildEditConfigParams(filePath, serverFolder, content, error)
		editForm := templates.TOMLEditForm(&configParams)
		editForm.Render(r.Context(), w)
		return
	}

	configParams := previewConfigEditParams(filePath, serverFolder, content)
	if action == "recheck" || action == "preview" || action == "" || configParams.SaveBlocked {
		templates.EditConfigPanel(&configParams).Render(r.Context(), w)
		return
	}

	// Write the content to the file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		errorMsg := templates.Warning(fmt.Sprintf("Failed to save file: %v", err))
		// Show form with save error
		configParams := buildEditConfigParams(filePath, serverFolder, content, errorMsg)
		editForm := templates.TOMLEditForm(&configParams)
		editForm.Render(r.Context(), w)
		return
	}
	success := templates.Success("File saved successfully")
	configParams = previewConfigEditParams(filePath, serverFolder, content)
	configParams.ErrorMsg = success
	editForm := templates.EditConfigPanel(&configParams)
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

	pkg.LogInfof("Preview request for instance: %s", instanceID)

	// Load config and find the instance to get the STL file path
	pkg.LogInfof("Checking config file: %s (encoded as: %s)", filePath, filePathEncoded)
	config, _, err := pkg.LoadConfigFromFile(models.CmdFlags{ConfigFile: filePath, Server: true, ServerFolder: globalServerFolder})
	if err != nil {
		pkg.LogWarnf("handlePreviewRequest Error loading config %s: %v", filePath, err)
		warning := templates.Warning(fmt.Sprintf("handlePreviewRequest Error loading config %s: %v", filePath, err))
		warning.Render(r.Context(), w)
		return
	}

	instances, err := pkg.GenerateInstanceConfigs(config)
	if err != nil {
		pkg.LogWarnf("Error generating instances for %s: %v", filePath, err)
		http.Error(w, "Error generating instances", http.StatusInternalServerError)
		return
	}

	pkg.LogInfof("Found %d instances in %s", len(instances), filePath)
	var targetInstance *models.InstanceConfig
	for _, instance := range instances {
		pkg.LogDebugf("Checking instance: %s", instance.UniqueID)
		if instance.UniqueID == instanceID {
			targetInstance = &instance
			pkg.LogInfof("Found matching instance: %s, STL path: %s", instance.UniqueID, instance.RunOutputPathV3)
			break
		}
	}

	if targetInstance == nil {
		pkg.LogWarnf("Instance not found: %s", instanceID)
		http.Error(w, fmt.Sprintf("Instance not found: %s", instanceID), http.StatusNotFound)
		return
	}

	// Check if STL file exists
	if _, err := os.Stat(targetInstance.RunOutputPathV3); os.IsNotExist(err) {
		pkg.LogWarnf("STL file not found: %s", targetInstance.RunOutputPathV3)
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

	pkg.LogInfof("Using STL file path: %s", targetInstance.RunOutputPathV3)
	pkg.LogDebugf("Encoded STL path: %s", encodedSTLPath)

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
		pkg.LogDebugf("handleSTLRequest - filePathEncoded: %s", filePathEncoded)
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

	pkg.LogInfof("STL handler - filePath: %s", filePath)

	// Check if STL file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		pkg.LogWarnf("STL file not found: %s", filePath)
		http.Error(w, "STL file not found", http.StatusNotFound)
		return
	}

	// Check file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		pkg.LogWarnf("Error getting file info: %v", err)
		http.Error(w, "Error accessing file", http.StatusInternalServerError)
		return
	}

	pkg.LogInfof("STL file size: %d bytes", fileInfo.Size())

	// Set headers for STL file
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Serve the STL file
	pkg.LogInfof("Serving STL file: %s", filePath)
	http.ServeFile(w, r, filePath)
}
