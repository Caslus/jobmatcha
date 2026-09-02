## Context

See proposal.md for motivation and specs/release-automation/spec.md for required behavior. The current release workflow starts from manually pushed tags, builds both builder stages for each target platform, and publishes the image before creating a GitHub Release.

## Goals / Non-Goals

**Goals:**

- Determine release versions and notes from Conventional Commits on main.
- Publish a versioned multi-architecture image before creating the matching GitHub Release.
- Run architecture-independent frontend work and Go cross-compilation on the native Buildx platform.
- Prevent invalid release-driving commits from merging.

**Non-Goals:**

- Rewriting historical commits or tags.
- Publishing frontend packages to npm.
- Supporting prerelease or maintenance branches in this change.

## Decisions

### Semantic-release owns release tags and GitHub Releases

Run semantic-release after a pull request has passed its required checks and merged to the protected main branch, with full Git history. Its commit analyzer and GitHub integration calculate the SemVer release, create the tag, generate notes, and create the GitHub Release. This avoids a manual version input and makes v1.0.0 the baseline once the repository owner removes the failed tag.

Alternative considered: retain tag-triggered publishing and ask contributors to choose tags. Rejected because it preserves the manual release failure mode.

### Image publishing runs inside the semantic-release publish lifecycle

Use an exec plugin to run Buildx publishing with `nextRelease.version` after release analysis and before the GitHub Release plugin. This makes the release version available for image tags while ensuring the public GitHub Release is not created until the image is available.

Alternative considered: publish on a separate `release` event workflow. Rejected because it makes image publication and release creation separate failure domains.

### Build stages use the Buildx builder platform

Declare `BUILDPLATFORM` before Docker stages and select it for the Node frontend and Go builder stages. The final runtime remains target-platform-specific; Go continues to set `GOOS` and `GOARCH` from Buildx target arguments.

Alternative considered: leave stages target-platform-specific and rely on QEMU. Rejected because the canceled run demonstrated impractical frontend build time under arm64 emulation.

### Conventional Commit validation is a required PR CI job

Use the pinned CommitMe GitHub Action to validate every pull-request commit against the release-recognized Conventional Commit types. It runs without installing tooling in the frontend workspace and does not update pull-request labels. Add the same format and SemVer mapping to AGENTS.md so automated contributors select release-meaningful commit types before CI enforces them.

Alternative considered: documentation-only enforcement. Rejected because semantic-release would silently skip or misclassify nonconforming merged commits.

## Risks / Trade-offs

- [The existing v1.0.0 tag remains] → The repository owner removes it before the first main-branch automated release so semantic-release can generate v1.0.0 from history.
- [Release tooling is Node-based in a non-Node monorepo root] → Keep the configuration at the repository root and install pinned semantic-release packages only in the ephemeral GitHub Actions runner with `npx`; do not add them to the frontend project.
- [GitHub token permissions differ by repository settings] → Retain explicit contents and packages write permissions and use the workflow-scoped GitHub token.

## Migration Plan

1. Merge the release automation change without creating a manual tag.
2. Remove the failed remote v1.0.0 tag.
3. Merge or push the first release-eligible Conventional Commit to main.
4. Verify the generated tag, GitHub Release, and both image platforms.
5. Roll back workflow behavior by disabling the main-branch release workflow; published immutable tags and images remain intact.
