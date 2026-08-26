# Frontend Guidelines

## Workflow

Run project tasks from the repository root whenever they exist:

- `mise run "Run Vite web app"` for local development.
- `mise run "Generate TanStack routes"` after adding, renaming, or removing a
  route file.
- `mise run "Check Vite web app"` for Biome checks.
- `mise run "Build Vite web app"` to produce the deployment artifact.

Do not edit `src/routeTree.gen.ts` or `src/types/api.gen.ts`; regenerate them
from route files and Go DTOs respectively.

## Application shape

This is a React 19 and TanStack Start application in SPA mode. Route files are
in `src/routes/`; feature code is in `src/features/`; shared UI primitives are
in `src/components/ui/`. There is no `App.tsx`, `main.tsx`, or Tailwind config
file to maintain.

- Use `@/` or `#/` aliases for imports from `src/`.
- Keep feature-specific components, hooks, and query logic together under the
  relevant feature directory.
- Build reusable primitives with the existing Tailwind CSS v4 and shadcn-style
  component conventions.
- Use Lucide for new interface icons.

## Data and state

- Use Ky through `src/lib/api.ts`; do not create another HTTP client.
- The production API base is same-origin `/api`. Vite proxies `/api` to the Go
  server during development. Do not introduce a hardcoded localhost origin.
- Use TanStack Query for server state and Zustand or component state for local
  UI state.
- Import all API request and response contracts from `@/types/api.gen`. Do not
  duplicate DTOs or use untyped `Record<string, unknown>` request payloads.

## Verification

There is no frontend unit-test runner configured. Run the Biome check and
production build for frontend changes. If a change warrants component tests,
add and document the test runner and its mise task as part of that work.
