package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestToolsPageRendersNightlyInstallOptionCheckedByDefault(t *testing.T) {
	var rendered bytes.Buffer
	err := ToolsPage(
		OpenSCADNavView{InstallSupported: true},
		[]ToolLink{{
			Title:                 "OpenSCAD",
			Slug:                  "openscad",
			SupportsInstallFlow:   true,
			InstallURL:            "/api/openscad/install",
			InstallNightlyDefault: true,
		}},
		"/",
		false,
	).Render(context.Background(), &rendered)
	if err != nil {
		t.Fatal(err)
	}

	html := rendered.String()
	if !strings.Contains(html, `name="install_nightly" value="true" checked`) {
		t.Fatalf("expected the nightly option to be checked by default, got %s", html)
	}
	if !strings.Contains(html, `hx-post="/api/openscad/install"`) {
		t.Fatalf("expected the OpenSCAD install endpoint, got %s", html)
	}
	if !strings.Contains(html, `id="tools-results-content"`) {
		t.Fatalf("expected a results section, got %s", html)
	}
}

func TestToolsPageShowsBOSL2PlatformCompatibilityMessage(t *testing.T) {
	var rendered bytes.Buffer
	err := ToolsPage(
		OpenSCADNavView{BOSL2InstallSupported: false},
		[]ToolLink{{
			Title:              "BOSL2",
			Slug:               "bosl2",
			SupportsUpdateFlow: true,
			CheckUpdatesURL:    "/api/openscad/libraries/bosl2/check",
			UpdateURL:          "/api/openscad/libraries/bosl2/update",
		}},
		"/",
		false,
	).Render(context.Background(), &rendered)
	if err != nil {
		t.Fatal(err)
	}

	html := rendered.String()
	if !strings.Contains(html, "BOSL2 installation is currently supported on macOS only.") {
		t.Fatalf("expected compatibility message, got %s", html)
	}
	if strings.Contains(html, `hx-post="/api/openscad/libraries/bosl2/update"`) {
		t.Fatalf("unsupported platforms must not show the BOSL2 update action, got %s", html)
	}
}
