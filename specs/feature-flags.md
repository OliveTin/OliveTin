# Spec: Feature flags

This spec describes how OliveTin opts into unfinished or optional product features for the whole installation.

---

## 1. Scope

Feature flags are **global** settings. They apply to every user and are not overridden by access control lists or policy.

They are distinct from policy options such as diagnostics or log list visibility, which may differ per user.

All feature-flagged functionality is **alpha / experimental**. Flags default to off. Security advisories and CVEs are not accepted for issues that only affect feature-flagged code; see the project security policy.

---

## 2. Configuration

Flags live under `features` in configuration.

- Each flag is a boolean.
- Omitted flags are **false**.
- Enabling a feature requires setting it to true explicitly.

Example:

```yaml
features:
  headerSearch: true
```

Unknown keys under `features` are ignored by the configuration loader (same as other unknown config keys).

---

## 3. Init projection

Every Init response includes a `features` object with the current flag values.

Clients must not assume a missing flag means enabled. Treat absent or false as off.

---

## 4. Server and client behavior

When a feature is off:

- The server skips work that exists only to support that feature (for example, building search hints when header search is off).
- The web UI hides controls and does not index or call APIs that exist only for that feature.

When a feature is on, normal access control still applies to the data and actions the feature exposes. Operators who enable a flag accept that the surface is experimental.

---

## 5. Header search

`features.headerSearch` controls the header QuickSearch control and Init search hints.

- Default: false (alpha / experimental).
- When false: Init omits search hints; the client does not show QuickSearch or populate the search index.
- When true: Init includes ACL-filtered search hints (unless guests must log in); the client shows QuickSearch after login is satisfied.

---

## 6. Adding a new flag

To add a flag:

1. Add a boolean field under `features` in configuration (default false).
2. Add the same field to the Init `features` message.
3. Project the value on every Init response.
4. Gate server work and UI behind the flag.
5. Document the flag as alpha / experimental and update this spec’s feature list.
6. Keep the security policy statement that CVEs are not accepted for feature-flagged functionality until the feature graduates (flag removed or stable default-on).
