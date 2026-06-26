#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SERVICE_DIR="${REPO_ROOT}/service"
VERSIONINFO_JSON="${SCRIPT_DIR}/versioninfo.json"
ICON_PATH="${SCRIPT_DIR}/OliveTin.ico"
MANIFEST_PATH="${SCRIPT_DIR}/OliveTin.exe.manifest"
GOVERSIONINFO_VERSION="${GOVERSIONINFO_VERSION:-v1.5.0}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [version]

Generate Windows resource (.syso) files for embedding in OliveTin.exe.

  version   Release version (e.g. 3.0.0 or v3.0.0). Defaults to VERSION env,
            then the latest git tag, then 0.0.0.
EOF
}

normalize_windows_version() {
  local raw="${1#v}"
  raw="${raw%%-*}"
  if [[ ! "${raw}" =~ ^[0-9]+(\.[0-9]+){0,3}$ ]]; then
    echo "0.0.0.0"
    return
  fi
  local -a parts=()
  IFS='.' read -r -a parts <<<"${raw}"
  local major="${parts[0]:-0}"
  local minor="${parts[1]:-0}"
  local patch="${parts[2]:-0}"
  local build="${parts[3]:-0}"
  printf '%s.%s.%s.%s' "${major}" "${minor}" "${patch}" "${build}"
}

resolve_version() {
  if [[ -n "${VERSION:-}" ]]; then
    echo "${VERSION}"
    return
  fi
  if [[ $# -gt 0 && -n "${1:-}" ]]; then
    echo "${1}"
    return
  fi
  if git -C "${REPO_ROOT}" describe --tags --abbrev=0 >/dev/null 2>&1; then
    git -C "${REPO_ROOT}" describe --tags --abbrev=0
    return
  fi
  echo "0.0.0"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ ! -f "${VERSIONINFO_JSON}" ]]; then
  echo "versioninfo.json not found: ${VERSIONINFO_JSON}" >&2
  exit 1
fi

if [[ ! -f "${ICON_PATH}" ]]; then
  echo "icon not found: ${ICON_PATH}" >&2
  exit 1
fi

WINDOWS_VERSION="$(normalize_windows_version "$(resolve_version "${1:-}")")"
echo "Generating Windows resources for version ${WINDOWS_VERSION}"

go install "github.com/josephspurrier/goversioninfo/cmd/goversioninfo@${GOVERSIONINFO_VERSION}"

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

(
  cd "${WORK_DIR}"
  goversioninfo \
    -64 \
    -platform-specific \
    -icon="${ICON_PATH}" \
    -manifest="${MANIFEST_PATH}" \
    -file-version="${WINDOWS_VERSION}" \
    -product-version="${WINDOWS_VERSION}" \
    "${VERSIONINFO_JSON}"
)

rm -f "${SERVICE_DIR}"/resource_windows_*.syso
mv "${WORK_DIR}"/resource_windows_*.syso "${SERVICE_DIR}/"
echo "Wrote Windows resource files to ${SERVICE_DIR}"
