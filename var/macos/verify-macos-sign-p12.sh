#!/usr/bin/env bash
# Preflight: ensure MACOS_SIGN_P12 has the full Developer ID chain
# (leaf + Developer ID G2 + Apple Root CA). Quill embeds an unsatisfiable
# designated requirement when Apple Root CA is missing — see signing.adoc.
set -euo pipefail

if [[ -z "${MACOS_SIGN_P12:-}" ]]; then
  echo "MACOS_SIGN_P12 unset; skipping macOS signing preflight."
  exit 0
fi

if [[ -z "${MACOS_SIGN_PASSWORD:-}" ]]; then
  echo "MACOS_SIGN_PASSWORD is required when MACOS_SIGN_P12 is set." >&2
  exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

p12_path="${tmpdir}/Certificates.p12"
# GitHub secrets may include trailing newlines; strip them before decode.
printf '%s' "${MACOS_SIGN_P12}" | tr -d '\n\r' | base64 -d >"${p12_path}"

pem_out="${tmpdir}/certs.pem"
extract_p12() {
  local extra_args=("${@}")
  openssl pkcs12 -in "${p12_path}" -nodes -passin "pass:${MACOS_SIGN_PASSWORD}" \
    "${extra_args[@]}" -out "${pem_out}" 2>/dev/null
}

if ! extract_p12; then
  # OpenSSL 3 may need -legacy for older P12 exports.
  extract_p12 -legacy
fi

cert_dir="${tmpdir}/certs"
mkdir -p "${cert_dir}"
awk -v dir="${cert_dir}" '
  /-----BEGIN CERTIFICATE-----/ { n++; file = sprintf("%s/cert-%02d.pem", dir, n) }
  n { print > file }
' "${pem_out}"

cert_count="$(find "${cert_dir}" -name 'cert-*.pem' | wc -l | tr -d ' ')"
if [[ "${cert_count}" -ne 3 ]]; then
  echo "MACOS_SIGN_P12 must contain exactly 3 certificates (leaf + G2 + Apple Root CA); found ${cert_count}." >&2
  echo "Rebuild the .p12 per docs/modules/dev/pages/signing.adoc." >&2
  exit 1
fi

found_apple_root=0
for cert in "${cert_dir}"/cert-*.pem; do
  subject="$(openssl x509 -in "${cert}" -noout -subject 2>/dev/null || true)"
  if [[ "${subject}" == *"Apple Root CA"* ]]; then
    found_apple_root=1
    break
  fi
done

if [[ "${found_apple_root}" -ne 1 ]]; then
  echo "MACOS_SIGN_P12 is missing Apple Root CA in the certificate chain." >&2
  echo "Rebuild the .p12 per docs/modules/dev/pages/signing.adoc." >&2
  exit 1
fi

echo "MACOS_SIGN_P12 preflight OK (3 certificates, including Apple Root CA)."
