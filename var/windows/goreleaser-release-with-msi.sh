#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TIMEOUT="${GORELEASER_TIMEOUT:-60m}"
ZIP_PATH="${REPO_ROOT}/dist/OliveTin-windows-amd64.zip"

if [[ -z "${VERSION:-}" ]]; then
  echo "VERSION is required to build the Windows MSI" >&2
  exit 1
fi

wait_for_stable_file() {
  local file="${1}"
  local last_size=-1
  local stable_count=0

  while [[ "${stable_count}" -lt 2 ]]; do
    if [[ ! -f "${file}" ]]; then
      stable_count=0
      last_size=-1
      sleep 1
      continue
    fi

    local size
    size="$(stat -c%s "${file}")"
    if [[ "${size}" -gt 0 && "${size}" == "${last_size}" ]]; then
      stable_count=$((stable_count + 1))
    else
      stable_count=0
    fi
    last_size="${size}"
    sleep 1
  done
}

goreleaser release --clean --timeout "${TIMEOUT}" "$@" &
goreleaser_pid=$!

while [[ ! -f "${ZIP_PATH}" ]]; do
  if ! kill -0 "${goreleaser_pid}" 2>/dev/null; then
    wait "${goreleaser_pid}"
    echo "GoReleaser exited before Windows archive was created: ${ZIP_PATH}" >&2
    exit 1
  fi
  sleep 1
done

wait_for_stable_file "${ZIP_PATH}"
"${SCRIPT_DIR}/build-msi.sh"
wait "${goreleaser_pid}"
