package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestNewBoardListOutput(t *testing.T) {
	boards := []*jira.Board{
		{ID: 1, Name: "Platform", Type: "scrum"},
		{ID: 2, Name: "Ops", Type: "kanban"},
	}

	out := newBoardListOutput(boards, 4, "TEST")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "boards[2]{id,name,type}:\n" +
		"  1,Platform,scrum\n" +
		"  2,Ops,kanban\n" +
		"total: 4\n" +
		"project: TEST\n" +
		"hint: jira sprint list"
	assert.Equal(t, want, string(encoded))
}

func TestNewBoardListOutputEmpty(t *testing.T) {
	out := newBoardListOutput(nil, 0, "TEST")
	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "boards: []\n" +
		"total: 0\n" +
		"project: TEST\n" +
		"hint: jira board list --help  # refine project"
	assert.Equal(t, want, string(encoded))
}
