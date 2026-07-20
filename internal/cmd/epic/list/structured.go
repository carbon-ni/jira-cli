package list

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

type epicListOutput struct {
	Epics   []epicRow `json:"epics" toon:"epics"`
	Total   int       `json:"total" toon:"total"`
	Project string    `json:"project" toon:"project"`
	Hint    string    `json:"hint" toon:"hint"`
}

type epicRow struct {
	Key      string `json:"key" toon:"key"`
	Summary  string `json:"summary" toon:"summary"`
	Status   string `json:"status" toon:"status"`
	Assignee string `json:"assignee" toon:"assignee"`
	Priority string `json:"priority" toon:"priority"`
	Start    string `json:"start" toon:"start"`
	End      string `json:"end" toon:"end"`
}

func newEpicListOutput(epics []*jira.Issue, project string) epicListOutput {
	rows := make([]epicRow, 0, len(epics))
	for _, epic := range epics {
		f := epic.Fields
		rows = append(rows, epicRow{
			Key:      epic.Key,
			Summary:  f.Summary,
			Status:   f.Status.Name,
			Assignee: f.Assignee.Name,
			Priority: f.Priority.Name,
			Start:    f.Created,
			End:      f.Updated,
		})
	}
	hint := "jira epic list --help"
	if len(rows) > 0 {
		hint = fmt.Sprintf("jira epic list %s", rows[0].Key)
	}
	return epicListOutput{Epics: rows, Total: len(rows), Project: project, Hint: hint}
}

func renderEpicListStructured(epics []*jira.Issue, project, format string) {
	if err := cmdutil.PrintStructured(newEpicListOutput(epics, project), format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "epic-list-render-failed",
				Message: fmt.Sprintf("Could not encode epics: %s", err),
			},
		}, format, false))
	}
}
