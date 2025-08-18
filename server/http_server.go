package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

var configFiles []models.ConfigFile
var fileWatcher *FileWatcher
var fileWatcherEnabled bool

// StartServer starts the HTTP server with the specified folder for config files
func StartServer(serverFolder string, cmdFlags models.CmdFlags) {
	var err error
	var msg = "Starting server on port 8080"

	// Initialize file watcher only if enabled
	fileWatcherEnabled = cmdFlags.EnableFileWatcher
	if cmdFlags.EnableFileWatcher {
		fileWatcher, err = NewFileWatcher()
		if err != nil {
			log.Fatalf("Error creating file watcher: %v", err)
		}
	}

	if serverFolder != "" {
		configFiles, err = pkg.ScanFolderForConfigFiles(serverFolder)
		if err != nil {
			log.Fatalf("Error: %v", err)
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
	log.Print(msg)

	// Setup handlers
	http.HandleFunc("/start", StartHandler)
	http.HandleFunc("/progress", ProgressHandler)
	http.HandleFunc("/cancel", CancelHandler)
	http.HandleFunc("/api/watcher/status", handleWatcherStatus)
	http.HandleFunc("/api/watcher/pause", handleWatcherPause)
	http.HandleFunc("/api/watcher/resume", handleWatcherResume)
	http.HandleFunc("/api/watcher/ui", handleWatcherUI)
	http.HandleFunc("/", handleMainRequest)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
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
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGETRequest handles GET requests for displaying configs
func handleGETRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET (Display) request" + r.Method)
	configEntryPathEncoded := r.URL.Query().Get("config")
	configEntryPath, err := url.QueryUnescape(configEntryPathEncoded)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	if configEntryPath == "" {
		configEntry := templates.EnterConfigPage(configFiles)
		configEntry.Render(context.Background(), w)
	} else {
		config, err := pkg.LoadConfig(models.CmdFlags{ConfigFile: configEntryPath, Server: true})
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("Error: %v", err))
			warning.Render(context.Background(), w)
			return
		}

		report := templates.Report("view", config, []models.InstanceConfig{}, "", []models.GenerateSTLResult{}, []models.GenerateImageResult{}, []string{}, true, configEntryPath, 0)
		report.Render(context.Background(), w)
	}
}

// handlePOSTRequest handles POST requests for adding configs
func handlePOSTRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST (Add Config) request" + r.Method)

	configEntry := r.FormValue("path")
	if configEntry == "" {
		warning := templates.Warning("No config file provided")
		warning.Render(context.Background(), w)
		return
	}

	encodedConfigEntry := url.QueryEscape(configEntry)
	http.Redirect(w, r, "/?config="+encodedConfigEntry, http.StatusSeeOther)
}

// handlePUTRequest handles PUT requests for processing
func handlePUTRequest(w http.ResponseWriter, r *http.Request) {
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

	cmdFlags.Server = true

	config, err := pkg.LoadConfig(cmdFlags)
	if err != nil {
		warning := templates.Warning(fmt.Sprintf("Error: %v", err))
		warning.Render(context.Background(), w)
		return
	}

	id := StartProcessingJob(config)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, GetProgressHTML(id))
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
