package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAISettingsPageData(t *testing.T) {
	root := t.TempDir()
	cfg := testMCPConfig(t, root)
	cfg.Detail = mcpDetailSummary
	cfg.AllowPush = true

	data := buildAISettingsPageData(cfg, 7203)
	if !data.Enabled {
		t.Fatal("expected enabled page")
	}
	if data.EndpointURL != "http://127.0.0.1:7203/mcp" {
		t.Fatalf("endpoint = %q", data.EndpointURL)
	}
	if !data.LoopbackOnly {
		t.Fatal("expected loopback detection")
	}
	if data.DetailLevel != mcpDetailSummary {
		t.Fatalf("detail = %q", data.DetailLevel)
	}
	if len(data.Roots) != 1 || !strings.EqualFold(data.Roots[0], filepath.Clean(root)) {
		t.Fatalf("roots = %v", data.Roots)
	}
	if len(data.Tools) != 11 {
		t.Fatalf("tools = %d, want 11", len(data.Tools))
	}
	for _, tool := range data.Tools {
		if tool.Gated {
			t.Fatalf("tool %q unexpectedly gated with push enabled", tool.Name)
		}
	}
	if len(data.Sources) == 0 {
		t.Fatal("expected setting provenance")
	}

	if len(data.Snippets) != 3 {
		t.Fatalf("snippets = %d, want 3", len(data.Snippets))
	}
	for _, snippet := range data.Snippets {
		if snippet.Format != "json" {
			continue
		}
		var decoded map[string]any
		if err := json.Unmarshal([]byte(snippet.Content), &decoded); err != nil {
			t.Fatalf("%s snippet is not valid JSON: %v\n%s", snippet.Title, err, snippet.Content)
		}
		if _, ok := decoded["mcpServers"]; !ok {
			t.Fatalf("%s snippet missing mcpServers", snippet.Title)
		}
	}
	if !strings.Contains(data.Snippets[2].Content, data.EndpointURL) {
		t.Fatalf("bridge snippet should reference the endpoint: %s", data.Snippets[2].Content)
	}
}

func TestBuildAISettingsPageDataReadOnly(t *testing.T) {
	cfg := testMCPConfig(t, t.TempDir())
	cfg.ReadOnly = true

	data := buildAISettingsPageData(cfg, 7203)
	active := 0
	gated := 0
	for _, tool := range data.Tools {
		if tool.Gated {
			gated++
			continue
		}
		active++
	}
	if active != 6 || gated != 5 {
		t.Fatalf("active=%d gated=%d, want 6/5", active, gated)
	}
	if !data.ReadOnly {
		t.Fatal("expected read-only flag")
	}
}

func TestAISettingsHandlerDisabled(t *testing.T) {
	previous := mcpToolsEnabled
	t.Cleanup(func() { mcpToolsEnabled = previous })
	mcpToolsEnabled = false

	recorder := httptest.NewRecorder()
	handleAISettingsPage(recorder, httptest.NewRequest(http.MethodGet, "/ai", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestAISettingsHandlerRejectsPost(t *testing.T) {
	previousEnabled := mcpToolsEnabled
	previousConfig := mcpServerConfig
	t.Cleanup(func() {
		mcpToolsEnabled = previousEnabled
		mcpServerConfig = previousConfig
	})
	mcpToolsEnabled = true
	mcpServerConfig = testMCPConfig(t, t.TempDir())

	recorder := httptest.NewRecorder()
	handleAISettingsPage(recorder, httptest.NewRequest(http.MethodPost, "/ai", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", recorder.Code)
	}
}

func TestMCPConfigNoticesForNonLoopback(t *testing.T) {
	flags := flagsForMCPTest(t.TempDir())
	flags.ServerAddress = "0.0.0.0"
	cfg := newMCPConfig(flags, "")
	found := false
	for _, notice := range cfg.Notices {
		if strings.Contains(notice, "no authentication") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected non-loopback warning, got %v", cfg.Notices)
	}
}
