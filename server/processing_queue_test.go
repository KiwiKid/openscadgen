package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestTestRunSummaryCountsPassingAndFailingTests(t *testing.T) {
	result := models.ProcessResult{STLResults: []models.GenerateSTLResult{
		{Tests: []models.GeometryTestResult{{Passed: true}, {Passed: false}}},
		{Tests: []models.GeometryTestResult{{Passed: true}}},
	}}
	summary := testRunSummary(result)
	if summary.Passing != 2 || summary.Failing != 1 {
		t.Fatalf("summary = %+v, want 2 passing and 1 failing", summary)
	}
	if !hasTestResults(result) {
		t.Fatal("expected test results to be detected")
	}
	if hasTestResults(models.ProcessResult{}) {
		t.Fatal("empty process result must not be treated as a test run")
	}
}

func TestLatestTestRunLoadsPersistedSummary(t *testing.T) {
	config := &models.Config{
		ConfigFile: filepath.Join(t.TempDir(), "config.toml"),
		Design:     models.DesignConfig{Version: "v0.1"},
	}
	if err := os.MkdirAll(pkg.GetOutputPaths(config).ExportFolderPath, 0o755); err != nil {
		t.Fatal(err)
	}
	want := models.TestRunSummary{Passing: 4, Failing: 2, RanAt: time.Now().Round(0)}
	if err := saveTestRunSummary(config, want); err != nil {
		t.Fatal(err)
	}

	got := latestTestRun(config)
	if got == nil || got.Passing != want.Passing || got.Failing != want.Failing || !got.RanAt.Equal(want.RanAt) {
		t.Fatalf("persisted summary = %+v, want %+v", got, want)
	}
}

func TestProcessingQueueBoundsPendingJobsAndPropagatesCancellation(t *testing.T) {
	queue := newProcessingQueue(1, 1)
	started := make(chan struct{})
	release := make(chan struct{})
	firstDone := make(chan struct{})
	if err := queue.submit("first", func(context.Context) {
		close(started)
		<-release
		close(firstDone)
	}); err != nil {
		t.Fatalf("submit first: %v", err)
	}
	<-started

	cancelled := make(chan bool, 1)
	if err := queue.submit("second", func(ctx context.Context) {
		cancelled <- ctx.Err() != nil
	}); err != nil {
		t.Fatalf("submit second: %v", err)
	}
	if err := queue.submit("overflow", func(context.Context) {}); err == nil {
		t.Fatal("expected bounded queue to reject overflow")
	}
	if !queue.cancel("second") {
		t.Fatal("expected queued job to be cancellable")
	}
	close(release)
	<-firstDone
	select {
	case wasCancelled := <-cancelled:
		if !wasCancelled {
			t.Fatal("queued job did not receive cancellation context")
		}
	case <-time.After(time.Second):
		t.Fatal("queued job was not run")
	}
}
