## Context

The scanner runs boards concurrently with one shared `http.Client` and a global semaphore. Greenhouse and Workable providers issue HTTP calls directly through that client; Workable can issue an additional request for each job description. Responses are currently consumed by provider methods without retaining rate-limit metadata. See `proposal.md` for motivation and `specs/scanner-rate-limit-management/spec.md` for behavior.

## Goals / Non-Goals

**Goals:**

- Coordinate all scan-originating HTTP calls by provider name before they reach the transport.
- Convert standard retry and rate-limit headers into a provider-specific, context-aware cooldown.
- Keep the current request timeout and global concurrency limit intact.
- Make time and HTTP behavior controllable in tests.

**Non-Goals:**

- Infer limits when the provider supplies no usable retry/reset time.
- Persist rate-limit state, expose it through APIs, or alter scan scheduling.
- Guarantee support for every provider-specific header convention beyond `Retry-After` and the requested `X-RateLimit-*` set.

## Decisions

### Use a provider-keyed HTTP wrapper

Introduce a small scanner-internal request coordinator/wrapper used by built-in provider calls. Before a request it waits on the cooldown held for that provider; after a response it updates that provider's cooldown from headers. On a `429` with usable cooldown timing, it waits and retries once before returning the response. This centralizes behavior, covers Workable's detail requests, and avoids duplicating parsing in every adapter.

Alternative: add rate-limit logic to each adapter. Rejected because every new request path can accidentally omit it and identical parsing would be repeated.

### Treat explicit retry instructions as authoritative

Parse `Retry-After` first, accepting delta-seconds and HTTP-date values. Otherwise, act on `X-RateLimit-Reset` only when `X-RateLimit-Remaining` parses as zero or below; accept Unix epoch seconds and HTTP-date resets. Ignore `X-RateLimit-Limit` for scheduling because it describes capacity but does not tell the scanner when it can resume.

Alternative: derive a paced request interval from limit and remaining. Rejected because the headers do not define a stable window start or provider quota scope.

### Scope state to one engine and provider identity

The coordinator belongs to an engine instance and maintains independent protected state per provider. It must raise, never shorten, an existing future cooldown when multiple concurrent responses provide timing. This avoids cross-provider blockage and preserves the most conservative known reset.

Alternative: global cooldown state. Rejected because an exhausted Workable quota must not pause Greenhouse work.

### Make delay context-aware and preserve client timeout semantics

Use a timer/select mechanism controlled by the caller context for cooldown waiting. The existing HTTP client's 30-second timeout remains an execution timeout after dispatch; it does not constrain a deliberate pre-request cooldown. A `429` retry is bounded to one additional dispatch so persistent provider failures remain visible rather than extending scans indefinitely.

Alternative: sleep directly. Rejected because it cannot promptly observe cancellation and makes tests slow.

## Risks / Trade-offs

- [Provider headers use an unsupported reset format] -> Ignore unusable timing rather than stall scans; add provider-specific support later with fixtures.
- [Concurrent responses race to change cooldown] -> Guard state and retain the latest known future time.
- [A long provider reset extends job duration] -> This is intentional quota compliance; context cancellation remains available to terminate the wait.
- [Validation/discovery use the same remote quota] -> Initial scope is scan fetch traffic; extend the wrapper deliberately if provider documentation confirms a shared quota.

## Migration Plan

1. Add the coordinator and header parsing behind the scanner's built-in provider request path.
2. Add transport-based unit tests for header parsing, shared-provider waiting, provider isolation, and cancellation.
3. Run the existing backend and repository validation suite. The change is in-memory only, so no data migration is required.

Rollback consists of removing the coordinator from the provider request path; no persisted state or API contract must be reverted.
