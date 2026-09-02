## Context

See proposal.md for motivation. The repository has a root Mise configuration, independent `web/` and `docs/` PNPM installations, and Go code in `server/`. GitHub Actions repeats validation commands and currently detects affected paths with a shell script. The contributor guide requires package-manager commands, rather than Mise tasks, for dependency installation.

## Goals / Non-Goals

**Goals:**

- Make named Mise tasks the source of truth for the work performed in CI jobs and local validation.
- Make daily development start all three independently served processes with one command.
- Preserve focused tasks and provide a non-breaking migration from the existing bracketed task names.
- Skip only deterministic builds whose declared outputs are fresh; always execute checks and tests.

**Non-Goals:**

- Convert the repository to a PNPM workspace, add Turborepo, or introduce remote build caching.
- Load dotenv files automatically, change application runtime configuration, or add local Git hooks.
- Use Mise tasks for dependency installation.

## Decisions

### Use namespaced TOML tasks with declared directories

Define canonical task names such as `web:build`, `server:test`, `docs:build`, `ci:application`, and `dev:all`. Each leaf task declares `dir` instead of embedding `cd` in shell commands. Retain aliases for the documented bracketed names during migration, then update docs and templates to the concise names.

TOML remains the task home because the graph is small and benefits from visible task metadata. Standalone `.mise/tasks` scripts remain available for future complex, platform-specific commands.

### Pin PNPM as a Mise tool; keep installs explicit

Add PNPM 10.23.0 to `[tools]`, matching the container build. CI and local setup use explicit `pnpm --dir <project> install --frozen-lockfile` commands and do not depend on Corepack or a developer-global PNPM version. This respects repository guidance that dependency installation is not a task.

### Separate task graph dependencies from sequential pipelines

Use `depends` for prerequisites, notably `web:e2e` depending on `web:build`. Define CI aggregate tasks using Mise task steps where ordering is meaningful and allow independent checks to run in parallel. Avoid calling `mise run` inside task shell strings so the graph remains inspectable with `mise tasks deps`.

`dev:all` launches `server:dev`, `web:dev`, and `docs:dev` in parallel with Mise-prefixed output. Development services retain their own watchers (Air, Vite, Astro); `mise watch` is not introduced for them.

### Add conservative freshness only to build leaves

Declare `sources` and `outputs` for the SPA, documentation, and API build tasks. Include each package manifest, lockfile, build config, and relevant source/assets; output paths are `web/dist/client`, `docs/dist`, and `server/dist/api`. Do not cache linting, type checking, unit/coverage tests, E2E tests, Docker builds, or dev tasks. CI can force builds if a runner ever restores build outputs in its cache.

### Use central dorny filters to select job tasks

Add a checked-in filter file with `application`, `docs`, and `workflow` scopes. Application covers web/server/container/coverage inputs; docs covers documentation and its build inputs; workflow covers Mise, CI workflow, and filter definitions and causes all relevant jobs to run. The `changes` job uses `dorny/paths-filter` with `pull-requests: read`; other jobs consume its boolean outputs and invoke only canonical `ci:*` tasks.

This replaces shell pattern matching and its full-history checkout. Pin the action to a reviewed commit SHA, consistent with the repository's existing pinned actions.

## Risks / Trade-offs

- [Alias support or task-name migration is incomplete] -> Verify `mise tasks` output and migrate every repository reference; keep legacy aliases until all references use namespaces.
- [Freshness rules omit an input] -> Start with broad source patterns, verify rebuild/skip behavior, and document `mise run --force` as an escape hatch.
- [A workflow/config-only edit skips a required job] -> Include shared tooling and filter files in a `workflow` scope that enables all dependent jobs; add representative path-filter tests or workflow assertions where practical.
- [Concurrent dev process output is noisy] -> Use Mise's prefixed output and retain separate service tasks for focused debugging.
- [Pinned third-party action becomes stale] -> Dependabot or routine dependency updates must maintain its SHA and version comment.

## Migration Plan

1. Add the namespaced task graph and aliases, then verify task discovery and dependency visualization.
2. Update GitHub Actions to install pinned tools, install dependencies explicitly, use path filters, and call the CI tasks.
3. Update developer documentation and PR-template commands.
4. Run the prescribed validation suite, including browser E2E and documentation build, and compare relevant path-filter scenarios before merging.
5. Roll back by restoring the prior Mise and workflow files; no database, API, or deployed-runtime migration is involved.
