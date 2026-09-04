## Why

The scanner sends concurrent adapter requests with only a global concurrency limit and ignores provider rate-limit responses. A provider can therefore reject a scan while it is temporarily exhausted, even when it supplies the time at which requests may safely resume.

## What Changes

- Add provider-scoped rate-limit coordination for scanner HTTP requests.
- Honor `Retry-After` and common `X-RateLimit-*` response headers to defer later requests for the affected provider until its advertised reset time.
- Retry a rate-limited scanner request once after its valid provider-advertised cooldown.
- Make deferred requests cancellable through their request context and leave other providers able to continue scanning.
- Preserve ordinary HTTP request timeouts and existing scan failure recording separately from rate-limit waiting.

## Capabilities

### New Capabilities

- `scanner-rate-limit-management`: Coordinate scanner requests according to provider-advertised rate-limit and retry timing.

### Modified Capabilities

- None.

## Impact

- Affects the backend scanner engine and built-in ATS provider HTTP request paths.
- Adds deterministic scanner/provider tests using response-header fixtures or transports.
- Does not change public API DTOs or frontend behavior.
