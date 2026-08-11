# internal/view

View renderers. Each domain gets its own file with a `<Domain>List` struct implementing `Render() error`.

## Boundaries

- Can import: `api`, `pkg/jira`, `pkg/jira/filter`.
- Cannot import: commands (`internal/cmd/*`).

## Pattern

Renderers are plain-text only (table/CSV). There is no interactive TUI layer anymore.

```go
type FooList struct {
    Project string
    Server  string
    Data    []*jira.Foo
    Display DisplayFormat
}

func (l *FooList) Render() error {
    if l.Display.CSV {
        return l.renderCSV(os.Stdout)
    }
    w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, 1, '\t', 0)
    return l.renderPlain(w, delimiter)
}
```

Canonical example: `issues.go` (`IssueList`).

Terminal-capability checks (`cmdutil.IsDumbTerminal`, `cmdutil.IsNotTTY`) live in `internal/cmdutil`, not here.

## Agent-facing output

Commands that emit structured TOON/JSON use `cmdutil.PrintStructured()`, not this layer. View renderers are for human-facing plain/CSV output only.
