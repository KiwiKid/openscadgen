#!/bin/bash

# Docker build script for openscadgen
# Usage: ./docker-build.sh [dev|prod]

set -e

MODE=${1:-prod}
IMAGE_NAME="openscadgen"
TAG="latest"

case $MODE in
  dev)
    echo "Building development image..."
    docker build -f Dockerfile.dev -t ${IMAGE_NAME}:dev .
    echo "Development image built: ${IMAGE_NAME}:dev"
    echo "Run with: docker run -it --rm -p 6767:6767 -v \$(pwd):/app ${IMAGE_NAME}:dev"
    ;;
  prod)
    echo "Building production image..."
    docker build -f Dockerfile -t ${IMAGE_NAME}:${TAG} .
    echo "Production image built: ${IMAGE_NAME}:${TAG}"
    echo "Run with: docker run -it --rm -p 6767:6767 -v \$(pwd)/examples:/app/examples:ro ${IMAGE_NAME}:${TAG}"
    ;;
  *)
    echo "Usage: $0 [dev|prod]"
    echo "  dev  - Build development image with live reloading"
    echo "  prod - Build production image (default)"
    exit 1
    ;;
esac

echo ""
echo "Quick start commands:"
echo "  Development: docker-compose up openscadgen-dev"
echo "  Production:  docker-compose up openscadgen"
echo "  OpenSCAD:     docker-compose --profile openscad-only up openscad"
