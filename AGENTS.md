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

### How to Run
- Run the server (dev):
  - From repo root: `go run ./service`
- Unit tests (Go):
  - From repo root: `cd service && make unittests`
- Integration tests (Mocha + Selenium):
  - Single test: `cd integration-tests && npx --yes mocha test/general.mjs`
  - All tests: `cd integration-tests && npx --yes mocha`

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

### API and Execution Flow (High-level)
1. Client calls Connect RPC (e.g., `Init`, `GetDashboard`, `StartAction`).
2. API translates requests to `executor.ExecutionRequest` and calls `Executor.ExecRequest`.
3. Executor runs a chain of steps: request binding → concurrency/rate/ACL checks → arg parsing → exec → post-exec → logging/triggering.
4. Logs are stored and can be fetched via `ExecutionStatus`/`GetLogs`.

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

### Troubleshooting
- API tests failing with content-type errors: ensure Connect handler is served under `/api/` and the client targets that base URL.
- Executor panics: check for nil `Binding/Action` and add guards in step functions.
- Integration timeouts: wait for `initial-marshal-complete` and use selectors matching the Vue UI.


