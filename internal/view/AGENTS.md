# internal/view

View renderers. Each domain gets its own file with a `<Domain>List` struct implementing `Render() error`.

## Boundaries

- Can import: `api`, `pkg/tui`, `pkg/jira`, `pkg/jira/filter`.
- Cannot import: commands (`internal/cmd/*`).

## Pattern

```go
type FooList struct {
    Project string
    Server  string
    Data    []*jira.Foo
    Display DisplayFormat
    Refresh tui.RefreshFunc
}

func (l *FooList) Render() error {
    // Plain/CSV path
    if l.Display.Plain || tui.IsDumbTerminal() { ... }
    // TUI path
    view := tui.NewTable(tui.WithSelectedFunc(...), ...)
    return view.Paint(l.data())
}
```

Canonical example: `issues.go` (`IssueList`).

## Agent-facing output

Commands that emit structured TOON/JSON use `cmdutil.PrintStructured()`, not this layer. View renderers are for human-facing TUI/plain/CSV output only.
