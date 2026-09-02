---
title: Troubleshooting
description: Diagnose common local and self-hosted Jobmatcha issues.
---

## I cannot sign in for the first time

Read `/data/initial-password` from the running container. The file exists only for a new database and is removed after you successfully change the password during onboarding. If the data volume already contains a configured instance, Jobmatcha does not create a new bootstrap password.

## The dashboard loads but requests fail

When using a reverse proxy, make sure the dashboard and `/api` share one public origin. For HTTPS deployments, set `COOKIE_SECURE=true`; for direct local HTTP, use the Compose default of `false`.

## The container is healthy but no dashboard appears

Check the container logs and confirm that the built frontend directory exists at `STATIC_DIR`. The Compose image includes it at `/app/web/dist/client`. A locally built binary needs `web/dist/client` or `../web/dist/client`, or an explicit `STATIC_DIR`.

## Scheduled scans do not run

Confirm that scheduled scanning is enabled, the cron expression has five fields, and the timezone is an IANA name. Use an on-demand scan to distinguish a scheduling problem from a provider or careers-page problem.

## A role is missing from the feed

Review location keywords, excluded keywords, selected work types, and the freshness setting. These rules can filter a role before ranking. Open a visible role's match analysis to verify how your include keywords and recency affect its score.

## Before asking for help

Capture the Jobmatcha commit, deployment method, non-sensitive logs, and the result of `GET /api/health`. Remove passwords, API keys, personal resume data, and session information from anything you share.
