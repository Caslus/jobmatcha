## Why

The local development, validation, and CI workflows repeat commands in multiple places and use hand-written change detection. This makes CI harder to reproduce locally and leaves tool and package-manager behavior less explicit than the project requires.

## What Changes

- Restructure Mise tasks into concise namespaces, retain compatibility aliases where useful, and run each task from its declared directory.
- Pin PNPM through Mise and remove implicit Corepack setup from CI while keeping dependency installation as explicit package-manager commands.
- Model prerequisite relationships with Mise task dependencies and add safe source/output freshness tracking for deterministic build tasks.
- Add a concurrent `dev:all` task for the API, SPA, and documentation site while retaining separate service tasks.
- Define CI-equivalent Mise tasks and have GitHub Actions invoke them rather than duplicating their command sequences.
- Replace shell-based changed-path detection with `dorny/paths-filter` and shared filters for application, documentation, and workflow/tooling changes.
- Update contributor-facing task references and local-development documentation.
- Do not introduce Turborepo, automatic dotenv loading, or Git-hook management in this change.

## Capabilities

### New Capabilities

None. This is developer tooling and workflow configuration; it does not change product behavior.

### Modified Capabilities

None.

## Impact

- `mise.toml`, GitHub Actions workflow configuration, and a new path-filter definition.
- Contributor documentation and PR validation instructions.
- CI dependencies gain a pinned `dorny/paths-filter` action; local tooling gains a pinned PNPM executable managed by Mise.
