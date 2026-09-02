## Context

The Jobs API currently loads roles without filtering their companies' active state. Company management currently exposes source state but no lifecycle APIs for source metadata. The completed career-board discovery change introduces board ownership and per-board source identity; this follow-up is intended to apply on top of that change.

## Goals / Non-Goals

**Goals:**

- Make disabled companies consistently invisible in the Jobs feed.
- Give users controlled correction and lifecycle actions for companies and individual boards.
- Keep scanner-provider validation authoritative for manually entered sources.
- Preserve already collected role history when sources are removed.

**Non-Goals:**

- Editing role history or creating an arbitrary free-form ATS adapter.
- Automatically merging companies or migrating roles between companies.
- Offering destructive role purging as part of ordinary source deletion.

## Decisions

### Filter visible jobs through company state

Make the role-list query join its company and constrain it to enabled companies before relevance, hidden-role, totals, and pagination are calculated. This prevents incompatible totals and avoids a frontend-only filter. Role detail remains addressable for historical links unless later requirements explicitly change that behavior.

Alternative considered: filtering Jobs after the API response. This gives incorrect pagination and briefly exposes disabled-company roles, so the server is the correct boundary.

### Treat provider and board identifier as validated source identity

Manual creation and edits accept provider, board identifier, and canonical URL together. The service normalizes and validates them via the existing provider registry, then enforces the global provider/identifier uniqueness rule. The modal refers to the identifier as the source name rather than adding a redundant label column.

Alternative considered: allowing free-form provider and URL strings. This would create sources the scanner cannot safely use and duplicate validation logic.

### Preserve roles on source deletion

Delete a board directly. Delete a company in a transaction that first deletes its boards, then removes the company while retaining roles. Role records already carry the company ID but do not require a live company for persistence; list queries will safely omit deleted-company roles through their join. Confirmation language will make this retention explicit.

Alternative considered: cascading deletion of roles. It is irreversible and conflicts with the app's historical job-tracking value.

### Use focused edit and confirmation modals

Company-level edit actions live alongside the existing company identity. Board add/edit/delete actions live in the board-management modal, with a lightweight confirmation modal for deletion. Successful mutations update the local modal state and invalidate the company list.

Alternative considered: inline editing within the compact list. It competes with selection, sorting, and source navigation and makes destructive confirmation harder to understand.

## Risks / Trade-offs

- [Legacy roles refer to a deleted company] -> Role list queries use an inner join on active companies; retained historical rows are not shown and do not break list rendering.
- [An edited board duplicates another source] -> Validate and enforce uniqueness transactionally before mutation.
- [A board edit appears to change a source but is invalid for its adapter] -> Normalize and validate through the adapter registry before persistence, returning an actionable field error.
- [Stacked change branch gets ahead of the discovery PR] -> Base implementation branch on the discovery feature branch or rebase after PR #45 merges.

## Migration Plan

No schema migration is required for the planned fields. Deploy the API and UI together. Existing boards remain valid; manual management is additive. Roll back application code by disabling the new controls while retaining any user-edited source records.
