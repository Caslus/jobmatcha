## Context

The scanner currently reads active companies and knows which providers are registered, but the authenticated API exposes no company-management endpoints. `Company` includes `Active` and `LastScannedAt`; the latter is updated only after a fetch that returns at least one role. Individual scan failures are retained only in aggregate scan-job JSON, which is unsuitable for a stable per-company status. See proposal.md and the `companies-management` delta spec for the user-facing contract.

## Goals / Non-Goals

**Goals:**

- Provide a route-level Companies experience rather than adding another panel to the dashboard's three-pane jobs workflow.
- Let users sort every visible column and efficiently manage multiple company states at once.
- Make source configuration, latest scan outcome, and job-feed freshness independently understandable.
- Preserve historical scan and discovery evidence across scans and restarts.
- Keep adapter support authoritative to the scanner registry rather than duplicating a frontend list.

**Non-Goals:**

- Adding, editing, or deleting company definitions.
- Implementing additional ATS adapters or probing every adapter independently of normal scans.
- Determining the employer's original posting timestamp; freshness is based on Jobmatcha's discovery time.
- Making the 30-day freshness threshold user-configurable in this change.

## Decisions

### Make Companies a protected route with header navigation

Add a `/companies` route and a visible Jobs/Companies navigation control in the authenticated header. This preserves a shareable URL and prevents the Companies list from competing with the dashboard's resizable job panes. The route uses the existing authentication guard pattern.

Alternative considered: render a local tab inside `/dashboard`. Rejected because it would require the Companies experience to inherit job-selection and panel state that it does not use.

### Expose a purpose-built company response and mutation API

Add an authenticated collection endpoint to list all companies, an item update endpoint limited to the active flag, and a bulk active-state update endpoint that accepts selected company IDs and one target state. A bulk update is transactional so it changes the requested set together or reports failure. The response combines persisted company data with server-derived `adapter_status` and `freshness_status`, plus timestamps and the latest error needed to explain a failing state. Update model DTO source first and regenerate Tygo so frontend types remain generated.

The service/handler layer determines adapter support from the scanner's registered providers; repository code remains responsible only for persistence. The mutation must reject unknown company IDs and invalid update bodies, then return the updated company state.

Alternative considered: expose raw `Company` records and determine statuses in the browser. Rejected because adapter registration is server runtime state and client-side derivation would duplicate business rules and hide scan evidence decisions.

### Sort the complete list in the client and bulk-update on the server

The collection is currently a bounded seeded registry, so the Companies route loads all entries and applies sorting in the browser across every field the UI displays: name, location, adapter, enabled state, role count, adapter status, freshness status, latest scan, and latest new-role discovery. Headers display distinct ascending/descending icons for the active sort, and the initial sort is role count descending.

The UI maintains an explicit selection set, provides bulk enable/disable controls only when selection is non-empty, and refreshes the company query after a successful bulk update. The API owns authorization, ID validation, and transactional persistence; it returns an error instead of silently applying only a subset when the requested update cannot be completed.

Alternative considered: implement server-side sorting and pagination immediately. Rejected because the registered-company set is small and already must be returned in full for bulk selection. The API response remains structured so server-side pagination can be added later if the registry grows.

### Aggregate stored role counts in the Companies response

Add a per-company role count to the company-list service response, calculated by the repository using an aggregate query rather than loading roles into application memory. This keeps the count authoritative, exposes no role details, and provides a useful default ordering signal. The response is regenerated through Tygo before frontend use.

### Prefer icon + tooltip status cells and a single active toggle

Use accessible icon buttons or labelled elements with tooltip text for adapter and freshness status. Keep the visible table compact while retaining status words and failure explanations on hover and keyboard focus. Replace separate enabled text and Enable/Disable action controls with one labelled switch that is disabled while its mutation is pending. Bulk controls remain separate and only appear after selection.

### Record distinct per-company scan and discovery timestamps

Extend the company persistence model with fields for last scan attempt, last successful scan, latest scan error, and last new-role discovery. Retain the existing `last_scanned_at` field as a compatibility-safe representation of the latest successful scan, or migrate callers to the explicit field if no API contract depends on it.

For each supported active company scan, record an attempt. A successful provider response, including an empty result, records success and clears the latest error. A provider resolution or fetch failure records the error without advancing successful-scan time. When at least one role is inserted for the first time, record the new-role-discovery timestamp. Provider updates to already-known roles do not refresh freshness.

Alternative considered: derive health and freshness from the latest aggregate scan job and role creation timestamps. Rejected because aggregate results are transient/ambiguous per company and do not distinguish successful empty scans from failures.

### Derive freshness after applying eligibility rules

The API returns `not_applicable` freshness for disabled or unsupported companies. For enabled supported companies, return `no_activity_yet` when no new role has been discovered, `fresh` when discovery was within 30 days, and `stale` otherwise. The UI only renders a warning treatment for `stale`; disabled entries remain visually neutral.

Alternative considered: use last successful scan as freshness. Rejected because a source can scan successfully without publishing new opportunities, which is the signal the user wants to investigate.

## Risks / Trade-offs

- [Existing records lack outcome and discovery history] -> Backfill remains nullable; old supported companies show neutral unknown/no-activity states until new scans provide evidence.
- [A currently unsupported company has historic roles] -> Return unsupported and no freshness warning because no action can make it fresh until an adapter is added.
- [Provider payloads omit original job posting dates] -> Label the indicator as “No new roles found in 30 days,” avoiding a claim about the employer's posting activity.
- [Toggle and a scan overlap] -> The current scan operates on an initial active-company snapshot; document and test that a toggle applies to subsequent scans.
- [Bulk action races with a list refresh] -> Disable duplicate submissions while the mutation is pending and refresh the authoritative company list after completion.
- [Status-only icons may be ambiguous without pointer access] -> Ensure tooltip content is also available through accessible labels and keyboard focus.

## Migration Plan

1. Add nullable scan-evidence fields through the existing GORM migration path; existing rows remain valid without a data backfill.
2. Deploy the backend endpoints and scan evidence updates before deploying the Companies route.
3. Regenerate frontend DTOs and deploy the UI once the endpoint is available.
4. Roll back the UI independently if necessary; new nullable data does not affect existing roles, settings, or scan-job APIs.
