package templates

import (
	"context"
	"strings"
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestProcessFormsOnlyRenderRunTestsForConfiguredTests(t *testing.T) {
	forms := []struct {
		name   string
		render func(models.ReportMeta) string
	}{
		{
			name: "slim",
			render: func(reportMeta models.ReportMeta) string {
				var rendered strings.Builder
				if err := SlimProcessForm(reportMeta).Render(context.Background(), &rendered); err != nil {
					t.Fatal(err)
				}
				return rendered.String()
			},
		},
		{
			name: "full",
			render: func(reportMeta models.ReportMeta) string {
				var rendered strings.Builder
				if err := ProcessFormV2(reportMeta).Render(context.Background(), &rendered); err != nil {
					t.Fatal(err)
				}
				return rendered.String()
			},
		},
	}

	for _, form := range forms {
		t.Run(form.name, func(t *testing.T) {
			if body := form.render(models.ReportMeta{}); strings.Contains(body, `name="run_tests"`) {
				t.Fatalf("Run tests should be hidden without configured tests: %s", body)
			}
			if body := form.render(models.ReportMeta{HasConfiguredTests: true}); !strings.Contains(body, `name="run_tests"`) {
				t.Fatalf("Run tests should be shown when tests are configured: %s", body)
			}
		})
	}
}
