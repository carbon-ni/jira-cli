package download

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// issueJSON mirrors a v3 issue payload; %s placeholders are the attachment
// content URLs served by the test content server.
const issueJSON = `{
	"key": "TEST-1",
	"fields": {
		"summary": "Has a design",
		"attachment": [
			{"id": "87891", "filename": "design.png", "mimeType": "image/png", "size": 35748, "content": "%s"},
			{"id": "87892", "filename": "notes.md", "mimeType": "text/markdown", "size": 2048, "content": "%s"}
		]
	}
}`

func newDownloadCmd(t *testing.T, format string) *cobra.Command {
	t.Helper()

	cmd := NewCmdDownload()
	cmd.Flags().Bool("debug", false, "Turn on debug output")

	if format != "" {
		cmd.Flags().String("format", cmdutil.FormatAuto, "")
		require.NoError(t, cmd.Flags().Set("format", format))
	}
	return cmd
}

// setupIssueServer starts a Jira-like API server plus a file content server.
func setupIssueServer(t *testing.T) *httptest.Server {
	t.Helper()

	contentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		switch r.URL.Path {
		case "/a":
			_, _ = w.Write([]byte("PNGDATA"))
		case "/b":
			_, _ = w.Write([]byte("MARKDOWN"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(contentServer.Close)

	issueServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/TEST-1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, issueJSON, contentServer.URL+"/a", contentServer.URL+"/b")
	}))
	t.Cleanup(issueServer.Close)

	return issueServer
}

func configureViper(t *testing.T, serverURL string) {
	t.Helper()

	viper.Reset()
	api.ResetClient()

	viper.Set("server", serverURL)
	viper.Set("project.key", "")
	viper.Set("auth_type", "cookie")
	viper.Set("cookies", "cloud.session=abc")
}

func mkAttachmentIssue(attachments []jira.Attachment) *jira.Issue {
	return &jira.Issue{
		Key:    "TEST-1",
		Fields: jira.IssueFields{Attachments: attachments},
	}
}

func TestDownloadWritesFileByID(t *testing.T) {
	issueServer := setupIssueServer(t)
	configureViper(t, issueServer.URL)

	dir := t.TempDir()
	dest := filepath.Join(dir, "design.png")

	cmd := newDownloadCmd(t, "")
	require.NoError(t, cmd.Flags().Set("output", dest))

	download(cmd, []string{"TEST-1", "87891"})

	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "PNGDATA", string(data))
}

func TestDownloadWritesFileByFilename(t *testing.T) {
	issueServer := setupIssueServer(t)
	configureViper(t, issueServer.URL)

	dir := t.TempDir()
	dest := filepath.Join(dir, "design.png")

	cmd := newDownloadCmd(t, "")
	require.NoError(t, cmd.Flags().Set("output", dest))

	download(cmd, []string{"TEST-1", "design.png"})

	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "PNGDATA", string(data))
}

func TestDownloadEmitsTOONEnvelope(t *testing.T) {
	issueServer := setupIssueServer(t)
	configureViper(t, issueServer.URL)

	dir := t.TempDir()
	dest := filepath.Join(dir, "design.png")

	cmd := newDownloadCmd(t, "")
	require.NoError(t, cmd.Flags().Set("output", dest))

	got := captureStdout(t, func() { download(cmd, []string{"TEST-1", "87891"}) })

	assert.Contains(t, got, "issue: TEST-1")
	assert.Contains(t, got, `attachmentId: "87891"`)
	assert.Contains(t, got, "filename: design.png")
	assert.Contains(t, got, "mimeType: image/png")
	assert.Contains(t, got, "size: 35748")
	assert.Contains(t, got, "savedTo: "+dest)
	assert.False(t, strings.HasSuffix(got, "\n"), "TOON must not end with a newline")
}

func TestDownloadAutoFormatPrintsHumanLine(t *testing.T) {
	issueServer := setupIssueServer(t)
	configureViper(t, issueServer.URL)

	dir := t.TempDir()
	dest := filepath.Join(dir, "design.png")

	cmd := newDownloadCmd(t, "auto")
	require.NoError(t, cmd.Flags().Set("output", dest))

	got := captureStdout(t, func() { download(cmd, []string{"TEST-1", "87891"}) })

	assert.Contains(t, got, "Downloaded design.png (34.9 KB)")
	assert.Contains(t, got, dest)
}

func TestFindAttachmentByID(t *testing.T) {
	iss := mkAttachmentIssue([]jira.Attachment{
		{ID: "87891", Filename: "design.png"},
		{ID: "87892", Filename: "notes.md"},
	})

	att, err := findAttachment(iss, "87892")
	require.NoError(t, err)
	assert.Equal(t, "notes.md", att.Filename)
}

func TestFindAttachmentByFilename(t *testing.T) {
	iss := mkAttachmentIssue([]jira.Attachment{
		{ID: "87891", Filename: "design.png"},
		{ID: "87892", Filename: "notes.md"},
	})

	att, err := findAttachment(iss, "design.png")
	require.NoError(t, err)
	assert.Equal(t, "87891", att.ID)
}

func TestFindAttachmentErrors(t *testing.T) {
	iss := mkAttachmentIssue([]jira.Attachment{
		{ID: "87891", Filename: "design.png"},
	})

	cases := []struct {
		name   string
		issue  *jira.Issue
		target string
	}{
		{name: "issue has no attachments", issue: mkAttachmentIssue(nil), target: "x"},
		{name: "unknown id", issue: iss, target: "99999"},
		{name: "unknown filename", issue: iss, target: "missing.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := findAttachment(tc.issue, tc.target)
			require.Error(t, err)
		})
	}
}

func TestFindAttachmentAmbiguousFilename(t *testing.T) {
	iss := mkAttachmentIssue([]jira.Attachment{
		{ID: "87891", Filename: "a.png"},
		{ID: "87892", Filename: "a.png"},
	})

	_, err := findAttachment(iss, "a.png")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestResolveDest(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name     string
		filename string
		output   string
		want     string
	}{
		{name: "defaults to filename in cwd", filename: "design.png", output: "", want: "design.png"},
		{name: "existing directory joins filename", filename: "design.png", output: dir, want: filepath.Join(dir, "design.png")},
		{name: "trailing slash joins filename", filename: "design.png", output: dir + string(os.PathSeparator), want: filepath.Join(dir, "design.png")},
		{name: "file path is used as is", filename: "design.png", output: filepath.Join(dir, "renamed.png"), want: filepath.Join(dir, "renamed.png")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveDest(tc.filename, tc.output)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
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
