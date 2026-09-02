## Purpose

Enable users to turn a public company careers URL into verified, selectable job-board sources without knowing the company's ATS configuration in advance.

## ADDED Requirements

### Requirement: A careers URL can be analyzed for job boards
The system SHALL allow an authenticated user to submit a public careers URL for discovery and SHALL return the discovery result without registering a company or enabling a scan.

#### Scenario: Analyzing a careers URL
- **WHEN** a user submits a reachable public careers URL
- **THEN** the system returns the discovered board candidates and their discovery outcomes without changing registered companies or scan sources

#### Scenario: URL cannot be analyzed
- **WHEN** the submitted URL cannot be fetched, is malformed, or resolves to a disallowed network destination
- **THEN** the system returns an actionable discovery failure and does not register a company or scan source

### Requirement: Known ATS URLs are recognized wherever publicly embedded
The system SHALL identify registered ATS providers from URLs found in fetched document links, URL-bearing document metadata, and raw inline script content. A recognized URL SHALL be normalized into the provider's board identifier and canonical board URL.

#### Scenario: Direct ATS link is discovered
- **WHEN** a careers page contains a recognized ATS URL in a link or metadata field
- **THEN** the result includes a candidate with the normalized provider and board identifier

#### Scenario: ATS URL is embedded in script content
- **WHEN** a careers page contains a recognized ATS URL in inline or minified script content
- **THEN** the result includes the same normalized candidate as if the URL had appeared in a visible link

### Requirement: Discovery follows relevant careers links within a bounded scope
The system SHALL inspect a bounded number of relevant career or recruiting links from the submitted page so that group-company career sites can yield distinct board candidates. Discovery SHALL not recursively crawl arbitrary links or the open web.

#### Scenario: Linked group-company board is discovered
- **WHEN** a submitted careers page links to a group-company careers page that contains a recognized board URL
- **THEN** the result includes the linked board as a distinct candidate and identifies the page on which it was found

#### Scenario: Irrelevant external link is encountered
- **WHEN** a page includes a non-careers external link that does not match a known provider
- **THEN** the system does not follow that link solely for discovery

### Requirement: Discovered boards are independently validated and presented for selection
The system SHALL validate candidates for supported providers before marking them ready to add. It SHALL deduplicate a board found through multiple paths, preserve the evidence URL, and present each discovered board independently for selection. Recognized providers without a supported scanner SHALL be presented as unsupported and SHALL not be enabled for scanning.

#### Scenario: Same board appears multiple times
- **WHEN** a recognized board is found on more than one inspected page
- **THEN** the system presents one candidate with the collected evidence rather than duplicate candidates

#### Scenario: Multiple group boards are found
- **WHEN** discovery finds distinct boards from a parent company and a group company
- **THEN** the system presents them as separately selectable sources and groups selected sources under the user-chosen employer by default

#### Scenario: Recognized board has no scanner adapter
- **WHEN** discovery identifies a provider that is recognized but not scannable
- **THEN** the result identifies it as unsupported and does not offer it as scan-ready

### Requirement: Selected boards are added as explicit scan sources
The system SHALL register only the candidates selected by the user. Each selected board SHALL retain its provider, board identifier, canonical URL, and its associated employer; enabling or disabling one board SHALL not change another board's state. A discovery submission SHALL use one shared employer name for all selected boards unless the user explicitly elects to add a source separately.

#### Scenario: User selects one of several candidates
- **WHEN** a user selects one discovered board and confirms adding it
- **THEN** the system registers only that board as a scan source

#### Scenario: User adds several sources to one employer
- **WHEN** a user selects several discovered boards and confirms adding them with a shared employer name
- **THEN** the system registers every selected board under that one company and makes the target company clear before confirmation

#### Scenario: Suggested employer name is available
- **WHEN** a discovered careers or board page exposes an `og:title` value
- **THEN** the discovery result uses it as the editable shared employer-name suggestion, preferring an already registered owner of a selected board when one exists

### Requirement: Discovery is presented as a focused source-management workflow
The Companies view SHALL present discovery in a compact, visually coherent workflow consistent with the Jobs experience. It SHALL provide a clear URL step, an editable shared employer-name step after results are available, compact selectable source rows with readable provider and validation status, and one primary action that names the target employer and selected-source count.

#### Scenario: Reviewing discovered sources
- **WHEN** discovery returns one or more candidates
- **THEN** the user can understand the proposed employer, selected-source count, and each source's readiness without repeated employer-name inputs

#### Scenario: User selects no candidates
- **WHEN** a user leaves all discovery candidates unselected
- **THEN** the system does not create a company or scan source

### Requirement: Discovery is safe and resource bounded
The system MUST block discovery requests to loopback, private, link-local, and otherwise non-public network destinations, including after redirects. It SHALL enforce request timeouts, response-size limits, redirect limits, inspected-page limits, and an allowlist of extraction and follow-link behavior sufficient to prevent user-supplied URLs from becoming an open crawler.

#### Scenario: Redirect reaches a private address
- **WHEN** a public submitted URL redirects to a private or loopback address
- **THEN** the system blocks the redirect and returns a discovery failure without making a request to that address

#### Scenario: Discovery exceeds a configured bound
- **WHEN** discovery reaches a configured page, time, redirect, or response-size limit
- **THEN** the system stops additional retrieval and returns any already discovered candidates together with an incomplete-result indication
