## ADDED Requirements

### Requirement: Company and career-board details can be maintained
The system SHALL allow an authenticated user to edit a registered company's name and location. It SHALL allow the user to add, edit, enable, disable, and delete an individual career board within a company. A manually added or edited board SHALL require a supported provider, a valid provider-specific board identifier, and a canonical URL, and the system SHALL reject an identity already registered to another company.

#### Scenario: Editing a company
- **WHEN** a user submits a valid new name or location for a company
- **THEN** the system persists the change and displays the updated company summary

#### Scenario: Manually adding a supported board
- **WHEN** a user provides a supported provider and a valid board identity for a company
- **THEN** the system validates the board and adds it as an independently manageable source

#### Scenario: Editing a board identity
- **WHEN** a user submits a valid provider, board identifier, and canonical URL for an existing board
- **THEN** the system validates and persists the updated source identity

#### Scenario: Duplicate board identity is submitted
- **WHEN** a user adds or edits a board to use a provider and identifier already owned by another company
- **THEN** the system rejects the request and leaves all existing board records unchanged

### Requirement: Source deletion is deliberate and preserves job history
The system SHALL require explicit confirmation before deleting a career board or company. Deleting a career board SHALL remove only that source. Deleting a company SHALL remove its owned career boards. Neither deletion SHALL remove the company's previously stored roles.

#### Scenario: Deleting one board
- **WHEN** a user confirms deletion of one career board from a company with multiple boards
- **THEN** the system removes only that board and preserves its sibling boards and the company's stored roles

#### Scenario: Deleting a company
- **WHEN** a user confirms deletion of a company
- **THEN** the system removes the company and its career boards while preserving its historical roles

#### Scenario: Cancelling a deletion
- **WHEN** a user dismisses a board or company deletion confirmation
- **THEN** the system does not remove any record

