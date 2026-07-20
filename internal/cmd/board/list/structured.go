package list

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

type boardListOutput struct {
	Boards  []boardRow `json:"boards" toon:"boards"`
	Total   int        `json:"total" toon:"total"`
	Project string     `json:"project" toon:"project"`
	Hint    string     `json:"hint" toon:"hint"`
}

type boardRow struct {
	ID   int    `json:"id" toon:"id"`
	Name string `json:"name" toon:"name"`
	Type string `json:"type" toon:"type"`
}

func newBoardListOutput(boards []*jira.Board, total int, project string) boardListOutput {
	rows := make([]boardRow, 0, len(boards))
	for _, board := range boards {
		rows = append(rows, boardRow{ID: board.ID, Name: board.Name, Type: board.Type})
	}
	hint := "jira sprint list"
	if len(rows) == 0 {
		hint = "jira board list --help  # refine project"
	}
	return boardListOutput{Boards: rows, Total: total, Project: project, Hint: hint}
}

func renderStructured(boards []*jira.Board, total int, project, format string) {
	if err := cmdutil.PrintStructured(newBoardListOutput(boards, total, project), format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "board-list-render-failed",
				Message: fmt.Sprintf("Could not encode boards: %s", err),
			},
		}, format, false))
	}
}
