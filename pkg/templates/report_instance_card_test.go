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
