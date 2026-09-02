# Jobmatcha Contributor Guide

## Project and runtime

Jobmatcha is a monorepo with a React/TanStack Start frontend in `web/` and a
Gin/SQLite backend in `server/`.

- Use `mise` to provision the pinned toolchain and run project tasks. Run
  `mise install` once, then `mise tasks` to discover available tasks.
- Prefer `mise run "<task name>"` over invoking a command directly whenever a
  matching task exists in `mise.toml`.
- Mise and the Go module use Go 1.26.6; Node 26.7.0, PNPM 10.23.0, and Air are also pinned.
- Do not alter generated output or build directories: `web/src/types/api.gen.ts`,
  `web/src/routeTree.gen.ts`, and `web/dist/`.

## Common workflow

Use these tasks from the repository root:

| Goal | Task |
| --- | --- |
| Run all local development services | `mise run dev:all` |
| Run the frontend in development | `mise run web:dev` |
| Build the static frontend | `mise run web:build` |
| Check frontend formatting and lint rules | `mise run web:check` |
| Typecheck the frontend | `mise run web:typecheck` |
| Regenerate route tree | `mise run web:generate-routes` |
| Run frontend tests | `mise run web:test` |
| Run frontend coverage | `mise run web:coverage` |
| Run frontend browser E2E tests | `mise run web:e2e` |
| Run the backend in development | `mise run server:dev` |
| Build the backend binary | `mise run server:build` |
| Regenerate API DTOs | `mise run server:generate-dtos` |
| Run focused API tests | `mise run server:test-api` |
| Run all backend tests | `mise run server:test` |
| Run backend coverage | `mise run server:coverage` |
| Run CI-equivalent coverage gate | `mise run coverage:gate` |
| Build the documentation site | `mise run docs:build` |
| Run focused CI contracts | `mise run ci:application`, `mise run ci:docs`, `mise run ci:e2e`, or `mise run ci:container` |
| Run every CI contract | `mise run ci:all` |

Do not use a task to install dependencies; use the appropriate package manager
only when dependencies actually need to change.

## Required validation before a PR update

For every application change, run these exact commands before opening or
updating a pull request:

```text
mise run web:check
mise run web:typecheck
mise run coverage:gate
mise run web:build
mise run web:e2e
```

`coverage:gate` is authoritative: it runs backend and frontend coverage and
enforces the same weighted thresholds as CI. Running `[Server] Run coverage`
alone only produces a report; it does not enforce the threshold.

Run `mise run docs:build` whenever `docs/`, README content, or other
user-facing documentation changes. When a validation command fails, inspect
and resolve its actual failure before handing off or updating the PR.

Add or update focused tests for changed API handlers, service decisions,
repository behavior, and user-visible frontend behavior. Regenerate DTOs and
routes when their source inputs change, then include the generated output in
the validation.

## GitHub issues and pull requests

Before creating or updating a GitHub issue or pull request, read and use the
applicable repository template:

- Pull requests: `.github/pull_request_template.md`
- Bug reports: `.github/ISSUE_TEMPLATE/bug_report.yml`
- Feature requests: `.github/ISSUE_TEMPLATE/feature_request.yml`

Keep the template's sections and answer every applicable prompt. For issues
created outside GitHub's form UI, reproduce the form labels as Markdown
headings so the report retains the same information.

Submit PR and issue bodies as real Markdown, preferably with `gh`'s
`--body-file` option. Never send literal escaped newline text such as `\\n`,
and do not rely on a one-line shell string to represent a multi-paragraph
description. Use blank lines around headings, lists, and code fences.

After creating or editing a PR or issue, read it back with `gh pr view` or
`gh issue view` and confirm that headings, lists, links, and code blocks render
as intended. Correct the body before reporting the link to the user.

## Commit messages and release versioning

Use Conventional Commits for every commit. The pull-request check enforces this
format and semantic-release derives versions from it:

- `fix: description` creates a patch release.
- `feat: description` creates a minor release.
- Add `BREAKING CHANGE: description` in the footer (or `!` after the type) to
  create a major release.
- Use other valid types such as `chore:`, `docs:`, `test:`, and `ci:` for work
  that does not itself require a release.

Use an optional scope when it makes the affected area clearer, for example
`feat(scanner): add provider retries`. Keep the subject imperative and concise.

## Architecture

### Backend (`server/`)

```
server/
├── cmd/api/            # startup, HTTP server, static SPA registration
├── internal/
│   ├── ai/             # LLM client and output parsing
│   ├── api/            # Gin handlers, routes, middleware, static serving
│   ├── model/          # domain structs and API DTO source for Tygo
│   ├── repository/     # GORM persistence only
│   ├── scanner/        # ATS providers and scan orchestration
│   ├── service/        # application business logic
│   └── util/           # shared backend utilities
├── migrations/
└── tygo.yaml
```

- Use `glebarez/sqlite` with GORM; it is backed by `modernc.org/sqlite`.
- Keep database access in `internal/repository` and business rules in
  `internal/service`. Handlers bind, authorize, call collaborators, and write
  HTTP responses.
- API DTOs live in `internal/model` and use the existing `snake_case` JSON
  convention. Update DTOs first, regenerate Tygo output, then use the
  generated frontend types.
- Use `slog` for operational logs and wrap returned errors with useful context.
- Keep request context flowing from Gin through service and repository calls.
- Use dependency injection rather than global database handles.

### Frontend (`web/`)

The frontend uses React 19, TanStack Start, TanStack Router, TanStack Query,
Tailwind CSS v4, Biome, Ky, and Zustand. It is a client-side SPA, not a plain
Vite entrypoint app: routes live in `src/routes/`, and `routeTree.gen.ts` is
generated from them.

- Place reusable UI primitives in `src/components/ui/` and feature code under
  `src/features/<feature>/`.
- Use `@/` or `#/` aliases for `src/` imports.
- Use Ky only through `src/lib/api.ts`; production requests are same-origin
  `/api`, while Vite proxies that path to the backend in development.
- Use TanStack Query for server state. API request and response types must come
  from `@/types/api.gen`, not locally duplicated interfaces or
  `Record<string, unknown>` payloads.
- Tailwind is configured through the Vite plugin; there is no
  `tailwind.config.js`.
- Vitest is configured for frontend unit tests. Run `[Web] Run tests` while
  iterating; the required PR coverage gate runs the suite with V8 coverage.
- Use Lucide for new interface icons and follow the existing Tailwind v4 and
  shadcn-style component conventions.

## Static deployment contract

`mise run web:build` creates `web/dist/client`. Gin serves that
directory and reserves `/api/*` for JSON endpoints; other GET and HEAD paths
fall back to `index.html` so deep links work. The TanStack Start server bundle
is not deployed or used.

For a packaged deployment, place `web/dist/client` alongside the backend or
set `STATIC_DIR` to its absolute path. Set runtime values such as `SERVER_PORT`,
`DB_PATH`, and `STATIC_DIR` through environment variables. Local `.env` files
may help a developer shell, but are not loaded by the application or Mise and
must never be committed. Install dependencies explicitly after `mise install`
with `pnpm --dir web install --frozen-lockfile` and `pnpm --dir docs install
--frozen-lockfile`; do not add dependency installation as a Mise task.

## Security and generated-code rules

- Never commit API keys, passwords, database files, or build outputs.
- Never hardcode service credentials or frontend API origins.
- Do not edit generated DTOs or the route tree manually. Regenerate them using
  the mise tasks after updating their sources.
