package list

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestRenderStructuredTOONNonEmpty(t *testing.T) {
	issues := []*jira.Issue{
		mkIssue("TEST-1", "Bug", "Fix login", "In Progress", "Ada", "High", "2026-05-01", "2026-05-02"),
		mkIssue("TEST-2", "Story", "Add export", "Open", "", "Normal", "2026-05-03", "2026-05-03"),
	}

	got := captureStdout(t, func() {
		RenderStructured(issues, "project = TEST", cmdutil.FormatTOON)
	})

	want := "issues[2]{key,type,summary,status,assignee,priority,created,updated}:\n" +
		"  TEST-1,Bug,Fix login,In Progress,Ada,High,2026-05-01,2026-05-02\n" +
		"  TEST-2,Story,Add export,Open,\"\",Normal,2026-05-03,2026-05-03\n" +
		"total: 2\n" +
		"query: project = TEST\n" +
		"hint: jira issue view TEST-1"
	assert.Equal(t, want, got)
	assert.False(t, strings.HasSuffix(got, "\n"), "TOON must not end with a newline")
}

func TestRenderStructuredTOONQuotesCommaInSummary(t *testing.T) {
	// Summary with a comma is quoted in the tabular cell (TOON §7.2 delimiter rule).
	issues := []*jira.Issue{
		mkIssue("TEST-1", "Bug", "hello, world", "Open", "Ada", "Low", "2026-05-01", "2026-05-01"),
	}

	got := captureStdout(t, func() {
		RenderStructured(issues, "project = TEST", cmdutil.FormatTOON)
	})

	assert.Contains(t, got, "TEST-1,Bug,\"hello, world\",Open,Ada,Low,2026-05-01,2026-05-01")
}

func TestRenderStructuredEmptyIsExplicitSuccess(t *testing.T) {
	// AXI: empty result is exit-0 success with query context, not a failure.
	got := captureStdout(t, func() {
		RenderStructured(nil, "project = TEST", cmdutil.FormatTOON)
	})

	want := "issues: []\n" +
		"total: 0\n" +
		"query: project = TEST\n" +
		"hint: jira issue list --help  # refine filters"
	assert.Equal(t, want, got)
}

func TestRenderStructuredJSONIsValid(t *testing.T) {
	issues := []*jira.Issue{
		mkIssue("TEST-1", "Bug", "Fix login", "Open", "Ada", "High", "2026-05-01", "2026-05-02"),
	}

	got := captureStdout(t, func() {
		RenderStructured(issues, "project = TEST", cmdutil.FormatJSON)
	})

	assert.True(t, strings.HasSuffix(got, "\n"), "JSON must end with one newline")

	var env issueListEnvelope
	require.NoError(t, json.Unmarshal([]byte(got), &env))
	assert.Equal(t, 1, env.Total)
	assert.Equal(t, "TEST-1", env.Issues[0].Key)
	assert.Equal(t, "jira issue view TEST-1", env.Hint)
}

func TestToIssueRowMapsFields(t *testing.T) {
	row := toIssueRow(mkIssue("K-1", "T", "S", "St", "A", "P", "C", "U"))
	assert.Equal(t, issueRow{Key: "K-1", Type: "T", Summary: "S", Status: "St", Assignee: "A", Priority: "P", Created: "C", Updated: "U"}, row)
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	fn()
	require.NoError(t, w.Close())
	os.Stdout = old

	var b bytes.Buffer
	_, _ = io.Copy(&b, r)
	return b.String()
}
