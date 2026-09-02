## 1. Native multi-platform builds

- [x] 1.1 Update Docker build stages to use the Buildx builder platform while preserving target Go cross-compilation, and verify `docker buildx build` accepts both linux/amd64 and linux/arm64 targets.

## 2. Automated release publication

- [x] 2.1 Add root-level semantic-release configuration and invoke pinned Semantic Release integrations ephemerally in CI, without adding release dependencies to the frontend workspace.
- [x] 2.2 Configure semantic-release to calculate main-branch SemVer releases, publish versioned GHCR images before the GitHub Release, and verify a dry run recognizes the existing release history.
- [x] 2.3 Replace the tag-triggered release workflow with a main-branch semantic-release workflow using full history, Buildx/QEMU setup, a finite timeout, and plain BuildKit logs; verify workflow YAML syntax and required permissions.

## 3. Commit convention enforcement

- [x] 3.1 Add a pinned CommitMe GitHub Action pull-request check for all commits, without adding frontend dependencies or pull-request labels.
- [x] 3.2 Document Conventional Commit syntax and its SemVer mapping in AGENTS.md, and verify the guidance agrees with the release configuration.

## 4. Validation

- [ ] 4.1 Run the applicable formatting, type, build, and release-config checks, and verify all changed workflow and tool configuration files are clean.
- [x] 4.2 Run `openspec validate automate-semantic-releases --strict` and verify the completed artifacts satisfy the release-automation scenarios.
