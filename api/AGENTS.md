# api

Singleton Jira client proxy. Every command goes through this layer, never directly to `pkg/jira.Client`.

## Responsibilities

- **Auth resolution**: env var → viper → .netrc → OS keyring → cookie files
- **Client singleton**: `api.Client(config)` creates once, `api.DefaultClient(debug)` retrieves. `api.ResetClient()` for tests only.
- **V2/V3 routing**: every `Proxy*()` checks `viper.GetString("installation")` and calls the right client method.

## Adding a proxy

```go
func ProxyDoThing(c *jira.Client, ...) (*jira.Result, error) {
    it := viper.GetString("installation")
    if it == jira.InstallationTypeLocal {
        return c.DoThingV2(...)
    }
    return c.DoThing(...)
}
```

## Tests

See `api/client_test.go` for patterns: `t.TempDir()`, `t.Setenv()`, no mocking needed for cookie/file resolution tests.
