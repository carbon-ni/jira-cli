package list

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// issueListEnvelope is the structured shape emitted by `jira issue list` for
// agent consumers. It carries a lean, decision-relevant row set plus totals and
// a copy-runnable next-step hint, so a single command round-trip is enough for
// the common "what should I look at next" decision.
type issueListEnvelope struct {
	Issues []issueRow `json:"issues" toon:"issues"`
	Total  int        `json:"total" toon:"total"`
	Query  string     `json:"query,omitempty" toon:"query,omitempty"`
	Hint   string     `json:"hint,omitempty" toon:"hint,omitempty"`
}

// issueRow is the tabular TOON row: uniform objects with primitive values only,
// which the encoder renders as a compact `issues[N]{...}:` table.
type issueRow struct {
	Key      string `json:"key" toon:"key"`
	Type     string `json:"type" toon:"type"`
	Summary  string `json:"summary" toon:"summary"`
	Status   string `json:"status" toon:"status"`
	Assignee string `json:"assignee" toon:"assignee"`
	Priority string `json:"priority" toon:"priority"`
	Created  string `json:"created" toon:"created"`
	Updated  string `json:"updated" toon:"updated"`
}

// RenderStructured emits an issue collection as deterministic structured output
// (TOON by default, JSON for compatibility) and exits 0. Per AXI, an empty
// result is an explicit success carrying query context, never a failure.
// Sprint and epic issue lists reuse this schema as their single source of truth.
func RenderStructured(issues []*jira.Issue, jql, format string) {
	env := issueListEnvelope{
		Issues: make([]issueRow, 0, len(issues)),
		Total:  len(issues),
		Query:  jql,
	}
	for _, iss := range issues {
		env.Issues = append(env.Issues, toIssueRow(iss))
	}
	env.Hint = nextHint(issues)

	if err := cmdutil.PrintStructured(env, format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "issue-list-render-failed",
				Message: fmt.Sprintf("Could not encode issues: %s", err),
			},
		}, format, false))
	}
}

func toIssueRow(iss *jira.Issue) issueRow {
	f := iss.Fields
	return issueRow{
		Key:      iss.Key,
		Type:     f.IssueType.Name,
		Summary:  f.Summary,
		Status:   f.Status.Name,
		Assignee: f.Assignee.Name,
		Priority: f.Priority.Name,
		Created:  f.Created,
		Updated:  f.Updated,
	}
}

// nextHint gives the agent a copy-runnable next command: view the first issue
// when results exist, otherwise point at filter help.
func nextHint(issues []*jira.Issue) string {
	if len(issues) == 0 {
		return "jira issue list --help  # refine filters"
	}
	return fmt.Sprintf("jira issue view %s", issues[0].Key)
}
