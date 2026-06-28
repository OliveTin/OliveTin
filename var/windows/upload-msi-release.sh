#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST_DIR="${DIST_DIR:-${REPO_ROOT}/dist}"
ARCH="${ARCH:-amd64}"
MSI_NAME="OliveTin-windows-${ARCH}.msi"
MSI_PATH="${DIST_DIR}/${MSI_NAME}"
TAG="${1:-}"

if [[ -z "${TAG}" ]]; then
  echo "Usage: $(basename "$0") <release-tag>" >&2
  exit 1
fi

if [[ ! -f "${MSI_PATH}" ]]; then
  echo "MSI not found: ${MSI_PATH}" >&2
  exit 1
fi

if ! command -v gh >/dev/null; then
  echo "gh is required to upload the MSI to GitHub releases" >&2
  exit 1
fi

checksums_path="${DIST_DIR}/checksums.txt"
new_checksum="$(cd "${DIST_DIR}" && sha256sum "${MSI_NAME}")"
if [[ -f "${checksums_path}" ]] && grep -qF " ${MSI_NAME}" "${checksums_path}"; then
  tmp="$(mktemp)"
  grep -vF " ${MSI_NAME}" "${checksums_path}" > "${tmp}" || true
  printf '%s\n' "${new_checksum}" >> "${tmp}"
  mv "${tmp}" "${checksums_path}"
elif [[ -f "${checksums_path}" ]]; then
  printf '%s\n' "${new_checksum}" >> "${checksums_path}"
else
  printf '%s\n' "${new_checksum}" > "${checksums_path}"
fi

gh release upload "${TAG}" "${MSI_PATH}" --clobber
if [[ -f "${checksums_path}" ]]; then
  gh release upload "${TAG}" "${checksums_path}" --clobber
fi

echo "Uploaded ${MSI_NAME} to release ${TAG}"
