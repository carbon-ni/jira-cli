package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestNewEpicListOutput(t *testing.T) {
	issues := []*jira.Issue{
		mkIssue("TEAM-1", "Epic", "Platform 2.0", "In Progress", "Ada", "High", "2026-07-01", "2026-07-14"),
		mkIssue("TEAM-2", "Epic", "Mobile App", "Open", "", "Medium", "2026-06-01", "2026-06-30"),
	}

	out := newEpicListOutput(issues, "TEAM")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "epics[2]{key,summary,status,assignee,priority,start,end}:\n" +
		"  TEAM-1,Platform 2.0,In Progress,Ada,High,2026-07-01,2026-07-14\n" +
		"  TEAM-2,Mobile App,Open,\"\",Medium,2026-06-01,2026-06-30\n" +
		"total: 2\n" +
		"project: TEAM\n" +
		"hint: jira epic list TEAM-1"
	assert.Equal(t, want, string(encoded))
}

func TestNewEpicListOutputEmpty(t *testing.T) {
	out := newEpicListOutput(nil, "TEAM")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "epics: []\ntotal: 0\nproject: TEAM\nhint: jira epic list --help"
	assert.Equal(t, want, string(encoded))
}

func mkIssue(key, itype, summary, status, assignee, priority, created, updated string) *jira.Issue {
	iss := &jira.Issue{Key: key}
	iss.Fields.Summary = summary
	iss.Fields.IssueType.Name = itype
	iss.Fields.Status.Name = status
	iss.Fields.Assignee.Name = assignee
	iss.Fields.Priority.Name = priority
	iss.Fields.Created = created
	iss.Fields.Updated = updated
	return iss
}
