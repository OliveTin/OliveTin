#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST_DIR="${DIST_DIR:-${REPO_ROOT}/dist}"
ARCH="${ARCH:-amd64}"
ZIP_NAME="OliveTin-windows-${ARCH}.zip"
ZIP_PATH="${DIST_DIR}/${ZIP_NAME}"
MSI_NAME="OliveTin-windows-${ARCH}.msi"
MSI_PATH="${DIST_DIR}/${MSI_NAME}"

if [[ ! -f "${ZIP_PATH}" ]]; then
  echo "Windows archive not found: ${ZIP_PATH}" >&2
  exit 1
fi

if ! command -v wixl >/dev/null || ! command -v wixl-heat >/dev/null; then
  echo "wixl and wixl-heat are required (install the wixl/msitools package)" >&2
  exit 1
fi

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

VERSION="${VERSION:-}"
if [[ -z "${VERSION}" ]]; then
  VERSION="$(git -C "${REPO_ROOT}" describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || true)"
fi
if [[ -z "${VERSION}" ]]; then
  echo "Could not determine release version; set VERSION explicitly" >&2
  exit 1
fi
WINDOWS_VERSION="$(normalize_windows_version "${VERSION}")"

STAGING="$(mktemp -d)"
APP_STAGING="$(mktemp -d)"
HEAT_WXS="$(mktemp)"
trap 'rm -rf "${STAGING}" "${APP_STAGING}" "${HEAT_WXS}"' EXIT

unzip -q "${ZIP_PATH}" -d "${STAGING}"
SOURCE_ROOT="${STAGING}/OliveTin-windows-${ARCH}"

if [[ ! -f "${SOURCE_ROOT}/OliveTin.exe" ]]; then
  echo "OliveTin.exe not found in ${SOURCE_ROOT}" >&2
  exit 1
fi

if [[ ! -f "${SOURCE_ROOT}/config.yaml" ]]; then
  echo "config.yaml not found in ${SOURCE_ROOT}" >&2
  exit 1
fi

mkdir -p "${APP_STAGING}/webui"
cp "${SOURCE_ROOT}/OliveTin.exe" "${APP_STAGING}/"
cp -a "${SOURCE_ROOT}/webui/." "${APP_STAGING}/webui/"

(
  cd "${APP_STAGING}"
  find . -type f | sed 's|^\./||'
) | wixl-heat \
  -p "" \
  --component-group CG.AppFiles \
  --var var.SourceDir \
  --directory-ref INSTALLDIR \
  --win64 \
  > "${HEAT_WXS}"

wixl \
  -v \
  -a x64 \
  -D "Version=${WINDOWS_VERSION}" \
  -D "Win64=yes" \
  -D "SourceDir=${APP_STAGING}" \
  -D "ConfigSource=${SOURCE_ROOT}/config.yaml" \
  -o "${MSI_PATH}" \
  "${SCRIPT_DIR}/OliveTin.wxs" \
  "${HEAT_WXS}"

echo "Built ${MSI_PATH}"
