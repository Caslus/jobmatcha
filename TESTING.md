# Testing and coverage

Use `mise install` once, then run these commands from the repository root:

- `mise run "[Coverage] Gate"` runs both suites and enforces at least 80% for core backend application layers and 70% across maintained application code.
- `mise run "[Web] Check"`, `mise run "[Web] Typecheck"`, and `mise run "[Web] Build"` validate the frontend.
- `mise run "[Server] Run all tests"` runs the complete backend suite.
- `mise run "[Web] Run browser E2E tests"` runs the production-build browser suite.

The gate reads Go and V8 reports directly and excludes generated files, test helpers, executable entrypoints, migrations, build output, and ATS provider adapters. Tests must be hermetic: use fixtures, fakes, or MSW; never call live external services.
