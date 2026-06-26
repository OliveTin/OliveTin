#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TIMEOUT="${GORELEASER_TIMEOUT:-60m}"

if [[ -z "${VERSION:-}" ]]; then
  echo "VERSION is required to build the Windows MSI" >&2
  exit 1
fi

goreleaser release --clean --timeout "${TIMEOUT}" --skip=checksum,publish "$@"

"${SCRIPT_DIR}/build-msi.sh"

goreleaser release --timeout "${TIMEOUT}" \
  --skip=validate,before,build,archive,nfpm,docker,sign,sbom,after,announce "$@"
