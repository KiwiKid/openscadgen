package pkg

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetOutputPaths(t *testing.T) {
	// Save and restore original logger after test
	originalLogger := logger
	originalOutput := log.Writer()
	// Redirect log output to discard during test
	log.SetOutput(ioutil.Discard)
	defer func() {
		logger = originalLogger
		log.SetOutput(originalOutput)
	}()

	// Create a test logger that won't output anything
	var logBuffer bytes.Buffer
	testLogger := log.New(&logBuffer, "", 0)
	logger = testLogger

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "openscadgen_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.toml")
	if err := os.WriteFile(configPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create test input file and directory structure
	designsDir := filepath.Join(tempDir, "designs")
	if err := os.MkdirAll(designsDir, 0755); err != nil {
		t.Fatalf("Failed to create designs dir: %v", err)
	}

	inputPath := filepath.Join(tempDir, "test_design.scad")
	if err := os.WriteFile(inputPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	relativeInputPath := filepath.Join(designsDir, "test_design.scad")
	if err := os.WriteFile(relativeInputPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create relative input file: %v", err)
	}

	testCases := []struct {
		name     string
		config   Config
		expected OutputPaths
	}{
		{
			name: "With specified output path",
			config: Config{
				ConfigFile: configPath,
				Design: DesignConfig{
					InputPath:  filepath.Join("designs", "test_design.scad"),
					OutputPath: "outputs",
					Version:    "v1.0",
				},
				Debug: false,
			},
			expected: OutputPaths{
				OutputPath:            filepath.Join(tempDir, "outputs", "v1.0"),
				ExportFolderPath:      filepath.Join(tempDir, "outputs", "v1.0"),
				LowQualityWarningPath: filepath.Join(tempDir, "outputs", "v1.0", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(tempDir, "outputs", "v1.0", "README.md"),
				LogOutputPath:         filepath.Join(tempDir, "outputs", "v1.0", "export_log_"+time.Now().Format("2006-01-02_15-04-05")+".log"),
			},
		},
		{
			name: "With derived output path",
			config: Config{
				ConfigFile: configPath,
				Design: DesignConfig{
					InputPath: filepath.Join("designs", "test_design.scad"),
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: OutputPaths{
				OutputPath:            filepath.Join(".", designsDir, "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(designsDir, "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(designsDir, "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(designsDir, "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(designsDir, "export", "v1.0", "test_design", "export_log.log"),
			},
		},
		{
			name: "With absolute input path",
			config: Config{
				ConfigFile: configPath,
				Design: DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: OutputPaths{
				OutputPath:            filepath.Join(".", filepath.Dir(inputPath), "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(inputPath), "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "export_log.log"),
			},
		},
		{
			name: "Relative config file path - should be relative to config location",
			config: Config{
				ConfigFile: filepath.Join("some", "nested", "path", "config.toml"),
				Design: DesignConfig{
					InputPath: filepath.Join("designs", "test_design.scad"),
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: OutputPaths{
				// All export paths should be relative to the config file directory (some/nested/path/)
				// rather than the current working directory
				OutputPath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "export_log.log"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create output directory if specified
			if tc.config.Design.OutputPath != "" {
				outputDir := filepath.Join(tempDir, tc.config.Design.OutputPath)
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatalf("Failed to create output dir: %v", err)
				}
			}

			// Reset log buffer between test cases
			logBuffer.Reset()
			result := getOutputPaths(tc.config)

			// Cannot compare log paths directly due to timestamp, so check other fields
			if tc.name == "With specified output path" {
				// For the timestamp-based log file, just check that the path has the correct prefix
				logPathPrefix := filepath.Join(tempDir, "outputs", "v1.0", "export_log_")
				if !filepath.HasPrefix(result.LogOutputPath, logPathPrefix) {
					t.Errorf("LogOutputPath does not have expected prefix.\nExpected prefix: %s\nGot: %s",
						logPathPrefix, result.LogOutputPath)
				}

				// Check other fields
				if result.OutputPath != tc.expected.OutputPath {
					t.Errorf("OutputPath = %s; want %s", result.OutputPath, tc.expected.OutputPath)
				}
				if result.ExportFolderPath != tc.expected.ExportFolderPath {
					t.Errorf("ExportFolderPath = %s; want %s", result.ExportFolderPath, tc.expected.ExportFolderPath)
				}
				if result.LowQualityWarningPath != tc.expected.LowQualityWarningPath {
					t.Errorf("LowQualityWarningPath = %s; want %s", result.LowQualityWarningPath, tc.expected.LowQualityWarningPath)
				}
				if result.ReadmePath != tc.expected.ReadmePath {
					t.Errorf("ReadmePath = %s; want %s", result.ReadmePath, tc.expected.ReadmePath)
				}
			} else {
				// For other cases, we can compare all fields directly
				if result.OutputPath != tc.expected.OutputPath {
					t.Errorf("OutputPath = %s; want %s", result.OutputPath, tc.expected.OutputPath)
				}
				if result.ExportFolderPath != tc.expected.ExportFolderPath {
					t.Errorf("ExportFolderPath = %s; want %s", result.ExportFolderPath, tc.expected.ExportFolderPath)
				}
				if result.LowQualityWarningPath != tc.expected.LowQualityWarningPath {
					t.Errorf("LowQualityWarningPath = %s; want %s", result.LowQualityWarningPath, tc.expected.LowQualityWarningPath)
				}
				if result.ReadmePath != tc.expected.ReadmePath {
					t.Errorf("ReadmePath = %s; want %s", result.ReadmePath, tc.expected.ReadmePath)
				}
				if result.LogOutputPath != tc.expected.LogOutputPath {
					t.Errorf("LogOutputPath = %s; want %s", result.LogOutputPath, tc.expected.LogOutputPath)
				}
			}
		})
	}
}
