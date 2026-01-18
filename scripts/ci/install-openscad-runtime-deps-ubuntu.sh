#!/usr/bin/env bash
set -euo pipefail

# Runtime deps for OpenSCAD binaries/AppImages on Ubuntu runners.
# Symptom: "error while loading shared libraries: libEGL.so.1: cannot open shared object file"
#
# Keep this minimal; add more deps only as we see concrete missing-libs errors in CI.

export DEBIAN_FRONTEND=noninteractive

sudo apt-get update -y
sudo apt-get install -y --no-install-recommends \
  libegl1 \
  libgl1

