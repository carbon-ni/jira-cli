package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestNewProjectListOutput(t *testing.T) {
	project := &jira.Project{Key: "TEST", Name: "Test Project", Type: "software"}
	project.Lead.Name = "Ada"

	out := newProjectListOutput([]*jira.Project{project})
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "projects[1]{key,name,type,lead}:\n" +
		"  TEST,Test Project,software,Ada\n" +
		"total: 1\n" +
		"hint: jira issue list --project TEST"
	assert.Equal(t, want, string(encoded))
}

func TestNewProjectListOutputEmpty(t *testing.T) {
	out := newProjectListOutput(nil)
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	assert.Equal(t, "projects: []\ntotal: 0\nhint: jira project list --help", string(encoded))
}
