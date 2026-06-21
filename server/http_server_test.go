package server

import (
	"encoding/base64"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestHandleImageRequest(t *testing.T) {
	// Create a temporary test image file
	tempDir := t.TempDir()
	testImagePath := filepath.Join(tempDir, "test.png")

	// Create a simple test file
	err := os.WriteFile(testImagePath, []byte("fake png content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Test the endpoint
	req, err := http.NewRequest("GET", "/images?config_path="+testImagePath, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handleImageRequest(rr, req)

	// Check the response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("Expected content type 'image/png', got '%s'", contentType)
	}

	if cacheControl := rr.Header().Get("Cache-Control"); cacheControl != "no-store, no-cache, must-revalidate, max-age=0" {
		t.Errorf("Expected no-store cache control, got %q", cacheControl)
	}
}

func TestHandleImageRequestNotFound(t *testing.T) {
	// Test with non-existent file
	req, err := http.NewRequest("GET", "/images?config_path=/nonexistent/file.png", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handleImageRequest(rr, req)

	// Check the response
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestHandleImageRequestMissingPath(t *testing.T) {
	// Test without path parameter
	req, err := http.NewRequest("GET", "/images", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handleImageRequest(rr, req)

	// Check the response
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandlePUTRequestRedirectsWhenInstanceSetChanges(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")
	scadPath := filepath.Join(tempDir, "sample.scad")

	if err := os.WriteFile(scadPath, []byte("cube([1,1,1]);\n"), 0o644); err != nil {
		t.Fatalf("failed to write scad file: %v", err)
	}

	initialConfig := `[openscadgen]
name = "sample"
version = "v0.1"
export_name_format = "{designFileName}_{size}"

[[openscadgen.input_paths]]
path = "./sample.scad"

[[openscadgen.instances]]
name = "first"
params = { size = 1 }
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loadedConfig, _, err := pkg.LoadConfigFromFile(models.CmdFlags{
		ConfigFile:   configPath,
		Server:       true,
		ServerFolder: tempDir,
	})
	if err != nil {
		t.Fatalf("failed to load initial config: %v", err)
	}
	instances, err := pkg.GenerateInstanceConfigs(loadedConfig)
	if err != nil {
		t.Fatalf("failed to generate initial instances: %v", err)
	}
	initialSignature := pkg.BuildInstanceSetSignature(instances)

	updatedConfig := initialConfig + `
[[openscadgen.instances]]
name = "second"
params = { size = 2 }
`
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0o644); err != nil {
		t.Fatalf("failed to update config: %v", err)
	}

	formData := make(url.Values)
	formData.Set("config_path", base64.StdEncoding.EncodeToString([]byte(configPath)))
	formData.Set("server_folder", base64.StdEncoding.EncodeToString([]byte(tempDir)))
	formData.Set("page_instances_signature", initialSignature)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handlePUTRequest(rr, req)

	expectedRedirect := pkg.BuildPageUrl(configPath, tempDir).PageURL
	if got := rr.Header().Get("HX-Redirect"); got != expectedRedirect {
		t.Fatalf("expected HX-Redirect %q, got %q", expectedRedirect, got)
	}
}

func TestHandleOpenSCADStatusWithoutBinary(t *testing.T) {
	emptyBin := t.TempDir()
	t.Setenv("PATH", emptyBin)

	req := httptest.NewRequest(http.MethodGet, "/api/openscad/status", nil)
	rr := httptest.NewRecorder()
	handleOpenSCADStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	got := mustParseOpenSCADDataAttrs(t, body)
	if got["available"] != "false" {
		t.Fatalf("expected data-available=false, got %#v", got)
	}
	if got["error"] == "" {
		t.Fatal("expected non-empty data-error when openscad missing")
	}
	if got["path"] != "" || got["version"] != "" {
		t.Fatalf("expected empty path/version, got %#v", got)
	}
}

func TestHandleOpenSCADStatusWithFakeBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake openscad stub not implemented for windows")
	}
	binDir := t.TempDir()
	fake := filepath.Join(binDir, "openscad")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho 'OpenSCAD fake 1.0'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	req := httptest.NewRequest(http.MethodGet, "/api/openscad/status", nil)
	rr := httptest.NewRecorder()
	handleOpenSCADStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rr.Code, rr.Body.String())
	}
	got := mustParseOpenSCADDataAttrs(t, rr.Body.String())
	if got["available"] != "true" {
		t.Fatalf("expected data-available=true, got %#v", got)
	}
	if got["error"] != "" {
		t.Fatalf("unexpected data-error: %q", got["error"])
	}
	if got["version"] == "" || !strings.Contains(got["version"], "OpenSCAD fake") {
		t.Fatalf("unexpected data-version: %q", got["version"])
	}
	if got["path"] == "" {
		t.Fatal("expected non-empty data-path")
	}
}

func TestHandleOpenSCADStatusMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/openscad/status", nil)
	rr := httptest.NewRecorder()
	handleOpenSCADStatus(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestHandleConfigGetReadErrorRendersHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rr := httptest.NewRecorder()

	handleConfigGet(rr, req, filepath.Join(t.TempDir(), "missing.toml"), "")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Edit TOML File") {
		t.Fatalf("expected editor page, got %q", body)
	}
	if !strings.Contains(body, "Failed to read config file") {
		t.Fatalf("expected read error in body, got %q", body)
	}
}

func TestHandleConfigPostInvalidTOMLRendersContext(t *testing.T) {
	form := url.Values{}
	form.Set("content", "l1\nl2\nl3\n[openscadgen")
	req := httptest.NewRequest(http.MethodPost, "/api/config", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handleConfigPost(rr, req, filepath.Join(t.TempDir(), "config.toml"), "")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Invalid TOML") {
		t.Fatalf("expected invalid TOML message, got %q", body)
	}
	if !strings.Contains(body, "Source context:") {
		t.Fatalf("expected TOML source context, got %q", body)
	}
}

func TestProgressHandlerClosedErrorJobRendersFailure(t *testing.T) {
	id := "job-error-closed"
	updates := make(chan string)
	close(updates)

	mu.Lock()
	progressMap[id] = updates
	jobErrorMap[id] = "boom"
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/progress?id="+id, nil)
	rr := httptest.NewRecorder()
	ProgressHandler(rr, req)

	if rr.Header().Get("X-Progress-Status") != "error" {
		t.Fatalf("expected error status, got %q", rr.Header().Get("X-Progress-Status"))
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Processing failed") || !strings.Contains(body, "boom") {
		t.Fatalf("expected rendered failure, got %q", body)
	}
}

func TestProgressHandlerHTMLErrorBatchRendersFailureAndCard(t *testing.T) {
	id := "job-error-html"
	updates := make(chan string, 2)
	updates <- `html:<div id="instance-test" hx-swap-oob="true">card</div>`
	updates <- "error: boom"

	mu.Lock()
	progressMap[id] = updates
	jobErrorMap[id] = "boom"
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/progress?id="+id, nil)
	rr := httptest.NewRecorder()
	ProgressHandler(rr, req)

	if rr.Header().Get("X-Progress-Status") != "error" {
		t.Fatalf("expected error status, got %q", rr.Header().Get("X-Progress-Status"))
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Processing failed") {
		t.Fatalf("expected failure message, got %q", body)
	}
	if !strings.Contains(body, `id="instance-test"`) {
		t.Fatalf("expected instance card HTML in response, got %q", body)
	}
}

func mustParseOpenSCADDataAttrs(t *testing.T, pageHTML string) map[string]string {
	t.Helper()
	const marker = `<div id="openscad-status"`
	i := strings.Index(pageHTML, marker)
	if i < 0 {
		t.Fatalf("missing %s in HTML", marker)
	}
	chunk := pageHTML[i:]
	j := strings.IndexByte(chunk, '>')
	if j < 0 {
		t.Fatal("unclosed openscad-status div")
	}
	openTag := chunk[:j]
	out := map[string]string{}
	for _, name := range []string{"available", "path", "version", "out-of-date", "error"} {
		key := `data-` + name + `="`
		k := strings.Index(openTag, key)
		if k < 0 {
			t.Fatalf("missing %s on status div: %q", key, openTag)
		}
		start := k + len(key)
		end := strings.IndexByte(openTag[start:], '"')
		if end < 0 {
			t.Fatalf("unterminated %s", key)
		}
		out[name] = html.UnescapeString(openTag[start : start+end])
	}
	return out
}
