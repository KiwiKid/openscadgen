# Build the application
build:
    go build .

# Run tests
test:
    go test ./... -v

# Generate templ files
generate:
    go run github.com/a-h/templ/cmd/templ generate

# Watch for changes and regenerate templ files
watch:
    go run github.com/a-h/templ/cmd/templ generate --watch

# Clean generated files
clean:
    rm -f pkg/templates/*_templ.go

# Build and test everything
all: generate build test

# Install dependencies
deps:
    go mod tidy
    go mod download

# Show templ version
templ-version:
    go run github.com/a-h/templ/cmd/templ version

# Run a simple example
example:
    ./openscadgen -c ./examples/small-tray/config.toml -ow

# Live reload with air (templ + go)
air:
    air -- -sf ./examples/ -p 6767

# Live reload templ files only
air-templ:
    go run github.com/a-h/templ/cmd/templ generate -
    -watch

# Live reload go files only (after templ is generated)
air-go:
    air -c .air-go.toml

# Generate templ files and run air
dev: generate air

# Clean and start fresh
dev-clean: clean generate air

# Run air with specific config
air-full:
    air -c .air.toml

# Stop any running air processes
air-stop:
    pkill -f air || true

# Show air help
air-help:
    air --help

# Test the build process
test-build: generate build

# Run openscadgen on all examples
examples:
    for config in examples/*/config.toml; do \
        echo "Processing $$config..."; \
        ./openscadgen -c "$$config" -ow; \
    done

# Docker commands
docker-build-dev:
    ./docker-build.sh dev

docker-build-prod:
    ./docker-build.sh prod

docker-dev:
    docker-compose up openscadgen-dev

docker-prod:
    docker-compose up openscadgen

docker-stop:
    docker-compose down

docker-clean:
    docker-compose down -v
    docker system prune -f

# Test Docker build
docker-test:
    docker run --rm openscadgen:dev openscadgen --version
    docker run --rm openscadgen:latest openscadgen --version 