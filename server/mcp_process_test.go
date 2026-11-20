package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProcessJobProgressLifecycle(t *testing.T) {
	job := newProcessJob("cup/config.toml", t.TempDir(), mcpDetailNormal)
	reporter := &mcpJobProgress{job: job}

	reporter.Construct(nil, 3)
	reporter.StartInstance("instance-1", "cup_obj", 0, 3)
	reporter.Update("rendering cup_obj")

	running := job.snapshot(mcpDetailNormal)
	if running["status"] != "running" || running["done"] != false {
		t.Fatalf("unexpected running snapshot: %v", running)
	}
	if running["total_instances"] != 3 {
		t.Fatalf("total_instances = %v", running["total_instances"])
	}
	if running["completed_instances"] != 0 {
		t.Fatalf("completed_instances = %v", running["completed_instances"])
	}
	if running["current_instance"] != "cup_obj" {
		t.Fatalf("current_instance = %v", running["current_instance"])
	}
	if _, err := time.Parse(time.RFC3339, running["started_at"].(string)); err != nil {
		t.Fatalf("started_at is not RFC3339: %v", running["started_at"])
	}
	if running["message"] != "rendering cup_obj" {
		t.Fatalf("message = %v", running["message"])
	}

	reporter.FinishInstance()
	afterOne := job.snapshot(mcpDetailNormal)
	if afterOne["completed_instances"] != 1 {
		t.Fatalf("completed_instances = %v", afterOne["completed_instances"])
	}

	reporter.Done()
	job.finish(mcpResult{
		OK:      true,
		Summary: map[string]any{"artifacts": 1, "successful": 1},
	}, "", nil)

	done := job.snapshot(mcpDetailNormal)
	if done["status"] != "done" || done["done"] != true {
		t.Fatalf("unexpected done snapshot: %v", done)
	}
	if done["ok"] != true {
		t.Fatalf("ok = %v", done["ok"])
	}
	if done["current_instance"] != "" {
		t.Fatalf("current_instance should clear on completion: %v", done["current_instance"])
	}
	if _, err := time.Parse(time.RFC3339, done["finished_at"].(string)); err != nil {
		t.Fatalf("finished_at is not RFC3339: %v", done["finished_at"])
	}
	result, ok := done["result"].(map[string]any)
	if !ok {
		t.Fatalf("result missing: %v", done["result"])
	}
	if result["artifacts"] != 1 {
		t.Fatalf("result artifacts = %v", result["artifacts"])
	}
}

func TestProcessJobFailureSnapshot(t *testing.T) {
	job := newProcessJob("cup/config.toml", t.TempDir(), mcpDetailNormal)
	reporter := &mcpJobProgress{job: job}
	reporter.Construct(nil, 2)
	reporter.StartInstance("i", "first", 0, 2)
	reporter.FinishInstance()
	reporter.Error(errTestRenderFailure)

	result := mcpResult{
		OK:      false,
		Summary: map[string]any{"error": "all instances failed to render"},
	}
	job.finish(result, "", errTestRenderFailure)

	done := job.snapshot(mcpDetailSummary)
	if done["ok"] != false {
		t.Fatalf("ok = %v", done["ok"])
	}
	if done["error"] != errTestRenderFailure.Error() {
		t.Fatalf("error = %v", done["error"])
	}
	if done["completed_instances"] != 1 || done["total_instances"] != 2 {
		t.Fatalf("counts = %v/%v", done["completed_instances"], done["total_instances"])
	}
}

func TestProcessStatusTool(t *testing.T) {
	job := newProcessJob("demo/config.toml", t.TempDir(), mcpDetailNormal)
	job.finish(mcpResult{OK: true, Summary: map[string]any{"artifacts": 2}}, "", nil)

	server := newMCPServer(testMCPConfig(t, t.TempDir()))
	payload := toolPayload(t, server, "process_status", map[string]any{"job_id": job.ID})
	if payload["status"] != "done" || payload["ok"] != true {
		t.Fatalf("unexpected status payload: %v", payload)
	}
	if payload["progress_tool"] != "process_status" {
		t.Fatalf("progress_tool = %v", payload["progress_tool"])
	}

	response := callTool(t, server, "process_status", map[string]any{"job_id": "does-not-exist"})
	var result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !result.IsError || !strings.Contains(result.Content[0].Text, "unknown or expired") {
		t.Fatalf("unexpected missing-job result: %+v", result)
	}
}

func TestProcessProgressHTTPEndpoint(t *testing.T) {
	previousEnabled := mcpToolsEnabled
	previousConfig := mcpServerConfig
	previousPort := mcpHTTPPort
	t.Cleanup(func() {
		mcpToolsEnabled = previousEnabled
		mcpServerConfig = previousConfig
		mcpHTTPPort = previousPort
	})

	mcpToolsEnabled = true
	mcpServerConfig = testMCPConfig(t, t.TempDir())
	mcpHTTPPort = 7203

	job := newProcessJob("demo/config.toml", t.TempDir(), mcpDetailNormal)
	job.finish(mcpResult{OK: true, Summary: map[string]any{"artifacts": 1}}, "", nil)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/mcp/process/"+job.ID, nil)
	handleMCPProcessStatus(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded["job_id"] != job.ID || decoded["status"] != "done" {
		t.Fatalf("unexpected payload: %v", decoded)
	}
	if decoded["progress_url"] != "http://127.0.0.1:7203/api/mcp/process/"+job.ID {
		t.Fatalf("progress_url = %v", decoded["progress_url"])
	}
	for _, key := range []string{"started_at", "elapsed_ms", "current_instance", "completed_instances", "total_instances"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("payload missing %q: %v", key, decoded)
		}
	}

	missing := httptest.NewRecorder()
	handleMCPProcessStatus(missing, httptest.NewRequest(http.MethodGet, "/api/mcp/process/nope", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing job status = %d, want 404", missing.Code)
	}

	wrongMethod := httptest.NewRecorder()
	handleMCPProcessStatus(wrongMethod, httptest.NewRequest(http.MethodPost, "/api/mcp/process/"+job.ID, nil))
	if wrongMethod.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", wrongMethod.Code)
	}

	mcpToolsEnabled = false
	disabled := httptest.NewRecorder()
	handleMCPProcessStatus(disabled, httptest.NewRequest(http.MethodGet, "/api/mcp/process/"+job.ID, nil))
	if disabled.Code != http.StatusNotFound {
		t.Fatalf("disabled status = %d, want 404", disabled.Code)
	}
}

func TestProgressURLAbsentWithoutHTTP(t *testing.T) {
	previousPort := mcpHTTPPort
	t.Cleanup(func() { mcpHTTPPort = previousPort })

	mcpHTTPPort = 0
	if url := progressURL("abc"); url != "" {
		t.Fatalf("stdio should not advertise a URL, got %q", url)
	}
	mcpHTTPPort = 7203
	if url := progressURL("abc"); url != "http://127.0.0.1:7203/api/mcp/process/abc" {
		t.Fatalf("url = %q", url)
	}
}

func TestWaitForProcessStartReturnsWhenTotalKnown(t *testing.T) {
	job := newProcessJob("demo/config.toml", t.TempDir(), mcpDetailNormal)
	go func() {
		time.Sleep(20 * time.Millisecond)
		job.setTotal(4)
	}()

	start := time.Now()
	waitForProcessStart(job, 2*time.Second)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("waited too long: %v", elapsed)
	}
	if job.snapshot(mcpDetailSummary)["total_instances"] != 4 {
		t.Fatalf("total_instances = %v", job.snapshot(mcpDetailSummary)["total_instances"])
	}
}

func TestWaitForProcessJob(t *testing.T) {
	job := newProcessJob("demo/config.toml", t.TempDir(), mcpDetailNormal)
	if waitForProcessJob(job, 0) {
		t.Fatal("unfinished job should not report done without a timeout")
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		job.finish(mcpResult{OK: true}, "", nil)
	}()
	if !waitForProcessJob(job, 2*time.Second) {
		t.Fatal("expected the job to finish before the timeout")
	}
}

// errTestRenderFailure is a stable error for job assertions.
var errTestRenderFailure = &testRenderError{}

type testRenderError struct{}

func (e *testRenderError) Error() string { return "all instances failed to render" }
