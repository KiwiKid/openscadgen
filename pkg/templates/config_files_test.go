package templates

import (
	"context"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestEnterConfigPageUsesVendoredTomSelect(t *testing.T) {
	var page strings.Builder
	if err := EnterConfigPage([]models.ConfigFile{{Path: "examples/demo/config.toml"}}, "/projects", false).Render(context.Background(), &page); err != nil {
		t.Fatalf("render config picker: %v", err)
	}

	body := page.String()
	for _, expected := range []string{
		`href="/static/vendor/tom-select-2.6.2.css"`,
		`src="/static/vendor/tom-select-2.6.2.min.js"`,
		`src="/static/vendor/hx-tomselect-2.0.0.js"`,
		`id="config-picker-group" hx-ext="tomselect"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("config picker missing %q", expected)
		}
	}
}

func TestSavedReportDoesNotReferenceTomSelectAssets(t *testing.T) {
	var report strings.Builder
	if err := Report("", &models.Config{}, nil, "", nil, nil, nil, models.ReportMeta{}, 0, nil).Render(context.Background(), &report); err != nil {
		t.Fatalf("render saved report: %v", err)
	}

	for _, forbidden := range []string{"tom-select", "hx-tomselect"} {
		if strings.Contains(report.String(), forbidden) {
			t.Errorf("saved report unexpectedly references %q", forbidden)
		}
	}
}
