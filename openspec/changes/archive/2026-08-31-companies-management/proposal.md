## Why

Jobmatcha already seeds and scans a set of companies, but users cannot see every registered company, control which sources participate in scans, or tell whether a company is scannable and producing fresh opportunities. A Companies view makes source coverage transparent and lets users manage it directly.

## What Changes

- Add a Companies area that lists every registered company and lets users enable or disable its inclusion in future scans.
- Let users sort the Companies list by any displayed company attribute and select companies for bulk enable or disable actions.
- Expose each company's configured ATS adapter and a user-facing adapter status: healthy, failing, unsupported, or unknown.
- Persist per-company scan outcome metadata so the adapter status reflects the latest scan attempt, including an actionable failure message when applicable.
- Surface a freshness indicator for enabled, supported companies that have had no newly discovered roles for 30 days; disabled companies never receive a stale warning.
- Treat companies with no discovered roles as a neutral “no activity yet” state rather than stale.
- Make sort direction unambiguous, use compact tooltip-backed status indicators, and consolidate each row's enabled state and action into one toggle.
- Show the number of stored roles for each company and initially rank the list by that count.

## Capabilities

### New Capabilities

- `companies-management`: Browse registered companies, control scan participation, and understand adapter availability, scan outcome, and job-feed freshness.

### Modified Capabilities

- None.

## Impact

- Backend company model, repository, scanner outcome recording, authenticated company endpoints, role-count aggregation, and API DTO generation.
- Frontend navigation, Companies route and sortable list UI, API client, query hooks, and generated API types.
- Backend and frontend tests covering list state, single and bulk toggle behavior, sorting, adapter statuses, and freshness boundaries.
