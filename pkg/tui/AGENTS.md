# pkg/tui

Pure rendering library wrapping `tview`/`tcell`. Zero internal dependencies (only `pkg/tui/primitive` subpackage).

## Boundaries

- No knowledge of Jira, API, auth, or commands.
- Input: `TableData` ([][]string). Output: painted terminal UI.
- Functional options pattern for configuration: `WithTableStyle`, `WithSelectedFunc`, `WithViewModeFunc`, `WithMoveFunc`, `WithRefreshFunc`, `WithCopyFunc`, `WithCopyKeyFunc`, `WithFixedColumns`.

## Components

- `table.go` — interactive table with select/view/move/copy/refresh actions
- `preview.go` — detail/preview pane
- `helper.go` — terminal detection (`IsDumbTerminal`, `IsNotTTY`)
- `text.go` — text rendering
- `screen.go` — screen management

## When to use

Any interactive list-view command. Wrap data in `TableData`, call `table.Paint()`. See `internal/view/issues.go` for canonical usage.
