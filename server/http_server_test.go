package server

import (
	"context"
	"encoding/base64"
	"fmt"
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

func TestHandleOpenSCADInstallUsesRequestedReleaseChannel(t *testing.T) {
	originalInstall := installOpenSCAD
	t.Cleanup(func() { installOpenSCAD = originalInstall })

	var installedNightly bool
	installOpenSCAD = func(_ context.Context, nightly bool) error {
		installedNightly = nightly
		return nil
	}
	request := func(form url.Values) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/openscad/install", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("HX-Target", "tools-results-content")
		rr := httptest.NewRecorder()
		handleOpenSCADInstall(rr, req)
		return rr
	}

	if rr := request(url.Values{"install_nightly": {"true"}}); rr.Code != http.StatusOK {
		t.Fatalf("expected nightly install to succeed, got %d: %s", rr.Code, rr.Body.String())
	}
	if !installedNightly {
		t.Fatal("expected the checked nightly option to request a nightly install")
	}

	if rr := request(url.Values{}); rr.Code != http.StatusOK {
		t.Fatalf("expected stable install to succeed, got %d: %s", rr.Code, rr.Body.String())
	}
	if installedNightly {
		t.Fatal("expected an unchecked nightly option to request a stable install")
	}
}

func TestHandleOpenSCADInstallRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/openscad/install", nil)
	rr := httptest.NewRecorder()
	handleOpenSCADInstall(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleOpenSCADLibraryActionInstallsBOSL2(t *testing.T) {
	originalInstall := installBOSL2
	originalServerFolder := globalServerFolder
	t.Cleanup(func() {
		installBOSL2 = originalInstall
		globalServerFolder = originalServerFolder
	})
	installed := false
	installBOSL2 = func(_ context.Context) error {
		installed = true
		return nil
	}
	globalServerFolder = t.TempDir()

	req := httptest.NewRequest(http.MethodPost, "/api/openscad/libraries/bosl2/update", nil)
	req.Header.Set("HX-Target", "tools-results-content")
	rr := httptest.NewRecorder()
	handleOpenSCADLibraryAction(rr, req)

	if !installed {
		t.Fatal("expected BOSL2 installer to run")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `class="message is-success"`) {
		t.Fatalf("expected full results response, got %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "BOSL2 installed") || !strings.Contains(rr.Body.String(), `class="message is-success"`) {
		t.Fatalf("unexpected response: %s", rr.Body.String())
	}
}

func TestHandleOpenSCADLibraryActionRejectsGETUpdate(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/openscad/libraries/bosl2/update", nil)
	rr := httptest.NewRecorder()
	handleOpenSCADLibraryAction(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
}

func TestHandleOpenSCADLibraryActionShowsManualCommandAfterInstallFailure(t *testing.T) {
	originalInstall := installBOSL2
	t.Cleanup(func() { installBOSL2 = originalInstall })
	installBOSL2 = func(_ context.Context) error { return fmt.Errorf("download timed out") }

	req := httptest.NewRequest(http.MethodPost, "/api/openscad/libraries/bosl2/update", nil)
	req.Header.Set("HX-Target", "tools-results-content")
	rr := httptest.NewRecorder()
	handleOpenSCADLibraryAction(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "download timed out") || !strings.Contains(rr.Body.String(), "Copy command") || !strings.Contains(rr.Body.String(), "set -e") || !strings.Contains(rr.Body.String(), "git clone --depth 1") {
		t.Fatalf("expected manual recovery command, got %s", rr.Body.String())
	}
}

func TestHandleOpenSCADFileOpensConfiguredSCADOnly(t *testing.T) {
	serverFolder := t.TempDir()
	configPath := filepath.Join(serverFolder, "config.toml")
	firstSCAD := filepath.Join(serverFolder, "first.scad")
	secondSCAD := filepath.Join(serverFolder, "second.scad")
	for _, path := range []string{firstSCAD, secondSCAD} {
		if err := os.WriteFile(path, []byte("cube(1);\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	config := `[openscadgen]
name = "example"
version = "v0.1"
export_name_format = "{designFileName}"

[[openscadgen.input_paths]]
path = "./first.scad"

[[openscadgen.input_paths]]
path = "./second.scad"
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	var opened string
	originalStarter := startOpenSCADFile
	startOpenSCADFile = func(path string) error {
		opened = path
		return nil
	}
	t.Cleanup(func() { startOpenSCADFile = originalStarter })

	makeRequest := func(sourcePath string) *httptest.ResponseRecorder {
		query := make(url.Values)
		query.Set("config_path", base64.StdEncoding.EncodeToString([]byte(configPath)))
		query.Set("server_folder", base64.StdEncoding.EncodeToString([]byte(serverFolder)))
		query.Set("source_path", base64.StdEncoding.EncodeToString([]byte(sourcePath)))
		req := httptest.NewRequest(http.MethodPost, "/api/openscad/open?"+query.Encode(), nil)
		rr := httptest.NewRecorder()
		handleOpenSCADFile(rr, req)
		return rr
	}

	if rr := makeRequest("./first.scad"); rr.Code != http.StatusOK {
		t.Fatalf("expected configured SCAD to open, got %d: %s", rr.Code, rr.Body.String())
	}
	if opened != firstSCAD {
		t.Fatalf("opened %q, want %q", opened, firstSCAD)
	}

	opened = ""
	if rr := makeRequest("./not-configured.scad"); rr.Code != http.StatusForbidden {
		t.Fatalf("expected unconfigured SCAD to be rejected, got %d: %s", rr.Code, rr.Body.String())
	}
	if opened != "" {
		t.Fatalf("unconfigured source started OpenSCAD for %q", opened)
	}

	if rr := makeRequest("../outside.scad"); rr.Code != http.StatusForbidden {
		t.Fatalf("expected path outside the server folder to be rejected, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleOpenSCADFileRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/openscad/open", nil)
	rr := httptest.NewRecorder()
	handleOpenSCADFile(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", rr.Code, rr.Body.String())
	}
}

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

func TestHandleEditPostBlocksSaveForBlockingInstanceValidation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")
	if err := os.WriteFile(filepath.Join(tempDir, "sample.scad"), []byte("cube(1);\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	original := `[openscadgen]
name = "sample"
version = "v0.1"
export_name_format = "{designFileName}"

[[openscadgen.input_paths]]
path = "./sample.scad"
`
	invalid := original + `
[[openscadgen.instances]]
name = "first"
params = { size = 1 }

[[openscadgen.instances]]
name = "second"
params = { size = 2 }
`
	if err := os.WriteFile(configPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	form := url.Values{"content": {invalid}, "action": {"save"}}
	req := httptest.NewRequest(http.MethodPost, "/api/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handleEditPost(rr, req, configPath, tempDir)

	body := rr.Body.String()
	if !strings.Contains(body, "Save-blocking error") || !strings.Contains(body, "disabled") {
		t.Fatalf("expected blocking feedback and disabled save, got %q", body)
	}
	got, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("blocking dry run wrote the config:\n%s", got)
	}
}

func TestHandlePUTRequestReturnsConfigValidationError(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")
	scadPath := filepath.Join(tempDir, "sample.scad")

	if err := os.WriteFile(scadPath, []byte("cube([1,1,1]);\n"), 0o644); err != nil {
		t.Fatalf("failed to write scad file: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("[openscadgen\n"), 0o644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	formData := make(url.Values)
	formData.Set("config_path", base64.StdEncoding.EncodeToString([]byte(configPath)))
	formData.Set("server_folder", base64.StdEncoding.EncodeToString([]byte(tempDir)))
	formData.Set("raw_config_file", base64.StdEncoding.EncodeToString([]byte("[openscadgen\n")))

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handlePUTRequest(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%q", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "LoadConfigFromFileError") {
		t.Fatalf("expected detailed load error, got %q", body)
	}
	if !strings.Contains(body, "Source context") && !strings.Contains(body, "line 1") {
		t.Fatalf("expected TOML error details, got %q", body)
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

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != `{"ok":true}` {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestHandleConfigOptionsRequestFiltersTopic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/config/options?topic=images", nil)
	rr := httptest.NewRecorder()
	handleConfigOptionsRequest(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "[openscadgen].images") {
		t.Fatalf("expected images option in response, got %q", body)
	}
	if strings.Contains(body, "[openscadgen].name") {
		t.Fatalf("expected non-images options to be filtered out, got %q", body)
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

func TestProgressHandlerCompletedHTMLBatchRendersFinalInstanceCard(t *testing.T) {
	id := "job-complete-html"
	updates := make(chan string, 1)
	updates <- `html:<div id="instance-chip_cover_sliced_chipWidth535" hx-swap-oob="true">render error</div>`

	mu.Lock()
	progressMap[id] = updates
	resultMap[id] = models.ProcessResult{}
	mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/progress?id="+id, nil)
	rr := httptest.NewRecorder()
	ProgressHandler(rr, req)

	if rr.Header().Get("X-Progress-Status") != "complete" {
		t.Fatalf("expected complete status, got %q", rr.Header().Get("X-Progress-Status"))
	}
	body := rr.Body.String()
	if !strings.Contains(body, "All instances completed!") {
		t.Fatalf("expected completion update, got %q", body)
	}
	if !strings.Contains(body, `id="instance-chip_cover_sliced_chipWidth535"`) {
		t.Fatalf("expected final instance OOB card in response, got %q", body)
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
