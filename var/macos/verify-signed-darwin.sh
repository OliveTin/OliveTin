#!/usr/bin/env bash
# Post-sign: fail if quill embedded the broken designated requirement
# certificate root[field.1.2.840.113635.100.6.2.6] (missing Apple Root in P12).
set -euo pipefail

if [[ -z "${MACOS_SIGN_P12:-}" ]]; then
  echo "MACOS_SIGN_P12 unset; skipping darwin signature check."
  exit 0
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST_DIR="${REPO_ROOT}/dist"

BAD_REQ='certificate root\[field\.1\.2\.840\.113635\.100\.6\.2\.6\]'
GOOD_REQ='certificate 1\[field\.1\.2\.840\.113635\.100\.6\.2\.6\]'

if ! command -v quill >/dev/null 2>&1; then
  echo "Installing quill..."
  go install github.com/anchore/quill/cmd/quill@latest
  export PATH="$(go env GOPATH)/bin:${PATH}"
fi

shopt -s nullglob
archives=("${DIST_DIR}"/OliveTin-darwin-*.tar.gz)
if [[ "${#archives[@]}" -eq 0 ]]; then
  echo "No OliveTin-darwin-*.tar.gz archives found under ${DIST_DIR}." >&2
  exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

checked=0
for archive in "${archives[@]}"; do
  name="$(basename "${archive}" .tar.gz)"
  extract_dir="${tmpdir}/${name}"
  mkdir -p "${extract_dir}"
  tar -xzf "${archive}" -C "${extract_dir}"

  # Prefer the top-level binary; archives also ship helper scripts named OliveTin.
  binary="${extract_dir}/${name}/OliveTin"
  if [[ ! -f "${binary}" ]]; then
    binary="$(find "${extract_dir}" -type f -path "*/OliveTin" ! -path "*/var/*" | head -n 1)"
  fi
  if [[ -z "${binary}" || ! -f "${binary}" ]]; then
    echo "OliveTin binary not found inside ${archive}." >&2
    exit 1
  fi
  if ! file "${binary}" | grep -qi 'Mach-O'; then
    echo "Expected a Mach-O binary at ${binary}, got: $(file "${binary}")" >&2
    exit 1
  fi


  echo "Checking designated requirement in ${archive}..."
  describe_out="$(quill describe "${binary}")"
  if echo "${describe_out}" | grep -qE "${BAD_REQ}"; then
    echo "Broken designated requirement in ${archive}:" >&2
    echo "  found certificate root[field.1.2.840.113635.100.6.2.6]" >&2
    echo "MACOS_SIGN_P12 is missing Apple Root CA. Rebuild per docs/modules/dev/pages/signing.adoc." >&2
    exit 1
  fi
  if ! echo "${describe_out}" | grep -qE "${GOOD_REQ}"; then
    echo "Expected designated requirement with certificate 1[...] not found in ${archive}." >&2
    echo "${describe_out}" >&2
    exit 1
  fi
  checked=$((checked + 1))
done

echo "Darwin signature check OK (${checked} archive(s); designated requirement uses certificate 1[...])."
