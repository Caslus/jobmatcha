## 1. Career-board persistence and scanner migration

- [x] 1.1 Add the career-board model, associations, repository operations, and migration/backfill from existing company ATS fields; verify existing configured companies receive one equivalent board and existing roles retain their company ownership.
- [x] 1.2 Move board-specific enabled state and scan-health evidence to the career-board lifecycle, update scanner orchestration to scan enabled supported boards, and verify focused scanner/service tests cover multiple boards for one company.
- [x] 1.3 Update company and board DTO sources, API handlers, and repository queries for independent board management; regenerate API DTOs with `mise run "[Server] Generate DTOs"` and verify generated output compiles.

## 2. Adapter-owned detection

- [x] 2.1 Add the optional adapter recognition and validation contracts plus registry support; verify a provider without discovery support remains scannable without being recognized.
- [x] 2.2 Implement Greenhouse URL recognition and normalization for both `boards.greenhouse.io` and `job-boards.greenhouse.io`, then validate a board through its public board endpoint; verify unit tests cover equivalent URL forms and invalid boards.
- [x] 2.3 Implement Workable URL recognition, normalization, and validation; verify unit tests cover supported URLs, non-matches, and failed validation.

## 3. Bounded careers-site discovery

- [x] 3.1 Build a discovery-specific HTTP client that rejects non-public destinations before connection and after redirects, and enforces protocol, timeout, redirect, and response-size limits; verify tests reject loopback/private redirect targets and oversized or timed-out responses.
- [x] 3.2 Extract and resolve candidate URLs from HTML links, metadata, script sources, and raw inline script text; verify tests recognize an ATS URL embedded in minified script content.
- [x] 3.3 Implement bounded relevant-link traversal, candidate deduplication, evidence collection, and incomplete-result reporting; verify a PayPay-style parent page plus linked group-careers page returns separate board candidates without following unrelated links.
- [x] 3.4 Validate supported candidates through their adapters and expose recognized-but-unsupported candidates without marking them scan-ready; verify tests cover valid, invalid, unsupported, duplicate, and partial-discovery outcomes.

## 4. Discovery and source-management experience

- [x] 4.1 Add authenticated APIs for non-persistent discovery and explicit registration of user-selected candidates; verify discovery alone writes no company or board records and selection writes only selected candidates.
- [x] 4.2 Update the Companies UI to submit a careers URL, display candidate provider/board/evidence/validation state, and require user selection before adding sources; verify `mise run "[Web] Check"` and `mise run "[Web] Build"` pass.
- [x] 4.3 Update the Companies UI to display and independently enable or disable each company's career boards; verify the backend API tests and frontend production build pass.

## 5. Integration verification

- [x] 5.1 Add migration, API, service, scanner, adapter, and discovery coverage for legacy single-board companies, multiple boards, group-company candidates, SSRF defenses, and zero-selection behavior; verify with `mise run "[Server] Run all tests"`.
- [x] 5.2 Run `mise run "[Web] Check"`, `mise run "[Web] Typecheck"`, and `mise run "[Web] Build"`, then manually verify the discovery result for a public PayPay-style careers URL and its separately selectable boards.

## 6. Shared employer grouping and discovery refinement

- [x] 6.1 Extend discovery metadata and registration handling to suggest one editable employer name (preferring an existing board owner, then `og:title`), and register all selected boards under that shared company by default; verify existing-company and metadata-suggestion cases.
- [x] 6.2 Redesign the discovery panel as a compact source-management workflow with a shared employer field, clear source selection and validation states, an explicit separate-registration override, and a target-aware primary action; verify frontend checks, typecheck, production build, and the PayPay-style flow.

## 7. Company-level list and board-detail modal

- [x] 7.1 Redesign the Companies list as a company-level summary showing location, board count, total jobs, aggregate freshness, and latest scan/discovery across all boards; move adapter and individual freshness details into the board-management modal; verify API and frontend checks.

## 8. Compact company index and board details

- [x] 8.1 Replace the full-width Companies table with a centered, compact company index and rebuild the board-management modal as a complete, polished per-board operational detail surface; verify frontend checks, typecheck, and production build.

## 9. Company index hierarchy and interaction refinement

- [x] 9.1 Reduce company-row metrics to the essential summary, replace cropped list activity text with modal detail, make board navigation visibly actionable, and align company selection/bulk actions with discovery controls; verify frontend checks, typecheck, and production build.

## 10. Company sorting and board-modal operational clarity

- [x] 10.1 Add compact company sorting, explained icon-based board statuses, relative board activity timestamps, and an immediately updating board enable switch; verify frontend checks, typecheck, and production build.
