package templates

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestImageURLsResolveToTheGeneratedExportImage(t *testing.T) {
	exportRoot := filepath.Join(t.TempDir(), "export", "v0.1")
	imagePath := filepath.Join(exportRoot, "img", "nested", "part-nice.png")
	if err := os.MkdirAll(filepath.Dir(imagePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(imagePath, []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name       string
		reportMode string
		resolve    func(t *testing.T, src string) string
	}{
		{
			name:       "saved report",
			reportMode: "",
			resolve: func(t *testing.T, src string) string {
				t.Helper()
				parsed, err := url.Parse(src)
				if err != nil {
					t.Fatal(err)
				}
				return parsed.Path
			},
		},
		{
			name:       "pages report",
			reportMode: "pages",
			resolve: func(t *testing.T, src string) string {
				t.Helper()
				return filepath.Join(exportRoot, filepath.FromSlash(src))
			},
		},
		{
			name:       "server generated result",
			reportMode: "view",
			resolve: func(t *testing.T, src string) string {
				t.Helper()
				parsed, err := url.Parse(src)
				if err != nil {
					t.Fatal(err)
				}
				relativePath := parsed.Query().Get("config_path")
				if filepath.IsAbs(relativePath) {
					t.Fatalf("server image URL exposes an absolute path: %q", src)
				}
				return filepath.Join(exportRoot, filepath.FromSlash(relativePath))
			},
		},
		{
			name:       "OOB update",
			reportMode: "complete",
			resolve: func(t *testing.T, src string) string {
				t.Helper()
				parsed, err := url.Parse(src)
				if err != nil {
					t.Fatal(err)
				}
				relativePath := parsed.Query().Get("config_path")
				if filepath.IsAbs(relativePath) {
					t.Fatalf("server image URL exposes an absolute path: %q", src)
				}
				return filepath.Join(exportRoot, filepath.FromSlash(relativePath))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := buildReportImageURL(tc.reportMode, imagePath, exportRoot, exportRoot, "cache-bust")
			resolved := tc.resolve(t, src)
			if resolved != imagePath {
				t.Fatalf("image URL %q resolves to %q, want generated image %q", src, resolved, imagePath)
			}
			if _, err := os.Stat(resolved); err != nil {
				t.Fatalf("image URL %q does not point to a real output image: %v", src, err)
			}
		})
	}
}

func TestReportAndOOBCardUseTheServerImagePath(t *testing.T) {
	exportRoot := filepath.Join(t.TempDir(), "export", "v0.1")
	imagePath := filepath.Join(exportRoot, "img", "part-nice.png")
	instance := models.InstanceConfig{
		Name: "default", AutoName: "part", UniqueID: "part", IsComplete: true,
		ImageResults: []models.GenerateImageResult{{OutputPath: imagePath, CameraName: "nice"}},
	}
	reportMeta := models.ReportMeta{IsServerMode: true, ServerFolder: exportRoot}
	expected := buildServerImageURL(imagePath, exportRoot, "")

	var report strings.Builder
	if err := Report("view", &models.Config{}, []models.InstanceConfig{instance}, exportRoot, nil, nil, nil, reportMeta, 0, nil).Render(context.Background(), &report); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(report.String(), `src="`+expected+`"`) {
		t.Fatalf("server report does not use generated image path %q", expected)
	}

	var card strings.Builder
	if err := InstanceCardV2(instance, exportRoot, "complete", nil, reportMeta).Render(context.Background(), &card); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(card.String(), `src="`+expected+`"`) {
		t.Fatalf("OOB card does not use generated image path %q", expected)
	}
}

func TestInstanceCardV2RendersFailedOpenSCADDetails(t *testing.T) {
	instance := models.InstanceConfig{
		Name:            "default",
		AutoName:        "broken-instance",
		UniqueID:        "broken-instance",
		ConfigError:     "error running openscad: exit status 1",
		RunOutputPathV3: "/tmp/export/broken-instance.stl",
		Params: map[string]interface{}{
			"width":   42,
			"version": "v0.1",
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
		`Copy params`,
		`data-openscad-params="width = 42;"`,
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected rendered instance card to contain %q\nbody=%s", needle, body)
		}
	}
}

func TestOpenSCADParamsFormatsAppliedValues(t *testing.T) {
	params := map[string]interface{}{
		"label":   "chip \"cover\"",
		"enabled": "true",
		"sizes":   []interface{}{42, "thin", false},
		"version": "v0.1",
	}

	got := openSCADParams(params)
	want := "enabled = true;\nlabel = \"chip \\\"cover\\\"\";\nsizes = [42, \"thin\", false];"
	if got != want {
		t.Fatalf("unexpected OpenSCAD parameters:\nwant: %s\n got: %s", want, got)
	}
}

func TestGetInstanceAppliedParamsPrefersRenderedParams(t *testing.T) {
	instance := models.InstanceConfig{
		Params: map[string]interface{}{"ignored": true, "width": 42},
		STLResults: []models.GenerateSTLResult{{
			AppliedParams: map[string]interface{}{"width": 42},
		}},
	}

	params := getInstanceAppliedParams(instance)
	if len(params) != 1 || params["width"] != 42 {
		t.Fatalf("expected rendered parameters, got %#v", params)
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
