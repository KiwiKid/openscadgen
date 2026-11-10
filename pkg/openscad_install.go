package pkg

import (
	"fmt"
	"runtime"
)

// openSCADInstallGuidance returns concise, platform-specific instructions for
// installing OpenSCAD and making the `openscad` command available on PATH.
func openSCADInstallGuidance() string {
	return openSCADInstallGuidanceFor(runtime.GOOS)
}

func openSCADInstallGuidanceFor(goos string) string {
	switch goos {
	case "windows":
		return `Install OpenSCAD with one of:
  winget install --id OpenSCAD.OpenSCAD -e
  https://openscad.org/downloads.html (enable "Add to PATH")

Open a new PowerShell window and verify:
  openscad --version

If it is still not found, add the folder containing openscad.exe (usually
C:\Program Files\OpenSCAD) to your user PATH, then restart the terminal/IDE.
Add the folder, not openscad.exe itself.`
	case "darwin":
		return `Install OpenSCAD with Homebrew:
  brew install --cask openscad

Homebrew installs the app but may not link the CLI. Link it once:
  ln -sf /Applications/OpenSCAD.app/Contents/MacOS/OpenSCAD "$(brew --prefix)/bin/openscad"

Then verify in a new terminal:
  openscad --version`
	case "linux":
		return `Install OpenSCAD with your package manager:
  Debian/Ubuntu: sudo apt update && sudo apt install openscad
  Fedora:        sudo dnf install openscad
  Arch:          sudo pacman -S openscad

For headless/CI machines also install the graphics runtime, e.g. on Ubuntu:
  sudo apt install libegl1 libgl1 libopengl0 libglx0

Then verify:
  openscad --version`
	default:
		return `Install OpenSCAD from https://openscad.org/downloads.html, then verify:
  openscad --version`
	}
}

func openSCADNotFoundError(err error) error {
	return fmt.Errorf("OpenSCAD not found in PATH.\n\n%s\n\nUnderlying error: %w", openSCADInstallGuidance(), err)
}
