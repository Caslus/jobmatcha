## Why

The manually-tagged v1.0.0 release stalled during an emulated arm64 frontend build and required cancellation. Releases should be reproducible from merged work, publish architecture-correct images, and derive their versions from enforced Conventional Commits.

## What Changes

- Run the architecture-independent frontend build and the cross-compiling Go build on the native Buildx builder platform.
- Replace tag-triggered manual publishing with semantic-release on the main branch.
- Publish multi-architecture GHCR images and create GitHub Releases only for successful semantic releases.
- Enforce Conventional Commit messages in pull-request CI and document the versioning convention for agents.
- Add bounded release execution and readable BuildKit progress.

## Capabilities

### New Capabilities

- `release-automation`: Automated SemVer release, GitHub Release, and multi-architecture container publication driven by validated Conventional Commits on main.

### Modified Capabilities

- None.

## Impact

Affected systems include the Dockerfile, GitHub Actions release and pull-request workflows, root-level release configuration, and contributor/agent guidance. The existing remote `v1.0.0` tag is outside this change and will be removed by the repository owner before the first automated release.
