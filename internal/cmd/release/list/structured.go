package list

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

type releaseListOutput struct {
	Releases []releaseRow `json:"releases" toon:"releases"`
	Total    int          `json:"total" toon:"total"`
	Project  string       `json:"project" toon:"project"`
	Hint     string       `json:"hint" toon:"hint"`
}

type releaseRow struct {
	ID       string `json:"id" toon:"id"`
	Name     string `json:"name" toon:"name"`
	Released bool   `json:"released" toon:"released"`
	Archived bool   `json:"archived" toon:"archived"`
}

func newReleaseListOutput(releases []*jira.ProjectVersion, project string) releaseListOutput {
	rows := make([]releaseRow, 0, len(releases))
	for _, release := range releases {
		rows = append(rows, releaseRow{
			ID:       release.ID,
			Name:     release.Name,
			Released: release.Released,
			Archived: release.Archived,
		})
	}
	hint := "jira release list --help"
	if len(rows) > 0 {
		hint = fmt.Sprintf("jira release list --project %s", project)
	}
	return releaseListOutput{Releases: rows, Total: len(rows), Project: project, Hint: hint}
}

func renderStructured(releases []*jira.ProjectVersion, project, format string) {
	if err := cmdutil.PrintStructured(newReleaseListOutput(releases, project), format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "release-list-render-failed",
				Message: fmt.Sprintf("Could not encode releases: %s", err),
			},
		}, format, false))
	}
}
