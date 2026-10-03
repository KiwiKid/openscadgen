package pkg

import (
	"testing"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func TestBuildReportMetaDetectsConfiguredTests(t *testing.T) {
	base := models.BuildReportMetaParams{Config: &models.Config{}}
	if meta := BuildReportMeta(base, models.Results{}); meta.HasConfiguredTests {
		t.Fatal("expected no configured tests")
	}

	base.Config.Design.Tests = []models.GeometryTest{{Type: "is-flat"}}
	if meta := BuildReportMeta(base, models.Results{}); !meta.HasConfiguredTests {
		t.Fatal("expected configured design test")
	}

	base.Config.Design.Tests = nil
	base.Instances = []models.InstanceConfig{{
		ExportImages: []models.ExportCameraCoordinates{{
			Tests: []models.GeometryTest{{Type: "color-in-bounding-box"}},
		}},
	}}
	if meta := BuildReportMeta(base, models.Results{}); !meta.HasConfiguredTests {
		t.Fatal("expected configured per-image test")
	}
}
