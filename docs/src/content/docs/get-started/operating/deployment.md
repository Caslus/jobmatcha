---
title: Deploy Jobmatcha
description: Run the container safely beyond local development.
---

Jobmatcha builds into one container: Gin serves the API and the built React application. SQLite, its WAL files, and the bootstrap password file (until the password changes) live in the persistent data directory.

## Docker Compose

Use the supplied [`compose.yaml`](https://github.com/Caslus/jobmatcha/blob/main/compose.yaml) as described in [Start with Docker Compose](/get-started/docker-compose/). It publishes port `8181`, mounts the `jobmatcha-data` volume at `/data`, and sets `COOKIE_SECURE=false` for direct local HTTP access.

## HTTPS and reverse proxies

For an internet-facing deployment, terminate TLS in a reverse proxy and set this environment variable in the Jobmatcha service:

```yaml
environment:
  COOKIE_SECURE: "true"
```

Serve the dashboard and `/api` at the same public origin. This lets the browser send its session cookie with API requests. Forward the proxy to Jobmatcha's port `8181`; do not expose a separate API origin unless you also redesign the authentication flow.

> [!WARNING]
> `COOKIE_SECURE=true` prevents browsers from sending the session cookie over plain HTTP. Keep it disabled only for trusted local HTTP development.

Use `GET /api/health` for container or proxy health checks. It returns a JSON status response without requiring authentication.

Continue with [upgrades and backups](/get-started/operating/upgrades-and-backups/) before relying on the instance for an active search.
