package pkg

import (
	"errors"
	"strings"
	"testing"
)

func TestOpenSCADInstallGuidanceForPlatforms(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{goos: "windows", want: "winget install --id OpenSCAD.OpenSCAD -e"},
		{goos: "darwin", want: "brew install --cask openscad"},
		{goos: "linux", want: "sudo apt install openscad"},
		{goos: "freebsd", want: "https://openscad.org/downloads.html"},
	}
	for _, tc := range tests {
		t.Run(tc.goos, func(t *testing.T) {
			got := openSCADInstallGuidanceFor(tc.goos)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("guidance for %s missing %q:\n%s", tc.goos, tc.want, got)
			}
			if !strings.Contains(got, "openscad --version") {
				t.Fatalf("guidance for %s missing verification command:\n%s", tc.goos, got)
			}
		})
	}
}

func TestOpenSCADNotFoundErrorIncludesGuidanceAndCause(t *testing.T) {
	cause := errors.New("executable file not found")
	err := openSCADNotFoundError(cause)
	if !errors.Is(err, cause) {
		t.Fatalf("error does not wrap its cause: %v", err)
	}
	if !strings.Contains(err.Error(), openSCADInstallGuidance()) {
		t.Fatalf("error missing platform guidance:\n%s", err)
	}
}
