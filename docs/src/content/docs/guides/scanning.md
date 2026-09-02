---
title: Scanning companies
description: Run scans on demand or on a schedule.
---

Jobmatcha begins with a seeded set of companies. Greenhouse and Workable are the supported scanners; unsupported seeded entries remain inactive. Start a scan immediately or enable scheduled scanning.

## Manage sources

Open **Companies** to manage the sources that feed your shortlist:

- Use **Discover boards** with a public careers URL to find supported job boards. Review the suggestions, choose the boards to keep, and add them under one company or separately.
- Open a company's board count to add a board manually, edit its provider, identifier, or canonical URL, toggle it, or remove it.
- Edit a company's name and location from its edit control. You can disable a company or individual board without deleting it.

Disabled companies and boards are skipped by future scans, and their roles disappear from the Jobs page. Removing a company or board also keeps its previously collected roles as historical records, but they remain out of the active shortlist.

## Scheduled scans

The schedule accepts standard five-field cron syntax:

```text
minute hour day-of-month month day-of-week
```

For example, `0 6 * * *` runs every day at 06:00. Set an [IANA timezone](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones), such as `America/Sao_Paulo`, so the schedule runs when you expect. Settings rejects invalid cron expressions and timezones. Onboarding currently stores schedule values without that validation, so verify or correct them in Settings after onboarding.

Only one scheduled scan runs at a time. This singleton protection applies to scheduled runs only; manually started scans are not globally serialized. Scan settings, scan history, roles, and preferences persist in the SQLite database, so restarting the container does not discard them.

If a provider scan fails, review the scan status in the dashboard and retry after confirming that the seeded entry uses Greenhouse or Workable.
