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
		nil,
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

func TestToolsPageRendersLibraryCheckboxes(t *testing.T) {
	var rendered bytes.Buffer
	err := ToolsPage(
		OpenSCADNavView{},
		nil,
		[]LibraryLink{
			{
				ID:               "bosl2",
				Name:             "BOSL2",
				Description:      "Helper library",
				Installed:        true,
				Version:          "2.0.785",
				Path:             "/tmp/OpenSCAD/libraries/BOSL2",
				InstallSupported: true,
				InstallURL:       "/api/openscad/libraries/bosl2/update",
			},
			{
				ID:               "future-lib",
				Name:             "Future Lib",
				Description:      "Not installed yet",
				InstallSupported: true,
				InstallURL:       "/api/openscad/libraries/future-lib/update",
			},
		},
		"/",
		false,
	).Render(context.Background(), &rendered)
	if err != nil {
		t.Fatal(err)
	}

	html := rendered.String()
	for _, needle := range []string{
		`id="openscad-library-list"`,
		`value="bosl2"`,
		`value="future-lib"`,
		`hx-post="/api/openscad/libraries/bosl2/update"`,
		`hx-post="/api/openscad/libraries/future-lib/update"`,
		"Installed",
		"Not installed",
		"2.0.785",
	} {
		if !strings.Contains(html, needle) {
			t.Fatalf("expected library list to contain %q, got %s", needle, html)
		}
	}
	if !strings.Contains(html, `checked disabled`) {
		t.Fatalf("expected the installed library checkbox to be checked and disabled, got %s", html)
	}
}

func TestOpenSCADLibraryListOOBRefreshIncludesSwapAttribute(t *testing.T) {
	var rendered bytes.Buffer
	err := OpenSCADLibraryList([]LibraryLink{{
		ID:               "bosl2",
		Name:             "BOSL2",
		Installed:        true,
		InstallSupported: true,
		InstallURL:       "/api/openscad/libraries/bosl2/update",
	}}, true).Render(context.Background(), &rendered)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.String(), `hx-swap-oob="outerHTML"`) {
		t.Fatalf("expected OOB swap marker, got %s", rendered.String())
	}
}
