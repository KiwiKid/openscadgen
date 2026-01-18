# Installation Instructions

## Quick Install

**Download the appropriate version for your system:**

| Platform | Architecture | Download |
|----------|--------------|----------|
| **Windows** | x86-64 (64-bit) | [openscadgen-Windows-amd64-v{{ .Version }}.zip]({{ .URL }}) |
| **Windows** | ARM64 | [openscadgen-Windows-arm64-v{{ .Version }}.zip]({{ .URL }}) |
| **macOS** | Intel (x86-64) | [openscadgen-macOS-Intel-v{{ .Version }}.tar.gz]({{ .URL }}) |
| **macOS** | Apple Silicon (ARM64) | [openscadgen-macOS-AppleSilicon-v{{ .Version }}.tar.gz]({{ .URL }}) |
| **Linux** | x86-64 (64-bit) | [openscadgen-Linux-amd64-v{{ .Version }}.tar.gz]({{ .URL }}) |
| **Linux** | ARM64 | [openscadgen-Linux-arm64-v{{ .Version }}.tar.gz]({{ .URL }}) |

## Platform-Specific Instructions

### Windows
1. Download the appropriate `.zip` file for your architecture
2. Extract the archive to a folder of your choice
3. Open Command Prompt or PowerShell in that folder
4. Run: `openscadgen.exe -h` to verify installation

**Note:** You may need to allow the executable in Windows Security settings.

### macOS
1. Download the appropriate `.tar.gz` file for your Mac
2. Open Terminal and navigate to the download folder
3. Extract: `tar -xzf openscadgen-macOS-*.tar.gz`
4. Make executable: `chmod +x openscadgen`
5. Move to PATH (optional): `sudo mv openscadgen /usr/local/bin/`
6. Test: `openscadgen -h`

**Note:** For Apple Silicon Macs, download the "AppleSilicon" version for best performance.

### Linux
1. Download the appropriate `.tar.gz` file for your architecture
2. Open terminal and navigate to the download folder
3. Extract: `tar -xzf openscadgen-Linux-*.tar.gz`
4. Make executable: `chmod +x openscadgen`
5. Move to PATH (optional): `sudo mv openscadgen /usr/local/bin/`
6. Test: `./openscadgen -h`

## Prerequisites

Before using openscadgen, ensure you have:
- **OpenSCAD** installed and available in your PATH
- A `.toml` configuration file for your project
- OpenSCAD files (`.scad`) you want to generate

## Quick Start

1. **Create a simple config file** (`config.toml`):
   ```toml
   [openscadgen]
   name = "my-design"
   input_path = "./my-design.scad"
   export_name_format = "design-{width}mm"
   version = "v1.0"
   
   [[openscadgen.instances]]
   params = { width = "10,20,30" }
   ```

2. **Run openscadgen**:
   ```bash
   ./openscadgen -c config.toml
   ```

## Package Managers

### Homebrew (macOS/Linux)
```bash
brew install kiwikid/tap/openscadgen
```

### Scoop (Windows)
```bash
scoop bucket add kiwikid https://github.com/kiwikid/scoop-bucket
scoop install openscadgen
```

## Troubleshooting

**"Permission denied" error:**
- Windows: Right-click → Properties → Unblock
- macOS/Linux: `chmod +x openscadgen`

**"Command not found" after moving to PATH:**
- Restart your terminal or run: `source ~/.bashrc` (Linux) / `source ~/.zshrc` (macOS)

**OpenSCAD not found:**
- Ensure OpenSCAD is installed and in your system PATH
- Test with: `openscad --version`

**OpenSCAD fails to start on Linux CI with `libEGL.so.1` missing:**
- Install EGL/OpenGL runtime libs (Ubuntu/Debian):

```bash
sudo apt-get update -y
sudo apt-get install -y --no-install-recommends libegl1 libgl1
```

**"File could not be run by the operating system":**
- Ensure you downloaded the correct architecture version for your system

**"Terminated by signal SIGKILL":**
- macOS: Check Privacy & Security settings and allow openscadgen to run

## Alternative Downloads

- **Source Code**: [Download ZIP](https://github.com/kiwikid/openscadgen/archive/refs/tags/v{{ .Version }}.zip)
- **Nightly Builds**: [Latest Development Version](https://github.com/kiwikid/openscadgen/actions)

## What's New in v{{ .Version }}

{{ .Changelog }}

---

**Need help?** Check the [documentation](https://github.com/kiwikid/openscadgen#readme) or [open an issue](https://github.com/kiwikid/openscadgen/issues). 