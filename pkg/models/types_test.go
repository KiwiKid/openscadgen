package models

/*
func TestGetInstanceConfigSaveLocation(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		instance       *InstanceConfig
		expectedOutput string
	}{
		{
			name: "Basic instance with default format",
			config: &Config{
				ConfigFile: filepath.Join("/path/to/config.toml"),
				Design: DesignConfig{
					OutputPath: "export",
					Version:   "v1.0",
					ExportNameFormat: "{designFileName}-{version}-{part_id_letter}",
				},
			},
			instance: &InstanceConfig{
				InputPath: InputPath{
					Path: filepath.Join("/path/to/design.scad"),
				},
				PartIDLetter: "A",
			},
			expectedOutput: filepath.Join("/path/to/export/v1_0/design-v1_0-A.stl"),
		},
		/*{
			name: "Instance with instance-specific export format",
			config: &Config{
				ConfigFile: filepath.Join("/path/to/config.toml"),
				Design: DesignConfig{
					OutputPath: "export",
					Version:   "v1.0",
					ExportNameFormat: "{designFileName}",
				},
			},
			instance: &InstanceConfig{
				InputPath:        filepath.Join("/path/to/design.scad"),
				ExportNameFormat: "{designFileName}-custom",
			},
			expectedOutput: filepath.Join("/path/to", "export", "v1.0", "design-custom.stl"),
		},
		{
			name: "Instance with parameters in export format",
			config: &Config{
				ConfigFile: filepath.Join("/path/to/config.toml"),
				Design: DesignConfig{
					OutputPath:        "export",
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}-{width}-{height}",
				},
			},
			instance: &InstanceConfig{
				InputPath: filepath.Join("/path/to/design.scad"),
				Params: map[string]interface{}{
					"width":  10,
					"height": 20,
				},
			},
			expectedOutput: filepath.Join("/path/to", "export", "v1.0", "design-10-20.stl"),
		},
		{
			name: "Instance with global parameters",
			config: &Config{
				ConfigFile: filepath.Join("/path/to/config.toml"),
				Design: DesignConfig{
					OutputPath:        "export",
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}-{globalParam}",
					GlobalParams: map[string]interface{}{
						"globalParam": "global",
					},
				},
			},
			instance: &InstanceConfig{
				InputPath: filepath.Join("/path/to/design.scad"),
			},
			expectedOutput: filepath.Join("/path/to", "export", "v1.0", "design-global.stl"),
		},
		{
			name: "Instance with quality parameter",
			config: &Config{
				ConfigFile: filepath.Join("/path/to/config.toml"),
				Quality:   "high",
				Design: DesignConfig{
					OutputPath:        "export",
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}-{quality}",
				},
			},
			instance: &InstanceConfig{
				InputPath: filepath.Join("/path/to/design.scad"),
			},
			expectedOutput: filepath.Join("/path/to", "export", "v1.0", "design-high.stl"),
		},
		{
			name: "Instance with absolute output path",
			config: &Config{
				ConfigFile: filepath.Join("/path/to/config.toml"),
				Design: DesignConfig{
					OutputPath: "/absolute/path/export",
					Version:   "v1.0",
				},
			},
			instance: &InstanceConfig{
				InputPath: filepath.Join("/path/to/design.scad"),
			},
			expectedOutput: filepath.Join("/absolute/path/export", "v1.0", "design.stl"),
		},
		{
			name: "Instance with multiple parameters in name",
			config: &Config{
				ConfigFile: filepath.Join("/path/to/config.toml"),
				Design: DesignConfig{
					OutputPath:        "export",
					Version:          "v1.0",
					ExportNameFormat: "{designFileName}-{param1}-{param2}-{param3}",
				},
			},
			instance: &InstanceConfig{
				InputPath: filepath.Join("/path/to/design.scad"),
				Params: map[string]interface{}{
					"param1": "value1",
					"param2": 42,
					"param3": true,
				},
			},
			expectedOutput: filepath.Join("/path/to", "export", "v1.0", "design-value1-42-true.stl"),
		}, */
/*
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exportNameFormat := tt.config.Design.ExportNameFormat
			if (exportNameFormat == "") {
				exportNameFormat = tt.instance.ExportNameFormat
			}
			if (exportNameFormat == "") {
				t.Errorf("exportNameFormat is empty")
			}
			result := GetInstanceConfigSaveLocation(tt.config, tt.instance.InputPath.Path, tt.instance.Name, exportNameFormat, tt.instance.Params, tt.instance.PartIDLetter)
			if result != tt.expectedOutput {
				t.Errorf("Expected %s, got %s", tt.expectedOutput, result)
			}
		})
	}
} */
