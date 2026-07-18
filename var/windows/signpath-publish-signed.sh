#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST_DIR="${DIST_DIR:-${REPO_ROOT}/dist}"
ARCH="${ARCH:-amd64}"
ZIP_NAME="OliveTin-windows-${ARCH}.zip"
MSI_NAME="OliveTin-windows-${ARCH}.msi"
CHECKSUMS_NAME="checksums.txt"

usage() {
  echo "Usage: $(basename "$0") <release-tag> <signed-zip-path> <signed-msi-path>" >&2
  exit 1
}

TAG="${1:-}"
SIGNED_ZIP="${2:-}"
SIGNED_MSI="${3:-}"

if [[ -z "${TAG}" || -z "${SIGNED_ZIP}" || -z "${SIGNED_MSI}" ]]; then
  usage
fi

if [[ ! -f "${SIGNED_ZIP}" ]]; then
  echo "Signed zip not found: ${SIGNED_ZIP}" >&2
  exit 1
fi

if [[ ! -f "${SIGNED_MSI}" ]]; then
  echo "Signed MSI not found: ${SIGNED_MSI}" >&2
  exit 1
fi

if ! command -v gh >/dev/null; then
  echo "gh is required to update the GitHub release" >&2
  exit 1
fi

mkdir -p "${DIST_DIR}"
cp -f "${SIGNED_ZIP}" "${DIST_DIR}/${ZIP_NAME}"
cp -f "${SIGNED_MSI}" "${DIST_DIR}/${MSI_NAME}"

checksums_path="${DIST_DIR}/${CHECKSUMS_NAME}"
checksums_backup="${DIST_DIR}/${CHECKSUMS_NAME}.orig"
if ! gh release download "${TAG}" --pattern "${CHECKSUMS_NAME}" --dir "${DIST_DIR}" --clobber; then
  echo "Failed to download ${CHECKSUMS_NAME} from release ${TAG}" >&2
  exit 1
fi
if [[ ! -f "${checksums_path}" ]]; then
  echo "${CHECKSUMS_NAME} not found after download from release ${TAG}" >&2
  exit 1
fi
cp -f "${checksums_path}" "${checksums_backup}"

update_checksum() {
  local file_name="${1}"
  local new_checksum
  new_checksum="$(cd "${DIST_DIR}" && sha256sum "${file_name}")"

  if [[ -f "${checksums_path}" ]] && grep -qF " ${file_name}" "${checksums_path}"; then
    local tmp
    tmp="$(mktemp)"
    grep -vF " ${file_name}" "${checksums_path}" > "${tmp}" || true
    printf '%s\n' "${new_checksum}" >> "${tmp}"
    mv "${tmp}" "${checksums_path}"
  else
    printf '%s\n' "${new_checksum}" >> "${checksums_path}"
  fi
}

update_checksum "${ZIP_NAME}"
update_checksum "${MSI_NAME}"

# Replace binaries first so a failed checksums upload leaves the draft recoverable.
gh release upload "${TAG}" \
  "${DIST_DIR}/${ZIP_NAME}" \
  "${DIST_DIR}/${MSI_NAME}" \
  --clobber

if ! gh release upload "${TAG}" "${checksums_path}" --clobber; then
  echo "Failed to upload updated ${CHECKSUMS_NAME}; restoring previous asset" >&2
  restore_dir="$(mktemp -d)"
  cp -f "${checksums_backup}" "${restore_dir}/${CHECKSUMS_NAME}"
  gh release upload "${TAG}" "${restore_dir}/${CHECKSUMS_NAME}" --clobber
  rm -rf "${restore_dir}"
  exit 1
fi

gh release edit "${TAG}" --draft=false

echo "Published signed ${ZIP_NAME} and ${MSI_NAME} on release ${TAG}"
