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
	h.updates <- "done 1"
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

func (h *HTMLProgressReporter) StartInstance(instanceId string, name string) {
	h.currentInstanceID = instanceId
	log.Printf("HTMLProgressReporter: Starting instance: %s", instanceId)
	h.updates <- fmt.Sprintf("Starting instance: %s", name)
}

func (h *HTMLProgressReporter) FinishInstance() {
	log.Printf("HTMLProgressReporter: FinishInstance called for instance: %s", h.currentInstanceID)

	// Only process if we have a current instance ID (instance was actually started)
	if h.currentInstanceID == "" {
		log.Printf("HTMLProgressReporter: No current instance ID, skipping FinishInstance")
		return
	}

	// Find the completed instance
	var completedInstance *models.InstanceConfig
	for i := range h.instances {
		if h.instances[i].ID == h.currentInstanceID {
			completedInstance = &h.instances[i]
			log.Printf("HTMLProgressReporter: Found matching instance: %s (ID: %s)", completedInstance.AutoName, completedInstance.ID)
			break
		}
	}

	if completedInstance != nil {
		log.Printf("HTMLProgressReporter: Found completed instance: %s", completedInstance.AutoName)

		// Send terminal-style update first
		//	h.updates <- fmt.Sprintf("Instance complete: %s", completedInstance.AutoName)

		// Generate HTML card for this instance using templ component
		var htmlCard strings.Builder
		// templates.InstanceCard(*completedInstance, h.outputPath).Render(context.Background(), &htmlCard)

		templates.InstanceCardV2(*completedInstance, h.outputPath, "complete", h.allParamNames, true, "").Render(context.Background(), &htmlCard)

		// Send HTML update
		htmlUpdate := "html:" + htmlCard.String()
		log.Printf("HTMLProgressReporter: Sending HTML update for instance: %s", completedInstance.AutoName)
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

		// Clear the current instance ID to prevent reuse
		h.currentInstanceID = ""
	} else {
		log.Printf("HTMLProgressReporter: Could not find completed instance for ID: %s", h.currentInstanceID)
		// Clear the current instance ID even if not found
		h.currentInstanceID = ""
	}
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
		result, err := pkg.Process(config, progressReporter, cancel, pkg.Operations{
			GenerateReport: false,
		}, true)
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
		success := templates.Success("Progress completed")
		success.Render(context.Background(), w)
		return
	}
	select {
	case msg, ok := <-updates:
		if !ok {
			// Processing is done, just return "done"
			w.Header().Set("X-Progress-Status", "done")
			done := templates.AllComplete()
			done.Render(context.Background(), w)

			// Clean up
			mu.Lock()
			delete(resultMap, id)
			delete(configMap, id)
			delete(progressMap, id)
			delete(cancelMap, id)
			delete(instanceResultsMap, id)
			mu.Unlock()
			return
		}

		// Check if this is an HTML update
		if strings.HasPrefix(msg, "html:") {
			// For HTML updates (instance completions), return HTMX OOB update
			htmlContent := strings.TrimPrefix(msg, "html:")

			// Get the current progress message (non-blocking)
			var progressMsg string
			select {
			case progressUpdate := <-updates:
				if strings.HasPrefix(progressUpdate, "html:") {
					// Another HTML update, combine them
					htmlContent += strings.TrimPrefix(progressUpdate, "html:")
					progressMsg = "Instance completed"
				} else {
					progressMsg = progressUpdate
				}
			default:
				progressMsg = "Instance completed"
			}

			// Check if processing is complete by checking if we have a result
			mu.Lock()
			result, hasResult := resultMap[id]
			mu.Unlock()

			if hasResult {
				allCompleteHTML := templates.AllComplete()
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "complete")
				allCompleteHTML.Render(context.Background(), w)
				templates.ProcessForm(result.ConfigFile).Render(context.Background(), w)
			} else {
				// Return progress div with HTMX attributes and HTML content as OOB update
				combinedHTML := fmt.Sprintf(`<div id="progress" hx-get="/progress?id=%s" hx-trigger="every 1s" hx-swap="outerHTML" class="notification is-info">%s</div> %s
`, id, progressMsg, htmlContent)
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "html")
				w.Write([]byte(combinedHTML))
			}
		} else {
			// Return HTMX-compatible progress update with hx-get for next poll
			progressHTML := fmt.Sprintf(`<div id="progress" hx-get="/progress?id=%s" hx-trigger="every 1s" hx-swap="outerHTML" class="notification is-info">%s</div>`, id, msg)
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("X-Progress-Status", "progress")
			w.Write([]byte(progressHTML))
		}
	}
	/*case <-time.After(120 * time.Second):
		// Return HTMX-compatible waiting update with hx-get for next poll
		waitingHTML := fmt.Sprintf(`<div id="progress" hx-get="/progress?id=%s" hx-trigger="every 1s" hx-swap="outerHTML" class="notification is-warning">waiting</div>`, id)
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Progress-Status", fmt.Sprintf("(hmm, its been over 2 minutes, still waiting for %s..)", msg))
		w.Write([]byte(waitingHTML))
	}*/
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
		result, err := pkg.Process(config, progressReporter, cancel, pkg.Operations{
			GenerateReport: false,
		}, true)
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
		<div id="progress" class="notification is-info"></div>[[GetProgressHTML]]

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
