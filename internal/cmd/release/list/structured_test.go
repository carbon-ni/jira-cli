package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestNewReleaseListOutput(t *testing.T) {
	releases := []*jira.ProjectVersion{
		{ID: "10", Name: "1.0.0", Released: true},
		{ID: "11", Name: "1.1.0", Archived: true},
	}

	out := newReleaseListOutput(releases, "TEST")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "releases[2]{id,name,released,archived}:\n" +
		"  \"10\",1.0.0,true,false\n" +
		"  \"11\",1.1.0,false,true\n" +
		"total: 2\n" +
		"project: TEST\n" +
		"hint: jira release list --project TEST"
	assert.Equal(t, want, string(encoded))
}

func TestNewReleaseListOutputEmpty(t *testing.T) {
	out := newReleaseListOutput(nil, "TEST")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	assert.Equal(t, "releases: []\ntotal: 0\nproject: TEST\nhint: jira release list --help", string(encoded))
}
