# jira-cli

Interactive Jira CLI in Go. `cobra`/`viper` for CLI, `tview`/`tcell` for TUI, `toon-go` for structured agent-facing output.

## Module routing

| Package | Guide | Role |
|---|---|---|
| `api/` | [AGENTS.md](api/AGENTS.md) | Singleton client proxy with auth resolution + V2/V3 routing |
| `pkg/jira/` | [AGENTS.md](pkg/jira/AGENTS.md) | HTTP client + domain types |
| `pkg/tui/` | [AGENTS.md](pkg/tui/AGENTS.md) | Pure TUI rendering library |
| `internal/view/` | [AGENTS.md](internal/view/AGENTS.md) | View renderers (issue list, sprint, epic, etc.) |
| `internal/query/` | — | JQL builder from CLI flags |
| `internal/cmdutil/` | — | Formatting, errors, config home |
| `internal/config/` | — | `jira init` wizard |
| `internal/cmd/` | — | Cobra commands (one pkg per command) |
| `pkg/adf/` | — | Atlassian Document Format → markdown |
| `pkg/md/jirawiki/` | — | Jira wiki markup parser |
| `pkg/jql/` | — | JQL string builder |

## Quick rules

- **All API calls go through `api.Proxy*()`**, never directly to `pkg/jira.Client`.
- **`pkg/tui` has zero internal dependencies** — don't add any.
- **`pkg/jira` ↔ `pkg/jira/filter` has a known cycle** — don't add cross-references.
- **Commands never import other commands** — shared logic goes in `internal/cmdutil/` or `internal/cmdcommon/`.
- **Default output is TOON** (machine-readable). Plain/CSV/JSON also supported.
- **Agent-facing data** uses structured envelopes with `cmdutil.PrintStructured()`.

## Testing

```bash
make test   # go test -race ./...
make lint   # golangci-lint
make ci     # lint + test
```

Table-driven, `testify/assert`, `t.TempDir()` + `t.Setenv()` for file tests. `api/client_test.go` is the cleanest pattern to follow.
