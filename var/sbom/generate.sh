#!/usr/bin/env bash
# Generate the per-release SBOM (CycloneDX + SPDX) for OliveTin.
#
# Called by GoReleaser (sboms, artifacts: any) with the working directory set to dist/.
# Usage: generate.sh <version> [dist-dir]
#
# One SBOM covers every archive/package/MSI because they all ship the same two things:
#   - the Go service binary: modules are read from its buildinfo, NOT from service/go.mod
#     (go.mod also carries tool deps like golangci-lint that are never linked);
#   - the webui bundle: npm runtime deps from frontend/package-lock.json (dev deps excluded).
# Container images get their own per-platform SBOM attestations from BuildKit (dockers_v2 sbom: true).
set -euo pipefail

VERSION="${1:?version is required}"
DIST_DIR="$(cd "${2:-.}" && pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# linux/amd64 is the reference: its module set is a superset of the other targets
# (darwin/windows only drop the Linux-only github.com/prometheus/procfs).
BINARY="${DIST_DIR}/OliveTin_linux_amd64_v1/OliveTin"
OUT="${DIST_DIR}/OliveTin-${VERSION}-sbom"

command -v syft >/dev/null || { echo "syft is required to generate SBOMs" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required to check SBOMs" >&2; exit 1; }
[[ -f "${BINARY}" ]] || { echo "Reference binary not found: ${BINARY}" >&2; exit 1; }

STAGE="$(mktemp -d)"
trap 'rm -rf "${STAGE}"' EXIT

mkdir -p "${STAGE}/bin" "${STAGE}/webui"
cp "${BINARY}" "${STAGE}/bin/OliveTin"
cp "${REPO_ROOT}/frontend/package.json" "${REPO_ROOT}/frontend/package-lock.json" "${STAGE}/webui/"

SYFT_JAVASCRIPT_INCLUDE_DEV_DEPENDENCIES=false syft scan "dir:${STAGE}" \
  --source-name OliveTin \
  --source-version "${VERSION}" \
  -o "cyclonedx-json=${OUT}.cdx.json" \
  -o "spdx-json=${OUT}.spdx.json"

# Sanity checks: catch a scan that picked up go.mod/dev deps or missed a whole ecosystem.
names="$(jq -r '.components[].name' "${OUT}.cdx.json")"

for required in github.com/expr-lang/expr @connectrpc/connect; do
  grep -qxF "${required}" <<<"${names}" || { echo "SBOM is missing expected component: ${required}" >&2; exit 1; }
done

for forbidden in github.com/golangci/golangci-lint/v2 vite eslint; do
  if grep -qxF "${forbidden}" <<<"${names}"; then
    echo "SBOM contains a build/dev-only component: ${forbidden}" >&2
    exit 1
  fi
done

echo "SBOM written: ${OUT}.cdx.json ($(wc -l <<<"${names}") components), ${OUT}.spdx.json"
