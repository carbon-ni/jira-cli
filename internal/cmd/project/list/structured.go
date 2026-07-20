package list

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

type projectListOutput struct {
	Projects []projectRow `json:"projects" toon:"projects"`
	Total    int          `json:"total" toon:"total"`
	Hint     string       `json:"hint" toon:"hint"`
}

type projectRow struct {
	Key  string `json:"key" toon:"key"`
	Name string `json:"name" toon:"name"`
	Type string `json:"type" toon:"type"`
	Lead string `json:"lead" toon:"lead"`
}

func newProjectListOutput(projects []*jira.Project) projectListOutput {
	rows := make([]projectRow, 0, len(projects))
	for _, project := range projects {
		rows = append(rows, projectRow{
			Key:  project.Key,
			Name: project.Name,
			Type: project.Type,
			Lead: project.Lead.Name,
		})
	}
	hint := "jira project list --help"
	if len(rows) > 0 {
		hint = fmt.Sprintf("jira issue list --project %s", rows[0].Key)
	}
	return projectListOutput{Projects: rows, Total: len(rows), Hint: hint}
}

func renderStructured(projects []*jira.Project, format string) {
	if err := cmdutil.PrintStructured(newProjectListOutput(projects), format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "project-list-render-failed",
				Message: fmt.Sprintf("Could not encode projects: %s", err),
			},
		}, format, false))
	}
}
