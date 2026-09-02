## Context

Companies currently store one `ats_type` and `ats_slug`, while the scanner registry exposes only fetch behavior. The proposed discovery behavior needs a user-supplied URL to lead to zero, one, or many independently selectable boards, including boards found through explicit group-company career links. See proposal.md and the delta specs for the behavioral contract.

## Goals / Non-Goals

**Goals:**

- Make Greenhouse and Workable immediately discoverable from ordinary company career pages and establish an adapter-owned pattern for future providers.
- Preserve employer identity when a corporate landing page links to group-company boards.
- Keep crawling deterministic, bounded, and safe for user-controlled URLs.
- Migrate existing registered companies without losing current scan state or roles.

**Non-Goals:**

- Building an unbounded web crawler, discovering companies without an input URL, or aggregating third-party board search indexes.
- Rendering JavaScript in the initial release; static markup and raw inline scripts are the supported evidence sources.
- Automatically creating employers, enabling sources, or merging group-company candidates without user confirmation.
- Adding new ATS fetch adapters beyond recognition and validation support for the existing Greenhouse and Workable adapters.

## Decisions

### Separate career boards from companies

Introduce a persistent `CareerBoard` (name finalized during implementation) owned by one Company. It contains the provider key, normalized external board identifier, canonical board URL, enabled state, and scan-health timestamps/failure evidence. Roles retain their company association; scans run per board and upsert roles under the owning company.

This replaces the company-level `ats_type` and `ats_slug` relationship. A unique key on provider plus board identifier prevents duplicate source registration, and a board-level active flag permits disabling one source without disabling its siblings.

The Companies page is a centered, compact company index, not a full-width data table or flattened board table. Each row uses company identity and location as its anchor, then presents only the job count, aggregate freshness, and an explicit, labelled board-navigation control. Latest scan and discovery values belong in the board-management modal, where they are complete and do not crop or compete with the company identity. Selection uses the same custom circular check control as discovery; select-all, clear-selection, enable, and disable use a cohesive compact action treatment so each is obviously interactive.

The board-management modal is the complete per-source operational detail surface. Each source uses a calm, distinct section rather than deeply nested outlines and exposes provider and board identifier, canonical URL, enabled state, adapter health, freshness, last scan, latest new-role discovery, and a failure detail when available. Adapter health and freshness pair a semantic icon with a concise label and visible explanation so states such as healthy and fresh are understandable without specialist knowledge. Activity timestamps use relative time for scanning, while retaining the complete timestamp in an accessible hover label. Board enablement updates the modal's local state immediately, then reconciles against the refreshed company list. Aggregate freshness is derived from the company boards, with failing taking precedence over stale, stale over fresh, and unknown used when no fresher state is available.

The compact company index keeps its sparse card-like layout while providing a small, explicit sort control. Users can order companies by job count, board count, name, location, or most recent activity and reverse the applicable order without adding per-column table affordances or widening each row.

Alternative considered: store an array of provider configurations on Company. This makes querying, uniqueness, board health, and incremental management harder in SQLite/GORM, so a normalized relation is preferred.

### Extend providers with optional discovery capabilities

Keep the existing fetch `Provider` contract focused on scanning. Add a narrowly scoped optional companion capability for URL recognition and validation. A recognizer receives a candidate URL and returns either no match or a normalized board identity and canonical URL; a validator confirms a supported board can be scanned without importing roles.

The discovery orchestrator extracts candidate URLs generically and asks the registry's registered recognizers to classify them. Thus Greenhouse owns both `boards.greenhouse.io` and `job-boards.greenhouse.io` forms, while Workable owns its own host forms. Future adapters become discoverable by registering their own recognizer rather than altering central conditionals.

Alternative considered: one shared table of regexes. It is simpler initially but splits provider knowledge from its fetch implementation and cannot easily express provider-specific normalization or validation.

### Use a bounded static discovery traversal

The discovery service fetches the submitted URL, resolves links against the final document URL, and extracts absolute URLs from anchors, URL-bearing metadata, script sources, and inline script text. It first classifies all extracted URLs against registered providers. It then considers only a small, configurable set of career/recruit/job-labelled links for a one-hop follow-up fetch; direct known-provider URLs are never fetched as ordinary pages before validation.

Candidates are deduplicated by provider and normalized board identity. Each retains evidence pages, and every candidate is either scan-ready after validation, recognized-but-unsupported, invalid, or incomplete when a traversal bound was hit. Discovery results are ephemeral until an explicit add request carries the user's selected candidates.

Alternative considered: headless-browser rendering first. It reaches JavaScript-only links but adds considerable runtime cost and operational fragility. It remains a later fallback for inconclusive static discovery.

### Treat user URLs as hostile network inputs

Use a discovery-specific HTTP client, not the scanner's general client. Validate the initial URL and every redirect target as HTTP(S) with a public DNS/IP destination; reject loopback, private, link-local, multicast, unspecified, and metadata-service ranges. Revalidate resolved addresses at connection time to reduce DNS-rebinding exposure. Disable ambient proxy behavior unless it can provide the same target validation.

Cap redirect count, total inspected pages, per-response byte count, total elapsed time, and extracted-link count. Accept only HTML-like responses for page inspection and do not submit forms, execute scripts, or fetch arbitrary external resources. Return partial candidates with an incomplete status when a safe bound is reached.

Alternative considered: relying on URL hostname checks. Hostname-only checks do not defend against redirects or DNS resolution to private addresses.

### Make selection an explicit two-step API/UI flow with a shared employer group

Provide a discovery operation that produces candidate data only, followed by an add operation accepting the selected candidates and one shared employer identity. The discovery response also carries an editable employer-name suggestion: first the owner of a matching registered board, then `og:title` discovered from the relevant careers or board page, then a document-title or board-identifier fallback. The registration UI sends that same chosen name with every selected board so the backend persists them as sources of one company.

The UI uses the Jobs experience as its visual reference: a contained dark surface, concise step labels, a prominent shared employer field once results are present, compact source rows with validation chips, and a single gradient primary action such as “Add 3 sources to PayPay”. Per-source employer inputs are deliberately omitted. An explicit “add separately” path remains available for the exceptional group-company case, but is not the default workflow.

This avoids accidental persistence and makes the PayPay pattern understandable: a user sees exactly which company will receive the selected sources. It also allows a later user decision on whether to add a discovered unsupported board for tracking versus omit it.

## Risks / Trade-offs

- [Static-only extraction misses runtime-built links] → Return an explicit inconclusive/incomplete state and defer headless rendering to a later change.
- [A site contains stale or unrelated ATS URLs] → Validate supported boards, show evidence, and require user selection before persistence.
- [Group sites link to legally distinct employers] → Group selected sources under the user's explicit shared employer choice by default; make separate registration an explicit override rather than inferring it.
- [Migration can affect scheduled scans] → Backfill one board per existing configured company, retain company role ownership, and deploy migration plus scanner changes together.
- [SSRF or excessive retrieval] → Use the dedicated validated transport and hard traversal/resource limits.
- [One company can publish duplicate roles on multiple boards] → Keep existing URL-based role deduplication within company scope; assess cross-board URL collisions during implementation tests.
- [A company-level summary can obscure one unhealthy board] → Keep per-board adapter and freshness evidence prominent in the board-management modal.

## Migration Plan

1. Add the board table and backfill one board from every company that has an existing adapter type and slug, retaining enabled state and scan evidence.
2. Update scanning and company APIs to read boards while preserving the existing company and role identifiers.
3. Regenerate API DTOs, release the discovery and selection UI, and verify existing scans against seeded Greenhouse and Workable companies.
4. Roll back application code only before a migration is applied. After migration, retain the board table and roll forward with a compatibility path rather than deleting source data.

## Open Questions

- Whether unsupported recognized boards should be persistable as disabled watch-list entries in the first release, or only displayed during discovery. This does not change the discovery or safety approach.
