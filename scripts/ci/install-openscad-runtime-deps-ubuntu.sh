#!/usr/bin/env bash
set -euo pipefail

# Runtime deps for OpenSCAD binaries/AppImages on Ubuntu runners.
# Symptom: "error while loading shared libraries: libEGL.so.1: cannot open shared object file"
#
# Keep this minimal-ish, but prefer "it works" over shaving one package.

export DEBIAN_FRONTEND=noninteractive

echo "[openscad-runtime-deps] uname: $(uname -a)"
if [[ -f /etc/os-release ]]; then
  echo "[openscad-runtime-deps] /etc/os-release:"
  cat /etc/os-release
fi

if ! command -v apt-get >/dev/null 2>&1; then
  echo "[openscad-runtime-deps] ERROR: apt-get not found. Are you running tests inside a container (alpine/distroless)?" >&2
  exit 2
fi

sudo apt-get update -y

# Prefer canonical package names, but fall back where distros rename/virtualize.
sudo apt-get install -y --no-install-recommends \
  libgl1 \
  libopengl0 \
  libglx0 \
  || true

if ! sudo apt-get install -y --no-install-recommends libegl1; then
  # Some images use mesa-provided naming.
  sudo apt-get install -y --no-install-recommends libegl1-mesa
fi

# Hard check: fail early with a useful message if we still can't see libEGL.so.1.
if command -v ldconfig >/dev/null 2>&1; then
  if ! (ldconfig -p 2>/dev/null | grep -Fq "libEGL.so.1"); then
    echo "[openscad-runtime-deps] ERROR: ldconfig can't see libEGL.so.1 after install" >&2
    ldconfig -p 2>/dev/null | grep -E "libEGL|libGLX|libOpenGL|libGL\.so" || true
    exit 3
  fi
else
  # Fallback check (covers some minimal images).
  if ! (test -e /usr/lib/x86_64-linux-gnu/libEGL.so.1 || test -e /usr/lib/aarch64-linux-gnu/libEGL.so.1 || test -e /usr/lib/libEGL.so.1); then
    echo "[openscad-runtime-deps] ERROR: libEGL.so.1 not found in common paths after install" >&2
    exit 3
  fi
fi

echo "[openscad-runtime-deps] OK: libEGL.so.1 present"

