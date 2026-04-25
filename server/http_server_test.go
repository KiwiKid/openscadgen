package server

import (
	"html"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
