package view

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// issueViewEnvelope is the structured shape emitted by `jira issue view` for
// agent consumers. It carries the decision-relevant issue facts plus the
// attachment list (id + filename + type + size) so an agent can pick one and
// download it with `jira issue download`.
type issueViewEnvelope struct {
	Key         string          `json:"key" toon:"key"`
	Summary     string          `json:"summary" toon:"summary"`
	Type        string          `json:"type" toon:"type"`
	Status      string          `json:"status" toon:"status"`
	Assignee    string          `json:"assignee" toon:"assignee"`
	Priority    string          `json:"priority" toon:"priority"`
	Attachments []attachmentRow `json:"attachments,omitempty" toon:"attachments,omitempty"`
	Hint        string          `json:"hint,omitempty" toon:"hint,omitempty"`
}

// attachmentRow is the tabular TOON row for one attachment.
type attachmentRow struct {
	ID       string `json:"id" toon:"id"`
	Filename string `json:"filename" toon:"filename"`
	MimeType string `json:"mimeType" toon:"mimeType"`
	Size     int64  `json:"size" toon:"size"`
}

// renderStructured emits the issue view as deterministic structured output
// (TOON by default, JSON for compatibility).
func renderStructured(iss *jira.Issue, format string) {
	env := issueViewEnvelope{
		Key:      iss.Key,
		Summary:  iss.Fields.Summary,
		Type:     iss.Fields.IssueType.Name,
		Status:   iss.Fields.Status.Name,
		Assignee: iss.Fields.Assignee.Name,
		Priority: iss.Fields.Priority.Name,
	}

	for _, att := range iss.Fields.Attachments {
		env.Attachments = append(env.Attachments, attachmentRow{
			ID:       att.ID,
			Filename: att.Filename,
			MimeType: att.MimeType,
			Size:     att.Size,
		})
	}
	if len(env.Attachments) > 0 {
		env.Hint = fmt.Sprintf("jira issue download %s <attachment-id>", iss.Key)
	}

	if err := cmdutil.PrintStructured(env, format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "issue-view-render-failed",
				Message: fmt.Sprintf("Could not encode issue: %s", err),
			},
		}, format, false))
	}
}
