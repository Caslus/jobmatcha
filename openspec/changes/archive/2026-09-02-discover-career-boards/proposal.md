## Why

Adding a company currently requires a user to know its ATS adapter and board slug, even when the company publishes a conventional careers site that links to one or more known job boards. A single corporate careers page can also lead to distinct group-company boards, so forcing one ATS source onto one company loses both coverage and employer identity.

## What Changes

- Add a bounded careers-site discovery flow that accepts a user-supplied URL, follows safe redirects, extracts public links and embedded URL strings, and identifies known ATS board URLs.
- Have scanner adapters declare their own recognized URL patterns and normalize matched URLs into adapter-specific board identifiers.
- Validate discovered, supported boards using the adapter before presenting them to the user, and show unsupported but recognized providers without enabling scans.
- Let users select discovered board candidates and add them to one user-chosen employer group before any company or scan source is persisted; allow a source to be registered separately only through an explicit override.
- Replace the single company-level ATS configuration with separately managed career-board sources so one company can own multiple boards.
- Present the Companies list as company-level operational summary, while keeping adapter and per-board freshness details inside each company's board-management modal.
- Present that summary as a centered, compact company index rather than a wide data table, and make the board-management modal the complete per-source operational detail surface.
- Keep the company index visually sparse, with an explicit board-navigation control and a consistent, discovery-style selection and bulk-action language.
- Let users sort the compact company index by its available operational summaries without turning it back into a wide table.
- Make the board-management modal explain scan health and freshness with semantic icons, readable status language, relative activity times, and immediately responsive board controls.
- Infer a suggested shared employer name from discovered page metadata and present discovery as a polished, compact source-management workflow consistent with the Jobs experience.
- Apply strict crawl, network, and response limits to discovery requests made from user-supplied URLs.

## Capabilities

### New Capabilities

- `career-board-discovery`: Discover, classify, validate, and present career-board candidates from a public careers URL.

### Modified Capabilities

- `companies-management`: Represent and manage one or more independent career-board sources per registered company instead of one company-level adapter configuration.

## Impact

- Backend model, migrations, repositories, scanner provider registry, scanner orchestration, API DTOs/routes, and frontend company-management flows.
- Existing Greenhouse and Workable adapters gain URL recognition and board-validation behavior; new adapters become discoverable by registering their patterns.
- Generated API DTO output and frontend route tree will be regenerated through the established project tasks, not edited directly.
