package server

import (
	"runtime"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

// openscadInstallSupported reports whether this platform can run the automated
// OpenSCAD install offered in the UI. Other platforms get manual instructions
// instead of a button that cannot work.
func openscadInstallSupported() bool {
	return runtime.GOOS == "darwin"
}

// openscadInstallSupportLabel names the platform that automation is available
// for, so the UI can explain why a button is missing.
func openscadInstallSupportLabel() string {
	return "Homebrew on macOS"
}

// buildOpenSCADNavView probes OpenSCAD (and BOSL2 when OpenSCAD is usable) and
// packages the result for the nav widget, the tools page and the warning
// banner. A missing or broken OpenSCAD instance is reported as state, not as an
// error, so pages keep rendering.
func buildOpenSCADNavView() templates.OpenSCADNavView {
	v := templates.OpenSCADNavView{
		InstallSupported:      openscadInstallSupported(),
		InstallSupportLabel:   openscadInstallSupportLabel(),
		BOSL2InstallSupported: openscadInstallSupported(),
		BOSL2ActionLabel:      "Install",
		ManualInstallCommand:  openSCADManualInstallCommand(false),
	}

	info, err := pkg.ProbeOpenSCAD()
	if err != nil {
		v.Available = false
		v.Error = err.Error()
		v.Reason = info.ProblemSummary()
		v.Hint = info.Hint
		return v
	}

	bosl2Available, bosl2Err := pkg.ProbeBOSL2(info.Path)
	if bosl2Available {
		v.BOSL2ActionLabel = "Upgrade"
	}

	v.Available = true
	v.Path = info.Path
	v.Version = info.Version
	v.OutOfDate = info.IsOutOfDate
	v.BOSL2Available = bosl2Available
	if bosl2Err != nil {
		v.BOSL2Error = bosl2Err.Error()
	}
	return v
}
