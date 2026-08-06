# Security Advisory Duplicates — Maintainer Guide

This document lists known duplicate security advisory clusters for [OliveTin/OliveTin](https://github.com/OliveTin/OliveTin). When triaging new reports, check here and [open advisories](https://github.com/OliveTin/OliveTin/security/advisories) before accepting.

**Duplicate policy:** the earliest reporter on the canonical advisory receives primary credit. Later reporters are credited on the canonical advisory when closed as duplicates. See [SECURITY.md](../SECURITY.md).

## Triage checklist

1. Search open advisories for the same component and attack path.
2. Match against clusters below.
3. If duplicate: close the newer advisory, link to canonical, add reporter to canonical credits.
4. If unique: accept, patch on a private branch, reassess CVSS with OliveTin context (see SECURITY.md — OliveTin is intentional RCE by design).
5. Merge fix to `next`, publish advisory, credit reporters in advisory body (not commit message).

---

## shellAfterCompleted command injection

**Canonical:** [GHSA-vc6p-m6vx-6cwq](https://github.com/OliveTin/OliveTin/security/advisories/GHSA-vc6p-m6vx-6cwq) — reporter **knight-yagami** (2026-03-04)

Untrusted command output or template variables interpolated into `shellAfterCompleted` and executed via `sh -c`.

| GHSA | Reporter | Status | Notes |
|------|----------|--------|-------|
| GHSA-vc6p-m6vx-6cwq | knight-yagami | closed (canonical) | Original report |
| GHSA-v5gc-hqpq-227p | 0xkakash1 | closed (duplicate) | Output template variant |
| GHSA-m7wr-wj5j-7459 | Ryu7zz | duplicate | Webhook `exec` → output → `shellAfterCompleted` |
| GHSA-cjxm-x848-6vmc | Yesuhei | duplicate | Missing shell safety on after-completion |
| GHSA-j9p9-36jc-2v8w | anushkavirgaonkar | duplicate | Same root cause |

**Fix:** shell-quote `output`/`exitCode` before template render; block `shellAfterCompleted` for webhook-tagged actions.

**CVSS note:** requires admin-configured `shellAfterCompleted` and attacker influence on output — typically PR:H not PR:N.

---

## OAuth2 state map memory exhaustion (DoS)

**Canonical:** [GHSA-xpxj-f2fm-rqch](https://github.com/OliveTin/OliveTin/security/advisories/GHSA-xpxj-f2fm-rqch) — reporter **knight-yagami** (2026-03-04)

Unauthenticated `/oauth/login` grows `registeredStates` without TTL or cap.

| GHSA | Reporter | Status | Notes |
|------|----------|--------|-------|
| GHSA-xpxj-f2fm-rqch | knight-yagami | closed (canonical) | Original report |
| GHSA-cj96-c55v-2f3c | Dredsen | duplicate | Same unbounded map |

**Fix:** TTL sweep (match 15-minute cookie MaxAge), max map size, cleanup on failed callback.

**CVSS note:** unauthenticated DoS — reported 7.5 is appropriate.

---

## URL argument type — unrestricted URI schemes (SSRF / file read)

**Canonical:** [GHSA-45pc-w4ph-hrq4](https://github.com/OliveTin/OliveTin/security/advisories/GHSA-45pc-w4ph-hrq4) — reporter **fg0x0** (2026-03-09)

`url` type accepts `file://`, `gopher://`, etc. Blocked in `shell:` mode but still validated weakly for `exec:` actions.

| GHSA | Reporter | Status | Notes |
|------|----------|--------|-------|
| GHSA-45pc-w4ph-hrq4 | fg0x0 | closed (canonical) | Original report |
| GHSA-cchg-25m4-q6rj | anushkavirgaonkar | duplicate | Same scheme validation gap |

**Fix:** allowlist `http`/`https` in `typeSafetyCheckUrl`.

**CVSS note:** admin must configure `exec:` action passing URL to external tool — PR:H.

---

## Custom `regex:` argument type in shell actions

**Canonical:** [GHSA-xc5w-4v5w-7x65](https://github.com/OliveTin/OliveTin/security/advisories/GHSA-xc5w-4v5w-7x65) — reporter **Ayantaker** (2026-05-06)

`regex:` types not in shell denylist; partial `MatchString` allows injection suffixes.

| GHSA | Reporter | Status | Notes |
|------|----------|--------|-------|
| GHSA-xc5w-4v5w-7x65 | Ayantaker | canonical | Missing denylist entry |
| GHSA-gvxq-7gvp-4ggr | anushkavirgaonkar | duplicate | Unanchored partial match |

**Fix:** deny `regex:` in shell mode; enforce full-string match for custom regex types.

---

## Shell denylist incomplete (post CVE-2026-27626)

**Canonical:** [GHSA-c26w-h42g-jfp9](https://github.com/OliveTin/OliveTin/security/advisories/GHSA-c26w-h42g-jfp9) — reporter **sec-reex** (2026-07-03)

CVE-2026-27626 added `password` to denylist only; `html`, `confirmation`, and choiceless `checkbox` still skip validation and are allowed in `shell:` actions.

**Fix:** extend `checkShellArgumentSafety` denylist.

---

## StartActionAndWait logs ACL bypass

**Canonical:** [GHSA-jm28-2wcr-qf3h](https://github.com/OliveTin/OliveTin/security/advisories/GHSA-jm28-2wcr-qf3h) — reporter **offset** (2026-03-12)

`StartActionAndWait` / `StartActionByGetAndWait` return full log output without `logs` ACL check.

No known duplicates.

**Fix:** apply `isLogEntryAllowed` before returning `LogEntry`.

**CVSS note:** requires authenticated user with `exec` but not `logs` — typically 4.3–5.3 not 6.5.

---

## Easy to confuse (not duplicates)

| Topic | Advisories | Distinction |
|-------|------------|-------------|
| OAuth2 state DoS vs OAuth2 auth bypass | GHSA-xpxj vs GHSA-3v7p | DoS fills state map; bypass spoofs `authHttpHeaderUsername` |
| `shellAfterCompleted` vs direct `shell` injection | GHSA-vc6p vs GHSA-49gm | Second-order via output vs first-order argument injection |
| `ValidateArgumentType` enumeration | GHSA-f637 vs GHSA-x6q3 | Same issue; GHSA-f637 published |
