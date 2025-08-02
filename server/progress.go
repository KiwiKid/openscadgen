package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

var (
	progressMap = make(map[string]chan string)
	cancelMap   = make(map[string]chan struct{})
	resultMap   = make(map[string]models.ProcessResult)
	configMap   = make(map[string]*models.Config)
	mu          sync.Mutex
)

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
		result, err := pkg.Process(config, pkg.NewTerminalProgressReporter(config), cancel)
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
				mu.Unlock()
			} else {
				w.Header().Set("X-Progress-Status", "done")
				w.Write([]byte("done"))
			}
			return
		}
		w.Header().Set("X-Progress-Status", "progress")
		w.Write([]byte(msg))
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
		result, err := pkg.Process(config, pkg.NewTerminalProgressReporter(config), cancel)
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
		<div id="progress"></div>
		<button onclick="fetch('/cancel?id=%s')">Cancel</button>
		<script>
		function poll() {
			fetch('/progress?id=%s').then(r => {
				const status = r.headers.get('X-Progress-Status');
				return r.text().then(msg => ({ status, msg }));
			}).then(({ status, msg }) => {
				if(status === 'complete') {
					// Complete report received, replace the entire page content
					document.documentElement.innerHTML = msg;
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
