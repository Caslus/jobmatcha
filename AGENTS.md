# Jobmatcha Contributor Guide

## Project and runtime

Jobmatcha is a monorepo with a React/TanStack Start frontend in `web/` and a
Gin/SQLite backend in `server/`.

- Use `mise` to provision the pinned toolchain and run project tasks. Run
  `mise install` once, then `mise tasks` to discover available tasks.
- Prefer `mise run "<task name>"` over invoking a command directly whenever a
  matching task exists in `mise.toml`.
- Mise currently provisions Go 1.26.6, Node 26.7.0, and Air. The Go module
  declares Go 1.25.10 as its minimum language version.
- Do not alter generated output or build directories: `web/src/types/api.gen.ts`,
  `web/src/routeTree.gen.ts`, and `web/dist/`.

## Common workflow

Use these tasks from the repository root:

| Goal | Task |
| --- | --- |
| Run the backend in development | `mise run "Run Go backend"` |
| Run the frontend in development | `mise run "Run Vite web app"` |
| Run focused API tests | `mise run "Run Go backend tests"` |
| Run all backend tests | `mise run "Run all Go tests"` |
| Regenerate API DTOs | `mise run "Generate TypeScript DTOs"` |
| Regenerate route tree | `mise run "Generate TanStack routes"` |
| Check frontend formatting and lint rules | `mise run "Check Vite web app"` |
| Build the static frontend | `mise run "Build Vite web app"` |
| Build the backend binary | `mise run "Build Go backend"` |

Do not use a task to install dependencies; use the appropriate package manager
only when dependencies actually need to change.

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
- There is no frontend test runner configured yet. For frontend changes, run
  the Biome check and production build; add a test runner and scripts before
  claiming component tests are available.

## Static deployment contract

`mise run "Build Vite web app"` creates `web/dist/client`. Gin serves that
directory and reserves `/api/*` for JSON endpoints; other GET and HEAD paths
fall back to `index.html` so deep links work. The TanStack Start server bundle
is not deployed or used.

For a packaged deployment, place `web/dist/client` alongside the backend or
set `STATIC_DIR` to its absolute path. Set runtime values such as `SERVER_PORT`,
`DB_PATH`, and `STATIC_DIR` through environment variables. Local `.env` files
may help a developer shell, but are not loaded by the application and must
never be committed.

## Security and generated-code rules

- Never commit API keys, passwords, database files, or build outputs.
- Never hardcode service credentials or frontend API origins.
- Do not edit generated DTOs or the route tree manually. Regenerate them using
  the mise tasks after updating their sources.
