package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestNewSprintListOutput(t *testing.T) {
	sprints := []*jira.Sprint{
		{ID: 7, Name: "Sprint 7", Status: "active", StartDate: "2026-07-01", EndDate: "2026-07-14", BoardID: 3},
	}

	out := newSprintListOutput(sprints, "TEST", "Platform")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "sprints[1]{id,name,state,start,end}:\n" +
		"  7,Sprint 7,active,2026-07-01,2026-07-14\n" +
		"total: 1\n" +
		"project: TEST\n" +
		"board: Platform\n" +
		"hint: jira sprint list 7"
	assert.Equal(t, want, string(encoded))
}

func TestNewSprintListOutputEmpty(t *testing.T) {
	out := newSprintListOutput(nil, "TEST", "Platform")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "sprints: []\ntotal: 0\nproject: TEST\nboard: Platform\nhint: jira sprint list --help"
	assert.Equal(t, want, string(encoded))
}
