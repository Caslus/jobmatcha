## Purpose

Give users a transparent, controllable view of every registered job source, including whether it can be scanned and whether it is still yielding new opportunities.

## ADDED Requirements

### Requirement: Registered companies are visible and manageable
The system SHALL provide authenticated users with a Companies view containing every registered company, including disabled companies. Each company entry SHALL show its name, location when available, configured ATS adapter, and whether it is enabled for future scans. The system SHALL allow a user to enable or disable a company, and the updated state SHALL govern its inclusion in subsequent scans.

#### Scenario: Viewing all registered companies
- **WHEN** an authenticated user opens the Companies view
- **THEN** the system displays every registered company with its current enabled state and source metadata

#### Scenario: Disabling a company
- **WHEN** a user disables an enabled company
- **THEN** the system persists the disabled state and excludes that company from later scans

#### Scenario: Enabling a company
- **WHEN** a user enables a disabled company
- **THEN** the system persists the enabled state and includes that company in later scans when its adapter is supported

### Requirement: Companies can be sorted by displayed information
The Companies view SHALL allow users to sort the full registered-company list by every displayed company attribute: company name, location, configured adapter, enabled state, role count, adapter status, freshness status, latest scan time, and latest new-role-discovery time. The system SHALL make the active sort attribute and direction apparent to the user with distinct ascending and descending indicators. The initial sort SHALL rank companies by role count descending.

#### Scenario: Sorting by adapter status
- **WHEN** a user selects adapter status as the sort attribute
- **THEN** the system reorders the full Companies list by adapter status and indicates the selected sort direction

#### Scenario: Reversing a sort
- **WHEN** a user selects the active sort attribute again
- **THEN** the system reverses the sort direction for the full Companies list

### Requirement: Company state and health signals are compact and explainable
The Companies view SHALL present enabled state as a single interactive toggle that updates the company state. Adapter and freshness statuses SHALL use a compact visual indicator with tooltip text that explains its status; a failing-adapter tooltip SHALL include the available failure detail.

#### Scenario: Inspecting a status indicator
- **WHEN** a user hovers or focuses an adapter or freshness indicator
- **THEN** the system displays that status's human-readable explanation

#### Scenario: Toggling a company
- **WHEN** a user changes an enabled-state toggle
- **THEN** the system persists the new state and the toggle reflects the returned value

### Requirement: Company role counts are visible
The system SHALL return and display the count of stored roles for every company. Users SHALL be able to sort by this count.

#### Scenario: Viewing role counts
- **WHEN** an authenticated user opens the Companies view
- **THEN** every company row shows its stored role count and the initial list is ordered by count descending

### Requirement: Companies can be enabled or disabled in bulk
The Companies view SHALL allow users to select one or more companies and apply a single bulk action to enable or disable all selected companies. The system SHALL change only the selected companies and SHALL report an error without changing unselected companies when a bulk update cannot be completed.

#### Scenario: Bulk disable selected companies
- **WHEN** a user selects multiple enabled companies and chooses bulk disable
- **THEN** the system disables every selected company and excludes each from future scans

#### Scenario: Bulk enable selected companies
- **WHEN** a user selects multiple disabled companies and chooses bulk enable
- **THEN** the system enables every selected company and includes each in future scans when its adapter is supported

#### Scenario: No companies selected
- **WHEN** no companies are selected
- **THEN** the system does not offer an actionable bulk enable or bulk disable operation

#### Scenario: Bulk update cannot complete
- **WHEN** a bulk company-state update cannot be completed
- **THEN** the system reports the failure and does not change any unselected company

### Requirement: Adapter availability and outcome are communicated
For every company, the system SHALL expose an adapter status that is one of `healthy`, `failing`, `unsupported`, or `unknown`. `unsupported` SHALL mean no scanner adapter is available for the company's configured adapter type. `unknown` SHALL mean the company has a supported adapter but no scan attempt has been recorded. `healthy` SHALL mean the latest scan attempt completed successfully. `failing` SHALL mean the latest scan attempt did not complete successfully and SHALL include the latest failure detail when available.

#### Scenario: Unsupported adapter
- **WHEN** a company is configured with an adapter that Jobmatcha does not provide
- **THEN** its adapter status is `unsupported`

#### Scenario: Latest scan succeeds
- **WHEN** a supported company's scan completes successfully, including a scan that finds zero roles
- **THEN** its adapter status is `healthy` and any prior failure detail is cleared

#### Scenario: Latest scan fails
- **WHEN** a supported company's scan attempt fails
- **THEN** its adapter status is `failing` and the Companies view presents the latest available failure detail

### Requirement: Job-feed freshness is indicated only where actionable
The system SHALL assess job-feed freshness only for enabled companies with supported adapters. A company SHALL be marked stale when Jobmatcha has not discovered a new role for that company in the preceding 30 days. A supported enabled company with no discovered roles SHALL display a neutral no-activity-yet state instead of a stale warning. Disabled companies SHALL not display a stale warning or freshness warning.

#### Scenario: Recent new role
- **WHEN** an enabled company with a supported adapter has a role newly discovered within 30 days
- **THEN** the Companies view reports the company's feed as fresh

#### Scenario: No new role for 30 days
- **WHEN** an enabled company with a supported adapter has previously had roles discovered but none in the preceding 30 days
- **THEN** the Companies view shows a stale indicator explaining that no new roles have been found in 30 days

#### Scenario: No roles have been discovered
- **WHEN** an enabled company with a supported adapter has never had a role discovered
- **THEN** the Companies view shows a neutral no-activity-yet state

#### Scenario: Disabled company has old activity
- **WHEN** a disabled company has not had a new role discovered for more than 30 days
- **THEN** the Companies view does not show a stale or freshness warning for that company

### Requirement: Per-company scan evidence is retained
The system SHALL retain the latest scan-attempt time, latest successful scan time, latest scan failure detail, and latest new-role-discovery time for each company. This evidence SHALL be updated independently for each company during a scan so that the Companies view does not infer an individual company's health from an aggregate scan job.

#### Scenario: Successful empty scan
- **WHEN** a supported company scan completes without returning new or existing roles
- **THEN** the system records a successful scan attempt for that company without recording a new-role discovery

#### Scenario: New role discovery
- **WHEN** a scan discovers one or more roles not previously known for a company
- **THEN** the system records the time of that discovery for the company
