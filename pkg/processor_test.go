package pkg

import (
	"bytes"
	//"fmt"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kiwikid/openscadgen/pkg/models"
)

// normalizePath converts a path to its canonical form for comparison
func normalizePath(path string) string {
	// Convert to absolute path and clean it
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return filepath.Clean(abs)
}

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
		config   models.Config
		expected models.OutputPaths
	}{
		{
			name: "with_specified_output_path",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        filepath.Join("designs", "test_design.scad"),
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{instanceName}",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join(tempDir, "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(tempDir, "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(tempDir, "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(tempDir, "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(tempDir, "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(tempDir, "export", "v1.0", "test_design", "report.html"),
			},
		},
		{
			name: "With derived output path",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: filepath.Join("designs", "test_design.scad"),
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join(tempDir, "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(tempDir, "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(tempDir, "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(tempDir, "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(tempDir, "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(tempDir, "export", "v1.0", "test_design", "report.html"),
			},
		},
		{
			name: "With absolute input path",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(inputPath), "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "report.html"),
			},
		},
		{
			name: "Relative config file path - should be relative to config location",
			config: models.Config{
				ConfigFile: filepath.Join("some", "nested", "path", "config.toml"),
				Design: models.DesignConfig{
					InputPath: filepath.Join("designs", "test_design.scad"),
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "test_design", "report.html"),
			},
		},
		{
			name: "Running from parent directory - ../../openscadgen -c ./folder/config.toml",
			config: models.Config{
				ConfigFile: filepath.Join("folder", "config.toml"),
				Design: models.DesignConfig{
					InputPath: "test_design.scad",
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join("folder", "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join("folder", "export", "v1.0"),
				LowQualityWarningPath: filepath.Join("folder", "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join("folder", "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join("folder", "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join("folder", "export", "v1.0", "test_design", "report.html"),
			},
		},
		{
			name: "Running from same directory - ./openscadgen -c ./folder/config.toml",
			config: models.Config{
				ConfigFile: filepath.Join(".", "folder", "config.toml"),
				Design: models.DesignConfig{
					InputPath: "test_design.scad",
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join(".", "folder", "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(".", "folder", "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(".", "folder", "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(".", "folder", "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(".", "folder", "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(".", "folder", "export", "v1.0", "test_design", "report.html"),
			},
		},
		{
			name: "Absolute config path - ./openscadgen -c /absolute/path/config.toml",
			config: models.Config{
				ConfigFile: "/absolute/path/config.toml",
				Design: models.DesignConfig{
					InputPath: "test_design.scad",
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join("/absolute/path", "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join("/absolute/path", "export", "v1.0"),
				LowQualityWarningPath: filepath.Join("/absolute/path", "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join("/absolute/path", "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join("/absolute/path", "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join("/absolute/path", "export", "v1.0", "test_design", "report.html"),
			},
		},
		{
			name: "Server mode - should follow same path logic",
			config: models.Config{
				ConfigFile: filepath.Join("folder", "config.toml"),
				Design: models.DesignConfig{
					InputPath: "test_design.scad",
					Version:   "v1.0",
				},
				Server: true,
				Debug:  false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join("folder", "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join("folder", "export", "v1.0"),
				LowQualityWarningPath: filepath.Join("folder", "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join("folder", "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join("folder", "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join("folder", "export", "v1.0", "test_design", "report.html"),
			},
		},
		/*{
			name: "With absolute input path",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:           filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(inputPath), "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "report.html"),
			},
		},*/
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
			result := getOutputPaths(&tc.config)

			// Cannot compare log paths directly due to timestamp, so check other fields
			if tc.name == "with_specified_output_path" {

				// Check other fields
				if normalizePath(result.OutputPath) != normalizePath(tc.expected.OutputPath) {
					t.Errorf("OutputPath = %s; want %s", result.OutputPath, tc.expected.OutputPath)
				}
				if normalizePath(result.ExportFolderPath) != normalizePath(tc.expected.ExportFolderPath) {
					t.Errorf("ExportFolderPath = %s; want %s", result.ExportFolderPath, tc.expected.ExportFolderPath)
				}
				if normalizePath(result.LowQualityWarningPath) != normalizePath(tc.expected.LowQualityWarningPath) {
					t.Errorf("LowQualityWarningPath = %s; want %s", result.LowQualityWarningPath, tc.expected.LowQualityWarningPath)
				}
				if normalizePath(result.ReadmePath) != normalizePath(tc.expected.ReadmePath) {
					t.Errorf("ReadmePath = %s; want %s", result.ReadmePath, tc.expected.ReadmePath)
				}
			} else {
				// For other cases, we can compare all fields directly
				if normalizePath(result.OutputPath) != normalizePath(tc.expected.OutputPath) {
					t.Errorf("OutputPath = %s; want %s", result.OutputPath, tc.expected.OutputPath)
				}
				if normalizePath(result.ExportFolderPath) != normalizePath(tc.expected.ExportFolderPath) {
					t.Errorf("ExportFolderPath = %s; want %s", result.ExportFolderPath, tc.expected.ExportFolderPath)
				}
				if normalizePath(result.LowQualityWarningPath) != normalizePath(tc.expected.LowQualityWarningPath) {
					t.Errorf("LowQualityWarningPath = %s; want %s", result.LowQualityWarningPath, tc.expected.LowQualityWarningPath)
				}
				if normalizePath(result.ReadmePath) != normalizePath(tc.expected.ReadmePath) {
					t.Errorf("ReadmePath = %s; want %s", result.ReadmePath, tc.expected.ReadmePath)
				}
				if normalizePath(result.LogOutputPath) != normalizePath(tc.expected.LogOutputPath) {
					t.Errorf("LogOutputPath = %s; want %s", result.LogOutputPath, tc.expected.LogOutputPath)
				}
				if normalizePath(result.ReportPath) != normalizePath(tc.expected.ReportPath) {
					t.Errorf("ReportPath = %s; want %s", result.ReportPath, tc.expected.ReportPath)
				}
			}
		})
	}
}

func TestLoadConfigFromFile_InvalidToml_ReturnsFatalErrThird(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.toml")
	// Unclosed table header — typical "expected '='" style decode failure
	content := "l1\nl2\nl3\nl4\nl5\nl6\n[openscadgen"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, warn, err := LoadConfigFromFile(models.CmdFlags{ConfigFile: path})
	if cfg != nil {
		t.Fatalf("expected nil config, got %+v", cfg)
	}
	if warn != nil {
		t.Fatalf("expected nil warning for fatal load, got %v", warn)
	}
	if err == nil {
		t.Fatal("expected fatal error")
	}
	if !strings.Contains(err.Error(), "not valid toml") {
		t.Fatalf("error should mention toml: %v", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("error should include path: %v", err)
	}
	if !strings.Contains(err.Error(), "Source context:") {
		t.Fatalf("error should include source context: %v", err)
	}
}

func TestLoadConfigFromFile_DirectoryArg_UsesConfigToml(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my_project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(projectDir, "config.toml")
	content := "[openscadgen]\nname = \"dir_arg_test\"\nexport_name_format = \"{designFileName}\"\n"
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := LoadConfigFromFile(models.CmdFlags{ConfigFile: projectDir})
	if err != nil {
		t.Fatalf("LoadConfigFromFile: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.ConfigFile != configPath {
		t.Fatalf("ConfigFile = %q; want %q", cfg.ConfigFile, configPath)
	}
	if cfg.Design.Name != "dir_arg_test" {
		t.Fatalf("Design.Name = %q", cfg.Design.Name)
	}
}

func TestTomlDecodeErrorSnippet(t *testing.T) {
	data := "a\nb\nc\nd\ne\nf\ng\n"
	decodeErr := fmt.Errorf(`toml: line 4 (last key "x"): bogus`)
	snip := tomlDecodeErrorSnippet(data, decodeErr)
	if !strings.Contains(snip, "   4 |") || !strings.Contains(snip, ">") {
		t.Fatalf("expected numbered context with marker, got:\n%s", snip)
	}
}

func TestProcess_RecordsFailedInstanceRunDetails(t *testing.T) {
	tmpDir := t.TempDir()
	if err := InitLogger(filepath.Join(tmpDir, "test.log")); err != nil {
		t.Fatalf("InitLogger: %v", err)
	}

	versionProbePath := filepath.Join(tmpDir, "openscad")
	if err := os.WriteFile(versionProbePath, []byte("#!/bin/sh\nif [ \"$1\" = \"--version\" ]; then\n  echo 'OpenSCAD 2024.01'\n  exit 0\nfi\necho 'openscad exploded' >&2\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write version probe: %v", err)
	}
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	scadPath := filepath.Join(tmpDir, "shape.scad")
	if err := os.WriteFile(scadPath, []byte("cube(1);\n"), 0o644); err != nil {
		t.Fatalf("write scad: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `[openscadgen]
name = "failure-case"
input_path = "shape.scad"
version = "v0.1"
export_name_format = "{designFileName}_{name}"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, _, err := LoadConfigFromFile(models.CmdFlags{
		ConfigFile: configPath,
	})
	if err != nil {
		t.Fatalf("LoadConfigFromFile: %v", err)
	}

	result, err := Process(config, &NoopProgress{}, nil, Operations{GenerateReport: false}, false)
	if err != nil {
		t.Fatalf("Process returned unexpected error: %v", err)
	}
	if len(result.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(result.Instances))
	}

	instance := result.Instances[0]
	if !instance.IsComplete {
		t.Fatal("expected failed instance to still be marked complete")
	}
	if instance.IsSuccessful {
		t.Fatal("expected failed instance to be marked unsuccessful")
	}
	if instance.ConfigError == "" {
		t.Fatal("expected failed instance to expose a summary error")
	}
	if len(instance.STLResults) != 1 {
		t.Fatalf("expected failed instance to keep STL result details, got %d", len(instance.STLResults))
	}
	if !strings.Contains(instance.STLResults[0].OutputLog, "openscad exploded") {
		t.Fatalf("expected STL output log to contain command output, got %q", instance.STLResults[0].OutputLog)
	}
	if len(result.STLResults) != 1 {
		t.Fatalf("expected 1 STL result, got %d", len(result.STLResults))
	}
}

func TestGenerateInstanceConfigs_NilConfig(t *testing.T) {
	_, err := GenerateInstanceConfigs(nil)
	if err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("expected nil config error, got %v", err)
	}
}

func TestGenerateDynamicInstances(t *testing.T) {
	t.Skip("Skipping unit tests - implementation details have changed, E2E tests cover functionality")
	// Initialize logger for testing
	err := InitLogger("memory")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

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

	// Create test input file
	inputPath := filepath.Join(tempDir, "test_design.scad")
	if err := os.WriteFile(inputPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Helper function to create a map of expected instances for easier lookup
	createExpectedInstancesMap := func(expectedParams []map[string]interface{}) map[string]map[string]interface{} {
		expectedMap := make(map[string]map[string]interface{})
		for _, params := range expectedParams {
			// Create a unique key based on the parameters
			key := fmt.Sprintf("%v", params)
			expectedMap[key] = params
		}
		return expectedMap
	}

	// Helper function to find a matching expected instance
	findMatchingExpectedInstance := func(instance models.InstanceConfig, expectedMap map[string]map[string]interface{}) (map[string]interface{}, bool) {
		// Create a key from the instance parameters
		instanceKey := fmt.Sprintf("%v", instance.Params)
		if expected, exists := expectedMap[instanceKey]; exists {
			return expected, true
		}
		return nil, false
	}

	testCases := []struct {
		name                string
		config              models.Config
		expectedParams      []map[string]interface{}
		expectedOutputPaths []string
	}{
		{
			name: "Single instance with no params",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name:   "default",
							Params: make(map[string]interface{}),
						},
					},
					ExportNameFormat: "{designFileName}_{version}_name_{instanceName}",
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"name":           "default",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0/test_design_v1.0_name_default.stl"),
			},
		},
		{
			name: "Multiple parameter combinations",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"width":  "10,20,30",
								"height": "5,15",
							},
						},
					},
					ExportNameFormat: "{designFileName}_{version}_name_{instanceName}_{width}_{height}",
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"height":         float64(5),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"height":         float64(15),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"height":         float64(5),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"height":         float64(15),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(30),
					"height":         float64(5),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(30),
					"height":         float64(15),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_10_15.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_20_5.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_20_15.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_30_15.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_10_5.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_30_5.stl"),
			},
		},
		{
			name: "Boolean values",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"enabled": "true,false",
							},
						},
					},
					ExportNameFormat: "{designFileName}_{version}_name_{instanceName}_{enabled}",
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"enabled":        true,
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"enabled":        false,
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_true.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_false.stl"),
			},
		},
		{
			name: "String values",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{type}",
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"type": "small,medium,large",
							},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"type":           "small",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"type":           "medium",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"type":           "large",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_small.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_medium.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_large.stl"),
			},
		},
		{
			name: "ignored_parameters - global param ignored",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{width}",
					GlobalParams: map[string]interface{}{
						"ignored_param_2": "4,5,6",
					},
					InputPaths: []models.InputPath{
						{
							Path:                       inputPath,
							IgnoreParamsWhenProcessing: "ignored_param_2",
						},
					},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"width": "10,20",
							},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_name_test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_name_test.stl"),
			},
		},
		{
			name: "ignored_parameters - input path params ignored",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{width}",
					InputPaths: []models.InputPath{
						{
							Path:                       inputPath,
							IgnoreParamsWhenProcessing: "ignored",
						},
					},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"ignored": "1,2,3",
								"width":   "10,20",
							},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_name_test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_name_test.stl"),
			},
		},
		{
			name: "global_parameters",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_global_${global}_global2_${global2}_name_${instanceName}",
					GlobalParams: map[string]interface{}{
						"global":  "value,value2",
						"global2": "value3,value4",
					},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"width": "10,20",
							},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"global":         "value",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"global":         "value",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"global":         "value",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"global":         "value",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"global":         "value2",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"global":         "value2",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"global":         "value2",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"global":         "value2",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value_global2_$value4_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value2_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value2_global2_$value4_name_$test.stl"),
			},
		},
		{
			name: "with_param_set_reference",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{width}_{height}",
					ParamSets: []models.ParamSet{
						{
							Name: "param_set_test",
							Params: map[string]interface{}{
								"width":  10,
								"height": 20,
							},
						},
						{
							Name: "param_set_test2",
							Params: map[string]interface{}{
								"width":  30,
								"height": 40,
							},
						},
					},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name:      "test",
							ParamSets: "param_set_test",
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"height":         float64(20),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_test_10_20.stl"),
			},
		},
		{
			name: "With Multiple Param Sets",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{instanceName}",
					ParamSets: []models.ParamSet{
						{
							Name: "base_params",
							Params: map[string]interface{}{
								"width":  10,
								"height": 20,
							},
						},
						{
							Name: "style_params",
							Params: map[string]interface{}{
								"color":   "red",
								"texture": "smooth",
							},
						},
					},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name:      "test",
							ParamSets: "base_params,style_params",
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"height":         float64(20),
					"color":          "red",
					"texture":        "smooth",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0/test_design_v1.0_test.stl"),
			},
		},
		{
			name: "with_param_set_types",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{instanceName}",
					ParamSets: []models.ParamSet{
						{
							Name: "mixed_params",
							Params: map[string]interface{}{
								"count":   5,
								"enabled": true,
								"type":    "test",
							},
						},
					},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name:      "test",
							ParamSets: "mixed_params",
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"count":          float64(5),
					"enabled":        true,
					"type":           "test",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_name_test.stl"),
			},
		},
		{
			name: "with_2_global_parameters",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_global_${global}_global2_${global2}_name_${instanceName}",
					GlobalParams: map[string]interface{}{
						"global":  "value,value2",
						"global2": "value3,value4",
					},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name:      "test",
							ParamSets: "global",
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"global":         "value",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"global":         "value2",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"global":         "value",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"global":         "value2",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value_global2_$value4_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value2_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value2_global2_$value4_name_$test.stl"),
			},
		},
		{
			name: "global param with comma-separated values",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_global_${global}_global2_${global2}_name_${instanceName}",
					GlobalParams:     map[string]interface{}{"foo": "a, b, c"},
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name:      "test",
							ParamSets: "global",
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"global":         "value",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"global":         "value2",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"global":         "value",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design",
					"global":         "value2",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value_global2_$value4_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value2_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_global_$value2_global2_$value4_name_$test.stl"),
			},
		},
		{
			name: "param_numberation_keys creates numbered keys",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}_{version}_name_{instanceName}",
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name:             "test",
							ParamsNumberated: map[string]interface{}{"foo": "a,b,c"},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"foo1":           "a",
					"foo2":           "b",
					"foo3":           "c",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1.0", "test_design_v1.0_name_test.stl"),
			},
		},
		{
			name:   "global-and-instance config file - all instance",
			config: models.Config{}, // Will be loaded in test
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "tree",
					"name":           "all",
					"version":        "v0.1",
				},
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "flag",
					"name":           "all",
					"version":        "v0.1",
				},
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "star",
					"name":           "all",
					"version":        "v0.1",
				},
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "circle",
					"name":           "all",
					"version":        "v0.1",
				},
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "triangle",
					"name":           "all",
					"version":        "v0.1",
				},
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "diamond",
					"name":           "all",
					"version":        "v0.1",
				},
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "square",
					"name":           "all",
					"version":        "v0.1",
				},
				{
					"designFileName": "own_your_place_tokens",
					"connector_type": "hexagon",
					"name":           "all",
					"version":        "v0.1",
				},
			},
			expectedOutputPaths: []string{}, // Will be validated based on export_name_format
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Special handling for config file loading test
			config := tc.config
			testInputPath := inputPath
			if tc.name == "global-and-instance config file - all instance" {
				configFilePath := "/Users/gregc/mine/making/3d-printing/openSCAD/openscadgen/tests/processor_test/config-inputs/global-and-instance/config.toml"
				loadedConfig, _, err := LoadConfigFromFile(models.CmdFlags{ConfigFile: configFilePath})
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				// Create the scad file if it doesn't exist
				configDir := filepath.Dir(configFilePath)
				scadPath := filepath.Join(configDir, "own_your_place_tokens.scad")
				if _, err := os.Stat(scadPath); os.IsNotExist(err) {
					if err := os.WriteFile(scadPath, []byte("// test file"), 0644); err != nil {
						t.Fatalf("Failed to create scad file: %v", err)
					}
				}
				// Update config to use absolute paths
				loadedConfig.ConfigFile = configFilePath
				if len(loadedConfig.Design.InputPaths) > 0 {
					absScadPath, _ := filepath.Abs(scadPath)
					loadedConfig.Design.InputPaths[0].Path = absScadPath
					testInputPath = absScadPath
				}
				// Find the "all" instance config
				var allInstanceConfig *models.ConfiguredInstanceConfig
				for i := range loadedConfig.Design.ConfiguredInstanceConfig {
					if loadedConfig.Design.ConfiguredInstanceConfig[i].Name == "all" {
						allInstanceConfig = &loadedConfig.Design.ConfiguredInstanceConfig[i]
						break
					}
				}
				if allInstanceConfig == nil {
					t.Fatalf("Could not find 'all' instance config in loaded config")
				}
				// Override config and use the "all" instance
				config = *loadedConfig
				config.Design.ConfiguredInstanceConfig = []models.ConfiguredInstanceConfig{*allInstanceConfig}
			}

			if len(config.Design.ConfiguredInstanceConfig) > 1 {
				t.Error("Only one dynamic instance config is supported for testing")
			}
			instances, exportLocation, err := GenerateInstances(&config, config.Design.ConfiguredInstanceConfig[0], models.InputPath{Path: testInputPath})
			if err != nil {
				t.Errorf("Error generating instances: %v", err)
			}

			if len(instances) != len(tc.expectedParams) {
				t.Errorf("Expected %d instances, got %d", len(tc.expectedParams), len(instances))
			}
			log.Printf("exportLocaiton: %s", exportLocation)

			// Create a map of expected instances for easier lookup
			expectedMap := createExpectedInstancesMap(tc.expectedParams)

			// Track used PartIDLetters to ensure uniqueness
			//usedPartIDLetters := make(map[string]bool)

			// Track which expected instances have been matched
			matchedExpected := make(map[string]bool)

			for _, instance := range instances {
				// Find matching expected instance
				expectedParams, found := findMatchingExpectedInstance(instance, expectedMap)
				if !found {
					t.Errorf("Unexpected instance with params: %+v", instance.Params)
					continue
				}

				// Mark this expected instance as matched
				expectedKey := fmt.Sprintf("%v", expectedParams)
				matchedExpected[expectedKey] = true

				// Compare map lengths
				if len(instance.Params) != len(expectedParams) {
					t.Errorf("Instance params length mismatch:\nExpected: %d\nGot: %d\nParams: %+v",
						len(expectedParams), len(instance.Params), instance.Params)
					continue
				}

				// Compare each key-value pair
				for k, expectedValue := range expectedParams {
					actualValue, exists := instance.Params[k]
					if !exists {
						t.Errorf("Instance missing expected key %q", k)
						continue
					}
					if !reflect.DeepEqual(actualValue, expectedValue) {
						t.Errorf("Instance param %q mismatch:\nExpected: %v (%T)\nGot: %v (%T)",
							k, expectedValue, expectedValue, actualValue, actualValue)
					}
				}

				// Verify PartIDLetter is set and unique
				//if instance.PartIDLetter == "" {
				//	t.Errorf("Instance has no PartIDLetter")
				//}
				//	if usedPartIDLetters[instance.PartIDLetter] {
				//		t.Errorf("Duplicate PartIDLetter found: %s", instance.PartIDLetter)
				//	}
				//	usedPartIDLetters[instance.PartIDLetter] = true

				// Verify InputPath is set correctly
				if instance.InputPath.Path != testInputPath {
					t.Errorf("Instance InputPath mismatch:\nExpected: %s\nGot: %s", testInputPath, instance.InputPath.Path)
				}
			}

			// Check if all expected instances were matched
			for key, params := range expectedMap {
				if !matchedExpected[key] {
					t.Errorf("Expected instance not found: %+v", params)
				}
			}

			// Check output paths (skip if expectedOutputPaths is empty)
			if len(tc.expectedOutputPaths) > 0 {
				if len(instances) != len(tc.expectedOutputPaths) {
					t.Errorf("Expected %d output paths, got %d instances", len(tc.expectedOutputPaths), len(instances))
				}

				// Create a map of expected paths for easier lookup
				expectedPaths := make(map[string]bool)
				for _, path := range tc.expectedOutputPaths {
					expectedPaths[filepath.Clean(path)] = true
				}

				// Check each instance's output path
				for _, instance := range instances {
					cleanPath := filepath.Clean(instance.OutputPathV2)
					if !expectedPaths[cleanPath] {
						t.Errorf("Unexpected output path: \n\n\t%s\nExpected one of:\n\t%s",
							instance.OutputPathV2,
							strings.Join(tc.expectedOutputPaths, "\n\t"))
					}
				}
			}

			// Verify OutputPathV2 is set correctly
			if len(instances) == 0 {
				t.Errorf("No instances were generated")
				return
			}
			instance := instances[0]
			if instance.OutputPathV2 == "" {
				t.Errorf("OutputPathV2 is empty")
			}
			// Skip format check for config file test case
			if tc.name != "global-and-instance config file - all instance" {
				if !strings.Contains(instance.OutputPathV2, "test_design_v1.0_name_default.stl") {
					t.Errorf("OutputPathV2 does not contain correct format: %s", instance.OutputPathV2)
				}
			}

			// Verify RunOutputPathV3 is set correctly and is relative to config.toml
			if instance.RunOutputPathV3 == "" {
				t.Errorf("RunOutputPathV3 is empty")
			}
			// Skip relative path check for config file test case
			if tc.name != "global-and-instance config file - all instance" {
				expectedRelPath := filepath.Join("export", "v1.0", "test_design_v1.0_name_default.stl")
				if normalizePath(instance.RunOutputPathV3) != normalizePath(expectedRelPath) {
					t.Errorf("RunOutputPathV3 = %s; want %s", instance.RunOutputPathV3, expectedRelPath)
				}
			}
		})
	}
}

func TestGenerateParamCombinations(t *testing.T) {

	// Test case 2: Single parameter with multiple values
	t.Run("Single parameter", func(t *testing.T) {
		paramCombos := map[string]interface{}{
			"color": "red,blue,green",
		}
		result, err := convertToParamCombinations(paramCombos, map[string]bool{}, []string{})
		if err != nil {
			t.Errorf("Error converting parameters: %v", err)
		}

		// Generate combinations from the parameter combinations
		combinations := generateCombinations(result)

		expectedCount := 3 // 3 values for the single parameter
		if len(combinations) != expectedCount {
			t.Errorf("Expected %d combinations, got %d", expectedCount, len(combinations))
		}

		// Check that each combination has the correct structure
		for i, combo := range combinations {
			if len(combo) != 1 {
				t.Errorf("Combination %d should have exactly 1 parameter, got %d", i, len(combo))
			}

			value, exists := combo["color"]
			if !exists {
				t.Errorf("Combination %d missing 'color' parameter", i)
			}

			expectedValues := map[string]bool{"red": true, "blue": true, "green": true}
			if !expectedValues[fmt.Sprintf("%v", value)] {
				t.Errorf("Unexpected value '%v' in combination %d", value, i)
			}
		}
	})

	// Test case 3: Multiple parameters with multiple values
	t.Run("Multiple parameters", func(t *testing.T) {
		paramCombos := map[string]interface{}{
			"color": "red,blue",
			"size":  "small,large",
			"shape": "circle,square",
		}
		result, err := convertToParamCombinations(paramCombos, map[string]bool{}, []string{})
		if err != nil {
			t.Errorf("Error converting parameters: %v", err)
		}

		// Generate combinations from the parameter combinations
		combinations := generateCombinations(result)

		expectedCount := 8 // 2 x 2 x 2 = 8 combinations
		if len(combinations) != expectedCount {
			t.Errorf("Expected %d combinations, got %d", expectedCount, len(combinations))
		}

		// Check that each combination has the correct structure
		for i, combo := range combinations {
			if len(combo) != 3 {
				t.Errorf("Combination %d should have exactly 3 parameters, got %d", i, len(combo))
			}

			// Check that each parameter exists
			for _, param := range []string{"color", "size", "shape"} {
				if _, exists := combo[param]; !exists {
					t.Errorf("Combination %d missing '%s' parameter", i, param)
				}
			}
		}

		// Check that all combinations are unique
		combinationsMap := make(map[string]bool)
		for _, combo := range combinations {
			key := fmt.Sprintf("%v-%v-%v", combo["color"], combo["size"], combo["shape"])
			if combinationsMap[key] {
				t.Errorf("Duplicate combination found: %s", key)
			}
			combinationsMap[key] = true
		}

		// Check that we have all expected combinations
		expectedCombinations := []string{
			"red-small-circle", "red-small-square", "red-large-circle", "red-large-square",
			"blue-small-circle", "blue-small-square", "blue-large-circle", "blue-large-square",
		}
		for _, expected := range expectedCombinations {
			parts := strings.Split(expected, "-")
			found := false
			for _, combo := range combinations {
				if fmt.Sprintf("%v", combo["color"]) == parts[0] &&
					fmt.Sprintf("%v", combo["size"]) == parts[1] &&
					fmt.Sprintf("%v", combo["shape"]) == parts[2] {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected combination not found: %s", expected)
			}
		}
	})

	// Test case 4: Parameters with empty values

}

func TestPathResolution(t *testing.T) {
	t.Skip("Skipping path resolution test - implementation details have changed")
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "openscadgen_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	subfolder := filepath.Join(tempDir, "subfolder")
	if err := os.MkdirAll(subfolder, 0755); err != nil {
		t.Fatalf("Failed to create subfolder: %v", err)
	}

	// Create test files
	configContent := `[openscadgen]
name = "Test Design"
description = "Test design for path resolution"
input_path = "design.scad"
version = "v1.0"
export_name_format = "{designFileName}_{version}_name_{instanceName}"

[[openscadgen.instances]]
name = "default"

[[openscadgen.images]]
name = "side"
coord = "0,0,0,90,0,0,600"
`
	configPath := filepath.Join(subfolder, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	designContent := `cube([10, 10, 10]);`
	designPath := filepath.Join(subfolder, "design.scad")
	if err := os.WriteFile(designPath, []byte(designContent), 0644); err != nil {
		t.Fatalf("Failed to create design file: %v", err)
	}

	// Test cases for different working directories
	testCases := []struct {
		name           string
		workingDir     string
		configPath     string
		expectedExport string
		expectedSTLs   []string
		expectedImages []string
	}{
		{
			name:           "Run from parent directory",
			workingDir:     tempDir,
			configPath:     "subfolder/config.toml",
			expectedExport: filepath.Join(subfolder, "export", "v1.0"),
			expectedSTLs:   []string{filepath.Join(subfolder, "export", "v1.0", "test_design_v1.0_name_default.stl")},
			expectedImages: []string{filepath.Join(subfolder, "export", "v1.0", "test_design_v1.0_name_default.png")},
		},
		{
			name:           "Run from subfolder directory",
			workingDir:     subfolder,
			configPath:     "config.toml",
			expectedExport: filepath.Join("export", "v1.0"),
			expectedSTLs:   []string{filepath.Join(subfolder, "export", "v1.0", "test_design_v1.0_name_default.stl")},
			expectedImages: []string{filepath.Join(subfolder, "export", "v1.0", "test_design_v1.0_name_default.png")},
		},
		{
			name:           "Run with absolute path",
			workingDir:     tempDir,
			configPath:     filepath.Join(subfolder, "config.toml"),
			expectedExport: filepath.Join(subfolder, "export", "v1.0"),
			expectedSTLs:   []string{filepath.Join(subfolder, "export", "v1.0", "test_design_v1.0_name_default.stl")},
			expectedImages: []string{filepath.Join(subfolder, "export", "v1.0", "test_design_v1.0_name_default.png")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save current working directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(originalDir)

			// Change to test working directory
			if err := os.Chdir(tc.workingDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			// Initialize logger first
			if err := InitLogger(filepath.Join(tc.workingDir, "test.log")); err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			// Load config
			config, _, err := LoadConfigFromFile(models.CmdFlags{
				ConfigFile: tc.configPath,
			})
			if err != nil {
				LogKeyValuePair("Config file", tc.configPath)
				LogKeyValuePair("Config file content", configContent)
				t.Fatalf("Failed to load config: %v", err)
			}

			if len(config.GetInputPaths()) == 0 {
				LogKeyValuePair("No input paths found in config", "")
				LogWarnWithCritical("No input paths found in config", true)
				t.Fatalf("No input paths found in config")
			} else if config.Debug {
				LogKeyValuePair(fmt.Sprintf("%d Input paths", len(config.Design.InputPaths)), "")
				for _, inputPath := range config.Design.InputPaths {
					LogKeyValuePair("", inputPath.Path)
				}
			}

			// Get output paths
			/*	outputPaths := getOutputPaths(config)

				// Verify export path
				expectedPath := filepath.Join(tc.workingDir, tc.expectedExport)
				if normalizePath(outputPaths.ExportFolderPath) != normalizePath(expectedPath) {
					t.Errorf("ExportFolderPath = %s; want %s", outputPaths.ExportFolderPath, expectedPath)
				}*/

			// Process the config
			result, err := Process(config, &NoopProgress{}, nil, Operations{
				GenerateReport: true,
			}, false)
			if err != nil {
				t.Fatalf("Failed to process config: %v", err)
			}

			if len(result.ImageResults) != len(tc.expectedImages) {
				t.Fatalf("Expected %d image results, got %d", len(tc.expectedImages), len(result.ImageResults))
			}

			if len(result.STLResults) != len(tc.expectedSTLs) {
				t.Fatalf("Expected %d STL results, got %d", len(tc.expectedSTLs), len(result.STLResults))
			}
			if len(result.STLResults) > 0 {
				for i, stlResult := range result.STLResults {
					if stlResult.OutputPath != tc.expectedSTLs[i] {
						t.Errorf("STL result output path = %s; want %s", stlResult.OutputPath, tc.expectedSTLs[i])
					}
				}
			}
			/*
				if len(config.Design.InputPaths) != 1 {
					t.Fatalf("Expected 1 input path, got %d", len(config.Design.InputPaths))
				}*/

			// Get instances from the config
			instances, exportLocation, err := GenerateInstances(config, config.Design.ConfiguredInstanceConfig[0], config.Design.InputPaths[0])
			if err != nil {
				t.Fatalf("Failed to generate instances: %v", err)
			}

			// Verify export directory was created
			/*	if _, err := os.Stat(outputPaths.ExportFolderPath); os.IsNotExist(err) {
					t.Errorf("Export directory was not created at %s", outputPaths.ExportFolderPath)
				}
			*/
			if _, err := os.Stat(exportLocation); os.IsNotExist(err) {
				t.Errorf("Export directory was not created at %s", exportLocation)
			}

			// Verify STL file was created with correct name
			expectedSTLName := "test_design_v1.0_name_default.stl"
			stlPath := filepath.Join(exportLocation, expectedSTLName)
			if _, err := os.Stat(stlPath); os.IsNotExist(err) {
				t.Errorf("STL file was not created at %s", stlPath)
			}

			// Verify input path is correct
			if config.Design.InputPath == "" {
				t.Errorf("Input path is empty")
			}

			// Verify the version format in the output path
			if !strings.Contains(exportLocation, "v1.0") {
				t.Errorf("Export folder path does not contain correct version format: %s", exportLocation)
			}

			// Verify OutputPathV2 is set correctly
			if len(instances) == 0 {
				t.Errorf("No instances were generated")
			}
			instance := instances[0]
			if instance.OutputPathV2 == "" {
				t.Errorf("OutputPathV2 is empty")
			}
			if !strings.Contains(instance.OutputPathV2, "test_design_v1.0_name_default.stl") {
				t.Errorf("OutputPathV2 does not contain correct format: %s", instance.OutputPathV2)
			}

			// Verify RunOutputPathV3 is set correctly and is relative to config.toml
			if instance.RunOutputPathV3 == "" {
				t.Errorf("RunOutputPathV3 is empty")
			}
			expectedRelPath := filepath.Join("export", "v1.0", "test_design_v1.0_name_default.stl")
			if normalizePath(instance.RunOutputPathV3) != normalizePath(expectedRelPath) {
				t.Errorf("RunOutputPathV3 = %s; want %s", instance.RunOutputPathV3, expectedRelPath)
			}
		})
	}
}

type Input struct {
	dynamicInstance models.ConfiguredInstanceConfig
	globalParams    map[string]interface{}
	paramSets       []models.ParamSet
	inputPath       models.InputPath
}

type Output struct {
	params          map[string]interface{}
	globalParamsMap map[string][]interface{}
	ignoredKeys     []string
}

type TestCase struct {
	name   string
	input  Input
	output Output
}

func TestGetAllParams(t *testing.T) {
	testCases := []TestCase{
		{
			name: "basic merge with no ignored params",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:      "inst1",
					Params:    map[string]interface{}{"foo": 1},
					ParamSets: "set1",
				},
				globalParams: map[string]interface{}{"bar": 2, "float": 1.5},
				paramSets: []models.ParamSet{
					{Name: "set1", Params: map[string]interface{}{"baz": 3}},
				},
				inputPath: models.InputPath{
					Params: map[string]interface{}{"qux": 4},
				},
			},
			output: Output{
				params: map[string]interface{}{
					"foo": float64(1),
					"baz": float64(3),
					"qux": float64(4),
				},
				globalParamsMap: map[string][]interface{}{
					"bar":   {float64(2)},
					"float": {float64(1.5)},
				},
				ignoredKeys: nil,
			},
		},
		{
			name: "global param overridden by param set",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:      "inst1",
					ParamSets: "set1",
				},
				globalParams: map[string]interface{}{"width": 10, "height": 20},
				paramSets: []models.ParamSet{
					{Name: "set1", Params: map[string]interface{}{"width": 15, "depth": 5}},
				},
				inputPath: models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"width": float64(15), // param set overrides global
					"depth": float64(5),  // param set only
				},
				globalParamsMap: map[string][]interface{}{
					"width":  {float64(10)},
					"height": {float64(20)},
				},
				ignoredKeys: nil,
			},
		},
		{
			name: "param set overridden by instance params",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:      "inst1",
					Params:    map[string]interface{}{"width": 25, "color": "red"},
					ParamSets: "set1",
				},
				globalParams: map[string]interface{}{"width": 10, "height": 20},
				paramSets: []models.ParamSet{
					{Name: "set1", Params: map[string]interface{}{"width": 15, "depth": 5}},
				},
				inputPath: models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"width": float64(25), // instance overrides param set
					"depth": float64(5),  // param set only
					"color": "red",       // instance only
				},
				globalParamsMap: map[string][]interface{}{
					"width":  {float64(10)},
					"height": {float64(20)},
				},
				ignoredKeys: nil,
			},
		},
		{
			name: "input path params override everything",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:      "inst1",
					Params:    map[string]interface{}{"width": 25, "color": "red"},
					ParamSets: "set1",
				},
				globalParams: map[string]interface{}{"width": 10, "height": 20},
				paramSets: []models.ParamSet{
					{Name: "set1", Params: map[string]interface{}{"width": 15, "depth": 5}},
				},
				inputPath: models.InputPath{
					Params: map[string]interface{}{"width": 30, "material": "plastic"},
				},
			},
			output: Output{
				params: map[string]interface{}{
					"width":    float64(30), // input path overrides instance
					"depth":    float64(5),  // param set only
					"color":    "red",       // instance only
					"material": "plastic",   // input path only
				},
				globalParamsMap: map[string][]interface{}{
					"width":  {float64(10)},
					"height": {float64(20)},
				},
				ignoredKeys: nil,
			},
		},
		{
			name: "multiple param sets with override hierarchy",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:      "inst1",
					Params:    map[string]interface{}{"quality": "high"},
					ParamSets: "base,style",
				},
				globalParams: map[string]interface{}{"size": "large", "color": "blue"},
				paramSets: []models.ParamSet{
					{Name: "base", Params: map[string]interface{}{"size": "medium", "weight": 100}},
					{Name: "style", Params: map[string]interface{}{"color": "red", "finish": "matte"}},
				},
				inputPath: models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"size":    "medium",     // base param set overrides global
					"weight":  float64(100), // base param set only
					"color":   "red",        // style param set overrides global
					"finish":  "matte",      // style param set only
					"quality": "high",       // instance overrides all
				},
				globalParamsMap: map[string][]interface{}{
					"size":  {"large"},
					"color": {"blue"},
				},
				ignoredKeys: nil,
			},
		},
		{
			name: "ignored parameters are excluded from all sources",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:      "inst1",
					Params:    map[string]interface{}{"width": 25, "ignored": "should_be_ignored"},
					ParamSets: "set1",
				},
				globalParams: map[string]interface{}{"width": 10, "ignored": "also_ignored"},
				paramSets: []models.ParamSet{
					{Name: "set1", Params: map[string]interface{}{"width": 15, "ignored": "ignored_too"}},
				},
				inputPath: models.InputPath{
					Params:                     map[string]interface{}{"width": 30, "ignored": "still_ignored"},
					IgnoreParamsWhenProcessing: "ignored",
				},
			},
			output: Output{
				params: map[string]interface{}{
					"width": float64(30), // only non-ignored params remain
				},
				globalParamsMap: map[string][]interface{}{
					"width": {float64(10)},
				},
				ignoredKeys: []string{"ignored"},
			},
		},
		{
			name: "string values with comma separation",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:   "inst1",
					Params: map[string]interface{}{"colors": "red,blue,green"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"colors": []interface{}{"red", "blue", "green"},
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "string values with comma separation",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:                "inst1",
					Params:              map[string]interface{}{"colors": "red,blue,green"},
					IgnoreCommaInParams: []string{"colors"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"colors": "red,blue,green",
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "global params with comma separation",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{},
				globalParams:    map[string]interface{}{"sizes": "small,medium,large"},
				paramSets:       []models.ParamSet{},
				inputPath:       models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{},
				globalParamsMap: map[string][]interface{}{
					"sizes": {"small", "medium", "large"},
				},
				ignoredKeys: nil,
			},
		},
		{
			name: "global params as array (TOML slice)",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{},
				globalParams: map[string]interface{}{
					"fonts": []interface{}{"Liberation Mono", "Liberation Sans"},
				},
				paramSets: []models.ParamSet{},
				inputPath: models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{},
				globalParamsMap: map[string][]interface{}{
					"fonts": {"Liberation Mono", "Liberation Sans"},
				},
				ignoredKeys: nil,
			},
		},
		{
			name: "param_numberation_keys creates numbered keys",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					ParamsNumberated: map[string]interface{}{"foo": "a,b,c"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"foo1": "a",
					"foo2": "b",
					"foo3": "c",
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "complete override hierarchy with all parameter types",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:      "inst1",
					Params:    map[string]interface{}{"width": 25, "enabled": true, "name": "instance_name"},
					ParamSets: "base,advanced",
				},
				globalParams: map[string]interface{}{
					"width":   10,
					"height":  20,
					"enabled": false,
					"name":    "global_name",
				},
				paramSets: []models.ParamSet{
					{
						Name: "base",
						Params: map[string]interface{}{
							"width":   15,
							"depth":   5,
							"enabled": true,
							"name":    "base_name",
						},
					},
					{
						Name: "advanced",
						Params: map[string]interface{}{
							"quality": "high",
							"name":    "advanced_name",
						},
					},
				},
				inputPath: models.InputPath{
					Params: map[string]interface{}{
						"width":    30,
						"material": "steel",
						"name":     "input_name",
					},
				},
			},
			output: Output{
				params: map[string]interface{}{
					"width":    float64(30),  // input path overrides all
					"depth":    float64(5),   // base param set only
					"enabled":  true,         // instance overrides base param set
					"quality":  "high",       // advanced param set only
					"material": "steel",      // input path only
					"name":     "input_name", // input path overrides all
				},
				globalParamsMap: map[string][]interface{}{
					"width":   {float64(10)},
					"height":  {float64(20)},
					"enabled": {false},
					"name":    {"global_name"},
				},
				ignoredKeys: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, globalParamsMap, ignoredKeys := getAllParams(tc.input.dynamicInstance, tc.input.globalParams, tc.input.paramSets, tc.input.inputPath)
			if !reflect.DeepEqual(params, tc.output.params) {
				t.Errorf("params mismatch:\nExpected: %#v\nGot: %#v", tc.output.params, params)
			}
			if !reflect.DeepEqual(globalParamsMap, tc.output.globalParamsMap) {
				t.Errorf("globalParamsMap mismatch:\nExpected: %#v\nGot: %#v", tc.output.globalParamsMap, globalParamsMap)
			}
			if !reflect.DeepEqual(ignoredKeys, tc.output.ignoredKeys) {
				t.Errorf("ignoredKeys mismatch:\nExpected: %#v\nGot: %#v", tc.output.ignoredKeys, ignoredKeys)
			}
		})
	}
}

func TestScanFolderForConfigFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanconfigtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Valid config file
	validPath := filepath.Join(tmpDir, "valid.toml")
	os.WriteFile(validPath, []byte("[openscadgen]\nname = 'foo'\n"), 0644)

	// No tag
	noTagPath := filepath.Join(tmpDir, "notag.toml")
	os.WriteFile(noTagPath, []byte("name = 'bar'\n"), 0644)

	// Too large
	largePath := filepath.Join(tmpDir, "large.toml")
	large := make([]byte, 2*1024*1024+1)
	copy(large, []byte("[openscadgen]\n"))
	os.WriteFile(largePath, large, 0644)

	// Non-text (binary, but still .toml)
	binPath := filepath.Join(tmpDir, "bin.toml")
	os.WriteFile(binPath, []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	// Create export folder and add config files there (should be excluded)
	exportDir := filepath.Join(tmpDir, "export")
	os.MkdirAll(exportDir, 0755)
	exportConfigPath := filepath.Join(exportDir, "export_config.toml")
	os.WriteFile(exportConfigPath, []byte("[openscadgen]\nname = 'export_config'\n"), 0644)

	// Create nested export folder
	nestedExportDir := filepath.Join(tmpDir, "project", "export", "v1.0")
	os.MkdirAll(nestedExportDir, 0755)
	nestedExportConfigPath := filepath.Join(nestedExportDir, "nested_export_config.toml")
	os.WriteFile(nestedExportConfigPath, []byte("[openscadgen]\nname = 'nested_export_config'\n"), 0644)

	// Create another valid config file outside export
	anotherValidPath := filepath.Join(tmpDir, "another_valid.toml")
	os.WriteFile(anotherValidPath, []byte("[openscadgen]\nname = 'another_valid'\n"), 0644)

	files, err := ScanFolderForConfigFiles(tmpDir)
	if err != nil {
		t.Fatalf("ScanFolderForConfigFiles error: %v", err)
	}

	// Should only find the two valid config files outside export folders
	expectedPaths := []string{"valid.toml", "another_valid.toml"}
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d: %v", len(files), files)
	}

	// Check that export folder files are excluded
	for _, file := range files {
		if strings.Contains(file.Path, "export") {
			t.Errorf("Found file in export folder (should be excluded): %s", file.Path)
		}
	}

	// Verify we have the expected files
	foundPaths := make(map[string]bool)
	for _, file := range files {
		foundPaths[file.Path] = true
	}
	for _, expectedPath := range expectedPaths {
		if !foundPaths[expectedPath] {
			t.Errorf("Expected file not found: %s", expectedPath)
		}
	}
}

func TestInstanceConstruction(t *testing.T) {
	t.Skip("Skipping instance construction test - implementation details have changed")
	// Initialize logger for testing
	err := InitLogger("memory")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "instanceconstructiontest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Valid config file
	validConfigPath := filepath.Join(tmpDir, "config.toml")
	os.WriteFile(validConfigPath, []byte(fmt.Sprintf(`[openscadgen]
name = 'foo'

export_name_format = "football_cards_{name}"

[[openscadgen.input_paths]]
path = "%s/football_cards.scad"

[[openscadgen.images]]
name = "nice"
	`, tmpDir)), 0644)

	validScadPath := filepath.Join(tmpDir, "football_cards.scad")
	os.WriteFile(validScadPath, []byte(`
		cube(10);
		`), 0644)

	config, _, err := LoadConfigFromFile(models.CmdFlags{
		ConfigFile: validConfigPath,
	})
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	result, err := Process(config, &NoopProgress{}, nil, Operations{
		GenerateReport: true,
	}, false)
	if err != nil {
		t.Fatalf("Failed to process config: %v", err)
	}

	if len(result.ImageResults) != 1 {
		t.Fatalf("Expected 1 image result, got %d", len(result.ImageResults))
	}

	if result.ImageResults[0].CameraName != "nice" {
		t.Fatalf("Expected image result name to be 'nice', got %s", result.ImageResults[0].CameraName)
	}

	if result.ImageResults[0].OutputPath != "export/v0.1/nice.png" {
		t.Fatalf("Expected image result output path to be 'export/v0.1/nice.png', got %s", result.ImageResults[0].OutputPath)
	}

	// instance specific reference

	if len(result.Instances) != 1 {
		t.Fatalf("Expected 1 instance, got %d", len(result.Instances))
	}

	if result.Instances[0].ImageResults[0].CameraName != "nice" {
		t.Fatalf("Expected instance export image name to be 'nice', got %s", result.Instances[0].ExportImages[0].CameraName)
	}

	if result.Instances[0].ImageResults[0].OutputPath != fmt.Sprintf("%s/export/v0.1/nice.png", tmpDir) {
		t.Fatalf("Expected instance export image output path to be 'export/v0.1/nice.png', got %s", result.Instances[0].ImageResults[0].OutputPath)
	}

}

func TestParseCameraNameValidDirections(t *testing.T) {
	// Test cases for valid camera names that should result in non-empty direction strings
	testCases := []struct {
		cameraName        string
		expectedDirection string
		expectedDistance  string
	}{
		{"nice", "nice", ""},
		{"nice-near", "nice", "near"},
		{"top", "top", ""},
		{"nice-far", "nice", "far"},
		{"front-near", "front", "near"},
		{"top-far", "top", "far"},
	}

	for _, tc := range testCases {
		t.Run(tc.cameraName, func(t *testing.T) {
			direction, distance := parseCameraName(tc.cameraName)

			// Check that direction is not empty
			if direction == "" {
				t.Errorf("Camera name '%s' resulted in empty direction, expected '%s'", tc.cameraName, tc.expectedDirection)
			}

			// Check that direction matches expected
			if direction != tc.expectedDirection {
				t.Errorf("Camera name '%s' resulted in direction '%s', expected '%s'", tc.cameraName, direction, tc.expectedDirection)
			}

			// Check that distance matches expected
			if distance != tc.expectedDistance {
				t.Errorf("Camera name '%s' resulted in distance '%s', expected '%s'", tc.cameraName, distance, tc.expectedDistance)
			}
		})
	}
}

func TestPopulateExportImages_ParamFilterCommaList(t *testing.T) {
	tmpDir := t.TempDir()

	// Minimal config: two global renderType values create obj + all instances; param_filter should match only "obj".
	configContent := `[openscadgen]
name = "test_design"
version = "v1.0"
export_name_format = "{designFileName}_{renderType}"
global_params = { renderType = "obj,all" }

[[openscadgen.input_paths]]
path = "./design.scad"

[[openscadgen.instances]]
name = "default"

[[openscadgen.images]]
name = "nice"
param_filter = { renderType = "obj,vertSlice" }
`

	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config.toml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "design.scad"), []byte("cube(1);"), 0644); err != nil {
		t.Fatalf("Failed to write design.scad: %v", err)
	}

	config, _, err := LoadConfigFromFile(models.CmdFlags{ConfigFile: configPath})
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	instances, _, err := GenerateInstances(config, config.Design.ConfiguredInstanceConfig[0], config.Design.InputPaths[0])
	if err != nil {
		t.Fatalf("Failed to generate instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("Expected 2 instances (obj, all), got %d", len(instances))
	}

	instances, err = populateExportImages(config, instances)
	if err != nil {
		t.Fatalf("populateExportImages error: %v", err)
	}

	var objFound, allFound bool
	for _, inst := range instances {
		rt, _ := inst.Params["renderType"].(string)
		switch rt {
		case "obj":
			objFound = true
			if len(inst.ExportImages) == 0 {
				t.Fatalf("Expected obj instance to have export images")
			}
			if inst.SkippedImageReason != "" {
				t.Fatalf("Did not expect SkippedImageReason for obj instance, got: %s", inst.SkippedImageReason)
			}
		case "all":
			allFound = true
			if len(inst.ExportImages) != 0 {
				t.Fatalf("Expected all instance to have 0 export images due to param_filter")
			}
			if inst.SkippedImageReason == "" {
				t.Fatalf("Expected SkippedImageReason for all instance")
			}
		}
	}
	if !objFound || !allFound {
		t.Fatalf("Did not find both obj and all instances")
	}
}

func TestGenerateInstances_InstanceParamOverridesGlobalParam(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `[openscadgen]
name = "small_handle_extender"
version = "v0.5"
export_name_format = "{designFileName}_{iName}"
global_params = { clipStyle = "clip" }

[[openscadgen.input_paths]]
path = "./small_handle_extender.scad"

[[openscadgen.instances]]
params = { iName = "small", holderLength = 50, holeDepth = 25, clipStyle = "noClip", holderRadius = 6, holeRadius = 3.6 }

[[openscadgen.instances]]
params = { iName = "small-stubby-holder", wedgeOut = 4, holderLength = 15, holeDepth = 14, clipDownOffset = 6, clipInwardPush = 7, holderRadius = 14, holeRadius = 6.5 }
`

	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config.toml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "small_handle_extender.scad"), []byte("cube(1);"), 0644); err != nil {
		t.Fatalf("Failed to write scad file: %v", err)
	}

	config, _, err := LoadConfigFromFile(models.CmdFlags{ConfigFile: configPath})
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Design.ConfiguredInstanceConfig) != 2 {
		t.Fatalf("Expected 2 configured instances, got %d", len(config.Design.ConfiguredInstanceConfig))
	}

	firstInstances, _, err := GenerateInstances(config, config.Design.ConfiguredInstanceConfig[0], config.Design.InputPaths[0])
	if err != nil {
		t.Fatalf("GenerateInstances for first instance failed: %v", err)
	}
	if len(firstInstances) != 1 {
		t.Fatalf("Expected 1 generated instance for first config, got %d", len(firstInstances))
	}
	if got := firstInstances[0].Params["clipStyle"]; got != "noClip" {
		t.Fatalf("Expected first instance clipStyle to override global value, got %#v", got)
	}

	secondInstances, _, err := GenerateInstances(config, config.Design.ConfiguredInstanceConfig[1], config.Design.InputPaths[0])
	if err != nil {
		t.Fatalf("GenerateInstances for second instance failed: %v", err)
	}
	if len(secondInstances) != 1 {
		t.Fatalf("Expected 1 generated instance for second config, got %d", len(secondInstances))
	}
	if got := secondInstances[0].Params["clipStyle"]; got != "clip" {
		t.Fatalf("Expected second instance to inherit global clipStyle, got %#v", got)
	}
}

func TestHTMLProgressReport(t *testing.T) {
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

	tmpDir, err := os.MkdirTemp("", "htmlreporttest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test config
	config := &models.Config{
		Design: models.DesignConfig{
			Name:             "Test Design",
			Version:          "v1.0",
			OutputPath:       tmpDir,
			ExportNameFormat: "test_{name}",
		},
		Debug: true,
	}

	// Create test instances
	instances := []models.InstanceConfig{
		{
			Name:         "instance1",
			AutoName:     "instance1",
			OutputPathV2: filepath.Join(tmpDir, "export", "v1.0", "test_instance1.stl"),
			Params: map[string]interface{}{
				"width":  10,
				"height": 20,
			},
		},
		{
			Name:         "instance2",
			AutoName:     "instance2",
			OutputPathV2: filepath.Join(tmpDir, "export", "v1.0", "test_instance2.stl"),
			Params: map[string]interface{}{
				"width":  15,
				"height": 25,
			},
		},
	}

	// Create test STL results
	stlResults := []models.GenerateSTLResult{
		{
			InstanceConfig: instances[0],
			OutputPath:     filepath.Join(tmpDir, "export", "v1.0", "test_instance1.stl"),
			Command:        "openscad -o test_instance1.stl test.scad",
			TimeTaken:      time.Millisecond * 100,
		},
		{
			InstanceConfig: instances[1],
			OutputPath:     filepath.Join(tmpDir, "export", "v1.0", "test_instance2.stl"),
			Command:        "openscad -o test_instance2.stl test.scad",
			TimeTaken:      time.Millisecond * 150,
		},
	}

	// Create test image results
	imageResults := []models.GenerateImageResult{
		{
			InstanceConfig: instances[0],
			OutputPath:     filepath.Join(tmpDir, "export", "v1.0", "test_instance1_nice.png"),
			CameraName:     "nice",
			CameraCoords:   "0,0,0,0,0,0,100",
			Command:        "openscad -o test_instance1_nice.png test.scad",
			TimeTaken:      time.Millisecond * 50,
		},
		{
			InstanceConfig: instances[1],
			OutputPath:     filepath.Join(tmpDir, "export", "v1.0", "test_instance2_nice.png"),
			CameraName:     "nice",
			CameraCoords:   "0,0,0,0,0,0,100",
			Command:        "openscad -o test_instance2_nice.png test.scad",
			TimeTaken:      time.Millisecond * 75,
		},
	}

	// Generate HTML report
	htmlContent, outputFile, err := GenerateOutputReport(config, instances, stlResults, imageResults, tmpDir, true, time.Second*2)
	if err != nil {
		t.Fatalf("Failed to generate HTML report: %v", err)
	}

	// Verify HTML content was generated
	if htmlContent == nil {
		t.Fatal("HTML content is nil")
	}

	// Read the generated HTML file
	htmlBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}
	htmlContentStr := string(htmlBytes)

	// Test 1: Verify instance count in HTML
	expectedInstanceCount := len(instances)
	if !strings.Contains(htmlContentStr, fmt.Sprintf("Instances (%d)", expectedInstanceCount)) {
		t.Errorf("HTML should contain instance count %d, but content was: %s", expectedInstanceCount, htmlContentStr)
	}

	// Test 2: Verify STL results count in HTML
	expectedSTLCount := len(stlResults)
	if !strings.Contains(htmlContentStr, fmt.Sprintf("STLs (%d)", expectedSTLCount)) {
		t.Errorf("HTML should contain STL results count %d", expectedSTLCount)
	}

	// Test 3: Verify image results count in HTML
	expectedImageCount := len(imageResults)
	expectedText := fmt.Sprintf("Images (%d)", expectedImageCount)
	if !strings.Contains(htmlContentStr, expectedText) {
		t.Errorf("HTML should contain image results count %d, looking for: %s", expectedImageCount, expectedText)
	}

	// Test 4: Verify STL file paths are present
	for _, stlResult := range stlResults {
		relativePath := strings.TrimPrefix(stlResult.OutputPath, tmpDir)
		relativePath = strings.TrimPrefix(relativePath, "/")
		if !strings.Contains(htmlContentStr, relativePath) {
			t.Errorf("HTML should contain STL path: %s", relativePath)
		}
	}

	// Test 5: Verify image file paths are present
	for _, imageResult := range imageResults {
		relativePath := strings.TrimPrefix(imageResult.OutputPath, tmpDir)
		relativePath = strings.TrimPrefix(relativePath, "/")
		if !strings.Contains(htmlContentStr, relativePath) {
			t.Errorf("HTML should contain image path: %s", relativePath)
		}
	}

	// Test 6: Verify instance names are present
	for _, instance := range instances {
		if !strings.Contains(htmlContentStr, instance.Name) {
			t.Errorf("HTML should contain instance name: %s", instance.Name)
		}
	}

	// Test 7: Verify camera names are present
	for _, imageResult := range imageResults {
		if !strings.Contains(htmlContentStr, imageResult.CameraName) {
			t.Errorf("HTML should contain camera name: %s", imageResult.CameraName)
		}
	}

	// Test 8: Verify design name is present
	if !strings.Contains(htmlContentStr, config.Design.Name) {
		t.Errorf("HTML should contain design name: %s", config.Design.Name)
	}

	// Test 9: Verify total processing time is present
	if !strings.Contains(htmlContentStr, "Total processing time:") {
		t.Error("HTML should contain total processing time")
	}

}

// TestDirectArrayParams tests the new direct array parameters functionality
func TestDirectArrayParams(t *testing.T) {
	testCases := []TestCase{
		{
			name: "basic direct array parameter with string format",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:              "inst1",
					Params:            map[string]interface{}{"supports": "[[20,80], [110,80]]"},
					DirectArrayParams: []string{"supports"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"supports": []interface{}{
						[]interface{}{20.0, 80.0},
						[]interface{}{110.0, 80.0},
					},
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "direct array parameter with TOML array format",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:              "inst1",
					Params:            map[string]interface{}{"supports": []interface{}{[]interface{}{20.0, 80.0}, []interface{}{110.0, 80.0}}},
					DirectArrayParams: []string{"supports"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"supports": []interface{}{
						[]interface{}{20.0, 80.0},
						[]interface{}{110.0, 80.0},
					},
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "complex nested array with multiple support points",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:              "inst1",
					Params:            map[string]interface{}{"supports": "[[[20,80], [110,80]], [[10,10], [110,10]], [[40,-60], [110,-60]]]"},
					DirectArrayParams: []string{"supports"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"supports": []interface{}{
						[]interface{}{[]interface{}{20.0, 80.0}, []interface{}{110.0, 80.0}},
						[]interface{}{[]interface{}{10.0, 10.0}, []interface{}{110.0, 10.0}},
						[]interface{}{[]interface{}{40.0, -60.0}, []interface{}{110.0, -60.0}},
					},
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "direct array parameter mixed with regular parameters",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:              "inst1",
					Params:            map[string]interface{}{"supports": "[[20,80], [110,80]]", "color": "red", "size": "large"},
					DirectArrayParams: []string{"supports"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"supports": []interface{}{
						[]interface{}{20.0, 80.0},
						[]interface{}{110.0, 80.0},
					},
					"color": "red",
					"size":  "large",
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "direct array parameter with comma splitting disabled",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:                "inst1",
					Params:              map[string]interface{}{"supports": "[[20,80], [110,80]]", "colors": "red,blue,green"},
					DirectArrayParams:   []string{"supports"},
					IgnoreCommaInParams: []string{"colors"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"supports": []interface{}{
						[]interface{}{20.0, 80.0},
						[]interface{}{110.0, 80.0},
					},
					"colors": "red,blue,green", // Should not be split due to ignore_comma_in_params
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "direct array parameter with invalid format falls back to regular processing",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:              "inst1",
					Params:            map[string]interface{}{"supports": "invalid_array_format"},
					DirectArrayParams: []string{"supports"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"supports": "invalid_array_format", // Should fall back to original value
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "empty direct array parameter",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:              "inst1",
					Params:            map[string]interface{}{"supports": "[]"},
					DirectArrayParams: []string{"supports"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"supports": []interface{}{}, // Empty array
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
		{
			name: "direct array parameter with mixed data types",
			input: Input{
				dynamicInstance: models.ConfiguredInstanceConfig{
					Name:              "inst1",
					Params:            map[string]interface{}{"data": "[[20, 80], [\"text\", true], [1.5, false]]"},
					DirectArrayParams: []string{"data"},
				},
				globalParams: map[string]interface{}{},
				paramSets:    []models.ParamSet{},
				inputPath:    models.InputPath{},
			},
			output: Output{
				params: map[string]interface{}{
					"data": []interface{}{
						[]interface{}{20.0, 80.0},
						[]interface{}{"text", true},
						[]interface{}{1.5, false},
					},
				},
				globalParamsMap: map[string][]interface{}{},
				ignoredKeys:     nil,
			},
		},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params, globalParamsMap, ignoredKeys := getAllParams(
				tc.input.dynamicInstance,
				tc.input.globalParams,
				tc.input.paramSets,
				tc.input.inputPath,
			)

			// Check params
			if !reflect.DeepEqual(params, tc.output.params) {
				t.Errorf("Params mismatch.\nExpected: %v\nGot: %v", tc.output.params, params)
			}

			// Check globalParamsMap
			if !reflect.DeepEqual(globalParamsMap, tc.output.globalParamsMap) {
				t.Errorf("GlobalParamsMap mismatch.\nExpected: %v\nGot: %v", tc.output.globalParamsMap, globalParamsMap)
			}

			// Check ignoredKeys
			if !reflect.DeepEqual(ignoredKeys, tc.output.ignoredKeys) {
				t.Errorf("IgnoredKeys mismatch.\nExpected: %v\nGot: %v", tc.output.ignoredKeys, ignoredKeys)
			}
		})
	}
}

// TestParseDirectArrayParam tests the parseDirectArrayParam helper function
func TestParseDirectArrayParam(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected []interface{}
		hasError bool
	}{
		{
			name:     "string array format",
			input:    "[[20,80], [110,80]]",
			expected: []interface{}{[]interface{}{20.0, 80.0}, []interface{}{110.0, 80.0}},
			hasError: false,
		},
		{
			name:     "already parsed array",
			input:    []interface{}{[]interface{}{20.0, 80.0}, []interface{}{110.0, 80.0}},
			expected: []interface{}{[]interface{}{20.0, 80.0}, []interface{}{110.0, 80.0}},
			hasError: false,
		},
		{
			name:     "empty array string",
			input:    "[]",
			expected: []interface{}{},
			hasError: false,
		},
		{
			name:     "single element array",
			input:    "[[20,80]]",
			expected: []interface{}{[]interface{}{20.0, 80.0}},
			hasError: false,
		},
		{
			name:     "mixed data types",
			input:    "[[20, 80], [\"text\", true], [1.5, false]]",
			expected: []interface{}{[]interface{}{20.0, 80.0}, []interface{}{"text", true}, []interface{}{1.5, false}},
			hasError: false,
		},
		{
			name:     "invalid format",
			input:    "invalid_format",
			expected: nil,
			hasError: true,
		},
		{
			name:     "non-array value",
			input:    42,
			expected: []interface{}{42},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseDirectArrayParam(tc.input)

			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected parseDirectArrayParam error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Result mismatch.\nExpected: %v\nGot: %v", tc.expected, result)
			}
		})
	}
}

func TestGetNiceName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard config path",
			input:    "example/name-in-here/config.toml",
			expected: "name-in-here",
		},
		{
			name:     "nested path",
			input:    "examples/small-tray/config.toml",
			expected: "small-tray",
		},
		{
			name:     "single level path",
			input:    "my-project/config.toml",
			expected: "my-project",
		},
		{
			name:     "deeply nested path",
			input:    "projects/3d-printing/openscad/my-design/config.toml",
			expected: "my-design",
		},
		{
			name:     "path with underscores",
			input:    "examples/carovan-cupboard-hook/config.toml",
			expected: "carovan-cupboard-hook",
		},
		{
			name:     "path with numbers",
			input:    "examples/football_cards/config.toml",
			expected: "football_cards",
		},
		{
			name:     "just config.toml",
			input:    "config.toml",
			expected: ".",
		},
		{
			name:     "empty path",
			input:    "",
			expected: ".",
		},
		{
			name:     "path ending with slash",
			input:    "example/name-in-here/",
			expected: "name-in-here",
		},
		{
			name:     "absolute path",
			input:    "/Users/gregc/mine/making/3d-printing/openSCAD/openscadgen/examples/small-tray/config.toml",
			expected: "small-tray",
		},
		{
			name:     "path with spaces",
			input:    "examples/my cool project/config.toml",
			expected: "my cool project",
		},
		{
			name:     "path with special characters",
			input:    "examples/project-v2.1/config.toml",
			expected: "project-v2.1",
		},
		{
			name:     "path without config.toml",
			input:    "example/name-in-here/some-other-file.toml",
			expected: "name-in-here",
		},
		{
			name:     "path with multiple config.toml",
			input:    "example/name-in-here/config.toml/extra/config.toml",
			expected: "extra",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetNiceName(tc.input)
			if result != tc.expected {
				t.Errorf("GetNiceName(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestGetImagePath(t *testing.T) {
	testCases := []struct {
		name  string
		input struct {
			runOutputImagePath string
			cameraName         string
		}
		expected string
	}{
		{
			name: "standard input",
			input: struct {
				runOutputImagePath string
				cameraName         string
			}{
				runOutputImagePath: "export/v0.1/nice.stl",
				cameraName:         "nice",
			},
			expected: "export/v0.1/nice-nice.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetImagePath(tc.input.runOutputImagePath, tc.input.cameraName)
			if result != tc.expected {
				t.Errorf("GetImagePath(%q, %q) = %q; want %q", tc.input.runOutputImagePath, tc.input.cameraName, result, tc.expected)
			}
		})
	}
}

func TestCheckRegexPattern(t *testing.T) {
	tests := []struct {
		name                     string
		configuredInstanceConfig models.ConfiguredInstanceConfig
		inputPath                models.InputPath
		config                   *models.Config
		expectedMatch            bool // true means should match (empty string returned), false means should not match (error string returned)
	}{
		{
			name: "match param_set name from instance param_sets",
			configuredInstanceConfig: models.ConfiguredInstanceConfig{
				Name:      "some-instance",
				ParamSets: "largeRectAnko",
				Params:    make(map[string]interface{}),
			},
			inputPath: models.InputPath{
				Path: "./test.scad",
			},
			config: &models.Config{
				RegexPattern: "largeRectAnko",
				Design: models.DesignConfig{
					ParamSets: []models.ParamSet{
						{
							Name: "largeRectAnko",
							Params: map[string]interface{}{
								"size": "largeRectAnko",
							},
						},
					},
				},
			},
			expectedMatch: true,
		},
		{
			name: "match param key from param_set",
			configuredInstanceConfig: models.ConfiguredInstanceConfig{
				Name:      "some-instance",
				ParamSets: "largeRectAnko",
				Params:    make(map[string]interface{}),
			},
			inputPath: models.InputPath{
				Path: "./test.scad",
			},
			config: &models.Config{
				RegexPattern: "size",
				Design: models.DesignConfig{
					ParamSets: []models.ParamSet{
						{
							Name: "largeRectAnko",
							Params: map[string]interface{}{
								"size": "largeRectAnko",
							},
						},
					},
				},
			},
			expectedMatch: true,
		},
		{
			name: "match param value from param_set",
			configuredInstanceConfig: models.ConfiguredInstanceConfig{
				Name:      "some-instance",
				ParamSets: "largeRectAnko",
				Params:    make(map[string]interface{}),
			},
			inputPath: models.InputPath{
				Path: "./test.scad",
			},
			config: &models.Config{
				RegexPattern: "largeRectAnko",
				Design: models.DesignConfig{
					ParamSets: []models.ParamSet{
						{
							Name: "largeRectAnko",
							Params: map[string]interface{}{
								"size": "largeRectAnko",
							},
						},
					},
				},
			},
			expectedMatch: true,
		},
		{
			name: "no match when pattern doesn't match anything",
			configuredInstanceConfig: models.ConfiguredInstanceConfig{
				Name:      "some-instance",
				ParamSets: "largeRectAnko",
				Params:    make(map[string]interface{}),
			},
			inputPath: models.InputPath{
				Path: "./test.scad",
			},
			config: &models.Config{
				RegexPattern: "nonexistent",
				Design: models.DesignConfig{
					ParamSets: []models.ParamSet{
						{
							Name: "largeRectAnko",
							Params: map[string]interface{}{
								"size": "largeRectAnko",
							},
						},
					},
				},
			},
			expectedMatch: false,
		},
		{
			name: "match param key from instance params",
			configuredInstanceConfig: models.ConfiguredInstanceConfig{
				Name:      "some-instance",
				ParamSets: "",
				Params: map[string]interface{}{
					"myParam": "value",
				},
			},
			inputPath: models.InputPath{
				Path: "./test.scad",
			},
			config: &models.Config{
				RegexPattern: "myParam",
				Design: models.DesignConfig{
					ParamSets: []models.ParamSet{},
				},
			},
			expectedMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkRegexPattern(tt.config, tt.configuredInstanceConfig, tt.inputPath)
			if tt.expectedMatch {
				if result != "" {
					t.Errorf("checkRegexPattern() = %q, expected empty string (match)", result)
				}
			} else {
				if result == "" {
					t.Errorf("checkRegexPattern() = empty string, expected non-empty error string (no match)")
				}
			}
		})
	}
}
