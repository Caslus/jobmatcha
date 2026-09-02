---
title: Develop Jobmatcha
description: Run the application locally, verify changes, and refresh documentation assets.
---

## Prerequisites

Use [Mise](https://mise.jdx.dev/) to provision the pinned Go, Node, PNPM, and Air versions:

```bash
mise install
pnpm --dir web install --frozen-lockfile
pnpm --dir docs install --frozen-lockfile
```

Dependency installation is deliberately explicit and is not a Mise task. Root
dotenv files are not loaded automatically by the application or Mise; set local
runtime values in your shell instead and never commit secrets.

## Daily workflow

Start every local service from the repository root:

```bash
mise run dev:all
```

It reserves the API, SPA, and documentation ports (8181, 3000, and 4321) before
starting any service. Use Ctrl-C to stop the whole group. If a port is already
in use, stop the existing service and run the command again.

For focused work, use `mise run server:dev`, `mise run web:dev`, or `mise run
docs:dev`. Stop the `dev:all` group with Ctrl-C.

The frontend proxies `/api` to the local backend. Run `mise tasks` to discover every available task, including the Starlight documentation commands.

## Checks

```bash
mise run web:check
mise run web:typecheck
mise run coverage:gate
mise run web:build
mise run web:e2e
mise run docs:build
```

`coverage:gate` is the CI-equivalent coverage check. It enforces the
project's weighted coverage thresholds; `server:coverage` only creates
a report and does not enforce those thresholds.

CI invokes `ci:application`, `ci:docs`, `ci:e2e`, and `ci:container`. Run
`mise run ci:all` to reproduce every CI contract locally.

API DTOs originate in `server/internal/model`. Update the source DTOs first, regenerate them with `mise run "[Server] Generate DTOs"`, and never edit generated frontend types or the route tree by hand.

## Screenshot refresh

The dashboard and role-detail images use fixture-only data: deterministic public-listing metadata, score-calibration variants, and a development-only Tokyo profile. Their captions describe the ranked dashboard/preferences and match-analysis views shown. Build the SPA, ensure Chromium is installed, then refresh the dashboard image:

```bash
pnpm --dir web exec playwright install chromium
mise run web:capture-readme-screenshots
```

Review the changed files in `docs/public/images/` before committing. The capture process never uses a production database or personal resume.

## Architecture

`server/cmd/api` composes the HTTP application. Handlers live in `server/internal/api`, business logic in `server/internal/service`, persistence in `server/internal/repository`, and ATS providers in `server/internal/scanner`. The client is a React/TanStack Start SPA in `web/`.
