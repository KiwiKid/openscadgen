# Docker Setup for OpenSCADGen

This project includes Docker configurations for both development and production use cases.

## Quick Start

### Development (with live reloading)
```bash
# Build and run development container
docker-compose up openscadgen-dev

# Or build manually
./docker-build.sh dev
docker run -it --rm -p 6767:6767 -v $(pwd):/app openscadgen:dev
```

### Production
```bash
# Build and run production container
docker-compose up openscadgen

# Or build manually
./docker-build.sh prod
docker run -it --rm -p 6767:6767 -v $(pwd)/examples:/app/examples:ro openscadgen:latest
```

## Available Images

### Development Image (`Dockerfile.dev`)
- **Purpose**: Development with live reloading
- **Includes**: Go toolchain, air, templ, gopls, golangci-lint
- **Features**: Hot reloading, volume mounting for live development
- **Port**: 6767 (openscadgen server)

### Production Image (`Dockerfile`)
- **Purpose**: Lightweight production deployment
- **Includes**: OpenSCAD, openscadgen binary, examples
- **Features**: Multi-stage build, non-root user, health checks
- **Size**: Optimized for minimal footprint

## Docker Compose Services

### `openscadgen-dev`
- Development service with live reloading
- Volume mounts source code for hot reloading
- Includes all development tools

### `openscadgen`
- Production service
- Read-only mount of examples
- Output directory mounting
- Auto-restart on failure

### `openscad` (profile: openscad-only)
- Standalone OpenSCAD service
- Useful for testing OpenSCAD without openscadgen

## Usage Examples

### Run a specific config file
```bash
# Development
docker run -it --rm -v $(pwd):/app openscadgen:dev openscadgen -c ./examples/small-tray/config.toml

# Production
docker run -it --rm -v $(pwd)/examples:/app/examples:ro openscadgen:latest openscadgen -c /app/examples/small-tray/config.toml
```

### Start server mode
```bash
# Development server
docker run -it --rm -p 6767:6767 -v $(pwd):/app openscadgen:dev openscadgen -sf /app/examples -p 6767

# Production server
docker run -it --rm -p 6767:6767 -v $(pwd)/examples:/app/examples:ro openscadgen:latest openscadgen -sf /app/examples -p 6767
```

### Build and test locally
```bash
# Build development image
./docker-build.sh dev

# Test the build
docker run --rm openscadgen:dev openscadgen --version

# Run tests
docker run --rm -v $(pwd):/app openscadgen:dev just test
```

## Development Workflow

1. **Start development environment**:
   ```bash
   docker-compose up openscadgen-dev
   ```

2. **Edit code** - Changes are automatically reloaded

3. **Test changes**:
   ```bash
   # In another terminal
   docker exec -it <container> just test
   ```

4. **Build production image**:
   ```bash
   ./docker-build.sh prod
   ```

## Environment Variables

- `GO_ENV`: Set to `development` or `production`
- `OPENSCAD_PATH`: Custom OpenSCAD executable path (if needed)

If you expose openscadgen on a public server, avoid custom OpenSCAD command paths and extra argument injection unless you explicitly launch with `--dangerously-skip-permissions`.

## Volume Mounts

### Development
- `.:/app` - Full source code mounting for live reloading

### Production
- `./examples:/app/examples:ro` - Read-only examples
- `./output:/app/output` - Output directory for generated files

## Ports

- `6767` - openscadgen server port

## Health Checks

The production image includes health checks that verify:
- openscadgen binary is executable
- Version command works
- Container is responsive

## Security

- Production image runs as non-root user (`appuser`)
- Minimal Alpine Linux base image
- No unnecessary packages in production
- Read-only filesystem where possible

## Troubleshooting

### Build Issues
```bash
# Clean build (no cache)
docker build --no-cache -f Dockerfile.dev -t openscadgen:dev .

# Check build logs
docker build -f Dockerfile.dev -t openscadgen:dev . 2>&1 | tee build.log
```

### Runtime Issues
```bash
# Check container logs
docker logs <container_name>

# Debug container
docker run -it --rm --entrypoint /bin/bash openscadgen:dev

# Check OpenSCAD installation
docker run --rm openscadgen:latest openscad --version
```

### Performance
```bash
# Monitor resource usage
docker stats <container_name>

# Check image size
docker images openscadgen
```
