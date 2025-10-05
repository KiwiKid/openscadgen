package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

var configFiles []models.ConfigFile
var fileWatcher *FileWatcher
var fileWatcherEnabled bool

// StartServer starts the HTTP server with the specified folder for config files
func StartServer(serverFolder string, cmdFlags models.CmdFlags) {
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

	msg += fmt.Sprintf("\n\n Running at http://localhost%s", port)
	log.Print(msg)

	// Setup handlers
	http.HandleFunc("/start", StartHandler)
	http.HandleFunc("/progress", ProgressHandler)
	http.HandleFunc("/cancel", CancelHandler)
	http.HandleFunc("/api/watcher/status", handleWatcherStatus)
	http.HandleFunc("/api/watcher/pause", handleWatcherPause)
	http.HandleFunc("/api/watcher/resume", handleWatcherResume)
	http.HandleFunc("/api/watcher/ui", handleWatcherUI)
	http.HandleFunc("/api/config", handleConfigRequest)
	http.HandleFunc("/api/open", handleOpenFile)
	http.HandleFunc("/api/edit", handleEditFile)
	http.HandleFunc("/images", handleImageRequest)
	http.HandleFunc("/static/", handleStaticFiles)
	http.HandleFunc("/", handleMainRequest)

	err = http.ListenAndServe(port, nil)
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
		log.Printf("Loading config for %s", configEntryPath)
		config, err := pkg.LoadConfig(models.CmdFlags{ConfigFile: configEntryPath, Server: true})
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("Error: %v", err))
			warning.Render(context.Background(), w)
			return
		}

		instances, err := pkg.GenerateInstanceConfigs(config)
		if err != nil {
			warning := templates.Warning(fmt.Sprintf("Error: %v", err))
			warning.Render(context.Background(), w)
			return
		}

		log.Printf("Generating report for %s with %d instances", configEntryPath, len(instances))

		report := templates.Report("view", config, instances, "", []models.GenerateSTLResult{}, []models.GenerateImageResult{}, []string{}, true, configEntryPath, 0)
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

	var useOOBUpdates bool
	if r.FormValue("regex") != "" {
		cmdFlags.RegexPattern = r.FormValue("regex")
		useOOBUpdates = true
	}

	config, err := pkg.LoadConfig(cmdFlags)
	if err != nil {
		warning := templates.Warning(fmt.Sprintf("Error: %v", err))
		warning.Render(context.Background(), w)
		return
	}

	id := StartProcessingJob(config)

	if useOOBUpdates {
		w.Header().Set("Content-Type", "text/html")
		instanceUpdates := templates.InstanceUpdate(id)
		instanceUpdates.Render(context.Background(), w)
	} else {
		w.Header().Set("Content-Type", "text/html")
		progressHTML := templates.GetProgressHTML(id)
		progressHTML.Render(context.Background(), w)
		//fmt.Fprintf(w, progressHTML)
	}
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

// handleConfigRequest handles GET and POST requests for config file operations
func handleConfigRequest(w http.ResponseWriter, r *http.Request) {
	configPath := r.URL.Query().Get("path")
	if configPath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	// Resolve absolute path
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to resolve config file path: %v", err), http.StatusBadRequest)
			return
		}
		configPath = absPath
	}

	switch r.Method {
	case "GET":
		handleConfigGet(w, r, configPath)
	case "POST":
		handleConfigPost(w, r, configPath)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleConfigGet reads and returns the config file content
func handleConfigGet(w http.ResponseWriter, r *http.Request, configPath string) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read config file: %v", err), http.StatusInternalServerError)
		return
	}

	form := templates.TOMLEditForm(configPath, string(content), nil)
	form.Render(r.Context(), w)
	return
}

// handleConfigPost validates TOML and updates the config file
func handleConfigPost(w http.ResponseWriter, r *http.Request, configPath string) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	configContent := r.FormValue("content")
	if configContent == "" {
		http.Error(w, "Missing 'content' form field", http.StatusBadRequest)
		return
	}

	// Validate TOML by attempting to decode it
	var testConfig models.Config
	_, err := toml.Decode(configContent, &testConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid TOML: %v", err), http.StatusBadRequest)
		return
	}

	// Write the content to the file
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write config file: %v", err), http.StatusInternalServerError)
		return
	}

	success := templates.Success("Config saved successfully")

	editForm := templates.TOMLEditForm(configPath, configContent, success)
	editForm.Render(r.Context(), w)

}

// handleImageRequest serves image files for the web interface
func handleImageRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	imagePath := r.URL.Query().Get("path")
	if imagePath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	// Resolve absolute path
	if !filepath.IsAbs(imagePath) {
		// If it's a relative path, try to resolve it from the current working directory
		absPath, err := filepath.Abs(imagePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to resolve image path: %v", err), http.StatusBadRequest)
			return
		}
		imagePath = absPath
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "Image file not found", http.StatusNotFound)
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
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Serve the file
	http.ServeFile(w, r, imagePath)
}

// handleStaticFiles serves static files from the static directory
func handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Serve the file
	http.ServeFile(w, r, absPath)
}

// handleOpenFile opens a file in the system's default editor
func handleOpenFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	// Resolve absolute path
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to resolve file path: %v", err), http.StatusBadRequest)
			return
		}
		filePath = absPath
	}

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
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}

	// Resolve absolute path
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to resolve file path: %v", err), http.StatusBadRequest)
			return
		}
		filePath = absPath
	}

	// Validate file extension
	if filepath.Ext(filePath) != ".toml" {
		http.Error(w, "Only .toml files are allowed", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		handleEditGet(w, r, filePath)
	case "POST":
		handleEditPost(w, r, filePath)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleEditGet loads and displays the TOML file in an editable form
func handleEditGet(w http.ResponseWriter, r *http.Request, filePath string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the edit form
	editForm := templates.TOMLEditForm(filePath, string(content), nil)
	editForm.Render(r.Context(), w)
}

// handleEditPost validates and saves the TOML file
func handleEditPost(w http.ResponseWriter, r *http.Request, filePath string) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		error := templates.Warning("Content cannot be empty")
		// Show form with error
		editForm := templates.TOMLEditForm(filePath, content, error)
		editForm.Render(r.Context(), w)
		return
	}

	// Validate TOML by attempting to decode it
	var testConfig models.Config
	_, err := toml.Decode(content, &testConfig)
	if err != nil {
		errorMsg := templates.Warning(fmt.Sprintf("Invalid TOML: %v", err))
		// Show form with validation error
		editForm := templates.TOMLEditForm(filePath, content, errorMsg)
		editForm.Render(r.Context(), w)
		return
	}

	// Write the content to the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		errorMsg := templates.Warning(fmt.Sprintf("Failed to save file: %v", err))
		// Show form with save error
		editForm := templates.TOMLEditForm(filePath, content, errorMsg)
		editForm.Render(r.Context(), w)
		return
	}

	// Success - redirect to view the file
	//http.Redirect(w, r, "/?config="+url.QueryEscape(filePath), http.StatusSeeOther)
	success := templates.Success("File saved successfully")
	editForm := templates.TOMLEditForm(filePath, content, success)
	editForm.Render(r.Context(), w)
}
