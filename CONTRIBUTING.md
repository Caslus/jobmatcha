# Contributing to Jobmatcha

Thanks for helping make job searching calmer and more useful.

## Before you start

- Search existing issues before opening a new one.
- Use the feature request form to explain the job-search problem before proposing a large change.
- Keep pull requests focused. Separate refactors from behavior changes when practical.
- For security vulnerabilities, use [private vulnerability reporting](SECURITY.md)—never open a public issue.

## Local setup

Follow the [development guide](docs/src/content/docs/developers/development.md) to install the pinned toolchain and run the frontend and backend. Use repository tasks through Mise whenever one exists.

Before opening a pull request, run the checks relevant to your change:

```bash
mise run "[Server] Run all tests"
mise run "[Web] Check"
mise run "[Web] Typecheck"
mise run "[Web] Run tests"
mise run "[Web] Build"
mise run "[Docs] Build"
```

## Project conventions

- Keep HTTP binding and responses in `server/internal/api`, business rules in `server/internal/service`, and persistence in `server/internal/repository`.
- Add ATS integrations through the scanner provider interface and include focused provider tests.
- Use generated API types in the frontend. Update DTO source structs first, then regenerate; never hand-edit `web/src/types/api.gen.ts` or `web/src/routeTree.gen.ts`.
- Do not commit secrets, local databases, build artifacts, or personal resume data.
- Refresh README screenshots only with the fictional documentation fixture described in the development guide.

## Pull requests

Describe the user-facing outcome, how you tested it, and any migration or deployment considerations. Screenshots are especially helpful for UI changes. By contributing, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
