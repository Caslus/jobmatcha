## 1. Company scan evidence and persistence

- [x] 1.1 Extend the company model and repository with nullable per-company scan attempt, successful scan, failure-detail, and new-role-discovery fields; verify repository tests cover storing and retrieving each field.
- [x] 1.2 Update scan orchestration to record attempts, successful empty responses, failures, and only newly inserted roles as discovery activity; verify focused scanner tests cover healthy, failing, empty, and newly discovered-role outcomes.
- [x] 1.3 Provide the scanner/service boundary with registered-adapter availability needed to classify companies without duplicating provider configuration; verify supported and unsupported adapters are classified from the registry.

## 2. Companies API contract

- [x] 2.1 Define company list and active-state update DTOs, including adapter and freshness statuses, timestamps, and safe failure detail; regenerate DTOs with `mise run "[Server] Generate DTOs"` and verify generated frontend types compile.
- [x] 2.2 Add authenticated endpoints to list every registered company, update one active state, and transactionally update selected active states in bulk, deriving adapter status and the 30-day freshness status server-side; verify handler tests cover authentication, all-company listing, successful single and bulk updates, invalid input, missing IDs, and no unintended partial update.
- [x] 2.3 Verify API freshness behavior for fresh, stale, no-activity-yet, disabled, and unsupported companies, including the rule that disabled companies return no warning state; verify focused API tests pass with `mise run "[Server] Run API tests"`.

## 3. Companies experience

- [x] 3.1 Add protected Companies-route navigation to the authenticated header while preserving the existing Jobs dashboard route; verify signed-out and incomplete-setup redirects match the dashboard guard behavior.
- [x] 3.2 Add typed API client methods, query keys, and TanStack Query hooks for company listing plus single and bulk active-state mutations; verify each mutation refreshes or updates the company list consistently.
- [x] 3.3 Build the Companies list with sort controls for every displayed attribute, row selection, bulk enable/disable controls, individual enable/disable controls, adapter status treatment, failure explanation, and freshness labels; verify component tests cover sort direction, bulk actions, all status combinations, and that disabled rows show no stale warning.
- [x] 3.4 Run frontend checks with `mise run "[Web] Check"` and `mise run "[Web] Build"`; fix any route-generation or type errors without manually editing generated files.

## 4. End-to-end verification

- [x] 4.1 Run `mise run "[Server] Run all tests"` and verify backend behavior remains compatible with existing scan, settings, and role workflows.
- [x] 4.2 Run the relevant frontend test task and manually verify the Companies route can sort each displayed attribute, bulk enable/disable selected companies, toggle one company, and clearly distinguishes unsupported, failing, unknown, fresh, stale, and no-activity-yet states.

## 5. Companies experience refinements

- [x] 5.1 Add a typed per-company stored-role count to the Companies API using repository aggregation; regenerate DTOs and cover counts in service/API tests.
- [x] 5.2 Refine the Companies table: default-sort by role count descending, add role-count sorting, and use distinct ascending/descending active-sort icons.
- [x] 5.3 Replace row enabled/action controls with one accessible toggle, and render adapter/freshness status as tooltip-backed icons including failure detail; cover interaction and all status-tooltip states in frontend tests.
- [x] 5.4 Run backend and frontend checks plus a production build without manually editing generated output.
