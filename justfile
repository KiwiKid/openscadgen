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

alias fonts := fonts-faces

# Recommended: ~hundreds of fonts — :lang=en + scalable outlines, no macOS “.” UI families, one entry per family (OpenSCAD picks default style).
fonts-small:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v fc-list >/dev/null 2>&1; then
        echo "fc-list not on PATH — install fontconfig (e.g. brew install fontconfig)" >&2
        exit 1
    fi
    echo "test_font = ["
    fc-list ':lang=en:scalable=true' -f '%{family}\n' \
        | tr ',' '\n' | sed 's/^[ \t]*//;s/[ \t]*$//' | grep -v '^$' | grep -v '^\.' \
        | sort -u | sed 's/\\/\\\\/g; s/"/\\"/g; s/.*/  "&",/'
    echo "]"

# Like fonts-small but keeps :style=… only for “regular”-ish weights (Regular, Plain, Roman, …). Still far smaller than fonts-faces.
fonts-faces-small:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v fc-list >/dev/null 2>&1; then
        echo "fc-list not on PATH — install fontconfig (e.g. brew install fontconfig)" >&2
        exit 1
    fi
    echo "test_font = ["
    fc-list ':lang=en:scalable=true' -f '%{family}\t%{style}\n' | awk -F'\t' 'NF>=2 {
        split($2, st, ",")
        style=st[1]; gsub(/^[ \t]+|[ \t]+$/, "", style)
        if (style == "") next
        s=tolower(style)
        if (s !~ /^(regular|plain|roman|normal|book|standard)$/) next
        n=split($1, a, ",")
        for (i=1; i<=n; i++) {
            fam=a[i]; gsub(/^[ \t]+|[ \t]+$/, "", fam)
            if (fam != "" && fam !~ /^\./) print fam ":style=" style
        }
    }' | sort -u | sed 's/\\/\\\\/g; s/"/\\"/g; s/.*/  "&",/'
    echo "]"

# Print TOML `test_font = [ ... ]` for OpenSCAD (Family:style=Style). Needs `fc-list` (brew install fontconfig).
# Optional lang: BCP 47 tag for fontconfig, e.g. `just fonts-faces en` → only fonts that declare :lang=en coverage
# (Latin/English UI fonts; many fonts support multiple languages, so this is “supports English”, not “English-only”).
fonts-faces lang="":
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v fc-list >/dev/null 2>&1; then
        echo "fc-list not on PATH — install fontconfig (e.g. brew install fontconfig)" >&2
        exit 1
    fi
    PAT=""
    if [[ -n "{{lang}}" ]]; then PAT=":lang={{lang}}"; fi
    echo "test_font = ["
    if [[ -n "$PAT" ]]; then
        fc-list "$PAT" -f '%{family}\t%{style}\n'
    else
        fc-list -f '%{family}\t%{style}\n'
    fi | awk -F'\t' 'NF>=2 {
        split($2, st, ",")
        style=st[1]; gsub(/^[ \t]+|[ \t]+$/, "", style)
        if (style == "") next
        n=split($1, a, ",")
        for (i=1; i<=n; i++) {
            fam=a[i]; gsub(/^[ \t]+|[ \t]+$/, "", fam)
            if (fam != "") print fam ":style=" style
        }
    }' | sort -u | sed 's/\\/\\\\/g; s/"/\\"/g; s/.*/  "&",/'
    echo "]"

# Same as fonts-faces with :lang=en (common “Latin / English” filter).
fonts-en:
    just fonts-faces en

# Same pool, one line per family name (comma-aliases split). Optional lang like fonts-faces.
fonts-families lang="":
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v fc-list >/dev/null 2>&1; then
        echo "fc-list not on PATH — install fontconfig (e.g. brew install fontconfig)" >&2
        exit 1
    fi
    PAT=""
    if [[ -n "{{lang}}" ]]; then PAT=":lang={{lang}}"; fi
    echo "test_font = ["
    if [[ -n "$PAT" ]]; then
        fc-list "$PAT" -f '%{family}\n'
    else
        fc-list -f '%{family}\n'
    fi | tr ',' '\n' | sed 's/^[ \t]*//;s/[ \t]*$//' | grep -v '^$' | sort -u | sed 's/\\/\\\\/g; s/"/\\"/g; s/.*/  "&",/'
    echo "]"

fonts-families-en:
    just fonts-families en

alias fonts-tiny := fonts-small