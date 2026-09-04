## 1. Rate-limit coordination

- [x] 1.1 Add a scanner-internal, provider-keyed request coordinator that waits for and updates cooldown state safely; verify focused scanner unit tests pass.
- [x] 1.2 Parse `Retry-After` delta-seconds and HTTP dates, and parse applicable `X-RateLimit-Remaining` and `X-RateLimit-Reset` values; verify table-driven tests cover valid, malformed, expired, and precedence cases.
- [x] 1.3 Make cooldown waits context-aware and ensure an already-cancelled request is not dispatched; verify a cancellation test observes the context error and no transport call.

## 2. Provider integration

- [x] 2.1 Route Greenhouse and all Workable scan HTTP calls, including markdown description requests, through the provider-keyed coordinator; verify provider fixture tests still pass.
- [x] 2.2 Preserve the existing HTTP-client execution timeout and global semaphore while applying cooldowns independently per provider; verify tests demonstrate one provider's cooldown does not defer another provider's request.

## 3. Validation

- [x] 3.1 Add deterministic transport-based tests for shared cooldown state, `429` retry timing, rate-limit header precedence, and concurrent response updates; verify `mise run server:test` passes.
- [x] 3.2 Retry a `429` once after a valid provider cooldown, close the discarded response safely, and add deterministic tests for successful retry, cancellation, and bounded repeated `429` behavior.
- [ ] 3.3 Run the required application validation commands after implementation (`mise run web:check`, `mise run web:typecheck`, `mise run coverage:gate`, `mise run web:build`, and `mise run web:e2e`) and resolve any failures.
