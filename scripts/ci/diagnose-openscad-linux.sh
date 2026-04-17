#!/usr/bin/env bash
set -euo pipefail

# Diagnostic helper for CI failures like:
#   /tmp/.mount_openscaXXXX/AppRun.wrapped: error while loading shared libraries: libEGL.so.1: cannot open shared object file
#
# Usage:
#   ./scripts/ci/diagnose-openscad-linux.sh
#
# Tip: If you're in GitHub Actions and using `container:`, run this INSIDE that container.

echo "=== openscad diagnose: basic env ==="
echo "date: $(date -u '+%Y-%m-%dT%H:%M:%SZ' || true)"
echo "uname: $(uname -a || true)"
echo "id: $(id || true)"

if [[ -f /etc/os-release ]]; then
  echo "--- /etc/os-release ---"
  cat /etc/os-release
fi

echo "--- container hints ---"
echo "/proc/1/cgroup:"
cat /proc/1/cgroup 2>/dev/null || true
echo "/.dockerenv present? $(test -f /.dockerenv && echo yes || echo no)"

echo
echo "=== openscad diagnose: openscad binary ==="
echo "PATH=$PATH"
echo "which openscad: $(command -v openscad || true)"
if [[ -e /usr/local/bin/openscad ]]; then
  echo "--- ls -la /usr/local/bin/openscad ---"
  ls -la /usr/local/bin/openscad || true
  echo "--- file /usr/local/bin/openscad ---"
  file /usr/local/bin/openscad || true
  echo "--- head -n 30 /usr/local/bin/openscad ---"
  head -n 30 /usr/local/bin/openscad 2>/dev/null || true
fi

echo
echo "=== openscad diagnose: ldconfig / libs ==="
if command -v ldconfig >/dev/null 2>&1; then
  echo "--- ldconfig -p | grep EGL/GL/OpenGL ---"
  ldconfig -p 2>/dev/null | grep -E "libEGL\.so\.1|libEGL\.so|libGLX\.so|libOpenGL\.so|libGL\.so" || true
else
  echo "ldconfig not found"
fi

echo "--- common libEGL locations ---"
ls -la /usr/lib/*/libEGL.so.1 /usr/lib/libEGL.so.1 2>/dev/null || true

echo
echo "=== openscad diagnose: run openscad --version (capture stderr) ==="
set +e
tmp_err="$(mktemp)"
openscad --version 1>/dev/null 2>"$tmp_err"
rc=$?
set -e
echo "openscad exit code: $rc"
echo "--- stderr ---"
cat "$tmp_err" || true

mount_path="$(grep -oE '/tmp/\.mount_[^/ ]+' "$tmp_err" | head -n1 || true)"
if [[ -n "${mount_path}" ]]; then
  echo
  echo "=== openscad diagnose: AppImage mount inspection ==="
  echo "mount_path=$mount_path"
  ls -la "$mount_path" || true
  echo "--- find top-level (first 2 levels) ---"
  find "$mount_path" -maxdepth 2 -type f -print 2>/dev/null | head -n 80 || true

  if [[ -e "$mount_path/AppRun.wrapped" ]]; then
    echo "--- file AppRun.wrapped ---"
    file "$mount_path/AppRun.wrapped" || true
    echo "--- ldd AppRun.wrapped ---"
    ldd "$mount_path/AppRun.wrapped" || true
  fi

  # Often the actual OpenSCAD binary is inside usr/bin.
  if [[ -e "$mount_path/usr/bin/openscad" ]]; then
    echo "--- file usr/bin/openscad ---"
    file "$mount_path/usr/bin/openscad" || true
    echo "--- ldd usr/bin/openscad ---"
    ldd "$mount_path/usr/bin/openscad" || true
  fi
fi

echo
echo "=== openscad diagnose: strace (optional) ==="
if command -v strace >/dev/null 2>&1; then
  set +e
  strace -f -e trace=file,openat openscad --version 2>&1 | grep -E "libEGL\.so\.1|ENOENT" | head -n 120
  set -e
else
  echo "strace not found (on Ubuntu: sudo apt-get update && sudo apt-get install -y strace)"
fi

rm -f "$tmp_err" || true
echo "=== openscad diagnose: done ==="

