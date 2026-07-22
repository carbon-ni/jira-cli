# internal/query

JQL query builder. Translates CLI flags into JQL strings via `pkg/jql`.

## Responsibilities

- `Issue` — builds JQL from issue list flags (type, status, priority, labels, dates, etc.)
- `Sprint` — builds JQL for sprint queries
- `IssueParams` / `SprintParams` — structured flag carriers

## Pattern

```go
q := query.NewIssue(project, cmd.Flags())
jql := q.Get()           // returns JQL string
params := q.Params()      // returns From, Limit for pagination
resp, _ := api.ProxySearch(client, jql, params.From, params.Limit)
```

## Adding a filter flag

1. Add the field to `IssueParams`
2. Add a case in `setBoolParams`/`setStringParams`
3. Add the JQL clause in `Issue.Get()`
4. Register the flag in `internal/cmd/issue/list/list.go` `SetFlags()`

## Tests

`issue_test.go` uses a mock `FlagParser` interface. Follow that pattern for new flags.
