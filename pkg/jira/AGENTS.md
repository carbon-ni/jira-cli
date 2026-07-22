# pkg/jira

Jira REST API client. HTTP layer + domain types. No CLI knowledge, no rendering.

## Boundaries

- **Domain types**: `Issue`, `SearchResult`, `User`, `Transition`, `CreateRequest`, etc. — defined here, consumed everywhere.
- **HTTP client**: `Client` with `Get/Post/Put/Delete` + V1/V2/V3 endpoint wrappers.
- **Auth**: `AuthType` enum (Bearer, Cookie, mTLS). Token passed via `Config`, never resolved here.
- **V1 endpoints**: special-cased (create issue) — avoid adding new V1 calls.

## Adding endpoints

1. Add the method on `Client` (e.g., `Client.GetFoo`)
2. Add a V2 variant if it differs (`Client.GetFooV2`)
3. Expose it in `api/client.go` as a `Proxy*()` that checks `viper.GetString("installation")`

## Known issue

`pkg/jira` ↔ `pkg/jira/filter` has an import cycle. Do not add new cross-references between these two.
