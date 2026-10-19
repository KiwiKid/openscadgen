package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestDuplicateGlobalParamsFromTOML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	contents := `[openscadgen]
name = "carovan_reinforcement_wedge"
version = "v0.1"
export_name_format = "{designFileName}_{partType}_{mode}"
global_params = { partType = ["wedge-sizer", "wedge", "frontFace"], mode = ["includeCutouts", "includeCutouts"] }
[[openscadgen.input_paths]]
path = "./shape.scad"
`
	if err := os.WriteFile(configPath, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "shape.scad"), []byte("cube(1);"), 0o644); err != nil {
		t.Fatal(err)
	}
	config, _, err := LoadConfigFromFile(models.CmdFlags{ConfigFile: configPath, Server: true})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Process(config, &NoopProgress{}, nil, Operations{GenerateReport: false}, true)
	if err == nil || !strings.Contains(err.Error(), `[openscadgen].global_params.mode: value "includeCutouts" is repeated at positions 1, 2`) {
		t.Fatalf("expected global mode diagnostic, got %v", err)
	}
	if strings.Contains(err.Error(), "REQUIRED EXPORT NAME FIELD MISSING") {
		t.Fatalf("all differing params are already in the export format: %v", err)
	}
}

func TestDuplicateParamDiagnostics(t *testing.T) {
	for _, scope := range []string{"global", "instance", "input", "param set", "comma", "literal comma", "direct array"} {
		t.Run(scope, func(t *testing.T) {
			params := map[string]interface{}{"mode": []interface{}{"includeCutouts", "includeCutouts"}}
			config := &models.Config{ConfigFile: "config.toml", Design: models.DesignConfig{ExportNameFormat: "{mode}"}}
			instance := models.InstanceConfig{Name: "front", InputPath: models.InputPath{Path: "shape.scad"}, Params: map[string]interface{}{"mode": "includeCutouts"}, RunOutputPathV3: "export/front.stl"}
			var want string
			switch scope {
			case "global":
				config.Design.GlobalParams = params
				want = "[openscadgen].global_params.mode"
			case "input":
				config.Design.InputPaths = []models.InputPath{{Path: "shape.scad", Params: params}}
				want = "[[openscadgen.input_paths]] #1 (path = \"shape.scad\").params.mode"
			case "param set":
				config.Design.ParamSets = []models.ParamSet{{Name: "cuts", Params: params}}
				instance.InputPath.ParamSets = "cuts"
				want = "[[openscadgen.param_sets]] #1 (name = \"cuts\").params.mode"
			default:
				configured := models.ConfiguredInstanceConfig{Name: "front", Params: params}
				if scope == "comma" || scope == "literal comma" {
					params["mode"] = "includeCutouts, includeCutouts"
				}
				if scope == "literal comma" {
					configured.IgnoreCommaInParams = []string{"mode"}
				}
				if scope == "direct array" {
					configured.DirectArrayParams = []string{"mode"}
				}
				config.Design.ConfiguredInstanceConfig = []models.ConfiguredInstanceConfig{configured}
				want = "[[openscadgen.instances]] #1 (name = \"front\").params.mode"
			}
			errors := ValidateInstances([]models.InstanceConfig{instance, instance}, config)
			if len(errors) != 1 {
				t.Fatalf("expected one collision, got %d", len(errors))
			}
			message := errors[0].Message
			if scope == "literal comma" || scope == "direct array" {
				if strings.Contains(message, "REPEATED PARAMETER VALUES") {
					t.Fatalf("literal parameter reported as repeated: %s", message)
				}
				return
			}
			for _, text := range []string{want, `value "includeCutouts" is repeated at positions 1, 2`, "Config: config.toml"} {
				if !strings.Contains(message, text) {
					t.Errorf("missing %q in diagnostic: %s", text, message)
				}
			}
		})
	}
}

func TestDuplicateParamDiagnosticsIgnoresUnrelatedValues(t *testing.T) {
	config := &models.Config{Design: models.DesignConfig{GlobalParams: map[string]interface{}{"mode": []string{"other", "other"}}}}
	instance := models.InstanceConfig{Params: map[string]interface{}{"mode": "includeCutouts"}}
	if got := duplicateParamDiagnostics(config, instance, instance); got != "" {
		t.Fatalf("unrelated repeated value reported: %s", got)
	}
}
