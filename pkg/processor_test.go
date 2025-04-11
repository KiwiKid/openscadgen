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
					ExportNameFormat: "{designFileName}-{version}",
				},
				Debug: false,
			},
			expected: models.OutputPaths{
				OutputPath:            filepath.Join(tempDir, "designs", "export", "v1_0", "test_design", "export", "v1_0"),
				ExportFolderPath:      filepath.Join(tempDir, "designs", "export", "v1_0"),
				LowQualityWarningPath: filepath.Join(tempDir, "designs", "export", "v1_0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(tempDir, "designs", "export", "v1_0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(tempDir, "designs", "export", "v1_0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(tempDir, "designs", "export", "v1_0", "test_design", "report.html"),
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
				OutputPath:            filepath.Join(designsDir, "export", "v1_0", "test_design", "export", "v1_0"),
				ExportFolderPath:      filepath.Join(designsDir, "export", "v1_0"),
				LowQualityWarningPath: filepath.Join(designsDir, "export", "v1_0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(designsDir, "export", "v1_0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(designsDir, "export", "v1_0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(designsDir, "export", "v1_0", "test_design", "report.html"),
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
				OutputPath:            filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "export", "v1_0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(inputPath), "export", "v1_0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "report.html"),
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
				OutputPath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1_0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1_0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1_0", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1_0", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1_0", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(filepath.Join("some", "nested", "path", "config.toml")), "export", "v1_0", "report.html"),
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
				OutputPath:           filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "export", "v1_0"),
				ExportFolderPath:      filepath.Join(filepath.Dir(inputPath), "export", "v1_0"),
				LowQualityWarningPath: filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "LOW_QUALITY_WARNING.md"),
				ReadmePath:            filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "README.md"),
				LogOutputPath:         filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "export_log.log"),
				ReportPath:            filepath.Join(filepath.Dir(inputPath), "export", "v1_0", "test_design", "report.html"),
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
			if tc.name == "With specified output path" {

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

func TestGenerateDynamicInstances(t *testing.T) {
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
					ExportNameFormat: "test_design_{version}",
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design.scad",
					"name":           "default",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0/test_design_v1_0.stl"),
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
					ExportNameFormat: "test_design_{version}_{name}_{width}_{height}",
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"height":         float64(5),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"height":         float64(15),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"height":         float64(5),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"height":         float64(15),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(30),
					"height":         float64(5),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(30),
					"height":         float64(15),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				//	filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_10_5.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_10_15.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_20_5.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_20_15.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_30_15.stl"),
				//	filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_20_15.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_10_5.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_30_5.stl"),
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
					ExportNameFormat: "test_design_{version}_{name}_{enabled}",
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design.scad",
					"enabled":        true,
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"enabled":        false,
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_true.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_false.stl"),
			},
		},
		{
			name: "String values",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_{version}_{name}_{type}",
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
					"designFileName": "test_design.scad",
					"type":           "small",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"type":           "medium",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"type":           "large",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_small.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_medium.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_large.stl"),
			},
		},
		{
			name: "ignored_parameters - global param ignored",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_{version}_{name}_{width}",
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
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_name_test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_name_test.stl"),
			},
		},
		{
			name: "ignored_parameters - input path params ignored",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_{version}_{name}_{width}",
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
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_name_test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_name_test.stl"),
			},
		},
		{
			name: "global_parameters",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_{version}_global_${global}_global2_${global2}_name_${name}",
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
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"global":         "value",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"global":         "value",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"global":         "value",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"global":         "value",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"global":         "value2",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"global":         "value2",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"global":         "value2",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"width":          float64(20),
					"global":         "value2",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value_global2_$value4_name_$test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value2_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value2_global2_$value4_name_$test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value_global2_$value4_name_$test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value2_global2_$value3_name_$test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_$value2_global2_$value4_name_$test.stl"),
			},
		},
		/*{
			name: "Instance naming with PartIDLetter",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath: inputPath,
					Version:   "v1.0",
					ConfiguredInstanceConfig: []models.ConfiguredInstanceConfig{
						{
							Name: "test",
						},
					},
					ExportNameFormat: "test_design_{version}_{name}_{part_id_letter}",
				},
			},
			expectedParams: []map[string]interface{}{
				{
					"designFileName": "test_design.scad",
					"name": "test",
					"version": "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_test_A.stl"),
			},
		},*/
		{
			name: "with_param_set_reference",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_w{width}_h{height}",
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
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"height":         float64(20),
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_w10_h20.stl"),
			},
		},
		{
			name: "With Multiple Param Sets",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_{version}_{name}",
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
					"designFileName": "test_design.scad",
					"width":          float64(10),
					"height":         float64(20),
					"color":          "red",
					"texture":        "smooth",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0/test_design_v1_0_test.stl"),
			},
		},
		{
			name: "with_param_set_types",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_{version}_name_{name}",
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
					"designFileName": "test_design.scad",
					"count":          float64(5),
					"enabled":        true,
					"type":           "test",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_name_test.stl"),
			},
		},
		{
			name: "with_2_global_parameters",
			config: models.Config{
				ConfigFile: configPath,
				Design: models.DesignConfig{
					InputPath:        inputPath,
					Version:          "v1.0",
					ExportNameFormat: "test_design_{version}_global_{global}_global2_{global2}_name_{name}",
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
					"designFileName": "test_design.scad",
					"global":         "value",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"global":         "value2",
					"global2":        "value3",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"global":         "value",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
				{
					"designFileName": "test_design.scad",
					"global":         "value2",
					"global2":        "value4",
					"name":           "test",
					"version":        "v1.0",
				},
			},
			expectedOutputPaths: []string{
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_value_global2_value3_name_test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_value_global2_value4_name_test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_value2_global2_value3_name_test.stl"),
				filepath.Join(tempDir, "v1_0", "test_design_v1_0_global_value2_global2_value4_name_test.stl"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if len(tc.config.Design.ConfiguredInstanceConfig) > 1 {
				t.Error("Only one dynamic instance config is supported for testing")
			}
			instances, err := generateInstances(&tc.config, tc.config.Design.ConfiguredInstanceConfig[0], models.InputPath{Path: inputPath}, tempDir)
			if err != nil {
				t.Errorf("Error generating instances: %v", err)
			}

			if len(instances) != len(tc.expectedParams) {
				t.Errorf("Expected %d instances, got %d", len(tc.expectedParams), len(instances))
			}

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
				if instance.InputPath.Path != inputPath {
					t.Errorf("Instance InputPath mismatch:\nExpected: %s\nGot: %s", inputPath, instance.InputPath)
				}
			}

			// Check if all expected instances were matched
			for key, params := range expectedMap {
				if !matchedExpected[key] {
					t.Errorf("Expected instance not found: %+v", params)
				}
			}

			// Check output paths
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
		})
	}
}

func TestGenerateParamCombinations(t *testing.T) {

	// Test case 2: Single parameter with multiple values
	t.Run("Single parameter", func(t *testing.T) {
		paramCombos := map[string]interface{}{
			"color": "red,blue,green",
		}
		result, err := convertToParamCombinations(paramCombos, map[string]bool{})
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
		result, err := convertToParamCombinations(paramCombos, map[string]bool{})
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
