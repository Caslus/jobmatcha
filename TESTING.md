# Testing and coverage

Use `mise install` once, then run these commands from the repository root:

- `mise run coverage:gate` runs both suites and enforces at least 80% for core backend application layers and 70% across maintained application code.
- `mise run web:check`, `mise run web:typecheck`, and `mise run web:build` validate the frontend.
- `mise run server:test` runs the complete backend suite.
- `mise run web:e2e` runs the production-build browser suite.

The gate reads Go and V8 reports directly and excludes generated files, test helpers, executable entrypoints, migrations, build output, and ATS provider adapters. Tests must be hermetic: use fixtures, fakes, or MSW; never call live external services.
