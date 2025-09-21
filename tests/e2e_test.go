package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

const (
	defaultDesign  = "default"
	defaultDesign2 = "default2"
	invalidDesign  = "invalid"
	defaultConfig  = `[openscadgen]
name = "test"
version = "1.0"
description = "Test design"
export_name_format = "{designFileName}-{instanceName}"
dont_use_manifold = true

[[openscadgen.input_paths]]
path = "default_design.scad"

[[openscadgen.images]]
name = "top"


`
	defaultConfigPath         = "config.toml"
	defaultShoppingConfigPath = "shopping-list.toml"
	shoppingListConfig        = `[openscadgen]
# name of the design, will be used in the name of output files
name = "list_generator"
export_name_format = "{designFileName}-{instanceName}"
input_path = "default_design_2.scad"

dont_use_manifold = true

[[openscadgen.instances]]
params = { name="shopping", title_text= "Shopping List", rows=20, cols=1 }
params_numbered = { text = "Apple,Avocados,Bananas,Oranges,Mushrooms,Butter,Butter Spread,Yogurt,Aioli,Cheese,Chicken,Coffee,Eggs,Garlic,Honey,Peanut Butter,Cereal,Pasta,Olive oil,Salt,Milk,Tomatoes,Mushrooms,Olive oil,Onions,Oranges,Pasta,Peanut butter,Peppers,Potatoes,Rice,Salt,Spinach,Sugar,Tea,Bread,Paper Towels,Toilet Paper"}

[[openscadgen.instances]]
params = { name="predrive", title_text= "Caravan PreDrive Checklist",  rows=20, cols=1}
params_numbered = { text = "Gas turned off,Legs up,Wheel up & secure,Water emptied,Inside Cabinets Latched"}


[[openscadgen.instances]]
params = { name="postdrive", title_text= "Caravan PostDrive Checklist",  rows=20, cols=1  }
params_numbered = { text = "Empty Fridge,Empty Trash"}

[[openscadgen.instances]]
params = { name="demo", title_text="Demo", rows=20, cols=1 }
params_numbered = { text = "first,second,third"}
`
	shoppingDesign = `
module shoppingList(name, title_text, text, rows, cols) {
    cube(10,10,10)
}

shoppingList(name="shopping", title_text="Shopping List", text="Apple,Avocados,Bananas,Oranges,Mushrooms,Butter,Butter Spread,Yogurt,Aioli,Cheese,Chicken,Coffee,Eggs,Garlic,Honey,Peanut Butter,Cereal,Pasta,Olive oil,Salt,Milk,Tomatoes,Mushrooms,Olive oil,Onions,Oranges,Pasta,Peanut butter,Peppers,Potatoes,Rice,Salt,Spinach,Sugar,Tea,Bread,Paper Towels,Toilet Paper", rows=20, cols=1);
`
)

func TestMain(m *testing.M) {
	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		os.Stderr.WriteString("Failed to get current directory: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Change to the parent directory to build
	parentDir := filepath.Dir(currentDir)
	if err := os.Chdir(parentDir); err != nil {
		os.Stderr.WriteString("Failed to change directory: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", "openscadgen", ".")
	if err := cmd.Run(); err != nil {
		os.Stderr.WriteString("Failed to build binary: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Get the absolute path to the binary
	binaryPath, err = filepath.Abs("openscadgen")
	if err != nil {
		os.Stderr.WriteString("Failed to get binary path: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Change back to the original directory
	if err := os.Chdir(currentDir); err != nil {
		os.Stderr.WriteString("Failed to change back to original directory: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()

	// Clean up
	//os.Remove(binaryPath)

	os.Exit(code)
}

type testCase struct {
	name          string
	configContent string
	scadContent   string
	scadContent2  string
	command       string
	expectedFiles []string
	shouldFail    bool
}

// printDirectoryContents prints the contents of a directory and its subdirectories
func printDirectoryContents(t *testing.T, dir string) {
	printDirectoryContentsRecursive(t, dir, 0)
}

// printDirectoryContentsRecursive is a helper function that recursively prints directory contents
func printDirectoryContentsRecursive(t *testing.T, dir string, depth int) {
	contents, err := os.ReadDir(dir)
	if err != nil {
		t.Errorf("Failed to read directory %s: %v", dir, err)
		return
	}

	// Print header only for the root directory
	if depth == 0 {
		t.Logf("\n===== Directory Contents:  =====\n (%s)\n", dir)
	}

	// Create indentation based on depth
	indent := strings.Repeat("  ", depth)

	for _, entry := range contents {
		if entry.IsDir() {
			t.Logf("%s📁 %s/\n", indent, entry.Name())
			// Recursively print subdirectory contents
			subDir := filepath.Join(dir, entry.Name())
			printDirectoryContentsRecursive(t, subDir, depth+1)
		} else {
			t.Logf("%s📄 %s\n", indent, entry.Name())
		}
	}

	// Print footer only for the root directory
	if depth == 0 {
		t.Logf("=============================\n")
	}
}

var scadFileInputExamples = map[string]string{
	defaultDesign: `
	module cubeDefault(size) {
    cube([size, size, size]);
}
	
cubeDefault(size=10);
	`,
	defaultDesign2: `
	module cylinderDefault(size) {
		cylinder(d=size, h=size);
	}

cylinderDefault(size=10);
`,
	invalidDesign: `
	module invalidDesign(size) {
		HMM_IS_THAT_A_VALID_METHOD_NAME___NO([size, size, size]);
	}

	invalidDesign(size=10);
	`,
}

func TestOpenSCADGenE2E(t *testing.T) {
	// Check if we're in a CI environment (headless)
	isCI := os.Getenv("CI") != "" || os.Getenv("DISPLAY") == ""

	testCases := []testCase{
		{
			name:          "0. Default Default design",
			configContent: defaultConfig,
			scadContent:   scadFileInputExamples[defaultDesign],
			command: func() string {
				if isCI {
					return binaryPath + " -c " + defaultConfigPath + " -ow --oe"
				}
				return binaryPath + " -c " + defaultConfigPath + " -ow"
			}(),
			expectedFiles: func() []string {
				if isCI {
					return []string{
						"config.toml",
						"default_design.scad",
						"default_design_2.scad",
						"shopping-list.toml",
						"export/1.0/default_design-default.stl",
						"export/1.0/report.html",
					}
				}
				return []string{
					"config.toml",
					"default_design.scad",
					"default_design_2.scad",
					"shopping-list.toml",
					"export/1.0/default_design-default.stl",
					"export/1.0/default_design-default-top.png",
					"export/1.0/report.html",
				}
			}(),
		},
		{
			name:          "1.Default design",
			configContent: defaultConfig,
			scadContent:   scadFileInputExamples[defaultDesign],
			command: func() string {
				if isCI {
					return binaryPath + " -c " + defaultConfigPath + " -ow --oe"
				}
				return binaryPath + " -c " + defaultConfigPath + " -ow"
			}(),
			expectedFiles: func() []string {
				if isCI {
					return []string{
						"config.toml",
						"default_design.scad",
						"default_design_2.scad",
						"shopping-list.toml",
						"export/1.0/default_design-default.stl",
						"export/1.0/report.html",
					}
				}
				return []string{
					"config.toml",
					"default_design.scad",
					"default_design_2.scad",
					"shopping-list.toml",
					"export/1.0/default_design-default.stl",
					"export/1.0/default_design-default-top.png",
					"export/1.0/report.html",
				}
			}(),
		},
		{
			name: "2.Default design 2",
			configContent: `
				[openscadgen]
				name = "test"
				version = "1.0"
				description = "Test design"
				export_name_format = "{designFileName}-{instanceName}-{size}"

				dont_use_manifold = true

				global_params = { size = "10,20" }

				[[openscadgen.input_paths]]
				path = "default_design.scad"
			`,
			scadContent: scadFileInputExamples[defaultDesign2],
			command:     binaryPath + " -c " + defaultConfigPath + " -ow",
			expectedFiles: []string{
				"config.toml",
				"default_design.scad",
				"default_design_2.scad",
				"shopping-list.toml",
				"export/1.0/default_design-default-10.stl",
				"export/1.0/default_design-default-20.stl",
				"export/1.0/report.html",
			},
		},
		{
			name: "3.Basic config with default output",
			configContent: `[openscadgen]
name = "test-project"
version = "1.0"
export_name_format = "{designFileName}-{instanceName}-{size}"

global_params = { size = "10,20" }

dont_use_manifold = true

[[openscadgen.input_paths]]       
path = "./default_design.scad"

[[openscadgen.instances]]
name = "instance-name"
`,
			scadContent: scadFileInputExamples[defaultDesign],
			command:     binaryPath + " -c ./config.toml",
			expectedFiles: []string{
				"default_design.scad",
				"config.toml",
				"default_design_2.scad",
				"shopping-list.toml",
				"export/1.0/report.html",
				"export/1.0/default_design-instance-name-10.stl",
				"export/1.0/default_design-instance-name-20.stl",
			},
			shouldFail: false,
		},
		{
			name: "4.Global config with multiple design and exports ",
			configContent: `[openscadgen]
name = "test-project"
version = "1.0"
export_name_format = "{designFileName}-{instanceName}-{size}"

global_params = { size = "10,20" }

dont_use_manifold = true

[[openscadgen.input_paths]]       
path = "./default_design.scad"

[[openscadgen.input_paths]]       
path = "./default_design_2.scad"

[[openscadgen.instances]]
name = "instance-name"

[[openscadgen.images]]
name = "top"

[[openscadgen.images]]
name = "bottom"
`,
			scadContent: scadFileInputExamples[defaultDesign],
			command: func() string {
				if isCI {
					return binaryPath + " -c ./config.toml --oe"
				}
				return binaryPath + " -c ./config.toml"
			}(),
			expectedFiles: func() []string {
				if isCI {
					return []string{
						"default_design.scad",
						"config.toml",
						"default_design_2.scad",
						"shopping-list.toml",
						"export/1.0/report.html",
						"export/1.0/default_design-instance-name-10.stl",
						"export/1.0/default_design-instance-name-20.stl",
						"export/1.0/default_design_2-instance-name-10.stl",
						"export/1.0/default_design_2-instance-name-20.stl",
					}
				}
				return []string{
					"default_design.scad",
					"config.toml",
					"default_design_2.scad",
					"shopping-list.toml",
					"export/1.0/report.html",
					"export/1.0/default_design-instance-name-10.stl",
					"export/1.0/default_design-instance-name-10-top.png",
					"export/1.0/default_design-instance-name-10-bottom.png",
					"export/1.0/default_design-instance-name-20.stl",
					"export/1.0/default_design-instance-name-20-top.png",
					"export/1.0/default_design-instance-name-20-bottom.png",
					"export/1.0/default_design_2-instance-name-10.stl",
					"export/1.0/default_design_2-instance-name-20.stl",
					"export/1.0/default_design_2-instance-name-10-top.png",
					"export/1.0/default_design_2-instance-name-10-bottom.png",
					"export/1.0/default_design_2-instance-name-20-top.png",
					"export/1.0/default_design_2-instance-name-20-bottom.png",
				}
			}(),
			shouldFail: false,
		},
		{
			name: "5.Run with regex pattern on instance name",
			configContent: `[openscadgen]
		name = "test-project"
		version = "1.0"
		export_name_format = "{designFileName}-{instanceName}"

		dont_use_manifold = true

		[[openscadgen.input_paths]]
		path = "./default_design.scad"

		[[openscadgen.instances]]
		name = "instance-name"

		[[openscadgen.instances]]
		name = "instance-name-2"
		`,
			scadContent: scadFileInputExamples[defaultDesign],
			command:     binaryPath + " -c ./config.toml -r instance-name-2 ",
			expectedFiles: []string{
				"default_design.scad",
				"config.toml",
				"default_design_2.scad",
				"shopping-list.toml",
				"export/1.0/report.html",
				"export/1.0/default_design-instance-name-2.stl",
			},
			shouldFail: false,
		},
		{
			name: "6.Run with regex pattern on path name",
			configContent: `[openscadgen]
		name = "test-project"
		version = "1.0"
		export_name_format = "{designFileName}-{instanceName}"

		dont_use_manifold = true

		[[openscadgen.input_paths]]
		path = "./default_design.scad"

		[[openscadgen.input_paths]]
		path = "./default_design_2.scad"

		[[openscadgen.instances]]
		name = "instance-name"

		[[openscadgen.instances]]
		name = "instance-name-2"
		`,
			scadContent:  scadFileInputExamples[defaultDesign],
			scadContent2: scadFileInputExamples[defaultDesign2],
			command:      binaryPath + " -c ./config.toml -r default_design_2 ",
			expectedFiles: []string{
				"default_design.scad",
				"default_design_2.scad",
				"config.toml",
				"shopping-list.toml",
				"export/1.0/report.html",
				"export/1.0/default_design_2-instance-name.stl",
				"export/1.0/default_design_2-instance-name-2.stl",
			},
			shouldFail: false,
		},
		{
			name: "7.Invalid config key",
			configContent: `[openscadgen]
		name = "test-project"
		INVAID_CONFIG_KEY = "INVALID_CONFIG_KEY"
		`,
			command:    binaryPath + " -c ./config.toml",
			shouldFail: true,
		},
		{
			name: "8.Multiple size parameters with different formats",
			configContent: `[openscadgen]
		name = "test-project"
		version = "1.0"
		export_name_format = "{designFileName}-{instanceName}-{size}"

		dont_use_manifold = true

		[[openscadgen.input_paths]]
		path = "./default_design.scad"

		[[openscadgen.instances]]
		name = "instance-name"
		params = { size = "40,50" }

		[[openscadgen.instances]]
		name = "instance-name"
		params = { size = "10,20,30" }
		`,
			scadContent: scadFileInputExamples[defaultDesign],
			command:     binaryPath + " -c ./config.toml",
			expectedFiles: []string{
				"default_design.scad",
				"config.toml",
				"default_design_2.scad",
				"shopping-list.toml",
				"export/1.0/report.html",
				"export/1.0/default_design-instance-name-40.stl",
				"export/1.0/default_design-instance-name-50.stl",
				"export/1.0/default_design-instance-name-10.stl",
				"export/1.0/default_design-instance-name-20.stl",
				"export/1.0/default_design-instance-name-30.stl",
			},
			shouldFail: false,
		},
		{
			name:          "9.invalid design",
			configContent: defaultConfig,
			scadContent:   scadFileInputExamples[invalidDesign],
			command:       binaryPath + " -c " + defaultConfigPath + " -ow",
			expectedFiles: []string{
				"config.toml",
				"default_design.scad",
				"export/1.0/report.html",
			},
			shouldFail: true,
		},
		{
			name:          "10.invalid design - with correct config	",
			configContent: defaultConfig,
			scadContent:   scadFileInputExamples[invalidDesign],
			command: func() string {
				if isCI {
					return binaryPath + " -c " + defaultConfigPath + " -ow -coe --oe"
				}
				return binaryPath + " -c " + defaultConfigPath + " -ow -coe"
			}(),
			expectedFiles: func() []string {
				if isCI {
					return []string{
						"config.toml",
						"default_design.scad",
						"default_design_2.scad",
						"export/1.0/report.html",
						"shopping-list.toml",
					}
				}
				return []string{
					"config.toml",
					"default_design.scad",
					"default_design_2.scad",
					"export/1.0/default_design-default-top.png",
					"export/1.0/report.html",
					"shopping-list.toml",
				}
			}(),
			shouldFail: false,
		},
		{
			name: "11.Invalid extra parameter in instances",
			configContent: `[openscadgen]
		name = "test-project"
		version = "1.0"
		export_name_format = "{designFileName}-{instanceName}"
		dont_use_manifold = true
		
		[[openscadgen.input_paths]]
		path = "./default_design.scad"

		[[openscadgen.instances]]
		name = "instance-name"
		size = "40,50"  # This is an invalid extra parameter
		`,
			scadContent: scadFileInputExamples[defaultDesign],
			command:     binaryPath + " -c ./config.toml",
			shouldFail:  true,
		},
		{
			name:          "12.Shopping list design",
			configContent: shoppingListConfig,
			scadContent:   shoppingDesign,
			command:       binaryPath + " -c " + defaultShoppingConfigPath + " -ow",
			expectedFiles: []string{
				"export/v0.1/report.html",
				"export/v0.1/default_design_2-predrive.stl",
				"export/v0.1/default_design_2-postdrive.stl",
				"export/v0.1/default_design_2-shopping.stl",
				"export/v0.1/default_design_2-demo.stl",
				"shopping-list.toml",
				"config.toml",
				"default_design_2.scad",
				"default_design.scad",
			},
			shouldFail: false,
		},
	}

	var onlyRunTestIndex = -1
	var extraArgs = " " //"-d"

	for index, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if onlyRunTestIndex != -1 {
				if onlyRunTestIndex != index {
					t.Skipf("Skipping test case: %s", tc.name)
					return
				}
			}
			t.Logf("\n\n\n======== Running test case: %s", tc.name)
			tempDir, err := os.MkdirTemp("", "openscadgen-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Change to temp directory
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(oldDir)
			os.Chdir(tempDir)

			// Write config file
			if err := os.WriteFile("config.toml", []byte(tc.configContent), 0644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Write SCAD file
			if err := os.WriteFile("default_design.scad", []byte(tc.scadContent), 0644); err != nil {
				t.Fatalf("Failed to write SCAD file: %v", err)
			}

			// Write SCAD file
			if err := os.WriteFile("default_design_2.scad", []byte(`
			cube(10,10,10);
			`), 0644); err != nil {
				t.Fatalf("Failed to write SCAD file: %v", err)
			}

			err = os.WriteFile(defaultShoppingConfigPath, []byte(shoppingListConfig), 0644)
			if err != nil {
				t.Fatalf("Failed to write shopping list config file: %v", err)
			}

			if tc.command == "" {
				t.Fatalf("Command is empty for %s", tc.name)
			}

			t.Logf("Running command: %s", tc.command+" "+extraArgs)

			// Run openscadgen
			cmd := exec.Command("sh", "-c", tc.command+" "+extraArgs)
			output, err := cmd.CombinedOutput()
			//if tc.logOutput {
			t.Logf("Command output:\n%s", output)
			//	}
			if tc.shouldFail {
				if err == nil {
					t.Errorf("Expected command to fail but it succeeded")
				} else {
					t.Logf("Command failed as expected: %v", err)
				}
				return
			}
			if err != nil {
				t.Errorf("Command failed: \n\nError: %+v\nOutput: %s", err, output)
				return
			}

			fileCount := 0
			// Count actual files in tempDir
			err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					fileCount++
				}
				return nil
			})
			if err != nil {
				t.Errorf("Failed to walk directory: %v", err)
			}

			// Verify expected files exist
			for _, file := range tc.expectedFiles {
				filePath := filepath.Join(tempDir, file)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Expected file :\n\t'%s' \n does not exist", file)
					printDirectoryContents(t, tempDir)
				}
			}
			if fileCount != len(tc.expectedFiles) {
				t.Errorf("Expected %d files but got %d", len(tc.expectedFiles), fileCount)
				printDirectoryContents(t, tempDir)
			}
		})
	}
	/*
		t.Run("confirm onlyRunTestIndex is set to run all tests", func(t *testing.T) {
			if onlyRunTestIndex != -1 {
				t.Errorf("onlyRunTestIndex is set to %d, expected -1", onlyRunTestIndex)
			}
		})
	*/
}
