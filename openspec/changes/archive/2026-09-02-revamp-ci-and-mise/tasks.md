## 1. Restructure Mise tooling

- [x] 1.1 Pin PNPM 10.23.0 alongside the existing toolchain and verify `mise install` exposes the pinned `pnpm` version without Corepack.
- [x] 1.2 Replace shell `cd` prefixes with namespaced `web:*`, `server:*`, and `docs:*` leaf tasks that declare `dir`, descriptions, and compatibility aliases; verify `mise tasks` exposes both the canonical names and aliases.
- [x] 1.3 Add conservative `sources` and `outputs` to the web, docs, and server build leaves, and verify unchanged inputs skip each build while a relevant changed source rebuilds it.
- [x] 1.4 Model web E2E and aggregate CI work with declared Mise task dependencies or task steps rather than nested `mise run` shell commands; verify `mise tasks deps` shows the intended prerequisites.
- [x] 1.5 Add `dev:all` to run server, web, and docs development services concurrently with readable prefixed output; verify all three local endpoints become available and Ctrl-C stops the group.
- [x] 1.6 Add a locally runnable container smoke-test task and a `ci:all` aggregate that cover the CI job contracts; verify each task names its prerequisites and cleans up its temporary container resources on success and failure.

## 2. Align GitHub Actions with the task graph

- [x] 2.1 Add a checked-in `dorny/paths-filter` filter definition for application, docs, and shared workflow/tooling changes; verify representative changes to each scoped path select every required job.
- [x] 2.2 Replace the shell-based detection job with a commit-pinned `dorny/paths-filter` action and the minimum required pull-request permission; verify its outputs gate the application, docs, E2E, and container jobs correctly.
- [x] 2.3 Update CI setup to rely on the PNPM version pinned by Mise and preserve explicit frozen-lockfile installation commands; verify no workflow step invokes `corepack enable`.
- [x] 2.4 Replace duplicated job validation command sequences with the corresponding `ci:*` Mise tasks while retaining coverage and failure-report artifact uploads; verify the workflow YAML remains valid and its task invocations reproduce local CI behavior.

## 3. Update contributor guidance

- [x] 3.1 Update the contributor guide, developer documentation, and PR template to use canonical namespaced tasks and document `dev:all`, focused CI tasks, and `ci:all`; verify every referenced task exists in `mise tasks`.
- [x] 3.2 Document the explicit dependency-install workflow with pinned PNPM and preserve the rule that root dotenv files are not automatically loaded by the application or Mise; verify no committed configuration enables `env._.file` or adds local secrets.

## 4. Validate the complete change

- [x] 4.1 Run `mise run "[Web] Check"`, `mise run "[Web] Typecheck"`, `mise run "[Coverage] Gate"`, `mise run "[Web] Build"`, and `mise run "[Web] Run browser E2E tests"`; resolve all failures and record the results.
- [x] 4.2 Run `mise run "[Docs] Build"`, task-discovery/dependency checks, and the container smoke test; verify no generated source files or build outputs are committed.
