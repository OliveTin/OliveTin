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

if [[ ! -f "${SCRIPT_DIR}/License.rtf" ]]; then
  echo "License.rtf not found: ${SCRIPT_DIR}/License.rtf" >&2
  exit 1
fi

INSTALLER_BANNER="${SCRIPT_DIR}/bitmaps/installer-banner.bmp"
INSTALLER_DIALOG="${SCRIPT_DIR}/bitmaps/installer-dialog.bmp"
if [[ ! -f "${INSTALLER_BANNER}" ]] || [[ ! -f "${INSTALLER_DIALOG}" ]]; then
  echo "Installer bitmaps not found: ${INSTALLER_BANNER} and ${INSTALLER_DIALOG}" >&2
  exit 1
fi

normalize_msi_version() {
  local raw="${1#v}"
  raw="${raw%%-*}"
  if [[ ! "${raw}" =~ ^[0-9]+(\.[0-9]+){0,3}$ ]]; then
    echo "Invalid MSI version (expected major[.minor[.patch[.build]]]): ${1}" >&2
    return 1
  fi
  local -a parts=()
  IFS='.' read -r -a parts <<<"${raw}"
  local major="${parts[0]:-0}"
  local minor="${parts[1]:-0}"
  local patch="${parts[2]:-0}"
  printf '%s.%s.%s' "${major}" "${minor}" "${patch}"
}

VERSION="${VERSION:-}"
if [[ -z "${VERSION}" ]]; then
  VERSION="$(git -C "${REPO_ROOT}" describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || true)"
fi
if [[ -z "${VERSION}" ]]; then
  echo "Could not determine release version; set VERSION explicitly" >&2
  exit 1
fi
MSI_VERSION="$(normalize_msi_version "${VERSION}")" || exit 1

STAGING="$(mktemp -d)"
APP_STAGING="$(mktemp -d)"
HEAT_WXS="$(mktemp)"
WIXL_EXT_STAGING=""
trap 'rm -rf "${STAGING}" "${APP_STAGING}" "${HEAT_WXS}" "${WIXL_EXT_STAGING}"' EXIT

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

if ! command -v objdump >/dev/null; then
  echo "objdump is required to verify OliveTin.exe Windows resources" >&2
  exit 1
fi
if ! objdump -h "${SOURCE_ROOT}/OliveTin.exe" | grep -q '[[:space:]]\.rsrc[[:space:]]'; then
  echo "OliveTin.exe is missing embedded Windows resources (.rsrc); use main: . in .goreleaser.yml" >&2
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

WIXL_EXT_DIR=""
for candidate in /usr/share/wixl*/ext; do
  if [[ -d "${candidate}/ui/bitmaps" ]]; then
    WIXL_EXT_DIR="${candidate}"
    break
  fi
done
if [[ -z "${WIXL_EXT_DIR}" ]]; then
  echo "wixl UI extension not found under /usr/share/wixl*/ext" >&2
  exit 1
fi

WIXL_EXT_STAGING="$(mktemp -d)"
cp -a "${WIXL_EXT_DIR}/." "${WIXL_EXT_STAGING}/"
cp "${INSTALLER_BANNER}" "${WIXL_EXT_STAGING}/ui/bitmaps/bannrbmp.bmp"
cp "${INSTALLER_DIALOG}" "${WIXL_EXT_STAGING}/ui/bitmaps/dlgbmp.bmp"

(
  cd "${SCRIPT_DIR}"
  wixl \
    -v \
    -a x64 \
    --ext ui \
    --extdir "${WIXL_EXT_STAGING}" \
    -D "Version=${MSI_VERSION}" \
    -D "Win64=yes" \
    -D "SourceDir=${APP_STAGING}" \
    -D "ConfigSource=${SOURCE_ROOT}/config.yaml" \
    -o "${MSI_PATH}" \
    OliveTin.wxs \
    "${HEAT_WXS}"
)

if ! msiinfo export "${MSI_PATH}" Media 2>/dev/null | grep -q '#cab1.cab'; then
  echo "MSI cabinet is not embedded (expected #cab1.cab in Media table); check EmbedCab in OliveTin.wxs" >&2
  exit 1
fi

echo "Built ${MSI_PATH}"
