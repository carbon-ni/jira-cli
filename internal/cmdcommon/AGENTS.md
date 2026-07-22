# internal/cmdcommon

Shared create-command logic. Used by `issue create` and `epic create`.

## Responsibilities

- **`CreateParams`**: single struct for all create flags — shared between issue/epic
- **`SetCreateFlags`**: registers flags on a cobra command. Pass `"Epic"` prefix for epic, empty for issue.
- **Interactive survey**: `GetNextAction`, `GetMetadata`, `HandleNoInput` — the metadata/action prompt loop
- **User resolution**: `GetRelevantUser` → `api.ProxyUserSearch` → `GetUserKeyForConfiguredInstallation`
- **Custom field validation**: `ValidateCustomFields`

## Boundaries

- Can import: `api`, `internal/cmdutil`, `pkg/jira`, `survey`, `cobra`, `viper`.
- Cannot import: individual commands (`internal/cmd/issue/*`, `internal/cmd/epic/*`).

## Adding fields to create

1. Add the field to `CreateParams`
2. Add a flag in `SetCreateFlags`
3. Wire it in both `issue create` and `epic create` callers
