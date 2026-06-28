#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
GORELEASER_CONFIG="${REPO_ROOT}/.goreleaser.yml"
TIMEOUT="${GORELEASER_TIMEOUT:-60m}"

if [[ -z "${VERSION:-}" ]]; then
  echo "VERSION is required to build the Windows MSI" >&2
  exit 1
fi

build_config="$(mktemp)"
trap 'rm -f "${build_config}"' EXIT
awk '/^checksum:/{print; print "  disable: true"; next}1' "${GORELEASER_CONFIG}" > "${build_config}"

goreleaser release -f "${build_config}" --clean --timeout "${TIMEOUT}" --skip=publish "$@"

"${SCRIPT_DIR}/build-msi.sh"

GORELEASER_SKIP_BUILD=1 goreleaser release --timeout "${TIMEOUT}" \
  --skip=validate,before,archive,nfpm,docker,sign,sbom,announce "$@"
