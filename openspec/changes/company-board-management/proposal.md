## Why

Users can discover and toggle sources but cannot correct company metadata, repair a board URL or identifier, add a known board directly, or remove an obsolete source. Disabled companies also remain visible in the Jobs feed, which makes source controls fail to match the user-facing job list.

## What Changes

- Exclude roles belonging to disabled companies from the Jobs list and its visible totals.
- Add company editing for the user-managed name and location fields, plus a confirmed company deletion flow.
- Add per-board editing, deletion, and manual registration for supported adapters without requiring a discovery run.
- Validate manual and edited board identities using the adapter registry, preserve provider/identifier uniqueness, and retain historical roles when a board or company is deleted.
- Extend API, backend, and frontend coverage for role visibility and company/board lifecycle operations.

## Capabilities

### New Capabilities

- `job-visibility`: Ensure the Jobs experience exposes roles only from currently enabled companies.

### Modified Capabilities

- `companies-management`: Allow companies and their independent career boards to be edited, manually created, and safely deleted.

## Impact

- Backend company, career-board, and role repositories; service and API DTO/route contracts; generated frontend DTOs.
- Companies UI, board-management modal, TanStack Query mutations, and confirmation states.
- No data migration is expected: current company and career-board fields hold the editable source identity. Deletion behavior will retain roles as historical records rather than cascading a destructive purge.
