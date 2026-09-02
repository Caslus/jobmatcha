## MODIFIED Requirements

### Requirement: Registered companies are visible and manageable
The system SHALL provide authenticated users with a Companies view containing every registered company, including disabled companies. The view SHALL use a centered, compact company index rather than a full-width data table. Each company entry SHALL show its name, location when available, an explicit board-navigation control with career-board count, total job count across its boards, aggregate freshness, and whether the company is enabled for future scans. The index SHALL provide a compact, visibly interactive sort control for ordering companies by job count, board count, name, location, or recent activity. Latest scan and discovery timestamps SHALL remain available in the board-management modal rather than competing with the company identity in every list row. Company selection and bulk actions SHALL use visually consistent, clearly interactive controls. Adapter and per-board freshness details SHALL be shown only in the company's board-management modal, where each career-board source is independently identifiable and manageable. The modal SHALL expose each board's provider, identifier, canonical URL, enabled state, adapter health, freshness, latest scan, latest new-role discovery, and failure detail when present. Adapter health and freshness SHALL use semantic icons and concise plain-language explanations. Board activity timestamps SHALL display relative time and expose their full timestamp accessibly. The system SHALL allow a user to enable or disable a company or one of its career-board sources; the updated state SHALL immediately be reflected in the board-management modal and SHALL govern inclusion of the affected sources in subsequent scans.

#### Scenario: Viewing all registered companies
- **WHEN** an authenticated user opens the Companies view
- **THEN** the system displays every registered company with its company-level operational summary and enabled state

#### Scenario: Viewing board-specific details
- **WHEN** a user opens a company's board-management modal
- **THEN** the system displays each board's source identity, URL, explained adapter health and freshness states, relative activity timestamps, and any scan failure independently

#### Scenario: Sorting registered companies
- **WHEN** a user selects a company-index sort option
- **THEN** the system orders the compact company index by the selected job count, board count, name, location, or recent activity criterion

#### Scenario: Disabling a company
- **WHEN** a user disables an enabled company
- **THEN** the system persists the disabled state and excludes that company's sources from later scans

#### Scenario: Enabling a company
- **WHEN** a user enables a disabled company
- **THEN** the system persists the enabled state and includes that company's enabled sources in later scans when their adapters are supported

#### Scenario: Managing one board without changing another
- **WHEN** a user changes the state of one career-board source associated with a company
- **THEN** the system immediately reflects the changed state, changes only that source's inclusion in later scans, and leaves the company's other sources unchanged
