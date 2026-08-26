# Backend Guidelines

## Workflow

Use repository tasks from the root whenever they exist:

- `mise run "Run Go backend"` for the Air development server.
- `mise run "Run Go backend tests"` for focused API tests.
- `mise run "Run all Go tests"` before handing off broader backend work.
- `mise run "Build Go backend"` for the production binary.
- `mise run "Generate TypeScript DTOs"` after changing exported API DTOs.

Mise provisions Go 1.26.6; `go.mod` declares 1.25.10. Do not edit the
generated frontend DTO file directly.

## Boundaries

`cmd/api` composes the application. `internal/api` owns HTTP binding,
authorization, response serialization, and static SPA delivery. Business rules
belong in `internal/service`; database queries belong in `internal/repository`;
providers and scan orchestration belong in `internal/scanner`.

- Keep request context flowing from Gin to service and repository calls.
- Handle errors explicitly, wrapping returned errors with actionable context.
- Use structured `slog` records for operational events and errors.
- Use dependency injection rather than global database handles.

## Data and contracts

- Use GORM with `glebarez/sqlite`, WAL mode, and a busy timeout.
- DTO structs in `internal/model` are the Tygo source of truth. Follow the
  existing `snake_case` JSON naming convention and give every exported API
  field an explicit JSON tag.
- After changing DTOs, regenerate types using the mise task and use the output
  from `@/types/api.gen` in the frontend.

## HTTP and deployment

- Register `/api/*` before static frontend routes. API misses must remain API
  404 responses and must never receive the SPA HTML fallback.
- The built SPA is at `web/dist/client`; `STATIC_DIR` overrides its location in
  a packaged deployment.
- Read runtime configuration from environment variables. Do not hardcode
  secrets or assume a `.env` loader exists.
