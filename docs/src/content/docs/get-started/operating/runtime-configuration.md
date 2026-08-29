---
title: Runtime configuration
description: Environment variables accepted by the Jobmatcha application.
---

Set runtime values with your container platform or process manager. Local `.env` files can help a development shell but are not loaded by Jobmatcha and must not be committed.

| Variable | Application default | Container value | Purpose |
| --- | --- | --- | --- |
| `SERVER_PORT` | `8181` | `8181` | HTTP listen port. |
| `DB_PATH` | `data/app.db` | `/data/app.db` | SQLite database path. |
| `STATIC_DIR` | checks `web/dist/client`, then `../web/dist/client` | `/app/web/dist/client` | Built SPA directory. |
| `COOKIE_SECURE` | `true` | unset; Compose sets `false` | Marks session cookies as HTTPS-only. Invalid values are treated as `true`. |
| `BOOTSTRAP_PASSWORD_FILE` | alongside `DB_PATH` | `/data/initial-password` | Optional path for the first-run password file. |

When `STATIC_DIR` is absent, the application checks the usual local frontend build locations. If it cannot find an `index.html`, it serves API routes only.
