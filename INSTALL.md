# Installation Instructions

## Quick Install

Download the appropriate release asset from the [releases page](https://github.com/kiwikid/openscadgen/releases). The archive includes this guide, so its links remain useful both on GitHub and after extraction.

| Platform | Architecture | Release asset |
|----------|--------------|----------|
| **Windows** | x86-64 (64-bit) | `openscadgen-Windows-Amd64-v<version>.zip` |
| **Windows** | ARM64 | `openscadgen-Windows-Arm64-v<version>.zip` |
| **macOS** | Intel (x86-64) | `openscadgen-macOS-Intel-v<version>.zip` |
| **macOS** | Apple Silicon (ARM64) | `openscadgen-macOS-AppleSilicon-v<version>.zip` |
| **Linux** | x86-64 (64-bit) | `openscadgen-Linux-Amd64-v<version>.zip` |
| **Linux** | ARM64 | `openscadgen-Linux-Arm64-v<version>.zip` |

## Platform-Specific Instructions

### Windows
1. Download the appropriate `.zip` file for your architecture
2. Extract the archive to a folder of your choice
3. Open Command Prompt or PowerShell in that folder
4. Run: `openscadgen.exe -h` to verify installation

**Note:** You may need to allow the executable in Windows Security settings.

### macOS
1. Download the appropriate `.zip` file for your Mac
2. Open Terminal and navigate to the download folder
3. Extract: `unzip openscadgen-macOS-*.zip`
4. Make executable: `chmod +x openscadgen`
5. Move to PATH (optional): `sudo mv openscadgen /usr/local/bin/`
6. Test: `openscadgen -h`

**Note:** For Apple Silicon Macs, download the "AppleSilicon" version for best performance.

### Linux
1. Download the appropriate `.zip` file for your architecture
2. Open terminal and navigate to the download folder
3. Extract: `unzip openscadgen-Linux-*.zip`
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
sudo apt-get install -y --no-install-recommends libegl1 libgl1 libopengl0 libglx0
```

**"File could not be run by the operating system":**
- Ensure you downloaded the correct architecture version for your system

**"Terminated by signal SIGKILL":**
- macOS: Check Privacy & Security settings and allow openscadgen to run

## Alternative Downloads

- **Source Code**: [Download ZIP](https://github.com/kiwikid/openscadgen/archive/refs/heads/main.zip)
- **Nightly Builds**: [Latest Development Version](https://github.com/kiwikid/openscadgen/actions)

## What's New

See the release page for the changelog for the version you downloaded.

---

**Need help?** Check the [documentation](https://github.com/kiwikid/openscadgen#readme) or [open an issue](https://github.com/kiwikid/openscadgen/issues).
