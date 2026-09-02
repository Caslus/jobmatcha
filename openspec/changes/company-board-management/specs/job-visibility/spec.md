## Purpose

Keep the Jobs experience consistent with source controls by showing roles only from companies currently enabled for job scanning.

## ADDED Requirements

### Requirement: Jobs exclude disabled companies
The system SHALL exclude roles belonging to disabled companies from the Jobs list, its pagination totals, and the visible unfiltered job count. Re-enabling a company SHALL make its existing non-hidden roles eligible to appear again without requiring a new scan.

#### Scenario: Disabled-company role is listed
- **WHEN** a role belongs to a disabled company
- **THEN** the Jobs list does not return that role or include it in list totals

#### Scenario: Company is re-enabled
- **WHEN** a disabled company with stored non-hidden roles is enabled
- **THEN** those roles are eligible to appear in the Jobs list again

