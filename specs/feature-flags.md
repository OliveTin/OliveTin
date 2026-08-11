# Spec: Feature flags

This spec describes how OliveTin opts into unfinished or optional product features for the whole installation.

---

## 1. Scope

Feature flags are **global** settings. They apply to every user and are not overridden by access control lists or policy.

They are distinct from policy options such as diagnostics or log list visibility, which may differ per user.

All feature-flagged functionality is **alpha / experimental**. Flags default to off. Private security reports for vulnerabilities that affect enabled experimental functionality are accepted under the project security policy.

---

## 2. Configuration

Operators configure flags in a dedicated section of the installation configuration.

- Each flag is a boolean.
- Omitted flags are **false**.
- Enabling a feature requires setting it to true explicitly.

Unknown keys in that section are ignored by the configuration loader (same as other unknown config keys).

---

## 3. Bootstrap projection

Every client bootstrap response includes the current feature flag values.

Clients must not assume a missing flag means enabled. Treat absent or false as off.

---

## 4. Server and client behavior

When a feature is off:

- The server skips work that exists only to support that feature (for example, building search hints when header search is off).
- The web UI hides controls and does not index or call APIs that exist only for that feature.

When a feature is on, normal access control still applies to the data and actions the feature exposes. Operators who enable a flag accept that the surface is experimental.

---

## 5. Header search

A header search flag controls the header search control and bootstrap search hints.

- Default: false (alpha / experimental).
- When false: bootstrap omits search hints; the client does not show header search or populate the search index.
- When true: bootstrap includes access-control-filtered search hints (unless guests must log in); the client shows header search after login is satisfied.

---

## 6. Adding a new flag

To add a flag:

1. Add a boolean setting in the feature-flags configuration section (default false).
2. Expose the same setting on every client bootstrap response.
3. Gate server work and UI behind the flag.
4. Document the flag as alpha / experimental and update this spec’s feature list.
5. Keep the security policy aligned: private reports remain accepted for enabled experimental functionality until the feature graduates (flag removed or stable default-on).
