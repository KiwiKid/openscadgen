package templates

import (
	"context"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestInstanceCardV2RendersFailedOpenSCADDetails(t *testing.T) {
	instance := models.InstanceConfig{
		Name:            "default",
		AutoName:        "broken-instance",
		UniqueID:        "broken-instance",
		ConfigError:     "error running openscad: exit status 1",
		RunOutputPathV3: "/tmp/export/broken-instance.stl",
		Params: map[string]interface{}{
			"width": 42,
		},
		IsComplete: true,
		STLResults: []models.GenerateSTLResult{
			{
				Command:   "openscad -o '/tmp/export/broken-instance.stl' '/tmp/input/broken.scad'",
				OutputLog: "ERROR: Parser error: syntax error in file broken.scad, line 17\nExecution aborted",
				Error:     "error running openscad: exit status 1",
			},
		},
	}

	reportMeta := models.ReportMeta{
		IsServerMode:          true,
		ConfigFilePathEncoded: "cfg",
		ServerFolderEncoded:   "srv",
		InstanceSetSignature:  "sig",
	}

	var rendered strings.Builder
	if err := InstanceCardV2(instance, "/tmp/export", "complete", []string{"width"}, reportMeta).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render instance card: %v", err)
	}

	body := rendered.String()
	for _, needle := range []string{
		`<details class="instance-card-v2__error-details">`,
		`<strong>Error:</strong> error running openscad: exit status 1`,
		`OpenSCAD command`,
		`Copy Command`,
		`openscad -o &#39;/tmp/export/broken-instance.stl&#39; &#39;/tmp/input/broken.scad&#39;`,
		`Full error log`,
		`Copy Error Log`,
		`ERROR: Parser error: syntax error in file broken.scad, line 17`,
		`Execution aborted`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected rendered instance card to contain %q\nbody=%s", needle, body)
		}
	}
}

func TestReportRendersOpenSCADButtonsForConfiguredSCADFiles(t *testing.T) {
	config := &models.Config{Design: models.DesignConfig{
		Name:             "example",
		Version:          "v0.1",
		ExportNameFormat: "{designFileName}",
		InputPaths: []models.InputPath{
			{Path: "./first.scad"},
			{Path: "./first.scad"},
			{Path: "nested/second.scad"},
			{Path: "notes.txt"},
		},
	}}
	reportMeta := models.ReportMeta{
		IsServerMode:          true,
		ConfigFilePathEncoded: "Y29uZmlnLnRvbWw=",
		ServerFolderEncoded:   "cHJvamVjdHM=",
	}

	var rendered strings.Builder
	if err := Report("view", config, nil, "", nil, nil, nil, reportMeta, 0, nil).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render report: %v", err)
	}
	body := rendered.String()
	if count := strings.Count(body, "Open first.scad"); count != 1 {
		t.Fatalf("expected one nav button for first.scad, got %d\nbody=%s", count, body)
	}
	if !strings.Contains(body, "Open second.scad") || !strings.Contains(body, "Open in OpenSCAD") {
		t.Fatalf("expected OpenSCAD buttons for configured SCAD files\nbody=%s", body)
	}
	if strings.Contains(body, "notes.txt") {
		t.Fatalf("non-SCAD input must not render an OpenSCAD button\nbody=%s", body)
	}
}
