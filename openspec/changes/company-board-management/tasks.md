## 1. Job visibility

- [x] 1.1 Filter role-list queries, visible totals, and pagination through enabled companies while retaining historical rows in storage; verify focused API tests cover disabled and re-enabled company roles.

## 2. Company and career-board lifecycle APIs

- [x] 2.1 Add validated company name/location update and confirmed company deletion operations that remove owned boards but retain roles; update DTO sources and regenerate frontend API types; verify API and service tests.
- [x] 2.2 Add validated manual career-board creation, identity update, and deletion operations using provider normalization/validation and global identity uniqueness; verify adapter, repository, service, and API tests cover valid, invalid, duplicate, and sibling-board cases.

## 3. Company source-management experience

- [x] 3.1 Add compact company edit controls plus board add/edit/delete controls and explicit confirmations, with optimistic/query-reconciled state updates; verify frontend interaction coverage.

## 4. Integration verification

- [x] 4.1 Run `mise run "[Server] Run all tests"`, `mise run "[Web] Run tests"`, `mise run "[Web] Check"`, `mise run "[Web] Typecheck"`, and `mise run "[Web] Build"`; manually verify a disabled company's roles disappear and that deleting a source preserves historical roles.
