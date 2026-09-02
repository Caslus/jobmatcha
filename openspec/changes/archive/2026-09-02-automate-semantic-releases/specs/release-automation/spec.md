## Purpose

Automate versioned application releases from validated main-branch changes and publish matching container images.

## ADDED Requirements

### Requirement: Releases derive their version from merged commits
The release system SHALL evaluate validated commits on the main branch using Conventional Commit semantics and SHALL create a release only when those commits require a semantic version increment.

#### Scenario: A feature is merged to main
- **WHEN** a validated `feat:` commit is added to main after the last release
- **THEN** the system creates the next minor semantic version release

#### Scenario: No releasable change is merged
- **WHEN** only commits that do not require a release are added to main
- **THEN** the system completes without creating a tag, GitHub Release, or container image version

### Requirement: Published container images match the release
For every created release, the system SHALL publish a multi-architecture GHCR image for linux/amd64 and linux/arm64 using tags that identify the released semantic version.

#### Scenario: A release is published
- **WHEN** a semantic release is created
- **THEN** matching linux/amd64 and linux/arm64 images are available under the release version tag before the GitHub Release is published

### Requirement: Release publication is observable and bounded
The release system SHALL impose a finite job timeout and SHALL emit plain build progress suitable for diagnosing multi-platform builds.

#### Scenario: A platform build stalls
- **WHEN** a release build stops making progress
- **THEN** the release job terminates within its configured timeout and its log identifies the active BuildKit step

### Requirement: Pull requests use release-recognized commit messages
The continuous-integration system SHALL reject pull requests whose commits do not conform to the Conventional Commit format accepted by the release system, using a dedicated GitHub Action without adding commit-validation dependencies to the frontend project.

#### Scenario: A pull request contains an invalid commit message
- **WHEN** pull-request validation examines a commit that is not a valid Conventional Commit
- **THEN** the commit-message validation job fails and prevents the pull request from satisfying CI
