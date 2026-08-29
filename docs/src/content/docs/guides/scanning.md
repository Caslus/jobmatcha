---
title: Scanning companies
description: Run scans on demand or on a schedule.
---

Jobmatcha seeds a fixed set of companies. At present, only the Greenhouse and Workable entries have implemented scanners; the other seeded entries remain inactive. The dashboard does not add or edit company career sites. Start a scan immediately or enable scheduled scanning.

## Scheduled scans

The schedule accepts standard five-field cron syntax:

```text
minute hour day-of-month month day-of-week
```

For example, `0 6 * * *` runs every day at 06:00. Set an [IANA timezone](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones), such as `America/Sao_Paulo`, so the schedule runs when you expect. Settings rejects invalid cron expressions and timezones. Onboarding currently stores schedule values without that validation, so verify or correct them in Settings after onboarding.

Only one scheduled scan runs at a time. This singleton protection applies to scheduled runs only; manually started scans are not globally serialized. Scan settings, scan history, roles, and preferences persist in the SQLite database, so restarting the container does not discard them.

If a provider scan fails, review the scan status in the dashboard and retry after confirming that the seeded entry uses Greenhouse or Workable.
