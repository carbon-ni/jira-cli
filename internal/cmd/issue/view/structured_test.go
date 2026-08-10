package view

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

func mkIssue(attachments []jira.Attachment) *jira.Issue {
	return &jira.Issue{
		Key: "TEST-1",
		Fields: jira.IssueFields{
			Summary:   "Has a design",
			IssueType: jira.IssueType{Name: "Task"},
			Status: struct {
				Name string `json:"name"`
			}{Name: "Open"},
			Assignee: struct {
				Name string `json:"displayName"`
			}{Name: "Ada"},
			Priority: struct {
				Name string `json:"name"`
			}{Name: "High"},
			Attachments: attachments,
		},
	}
}

func TestRenderStructuredTOONIncludesAttachments(t *testing.T) {
	iss := mkIssue([]jira.Attachment{
		{ID: "87891", Filename: "design.png", MimeType: "image/png", Size: 35748},
		{ID: "87892", Filename: "notes.md", MimeType: "text/markdown", Size: 2048},
	})

	got := captureStdout(t, func() {
		renderStructured(iss, cmdutil.FormatTOON)
	})

	assert.Contains(t, got, "key: TEST-1")
	assert.Contains(t, got, "summary: Has a design")
	assert.Contains(t, got, "attachments[2]{id,filename,mimeType,size}:")
	assert.Contains(t, got, `"87891",design.png,image/png,35748`)
	assert.Contains(t, got, `"87892",notes.md,text/markdown,2048`)
	assert.Contains(t, got, "hint: jira issue download TEST-1 <attachment-id>")
	assert.False(t, strings.HasSuffix(got, "\n"), "TOON must not end with a newline")
}

func TestRenderStructuredTOONOmitsAttachmentsWhenEmpty(t *testing.T) {
	got := captureStdout(t, func() {
		renderStructured(mkIssue(nil), cmdutil.FormatTOON)
	})

	assert.Contains(t, got, "key: TEST-1")
	assert.NotContains(t, got, "attachments")
	assert.NotContains(t, got, "hint: jira issue download")
}

func TestRenderStructuredJSONIsValid(t *testing.T) {
	iss := mkIssue([]jira.Attachment{
		{ID: "87891", Filename: "design.png", MimeType: "image/png", Size: 35748},
	})

	got := captureStdout(t, func() {
		renderStructured(iss, cmdutil.FormatJSON)
	})

	require.True(t, json.Valid([]byte(got)), "output must be valid JSON: %q", got)

	var env map[string]any
	require.NoError(t, json.Unmarshal([]byte(got), &env))
	assert.Equal(t, "TEST-1", env["key"])
	atts, ok := env["attachments"].([]any)
	require.True(t, ok)
	require.Len(t, atts, 1)
	first := atts[0].(map[string]any)
	assert.Equal(t, "87891", first["id"])
	assert.Equal(t, "design.png", first["filename"])
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

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
