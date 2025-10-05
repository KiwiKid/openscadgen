# Nix Setup for OpenSCADGen

This project includes a Nix flake that provides both OpenSCAD and openscadgen in a reproducible environment.

## Quick Start

### Using Nix (without flakes)
```bash
# Enter development environment
nix-shell

# Build openscadgen
nix-build -A packages.openscadgen
```

### Using Nix Flakes
```bash
# Enter development environment
nix develop

# Build openscadgen
nix build .#openscadgen

# Run openscadgen
nix run .#openscadgen -- -c ./examples/small-tray/config.toml

# Run OpenSCAD
nix run .#openscad
```

### Using direnv (recommended)
```bash
# Install direnv if you haven't already
nix profile install nixpkgs#direnv

# Allow direnv in this directory
direnv allow

# Now the environment will be automatically activated when you enter the directory
```

## What's Included

- **OpenSCAD**: The 3D modeling tool
- **openscadgen**: Your parametric model generator
- **Go toolchain**: For development
- **templ**: For template generation
- **air**: For live reloading during development
- **just**: For task running
- **Development tools**: gopls, golangci-lint, etc.

## Development Workflow

1. Enter the development environment: `nix develop` or use direnv
2. Generate templates: `just generate`
3. Start development with live reload: `just dev`
4. Build: `just build`
5. Test: `just test`

## System-wide Installation

To install openscadgen system-wide:

```bash
# Install to your user profile
nix profile install .#openscadgen

# Or add to your system configuration
# (add to your NixOS configuration.nix)
environment.systemPackages = [ inputs.openscadgen.packages.${system}.openscadgen ];
```

## Flake Commands

```bash
# Show available packages
nix flake show

# Update dependencies
nix flake update

# Format the flake
nix fmt
```
