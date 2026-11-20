package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

const (
	// mcpProcessJobRetention keeps a finished job queryable long enough for a
	// client to poll after a slow render.
	mcpProcessJobRetention = 15 * time.Minute
	maxMCPProcessJobs      = 32
	maxMCPProcessWait      = 300
)

// mcpProcessJob tracks one asynchronous process_config run. Clients poll the
// snapshot either through the progress URL (HTTP) or the process_status tool
// (any transport).
type mcpProcessJob struct {
	mu sync.Mutex

	ID         string
	ConfigPath string
	Root       string
	Detail     string

	StartedAt  time.Time
	FinishedAt time.Time
	Total      int
	Completed  int
	Current    string
	Message    string
	Err        string

	Result  mcpResult
	Warning string

	done     chan struct{}
	doneOnce sync.Once
}

func (j *mcpProcessJob) markDone() {
	j.doneOnce.Do(func() { close(j.done) })
}

// setTotal records how many instances will be rendered.
func (j *mcpProcessJob) setTotal(total int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if total >= 0 {
		j.Total = total
	}
}

// startInstance records the instance currently being rendered.
func (j *mcpProcessJob) startInstance(name string, count int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Current = name
	if count > 0 {
		j.Total = count
	}
}

func (j *mcpProcessJob) finishInstance() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Completed++
	if j.Total > 0 && j.Completed > j.Total {
		j.Completed = j.Total
	}
}

func (j *mcpProcessJob) setMessage(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Message = message
}

func (j *mcpProcessJob) setError(err error) {
	if err == nil {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Err = err.Error()
}

// finish stores the terminal payload and releases any waiters.
func (j *mcpProcessJob) finish(result mcpResult, warning string, processErr error) {
	j.mu.Lock()
	j.FinishedAt = time.Now()
	j.Result = result
	j.Warning = warning
	j.Current = ""
	if processErr != nil {
		j.Err = processErr.Error()
	}
	j.mu.Unlock()
	j.markDone()
}

// snapshot renders the current progress. shape is stable across running and
// finished states so a polling client can branch on status/done.
func (j *mcpProcessJob) snapshot(level string) map[string]any {
	j.mu.Lock()
	defer j.mu.Unlock()

	elapsed := time.Since(j.StartedAt)
	if !j.FinishedAt.IsZero() {
		elapsed = j.FinishedAt.Sub(j.StartedAt)
	}

	out := map[string]any{
		"ok":                  true,
		"job_id":              j.ID,
		"status":              "running",
		"done":                false,
		"config_path":         j.ConfigPath,
		"started_at":          j.StartedAt.UTC().Format(time.RFC3339),
		"elapsed_ms":          elapsed.Milliseconds(),
		"current_instance":    j.Current,
		"completed_instances": j.Completed,
		"total_instances":     j.Total,
		"detail":              level,
	}
	if j.Message != "" {
		out["message"] = j.Message
	}
	if j.FinishedAt.IsZero() {
		return out
	}

	out["status"] = "done"
	out["done"] = true
	out["finished_at"] = j.FinishedAt.UTC().Format(time.RFC3339)
	out["current_instance"] = ""
	out["ok"] = j.Result.OK
	out["result"] = j.Result.render(level)
	if j.Err != "" {
		out["error"] = j.Err
	}
	if j.Warning != "" {
		out["warning"] = j.Warning
	}
	return out
}

// mcpJobProgress adapts the renderer's progress callbacks onto a job.
type mcpJobProgress struct {
	job *mcpProcessJob
}

func (p *mcpJobProgress) Update(msg string) { p.job.setMessage(msg) }
func (p *mcpJobProgress) Done()             { p.job.setMessage("render complete") }
func (p *mcpJobProgress) Error(err error)   { p.job.setError(err) }
func (p *mcpJobProgress) Construct(instances []models.InstanceConfig, nonSkippedInstances int) {
	p.job.setTotal(nonSkippedInstances)
}
func (p *mcpJobProgress) StartInstance(instanceID string, name string, instanceIndex int, instanceCount int) {
	p.job.startInstance(name, instanceCount)
}
func (p *mcpJobProgress) FinishInstance() { p.job.finishInstance() }

var (
	mcpProcessJobsMu sync.Mutex
	mcpProcessJobs   = make(map[string]*mcpProcessJob)
)

func newProcessJob(configPath string, root string, detail string) *mcpProcessJob {
	job := &mcpProcessJob{
		ID:         uuid.NewString(),
		ConfigPath: configPath,
		Root:       root,
		Detail:     detail,
		StartedAt:  time.Now(),
		done:       make(chan struct{}),
	}

	mcpProcessJobsMu.Lock()
	defer mcpProcessJobsMu.Unlock()
	cleanupProcessJobsLocked(time.Now())
	mcpProcessJobs[job.ID] = job
	return job
}

func lookupProcessJob(id string) (*mcpProcessJob, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, false
	}
	mcpProcessJobsMu.Lock()
	defer mcpProcessJobsMu.Unlock()
	cleanupProcessJobsLocked(time.Now())
	job, ok := mcpProcessJobs[id]
	return job, ok
}

// cleanupProcessJobsLocked expires finished jobs and bounds the registry.
func cleanupProcessJobsLocked(now time.Time) {
	for id, job := range mcpProcessJobs {
		job.mu.Lock()
		finished := !job.FinishedAt.IsZero()
		expired := finished && now.Sub(job.FinishedAt) > mcpProcessJobRetention
		job.mu.Unlock()
		if expired {
			delete(mcpProcessJobs, id)
		}
	}
	if len(mcpProcessJobs) <= maxMCPProcessJobs {
		return
	}

	// Evict the oldest finished jobs first; running jobs are never dropped.
	type candidate struct {
		id       string
		finished time.Time
	}
	finishedJobs := make([]candidate, 0, len(mcpProcessJobs))
	for id, job := range mcpProcessJobs {
		job.mu.Lock()
		finishedAt := job.FinishedAt
		job.mu.Unlock()
		if !finishedAt.IsZero() {
			finishedJobs = append(finishedJobs, candidate{id: id, finished: finishedAt})
		}
	}
	for len(mcpProcessJobs) > maxMCPProcessJobs && len(finishedJobs) > 0 {
		oldest := 0
		for index := 1; index < len(finishedJobs); index++ {
			if finishedJobs[index].finished.Before(finishedJobs[oldest].finished) {
				oldest = index
			}
		}
		delete(mcpProcessJobs, finishedJobs[oldest].id)
		finishedJobs = append(finishedJobs[:oldest], finishedJobs[oldest+1:]...)
	}
}

// runProcessJob renders in the background and records the terminal payload.
func runProcessJob(job *mcpProcessJob, config *models.Config, root string, warning string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result := mcpResult{
				OK: false,
				Summary: map[string]any{
					"config_path": job.ConfigPath,
					"error":       fmt.Sprintf("render panicked: %v", recovered),
				},
			}
			job.finish(result, warning, fmt.Errorf("render panicked: %v", recovered))
		}
	}()

	cancel := make(chan struct{})
	timer := time.AfterFunc(maxMCPProcessDuration, func() { close(cancel) })
	defer timer.Stop()

	result, processErr := pkg.Process(config, &mcpJobProgress{job: job}, cancel, pkg.Operations{GenerateReport: false}, true)
	payload := buildProcessPayload(root, result, processErr, warning)
	job.finish(payload, warning, processErr)
}

// waitForProcessJob waits up to timeout for completion. A non-positive timeout
// checks once without blocking.
func waitForProcessJob(job *mcpProcessJob, timeout time.Duration) bool {
	if timeout <= 0 {
		select {
		case <-job.done:
			return true
		default:
			return false
		}
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-job.done:
		return true
	case <-timer.C:
		return false
	}
}

// progressURL is the polled HTTP address for a job. It is empty when the HTTP
// transport is not mounted (stdio mode), where process_status is the only poll.
func progressURL(id string) string {
	if mcpHTTPPort <= 0 {
		return ""
	}
	return fmt.Sprintf("http://127.0.0.1:%d/api/mcp/process/%s", mcpHTTPPort, id)
}

// waitForProcessStart gives the background job a short window to report its
// instance total, so the first response already carries total_instances.
func waitForProcessStart(job *mcpProcessJob, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		job.mu.Lock()
		total := job.Total
		finished := !job.FinishedAt.IsZero()
		job.mu.Unlock()
		if total > 0 || finished || time.Now().After(deadline) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// withProgressLinks decorates a snapshot with the polling endpoint and tool.
func withProgressLinks(payload map[string]any) map[string]any {
	id, _ := payload["job_id"].(string)
	payload["progress_tool"] = "process_status"
	if url := progressURL(id); url != "" {
		payload["progress_url"] = url
	}
	return payload
}

// handleMCPProcessStatus serves GET /api/mcp/process/<job id>. It returns the
// same snapshot as the process_status tool so scripts and dashboards can poll
// without speaking MCP.
func handleMCPProcessStatus(w http.ResponseWriter, r *http.Request) {
	if !mcpToolsEnabled {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "progress endpoint only accepts GET", http.StatusMethodNotAllowed)
		return
	}

	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/mcp/process/"), "/")
	if id == "" {
		http.Error(w, "missing job id", http.StatusBadRequest)
		return
	}
	job, ok := lookupProcessJob(id)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"ok":false,"error":"unknown or expired process job"}` + "\n"))
		return
	}

	level := mcpServerConfig.detailLevel(r.URL.Query().Get("detail"))
	payload := withProgressLinks(job.snapshot(level))

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_ = writeJSONResponse(w, payload)
}

// writeJSONResponse writes payload as JSON.
func writeJSONResponse(w http.ResponseWriter, payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return err
	}
	_, err = w.Write(append(encoded, '\n'))
	return err
}
