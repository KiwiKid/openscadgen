package server

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

// ConfigInfo tracks a loaded configuration and its metadata
type ConfigInfo struct {
	Path         string
	Config       *models.Config
	LastModified time.Time
	Instances    []models.InstanceConfig
}

// FileWatcher manages file watching and change detection
type FileWatcher struct {
	watcher      *fsnotify.Watcher
	configs      map[string]*ConfigInfo
	mu           sync.RWMutex
	stopChan     chan struct{}
	isWatching   bool
	serverFolder string
}

// NewFileWatcher creates a new file watcher instance
func NewFileWatcher() (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:    watcher,
		configs:    make(map[string]*ConfigInfo),
		stopChan:   make(chan struct{}),
		isWatching: false,
	}, nil
}

// StartWatching begins watching all config files in the specified folder
func (fw *FileWatcher) StartWatching(serverFolder string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.isWatching {
		return nil // Already watching
	}

	// Store server folder for later use
	fw.serverFolder = serverFolder

	// Load initial configs
	configFiles, err := pkg.ScanFolderForConfigFiles(serverFolder)
	if err != nil {
		return err
	}

	// Load and cache each config
	for _, configFile := range configFiles {
		// configFile.Path is relative to serverFolder, so join them
		fullConfigPath := filepath.Join(serverFolder, configFile.Path)
		if err := fw.loadAndWatchConfig(fullConfigPath, serverFolder); err != nil {
			pkg.LogWarnf("Error loading config %s: %v", fullConfigPath, err)
			continue
		}
	}

	// Process all loaded configs initially
	for _, configInfo := range fw.configs {
		pkg.LogInfof("Processing initial config: %s", configInfo.Path)
		StartProcessingJob(configInfo.Config)
	}

	// Start watching for changes
	go fw.watchLoop()
	fw.isWatching = true

	pkg.LogInfof("Started watching %d config files", len(fw.configs))
	return nil
}

// StopWatching stops the file watcher
func (fw *FileWatcher) StopWatching() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isWatching {
		return
	}

	close(fw.stopChan)
	fw.watcher.Close()
	fw.isWatching = false
	pkg.LogInfof("Stopped file watching")
}

// loadAndWatchConfig loads a config file and starts watching it
func (fw *FileWatcher) loadAndWatchConfig(configPath string, serverFolder string) error {
	// Get file info for modification time
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return err
	}

	// Load the config
	cmdFlags := models.CmdFlags{ConfigFile: configPath, Server: true, ServerFolder: serverFolder}
	config, _, err := pkg.LoadConfigFromFile(cmdFlags)
	if err != nil {
		return err
	}

	// Create config info
	configInfo := &ConfigInfo{
		Path:         configPath,
		Config:       config,
		LastModified: fileInfo.ModTime(),
		Instances:    []models.InstanceConfig{}, // Will be populated when instances are generated
	}

	// Add to watcher
	if err := fw.watcher.Add(filepath.Dir(configPath)); err != nil {
		return err
	}

	// Cache the config
	fw.configs[configPath] = configInfo
		pkg.LogInfof("Loaded and watching config: %s", configPath)

	return nil
}

// watchLoop handles file system events
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleFileEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			pkg.LogErrorf("File watcher error: %v", err)

		case <-fw.stopChan:
			return
		}
	}
}

// handleFileEvent processes a file system event
func (fw *FileWatcher) handleFileEvent(event fsnotify.Event) {
	// Only care about config.toml files
	if filepath.Base(event.Name) != "config.toml" {
		return
	}

	// Check if this is a config we're watching
	fw.mu.RLock()
	configInfo, exists := fw.configs[event.Name]
	fw.mu.RUnlock()

	if !exists {
		return
	}

	// Debounce rapid changes
	time.Sleep(100 * time.Millisecond)

	// Check if file was actually modified
	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		pkg.LogWarnf("Error checking file modification: %v", err)
		return
	}

	if fileInfo.ModTime().Equal(configInfo.LastModified) {
		return // No actual change
	}

	pkg.LogInfof("Config file changed: %s", event.Name)
	fw.handleConfigChange(event.Name)
}

// handleConfigChange processes a config file change
func (fw *FileWatcher) handleConfigChange(configPath string) {
	fw.mu.RLock()
	serverFolder := fw.serverFolder
	fw.mu.RUnlock()
	
	// Load the new config
	cmdFlags := models.CmdFlags{ConfigFile: configPath, Server: true, ServerFolder: serverFolder}
	newConfig, _, err := pkg.LoadConfigFromFile(cmdFlags)
	if err != nil {
		pkg.LogWarnf("Error loading changed config %s: %v", configPath, err)
		return
	}

	// Get file info for new modification time
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		pkg.LogWarnf("Error getting file info: %v", err)
		return
	}

	// Compare with old config
	fw.mu.Lock()
	oldConfigInfo := fw.configs[configPath]
	fw.mu.Unlock()

	if oldConfigInfo == nil {
		pkg.LogWarnf("No old config found for %s", configPath)
		return
	}

	// Update the cached config
	fw.mu.Lock()
	fw.configs[configPath] = &ConfigInfo{
		Path:         configPath,
		Config:       newConfig,
		LastModified: fileInfo.ModTime(),
		Instances:    []models.InstanceConfig{}, // Will be populated when instances are generated
	}
	fw.mu.Unlock()

	// Compare and log changes
	changes := fw.compareConfigs(oldConfigInfo.Config, newConfig)
	if len(changes) > 0 {
		pkg.LogInfof("Config changes detected in %s:", configPath)
		for _, change := range changes {
			pkg.LogInfof("  - %s", change)
		}
		// Trigger regeneration for the changed config
		pkg.LogStagef("watcher", "Starting processing for changed config: %s", configPath)
		StartProcessingJob(newConfig)
	} else {
		pkg.LogInfof("No meaningful changes detected in %s", configPath)
	}
}

// compareConfigs compares two configs and returns a list of changes
func (fw *FileWatcher) compareConfigs(oldConfig, newConfig *models.Config) []string {
	var changes []string

	// Compare basic config properties
	if oldConfig.Design.Name != newConfig.Design.Name {
		changes = append(changes, "Name changed")
	}

	if oldConfig.Design.Version != newConfig.Design.Version {
		changes = append(changes, "Version changed")
	}

	// Compare configured instances
	oldInstances := make(map[string]models.ConfiguredInstanceConfig)
	for _, instance := range oldConfig.Design.ConfiguredInstanceConfig {
		oldInstances[instance.Name] = instance
	}

	newInstances := make(map[string]models.ConfiguredInstanceConfig)
	for _, instance := range newConfig.Design.ConfiguredInstanceConfig {
		newInstances[instance.Name] = instance
	}

	// Check for added instances
	for name := range newInstances {
		if _, exists := oldInstances[name]; !exists {
			changes = append(changes, "Instance added: "+name)
		}
	}

	// Check for removed instances
	for name := range oldInstances {
		if _, exists := newInstances[name]; !exists {
			changes = append(changes, "Instance removed: "+name)
		}
	}

	// Check for modified instances
	for name, newInstance := range newInstances {
		if oldInstance, exists := oldInstances[name]; exists {
			if fw.configuredInstancesDiffer(oldInstance, newInstance) {
				changes = append(changes, "Instance modified: "+name)
			}
		}
	}

	return changes
}

// configuredInstancesDiffer compares two configured instances and returns true if they differ
func (fw *FileWatcher) configuredInstancesDiffer(old, new models.ConfiguredInstanceConfig) bool {
	// Compare basic properties
	if old.Name != new.Name {
		return true
	}

	if old.Description != new.Description {
		return true
	}

	// Compare parameters (simplified comparison)
	if len(old.Params) != len(new.Params) {
		return true
	}

	for key, oldValue := range old.Params {
		if newValue, exists := new.Params[key]; !exists || oldValue != newValue {
			return true
		}
	}

	// Compare param sets
	if old.ParamSets != new.ParamSets {
		return true
	}

	return false
}

// GetWatchedConfigs returns a copy of all watched configs
func (fw *FileWatcher) GetWatchedConfigs() map[string]*ConfigInfo {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	result := make(map[string]*ConfigInfo)
	for path, config := range fw.configs {
		result[path] = config
	}
	return result
}

// IsWatching returns whether the file watcher is currently active
func (fw *FileWatcher) IsWatching() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.isWatching
}
