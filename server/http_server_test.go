package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	req, err := http.NewRequest("GET", "/images?path="+testImagePath, nil)
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
	req, err := http.NewRequest("GET", "/images?path=/nonexistent/file.png", nil)
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
