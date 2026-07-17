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
	jobErrorMap        = make(map[string]string)
	mu                 sync.Mutex
)

func cleanupJobState(id string) {
	delete(resultMap, id)
	delete(configMap, id)
	delete(progressMap, id)
	delete(cancelMap, id)
	delete(instanceResultsMap, id)
	delete(jobErrorMap, id)
}

func trimProgressError(msg string) string {
	return strings.TrimSpace(strings.TrimPrefix(msg, "error:"))
}

// HTMLProgressReporter implements ProgressReporter to send HTML updates
type HTMLProgressReporter struct {
	updates           chan<- string
	config            *models.Config
	jobID             string
	instances         []models.InstanceConfig
	allParamNames     []string
	outputPath        string
	configInputPath   string
	instanceIndex     int
	currentInstanceID string // To track which instance is currently being processed
}

func NewHTMLProgressReporter(updates chan<- string, config *models.Config, jobID string, configInputPath string) *HTMLProgressReporter {
	return &HTMLProgressReporter{
		updates:         updates,
		config:          config,
		jobID:           jobID,
		configInputPath: configInputPath,
	}
}

func (h *HTMLProgressReporter) Update(msg string) {
	// Send terminal-style update
	h.updates <- msg
}

func (h *HTMLProgressReporter) Done() {
	h.updates <- "Complete"
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

	var instanceCount int
	paramSet := make(map[string]bool)
	for _, instance := range instances {
		for paramName := range instance.Params {
			paramSet[paramName] = true
		}
		if instance.SkippedReason == "" {
			instanceCount++
		}
	}

	h.allParamNames = make([]string, 0, len(paramSet))
	for paramName := range paramSet {
		h.allParamNames = append(h.allParamNames, paramName)
	}

	if instanceCount == 0 {
		if h.config.RegexPattern == "" {
			h.updates <- "No instances queued for processing"
		} else {
			h.updates <- fmt.Sprintf("%s - Did not match any instances", h.config.RegexPattern)
		}
	} else {

		h.updates <- fmt.Sprintf("Processing %d instances", instanceCount)
	}
}

func (h *HTMLProgressReporter) StartInstance(instanceId string, name string, instanceIndex int, instanceCount int) {
	h.currentInstanceID = instanceId
	pkg.LogStagef("progress", "Starting instance: %s", instanceId)

	h.updates <- fmt.Sprintf("Starting instance %s", name)
}

func (h *HTMLProgressReporter) FinishInstance() {
	pkg.LogStagef("progress", "FinishInstance called for instance: %s", h.currentInstanceID)

	// Only process if we have a current instance ID (instance was actually started)
	if h.currentInstanceID == "" {
		pkg.LogWarnWithCritical("progress: No current instance ID, skipping FinishInstance", false)
		return
	}

	// Find the completed instance
	var completedInstance *models.InstanceConfig
	for i := range h.instances {
		if h.instances[i].ID == h.currentInstanceID {
			completedInstance = &h.instances[i]
			pkg.LogInfof("progress: Found matching instance: %s (ID: %s)", completedInstance.AutoName, completedInstance.ID)
			break
		}
	}

	if completedInstance != nil {
		pkg.LogInfof("progress: Found completed instance: %s (ID: %s, UniqueID: %s)", completedInstance.AutoName, completedInstance.ID, completedInstance.UniqueID)

		// Send terminal-style update first
		//	h.updates <- fmt.Sprintf("Instance complete: %s", completedInstance.AutoName)

		// Generate HTML card for this instance using templ component
		var htmlCard strings.Builder
		// templates.InstanceCard(*completedInstance, h.outputPath).Render(context.Background(), &htmlCard)

		reportMeta := pkg.BuildReportMeta(models.BuildReportMetaParams{
			IsServerMode:         true,
			ConfigFilePath:       h.config.ConfigFile,
			ServerFolder:         h.config.ServerFolder,
			Instances:            h.instances,
			TotalQueuedInstances: h.config.TotalQueuedInstances,
		}, models.Results{
			TimeTake: 0,
		})

		templates.InstanceCardV2(*completedInstance, h.outputPath, "complete", h.allParamNames, reportMeta).Render(context.Background(), &htmlCard)

		// Send HTML update
		htmlUpdate := "html:" + htmlCard.String()
		pkg.LogInfof("progress: Sending HTML update for instance: %s (UniqueID: %s, HTML ID will be: instance-%s)", completedInstance.AutoName, completedInstance.UniqueID, completedInstance.UniqueID)
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
		pkg.LogWarnWithCritical(fmt.Sprintf("progress: Could not find completed instance for ID: %s", h.currentInstanceID), false)
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

	config, _, err := pkg.LoadConfigFromFile(flags)
	if err != nil {
		http.Error(w, "Error loading config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store config for later use
	mu.Lock()
	configMap[id] = config
	mu.Unlock()

	go func() {
		progressReporter := NewHTMLProgressReporter(updates, config, id, flags.ServerModeConfigFile)
		result, err := pkg.Process(config, progressReporter, cancel, pkg.Operations{
			GenerateReport: false,
		}, true)
		if err != nil {
			mu.Lock()
			jobErrorMap[id] = err.Error()
			mu.Unlock()
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
			mu.Lock()
			jobErr, hasErr := jobErrorMap[id]
			cleanupJobState(id)
			mu.Unlock()
			if hasErr {
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "error")
				templates.ProgressFailure(jobErr).Render(context.Background(), w)
				return
			}

			// Processing is done, just return "done"
			w.Header().Set("X-Progress-Status", "done")
			done := templates.AllComplete()
			done.Render(context.Background(), w)
			return
		}

		// Check if this is an instance start HTML update
		if strings.HasPrefix(msg, "instance-start-html:") {
			// Render instance start template directly
			htmlContent := strings.TrimPrefix(msg, "instance-start-html:")
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("X-Progress-Status", "progress")
			w.Write([]byte(htmlContent))
			return
		}
		// Check if this is an HTML update
		if strings.HasPrefix(msg, "html:") {
			// For HTML updates (instance completions), return HTMX OOB update
			htmlContent := strings.TrimPrefix(msg, "html:")

			// Drain all available HTML updates from the channel
			// This ensures we don't miss any instance updates
			htmlUpdates := []string{htmlContent}
			var progressMsg string
			draining := true
			for draining {
				select {
				case progressUpdate := <-updates:
					if strings.HasPrefix(progressUpdate, "html:") {
						// Another HTML update, add it to the list
						htmlUpdates = append(htmlUpdates, strings.TrimPrefix(progressUpdate, "html:"))
					} else {
						// Non-HTML update, use it as progress message and stop draining
						progressMsg = progressUpdate
						draining = false
					}
				default:
					// No more updates available, stop draining
					draining = false
					if progressMsg == "" {
						progressMsg = "Instance completed"
					}
				}
			}

			// Combine all HTML updates with proper spacing
			htmlContent = strings.Join(htmlUpdates, "")
			log.Printf("ProgressHandler: Batching %d HTML updates for OOB swap", len(htmlUpdates))

			// Check if processing is complete by checking if we have a result
			mu.Lock()
			_, hasResult := resultMap[id]
			jobErr, hasErr := jobErrorMap[id]
			mu.Unlock()

			if strings.HasPrefix(progressMsg, "error:") || hasErr {
				errMsg := trimProgressError(progressMsg)
				if errMsg == "" {
					errMsg = jobErr
				}
				mu.Lock()
				cleanupJobState(id)
				mu.Unlock()
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "error")
				templates.ProgressFailure(errMsg).Render(context.Background(), w)
				w.Write([]byte(htmlContent))
			} else if hasResult {
				log.Printf("ProgressHandler: Returning progress complete")
				// Return OOB updates for completion
				progressComplete := templates.ProgressComplete()
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "complete")
				progressComplete.Render(context.Background(), w)
			} else {
				log.Printf("ProgressHandler: Returning progress update for instance: %s", progressMsg)
				// Return progress update and instance HTML as OOB updates
				progressUpdate := templates.ProgressUpdate(progressMsg, id, true)
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "html")
				progressUpdate.Render(context.Background(), w)
				w.Write([]byte(htmlContent))
			}
		} else if msg == "Complete" {
			// "Complete" message signals processing is done
			// Check if we have a result to confirm completion
			mu.Lock()
			_, hasResult := resultMap[id]
			mu.Unlock()

			if hasResult {
				// Processing is complete
				progressComplete := templates.ProgressComplete()
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "complete")
				progressComplete.Render(context.Background(), w)
			} else {
				// Still processing, show the message and keep polling
				progressUpdate := templates.ProgressUpdate(msg, id, true)
				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("X-Progress-Status", "progress")
				progressUpdate.Render(context.Background(), w)
			}
		} else if strings.HasPrefix(msg, "error:") {
			mu.Lock()
			cleanupJobState(id)
			mu.Unlock()
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("X-Progress-Status", "error")
			templates.ProgressFailure(trimProgressError(msg)).Render(context.Background(), w)
		} else {
			// Return HTMX-compatible progress update with hx-get for next poll
			progressUpdate := templates.ProgressUpdate(msg, id, true)
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("X-Progress-Status", "progress")
			progressUpdate.Render(context.Background(), w)
		}
		/*default:
		// No update available yet, return current state to keep polling
		progressUpdate := templates.ProgressUpdate("Processing...", id, true)
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Progress-Status", "waiting")
		progressUpdate.Render(context.Background(), w)*/
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
		cleanupJobState(id)
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
		progressReporter := NewHTMLProgressReporter(updates, config, id, config.ServerModeConfigFile)
		result, err := pkg.Process(config, progressReporter, cancel, pkg.Operations{
			GenerateReport: false,
		}, true)
		if err != nil {
			log.Printf("Processing error: %v", err)
			mu.Lock()
			jobErrorMap[id] = err.Error()
			mu.Unlock()
			updates <- "error: " + err.Error()
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
		<div id="progress" class="notification is-info" hx-get="/progress?id=%s" hx-trigger="every 1s" hx-swap="outerHTML"></div>[[GetProgressHTML]]

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
	`, id, id, id)
}
