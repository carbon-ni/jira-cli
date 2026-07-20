package list

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

type sprintListOutput struct {
	Sprints []sprintRow `json:"sprints" toon:"sprints"`
	Total   int         `json:"total" toon:"total"`
	Project string      `json:"project" toon:"project"`
	Board   string      `json:"board" toon:"board"`
	Hint    string      `json:"hint" toon:"hint"`
}

type sprintRow struct {
	ID    int    `json:"id" toon:"id"`
	Name  string `json:"name" toon:"name"`
	State string `json:"state" toon:"state"`
	Start string `json:"start" toon:"start"`
	End   string `json:"end" toon:"end"`
}

func newSprintListOutput(sprints []*jira.Sprint, project, board string) sprintListOutput {
	rows := make([]sprintRow, 0, len(sprints))
	for _, sprint := range sprints {
		rows = append(rows, sprintRow{
			ID:    sprint.ID,
			Name:  sprint.Name,
			State: sprint.Status,
			Start: sprint.StartDate,
			End:   sprint.EndDate,
		})
	}
	hint := "jira sprint list --help"
	if len(rows) > 0 {
		hint = fmt.Sprintf("jira sprint list %d", rows[0].ID)
	}
	return sprintListOutput{Sprints: rows, Total: len(rows), Project: project, Board: board, Hint: hint}
}

func renderStructured(sprints []*jira.Sprint, project, board, format string) {
	if err := cmdutil.PrintStructured(newSprintListOutput(sprints, project, board), format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "sprint-list-render-failed",
				Message: fmt.Sprintf("Could not encode sprints: %s", err),
			},
		}, format, false))
	}
}
