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