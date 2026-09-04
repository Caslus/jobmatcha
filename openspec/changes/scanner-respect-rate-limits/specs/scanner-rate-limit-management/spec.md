## Purpose

Coordinate scanner requests with provider-advertised limits so scheduled scans avoid known temporary request exhaustion without stopping unrelated adapters.

## ADDED Requirements

### Requirement: Scanner observes provider rate-limit timing
The scanner SHALL observe rate-limit timing advertised on provider HTTP responses. It SHALL use a valid `Retry-After` value as the preferred delay signal and otherwise use `X-RateLimit-Reset` when `X-RateLimit-Remaining` indicates no requests remain. A malformed, missing, or expired timing value SHALL not delay later requests.

#### Scenario: Provider quota is exhausted with a reset time
- **WHEN** a provider response reports zero remaining requests and a future valid `X-RateLimit-Reset` time
- **THEN** subsequent scanner requests to that provider wait until the reset time

#### Scenario: Retry timing takes precedence
- **WHEN** a provider response includes both a valid `Retry-After` value and rate-limit reset headers
- **THEN** subsequent scanner requests to that provider use the `Retry-After` delay

#### Scenario: Rate-limit headers cannot be interpreted
- **WHEN** a provider response has missing, malformed, or expired rate-limit timing headers
- **THEN** the scanner does not impose a rate-limit delay based on those headers

### Requirement: Scanner isolates provider rate-limit delays
The scanner SHALL coordinate rate-limit delays independently for each provider. A delay for one provider SHALL not prevent requests for another provider from proceeding, subject to the scanner's existing concurrency controls.

#### Scenario: One provider is cooling down
- **WHEN** requests for one provider are deferred until its advertised reset time
- **THEN** the scanner continues eligible requests for a different provider

### Requirement: Deferred scanner requests respect cancellation
The scanner SHALL stop waiting for a provider cooldown when the request context is cancelled or reaches its deadline, and SHALL return the context error without sending that deferred request.

#### Scenario: Scan is cancelled during a provider cooldown
- **WHEN** a scanner request is waiting for a provider reset and its context is cancelled
- **THEN** the request returns the context cancellation error and no HTTP request is made after cancellation

### Requirement: Scanner preserves distinct request-timeout behavior
The scanner SHALL retain its configured HTTP request timeout independently of provider rate-limit delays. A rate-limit delay SHALL not be reported as an HTTP request timeout.

#### Scenario: Request starts after a rate-limit delay
- **WHEN** a scanner request waits for a provider cooldown and then starts
- **THEN** its HTTP request timeout applies from the request execution and not from the start of the cooldown
