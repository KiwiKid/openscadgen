package pkg

import (
	"bytes"
	//"fmt"
	"reflect"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kiwikid/openscadgen/pkg/models"
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
		config   models.Config
		expected models.OutputPaths
	}{
		{
			name: "With specified output path",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:  filepath.Join("designs", "test_design.scad"),
					OutputPath: "outputs",
					Version:    "v1.0",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join(tempDir, "outputs", "v1.0"),
				ExportFolderPath:      filepath.Join(tempDir, "outputs", "v1.0"),
				LowQualityWarningPath: filepath.Join(tempDir, "outputs", "v1.0", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(tempDir, "outputs", "v1.0", "README.md"),
				LogOutputPath:         filepath.Join(tempDir, "outputs", "v1.0", "export_log_"+time.Now().Format("2006-01-02_15-04-05")+".log"),
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
				OutputPath:            filepath.Join(".", designsDir, "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(designsDir, "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(designsDir, "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(designsDir, "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(designsDir, "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(designsDir, "export", "v1.0", "test_design", "report.html"),
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
				OutputPath:            filepath.Join(".", filepath.Dir(inputPath), "export", "v1.0", "test_design", "export", "v1.0"),
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
				OutputPath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1.0", "report.html"),
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
				OutputPath:            filepath.Join(".", filepath.Dir(inputPath), "export", "v1.0", "test_design", "export", "v1.0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(inputPath), "export", "v1.0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(inputPath), "export", "v1.0", "test_design", "report.html"),
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
				if result.ReportPath != tc.expected.ReportPath {
					t.Errorf("ReportPath = %s; want %s", result.ReportPath, tc.expected.ReportPath)
				}
			}
		})
	}
}

func TestGenerateDynamicInstances(t *testing.T) {
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
		name           string
		config         models.Config
		expectedParams []map[string]interface{}
	}{
		{
			name: "Single instance with no params",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
						{
							Name:   "default",
							Params: make(map[string]interface{}),
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"name":           "default",

				},
			},
		},
		{
			name: "Multiple parameter combinations",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"width":  []int{10,20,30},
								"height": []int{5,15},
							},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"height":         float64(5),
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"width":          float64(10),
					"height":         float64(15),
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"height":         float64(5),
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"height":         float64(15),
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"width":          float64(30),
					"height":         float64(5),
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"width":          float64(30),
					"height":         float64(15),
					"name":           "test",
				},
			},
		},
		{
			name: "Numeric ranges",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"count": "1-3",
							},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"count":          float64(1),
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"count":          float64(2),
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"count":          float64(3),
					"name":           "test",
				},
			},
		},
		{
			name: "Boolean values",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
						{
							Name: "test",
							Params: map[string]interface{}{
								"enabled": "true,false",
							},
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"enabled":        true,
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"enabled":        false,
					"name":           "test",
				},
			},
		},
		{
			name: "String values",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
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
					"type":          "small",
					"name":          "test",
				},
				{
					"designFileName": "test_design",
					"type":          "medium",
					"name":          "test",
				},
				{
					"designFileName": "test_design",
					"type":          "large",
					"name":          "test",
				},
			},
		},
		{
			name: "Ignored parameters",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					InputPaths: []models.InputPath{
						{
							Path:                           inputPath,
							IgnoreParamsWhenProcessing: "ignored",
						},
					},
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
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
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"name":           "test",
				},
			},
		},
		{
			name: "Global parameters",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					GlobalParams: map[string]interface{}{
						"global": "value",
					},
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
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
					"name":           "test",
				},
				{
					"designFileName": "test_design",
					"width":          float64(20),
					"global":         "value",
					"name":           "test",
				},
			},
		},
		{
			name: "Instance naming with PartIDLetter",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					DynamicInstanceConfig: []models.DynamicInstanceConfig{
						{
							Name: "test",
						},
						{
							Name: "test2",
						},
					},
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design",
					"name": "test2",
				},
				{
					"designFileName": "test_design",
					"name": "test",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instances := generateInstances(&tc.config)

			if len(instances) != len(tc.expectedParams) {
				t.Errorf("Expected %d instances, got %d", len(tc.expectedParams), len(instances))
			}

			// Create a map of expected instances for easier lookup
			expectedMap := createExpectedInstancesMap(tc.expectedParams)

			// Track used PartIDLetters to ensure uniqueness
			usedPartIDLetters := make(map[string]bool)

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
				if instance.PartIDLetter == "" {
					t.Errorf("Instance has no PartIDLetter")
				}
				if usedPartIDLetters[instance.PartIDLetter] {
					t.Errorf("Duplicate PartIDLetter found: %s", instance.PartIDLetter)
				}
				usedPartIDLetters[instance.PartIDLetter] = true

				// Verify InputPath is set correctly
				if instance.InputPath != inputPath {
					t.Errorf("Instance InputPath mismatch:\nExpected: %s\nGot: %s", inputPath, instance.InputPath)
				}
			}

			// Check if all expected instances were matched
			for key, params := range expectedMap {
				if !matchedExpected[key] {
					t.Errorf("Expected instance not found: %+v", params)
				}
			}
		})
	}
}
