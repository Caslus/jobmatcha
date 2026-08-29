---
title: Develop Jobmatcha
description: Run the application locally, verify changes, and refresh documentation assets.
---

## Prerequisites

Use [Mise](https://mise.jdx.dev/) to provision the pinned Go, Node, and Air versions:

```bash
mise install
corepack enable
pnpm --dir web install --frozen-lockfile
pnpm --dir docs install --frozen-lockfile
```

## Daily workflow

Run the API and SPA in separate terminals from the repository root:

```bash
mise run "[Server] Run dev"
mise run "[Web] Run dev"
```

The frontend proxies `/api` to the local backend. Run `mise tasks` to discover every available task, including the Starlight documentation commands.

## Checks

```bash
mise run "[Server] Run all tests"
mise run "[Web] Check"
mise run "[Web] Typecheck"
mise run "[Web] Run tests"
mise run "[Web] Build"
mise run "[Docs] Build"
```

API DTOs originate in `server/internal/model`. Update the source DTOs first, regenerate them with `mise run "[Server] Generate DTOs"`, and never edit generated frontend types or the route tree by hand.

## Screenshot refresh

The dashboard and role-detail images use fixture-only data: deterministic public-listing metadata, score-calibration variants, and a development-only Tokyo profile. Their captions describe the ranked dashboard/preferences and match-analysis views shown. Build the SPA, ensure Chromium is installed, then refresh the dashboard image:

```bash
pnpm --dir web exec playwright install chromium
mise run "[Web] Capture README screenshots"
```

Review the changed files in `docs/public/images/` before committing. The capture process never uses a production database or personal resume.

## Architecture

`server/cmd/api` composes the HTTP application. Handlers live in `server/internal/api`, business logic in `server/internal/service`, persistence in `server/internal/repository`, and ATS providers in `server/internal/scanner`. The client is a React/TanStack Start SPA in `web/`.
