## OliveTin – Agent Guide

This document helps AI agents contribute effectively to OliveTin.

If you are looking for OliveTin's AI policy, you can find it in `AI.md`.

### Project Overview
- **Service (Go)**: `service/` with business logic under `service/internal/*`
  - API (Connect RPC): `service/internal/api`
  - Command execution: `service/internal/executor`
  - HTTP frontends/proxy: `service/internal/httpservers`
  - Config/types/entities: `service/internal/config`, `service/internal/entities`
- **Frontend (Vue 3)**: `frontend/` (served by the service)
- **Integration tests**: `integration-tests/`
- **Protos/Generated**: `proto/`, `service/gen/...`
- **Specs**: `specs/` — Markdown specs that define how code should behave in human-readable form. When changing behavior in a spec-covered area, keep implementation and tests aligned with the spec; do not reference code or symbols in specs (English only).

### How to Run
- Run the server (dev):
  - From repo root: `go run ./service`
- Unit tests (Go):
  - From repo root: `cd service && make unittests`
- Code style (after editing code in `service/`):
  - From repo root: `cd service && make codestyle` (runs `go fmt` and `golangci-lint`; install linter via `make go-tools`)
- Documentation checks:
  - From repo root: `make docs-check` (validates documented config key casing and local AsciiDoc anchor links)
- Integration tests (Mocha + Selenium):
  - All tests: `make it` (from repo root; builds webui + service binary, then runs Mocha)
  - Single test: `cd integration-tests && npx --yes mocha tests/general/general.mjs`

### Test Notes and Gotchas
- The top-level Makefile does not expose `unittests`; use `cd service && make unittests`.
- Connect RPC API must be mounted correctly; in tests, create the handler via `GetNewHandler(ex)` and serve under `/api/`.
- Frontend “ready” state: the app sets `document.body` attribute `loaded-dashboard="<name>"` when loading a dashboard. Integration helpers that test dashboard functionality  wait for this before selecting elements. Certain conditions enforcing login will mean that this attribute is not set until a user is logged in.
- Modern UI uses Vue components:
  - Action buttons are rendered as `.action-button button`.
  - Logs and Diagnostics are Vue router links available via `/logs` and `/diagnostics`.
  - Some legacy DOM ids (e.g., `contentActions`) no longer exist; prefer class-based selectors.
- Hidden UI features:
  - Footer visibility is controlled by `showFooter` from Init API; tests may assert the footer is absent when config disables it.

### Coding Standards (Go)
- Avoid adding superflous comments that explain what the code is doing. Comments are only to describe business logic decisions.
- Prefer clear, descriptive names; avoid 1–2 letter identifiers.
- Use early returns and handle edge cases first.
- Do not swallow errors; propagate or log meaningfully.
- Match existing formatting; avoid unrelated reformatting.
- Be safe around nils in executor steps (e.g., guard `req.Binding` and `req.Binding.Action`).
- Cyclomatic complexity over 5 is not permitted (`gocyclo` `min-complexity: 5` in `service/.golangci.yml`).

### API and Execution Flow (High-level)
1. Client calls Connect RPC (e.g., `Init`, `GetDashboard`, `StartAction`).
2. API translates requests to `executor.ExecutionRequest` and calls `Executor.ExecRequest`.
3. Executor runs a chain of steps: request binding → concurrency/rate/ACL checks → arg parsing → exec → post-exec → logging/triggering.
4. Logs are stored and can be fetched via `ExecutionStatus`/`GetLogs`.

### Machine-readable API
- The OliveTin API is completely machine readable: it is defined in Protobuf and served with Connect RPC (Connect JSON/binary and gRPC-Web).
- Source of truth: `proto/olivetin/api/v1/olivetin.proto` (service `olivetin.api.v1.OliveTinApiService`), online at https://github.com/OliveTin/OliveTin/tree/main/proto
- Prefer reading the proto over guessing endpoints. Each RPC is callable as `POST /api/<MethodName>` with a protobuf-JSON body (lowerCamelCase fields, eg `{"bindingId": "..."}`).
- After changing the proto, regenerate code with `cd proto && make` (buf) — generated code lives in `service/gen/` and `frontend/resources/scripts/gen/`.
- User docs for the API: `docs/modules/ROOT/pages/api/intro.adoc`. LLM-oriented docs index: https://docs.olivetin.app/llm.txt

### Common Tasks
- Add/modify actions: update `config.yaml` and ensure `executor.RebuildActionMap()` is called when needed.
- Adjust dashboard rendering: see `service/internal/api/dashboards.go` and `apiActions.go`.
- Frontend behavior:
  - Router: `frontend/resources/vue/router.js`
  - Main shell/layout: `frontend/resources/vue/App.vue`
  - Action button behavior: `frontend/resources/vue/ActionButton.vue`

### Contributing Checklist
- Review the contributing guidelines at `CONTRIBUTING.adoc`.
- Review the AI guidance in `AI.md`.
- Review the pull request template at `.github/PULL_REQUEST_TEMPLATE.md`.
- When changing behaviour covered by a spec in `specs/`, ensure implementation and tests match the spec.
- If operators need to act when upgrading, add an upgrade note (see "Upgrade Notes" below).

### Upgrade Notes (changes that need operator action)
When a change means operators may need to do something when they upgrade (a breaking behaviour change, a config change, or a security fix that stops a previously working setup), add an upgrade note in the same PR as the change.

- Add the note to `docs/modules/ROOT/pages/upgrade/upgrade_notes.adoc`, under the `== Next release (unreleased)` heading. If that heading does not exist, create it directly after the page's introduction paragraph, above the newest version section.
- Each note is an anchor line plus a level-3 heading, then what changed, who is affected, and what to do:
  ```asciidoc
  [#short-kebab-case-anchor]
  === Short description of the change

  What changed, and who is affected. What operators need to change.
  ```
- Link to the note from related docs pages with `xref:upgrade/upgrade_notes.adoc#short-kebab-case-anchor[Upgrade Notes]`. The anchor stays the same when the note moves under a version heading, so links keep working.
- Do **not** write a version number and do **not** rename the unreleased heading by hand. When a 3k release is tagged, `.github/workflows/release-upgrade-notes.yml` runs `scripts/release_upgrade_notes.py`, which moves every note that was in the unreleased section at the tagged commit under a new `== <version>` heading on `next`. Notes merged after the release was cut stay unreleased. If the workflow fails, rerun it with `workflow_dispatch` and the release tag.
- Do **not** mark the change with a `BREAKING CHANGE:` footer or a `!` in the commit type. semantic-release treats either as a major version bump.
- docs.olivetin.app is built from `next`, so a note is public as soon as it merges. For security fixes the note goes in the private advisory fork PR, and should describe the behaviour change without exploit details.

### Security Fixes
Security advisory fixes follow the process in `SECURITY.md`:

- Work in the advisory's private temporary fork (`OliveTin/OliveTin-ghsa-xxxx-xxxx-xxxx`), on a branch from its `next`, and open the PR against the fork's `next`. Never push fixes for unpublished advisories to the public repository, and never describe unfixed advisories in public files.
- Commit subjects use `security: GHSA-xxxx-xxxx-xxxx (SEV) short description`, where `SEV` is `LOW`, `MED`, `HIGH` or `CRIT`. Do not name the reporter in commits; credit goes in the advisory.
- Supporting commits in the same PR (for example docs or an upgrade note) use their normal type, and may mention the advisory ID.
- `security:` commits trigger a patch release and are listed under **Security** in the release notes. Do not use `fix:` for advisory fixes, or they will be listed as bug fixes.
- When an advisory is closed as not a vulnerability but still worth fixing, the fix is an ordinary public `fix:` commit against `next`, without the advisory ID.
- Every triage or draft advisory carries a `[status:...]` tag at the end of its summary. Update it with `scripts/ghsa_status.py <GHSA-ID> <status>` whenever you change an advisory's state, ask the reporter a question, open a fork PR, or merge a fix; clear it with `clear` before publishing, closing, or requesting a CVE. The statuses and when to use them are listed in `.github/SECURITY_ADVISORY_DUPLICATES.md` under "Status tags". Run `scripts/ghsa_status.py --list` to check them.

### Branch Naming
Use conventional-commit-style branch names with a type prefix, optional issue reference, and a short kebab-case description:

- `feat/[#123]-add-justification-prompt`
- `fix/[#456]-websocket-reconnect`

Do **not** use `feat-...`, `feature/...`, or other variants. Omit the `[#<issue>]` segment only when there is no linked issue.

### Troubleshooting
- API tests failing with content-type errors: ensure Connect handler is served under `/api/` and the client targets that base URL.
- Executor panics: check for nil `Binding/Action` and add guards in step functions.
- Integration timeouts: wait for `loaded-dashboard` and use selectors matching the Vue UI.
