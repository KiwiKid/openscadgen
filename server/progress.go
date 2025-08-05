package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

var (
	progressMap        = make(map[string]chan string)
	cancelMap          = make(map[string]chan struct{})
	resultMap          = make(map[string]models.ProcessResult)
	configMap          = make(map[string]*models.Config)
	instanceResultsMap = make(map[string][]models.InstanceConfig) // Track individual instance results
	mu                 sync.Mutex
)

// HTMLProgressReporter implements ProgressReporter to send HTML updates
type HTMLProgressReporter struct {
	updates           chan<- string
	config            *models.Config
	jobID             string
	instances         []models.InstanceConfig
	allParamNames     []string
	outputPath        string
	instanceIndex     int
	currentInstanceID string // To track which instance is currently being processed
}

func NewHTMLProgressReporter(updates chan<- string, config *models.Config, jobID string) *HTMLProgressReporter {
	return &HTMLProgressReporter{
		updates: updates,
		config:  config,
		jobID:   jobID,
	}
}

func (h *HTMLProgressReporter) Update(msg string) {
	// Send terminal-style update
	h.updates <- msg
}

func (h *HTMLProgressReporter) Done() {
	h.updates <- "done"
}

func (h *HTMLProgressReporter) Error(err error) {
	h.updates <- "error: " + err.Error()
}

func (h *HTMLProgressReporter) Construct(instances []models.InstanceConfig, nonSkippedInstances int) {
	h.instances = instances

	// Calculate output path similar to pkg.getOutputPaths
	configDir := filepath.Dir(h.config.ConfigFile)
	versionPath := h.config.Design.Version
	var designName string
	if len(h.config.GetInputPaths()) > 0 {
		designName = strings.TrimSuffix(filepath.Base(h.config.GetInputPaths()[0].Path), ".scad")
	} else {
		designName = "test_design"
	}

	exportNameFormat := h.config.Design.ExportNameFormat
	hasExportPrefix := strings.HasPrefix(exportNameFormat, "export/") || strings.HasPrefix(exportNameFormat, "/export")

	var baseDir string
	if len(h.config.GetInputPaths()) > 0 {
		inputPath := h.config.GetInputPaths()[0].Path
		if filepath.IsAbs(inputPath) {
			baseDir = filepath.Dir(inputPath)
		} else {
			baseDir = configDir
		}
	} else {
		baseDir = configDir
	}

	var baseExportPath string
	if hasExportPrefix {
		baseExportPath = baseDir
	} else {
		baseExportPath = filepath.Join(baseDir, "export", versionPath, designName)
	}

	h.outputPath = filepath.Join(baseExportPath, "export", versionPath)

	// Extract all parameter names from instances
	paramSet := make(map[string]bool)
	for _, instance := range instances {
		for paramName := range instance.Params {
			paramSet[paramName] = true
		}
	}

	h.allParamNames = make([]string, 0, len(paramSet))
	for paramName := range paramSet {
		h.allParamNames = append(h.allParamNames, paramName)
	}

	h.updates <- fmt.Sprintf("Constructed progress for %d instances", len(instances))
}

func (h *HTMLProgressReporter) StartInstance(instanceId string) {
	h.currentInstanceID = instanceId
	log.Printf("HTMLProgressReporter: Starting instance: %s", instanceId)
	h.updates <- fmt.Sprintf("Starting instance: %s", instanceId)
}

func (h *HTMLProgressReporter) FinishInstance() {
	log.Printf("HTMLProgressReporter: FinishInstance called for instance: %s", h.currentInstanceID)

	// Find the completed instance
	var completedInstance *models.InstanceConfig
	for i := range h.instances {
		if h.instances[i].ID == h.currentInstanceID {
			completedInstance = &h.instances[i]
			break
		}
	}

	if completedInstance != nil {
		log.Printf("HTMLProgressReporter: Found completed instance: %s", completedInstance.Name)

		// Send terminal-style update first
		h.updates <- fmt.Sprintf("Instance complete: %s", completedInstance.Name)

		// Generate HTML card for this instance using templ component
		var htmlCard strings.Builder
		templates.InstanceCard(*completedInstance, h.outputPath).Render(context.Background(), &htmlCard)

		// Send HTML update
		htmlUpdate := "html:" + htmlCard.String()
		log.Printf("HTMLProgressReporter: Sending HTML update for instance: %s", completedInstance.Name)
		h.updates <- htmlUpdate

		// Store the completed instance
		mu.Lock()
		if instanceResultsMap[h.jobID] == nil {
			instanceResultsMap[h.jobID] = make([]models.InstanceConfig, 0)
		}
		instanceResultsMap[h.jobID] = append(instanceResultsMap[h.jobID], *completedInstance)
		mu.Unlock()

		h.instanceIndex++

		// Add a small delay to make updates visible
		time.Sleep(500 * time.Millisecond)
	} else {
		log.Printf("HTMLProgressReporter: Could not find completed instance for ID: %s", h.currentInstanceID)
	}

	h.updates <- "Instance complete"
}

// StartHandler handles starting a new processing job
func StartHandler(w http.ResponseWriter, r *http.Request) {
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

	// Store config for later use
	mu.Lock()
	configMap[id] = config
	mu.Unlock()

	go func() {
		progressReporter := NewHTMLProgressReporter(updates, config, id)
		result, err := pkg.Process(config, progressReporter, cancel)
		if err != nil {
			updates <- "error: " + err.Error()
		} else {
			// Store the result for later use
			mu.Lock()
			resultMap[id] = result
			mu.Unlock()
		}
		close(updates)
	}()

	// Return the ID to the client, which will poll /progress?id=...
	w.Write([]byte(id))
}

// ProgressHandler handles progress updates for a job
func ProgressHandler(w http.ResponseWriter, r *http.Request) {
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
			// Processing is done, return the complete report
			mu.Lock()
			result, hasResult := resultMap[id]
			config, hasConfig := configMap[id]
			mu.Unlock()

			if hasResult && hasConfig {
				// Generate the complete report
				report := templates.Report("complete", config, result.Instances, result.ExportLocation, result.STLResults, result.ImageResults, []string{}, true, config.ConfigFile, result.TotalTimeTaken)
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "complete")
				report.Render(context.Background(), w)

				// Clean up
				mu.Lock()
				delete(resultMap, id)
				delete(configMap, id)
				delete(progressMap, id)
				delete(cancelMap, id)
				delete(instanceResultsMap, id)
				mu.Unlock()
			} else {
				w.Header().Set("X-Progress-Status", "done")
				w.Write([]byte("done"))
			}
			return
		}

		// Check if this is an HTML update
		if strings.HasPrefix(msg, "html:") {
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("X-Progress-Status", "html")
			w.Write([]byte(strings.TrimPrefix(msg, "html:")))
		} else {
			w.Header().Set("X-Progress-Status", "progress")
			w.Write([]byte(msg))
		}
	case <-time.After(2 * time.Second):
		w.Header().Set("X-Progress-Status", "waiting")
		w.Write([]byte("waiting"))
	}
}

// CancelHandler handles cancellation of a job
func CancelHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	mu.Lock()
	cancel, ok := cancelMap[id]
	if ok {
		close(cancel)
		delete(cancelMap, id)
		delete(progressMap, id)
		delete(resultMap, id)
		delete(configMap, id)
		delete(instanceResultsMap, id)
	}
	mu.Unlock()
	w.Write([]byte("cancelled"))
}

// StartProcessingJob starts a new processing job and returns the job ID
func StartProcessingJob(config *models.Config) string {
	id := uuid.New().String()
	updates := make(chan string, 10)
	cancel := make(chan struct{})
	mu.Lock()
	progressMap[id] = updates
	cancelMap[id] = cancel
	configMap[id] = config
	mu.Unlock()

	go func() {
		progressReporter := NewHTMLProgressReporter(updates, config, id)
		result, err := pkg.Process(config, progressReporter, cancel)
		if err != nil {
			log.Printf("Processing error: %v", err)
		} else {
			// Store the result for later use
			mu.Lock()
			resultMap[id] = result
			mu.Unlock()
		}
		close(updates)
	}()

	return id
}

// GetProgressHTML returns HTML for progress tracking
func GetProgressHTML(id string) string {
	return fmt.Sprintf(`
		<div id="progress" class="notification is-info"></div>
		<div id="instances-container">
			<div class="columns is-multiline" id="instances-grid">
				<!-- Instances will be added here via HTMX -->
			</div>
		</div>
		<button class="button is-danger" onclick="fetch('/cancel?id=%s')">Cancel</button>
		<script>
		function poll() {
			fetch('/progress?id=%s').then(r => {
				const status = r.headers.get('X-Progress-Status');
				return r.text().then(msg => ({ status, msg }));
			}).then(({ status, msg }) => {
				if(status === 'complete') {
					// Complete report received, replace the entire page content
					document.documentElement.innerHTML = msg;
				} else if(status === 'html') {
					// HTML update received, append to instances grid
					const grid = document.querySelector('#instances-grid');
					grid.insertAdjacentHTML('beforeend', msg);
					// Continue polling
					setTimeout(poll, 1000);
				} else {
					document.getElementById('progress').innerText = msg;
					if(status !== 'done' && status !== 'cancelled' && !msg.startsWith('error:')) {
						setTimeout(poll, 1000);
					}
				}
			}).catch(err => {
				console.error('Progress polling error:', err);
				setTimeout(poll, 1000);
			});
		}
		poll();
		</script>
	`, id, id)
}
