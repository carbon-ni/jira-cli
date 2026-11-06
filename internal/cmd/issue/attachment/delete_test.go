package attachment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestDeleteOnlyMatchingIssueAttachment(t *testing.T) {
	for _, tc := range []struct {
		name, installation, issueBody string
		getStatus, deleteStatus       int
		want                          string
		deletes                       int
	}{
		{"cloud", jira.InstallationTypeCloud, `{"key":"TEST-1","fields":{"attachment":[{"id":"91151","filename":"test.txt"}]}}`, 200, 204, "", 1},
		{"toon", jira.InstallationTypeCloud, `{"key":"TEST-1","fields":{"attachment":[{"id":"91151","filename":"test.txt"}]}}`, 200, 204, "", 1},
		{"auto", jira.InstallationTypeCloud, `{"key":"TEST-1","fields":{"attachment":[{"id":"91151","filename":"test.txt"}]}}`, 200, 204, "", 1},
		{"server", jira.InstallationTypeLocal, `{"key":"TEST-1","fields":{"attachment":[{"id":"91151","filename":"test.txt"}]}}`, 200, 204, "", 1},
		{"wrong attachment", jira.InstallationTypeCloud, `{"key":"TEST-1","fields":{"attachment":[{"id":"123","filename":"test.txt"}]}}`, 200, 204, "not found on issue", 0},
		{"fetch failed", jira.InstallationTypeCloud, `{}`, 404, 204, "", 0},
		{"forbidden", jira.InstallationTypeCloud, `{"key":"TEST-1","fields":{"attachment":[{"id":"91151","filename":"test.txt"}]}}`, 200, 403, "Delete Own Attachments", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			api.ResetClient()
			t.Cleanup(func() { viper.Reset(); api.ResetClient() })
			version := "3"
			if tc.installation == jira.InstallationTypeLocal {
				version = "2"
			}
			deletes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "session=fake", r.Header.Get("Cookie"))
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/rest/api/"+version+"/issue/TEST-1":
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tc.getStatus)
					_, _ = io.WriteString(w, tc.issueBody)
				case r.Method == http.MethodDelete && r.URL.Path == "/rest/api/"+version+"/attachment/91151":
					deletes++
					w.WriteHeader(tc.deleteStatus)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			viper.Set("server", server.URL)
			viper.Set("installation", tc.installation)
			viper.Set("auth_type", "cookie")
			viper.Set("cookies", "session=fake")
			viper.Set("project.key", "TEST")
			cmd := newDeleteCommand()
			cmd.SetContext(context.Background())
			format := "json"
			if tc.name == "toon" || tc.name == "auto" {
				format = tc.name
			}
			cmd.Flags().String("format", format, "")
			cmd.Flags().Bool("debug", true, "")
			output, err := os.CreateTemp(t.TempDir(), "result")
			require.NoError(t, err)
			defer func() { _ = output.Close() }()
			previous := os.Stdout
			os.Stdout = output
			defer func() { os.Stdout = previous }()

			if tc.getStatus == 200 && tc.want == "" {
				cmd.SetArgs([]string{"1", "91151"})
				err = cmd.Execute()
			} else {
				err = deleteAttachment(cmd, []string{"1", "91151"})
			}

			os.Stdout = previous
			assert.Equal(t, tc.deletes, deletes)
			if tc.want != "" || tc.getStatus != 200 {
				require.Error(t, err)
				if tc.want != "" {
					assert.Contains(t, err.Error(), tc.want)
				}
				return
			}
			require.NoError(t, err)
			data, err := os.ReadFile(output.Name())
			require.NoError(t, err)
			if format == "toon" {
				assert.Contains(t, string(data), `attachmentId: "91151"`)
				assert.Contains(t, string(data), "deleted: true")
				return
			}
			if format == "auto" {
				assert.Contains(t, string(data), "Deleted attachment 91151 (test.txt) from TEST-1; issue unchanged")
				return
			}
			assert.JSONEq(t, `{"issue":"TEST-1","attachmentId":"91151","filename":"test.txt","deleted":true}`, string(data))
		})
	}
}

func TestAttachmentCommandDiscoveryAndArity(t *testing.T) {
	cmd := NewCmdAttachment()
	var output bytes.Buffer
	cmd.SetOut(&output)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, output.String(), "delete")
	for _, args := range [][]string{nil, {"TEST-1"}, {"TEST-1", "91151", "extra"}} {
		t.Run(fmt.Sprint(args), func(t *testing.T) { assert.Error(t, newDeleteCommand().Args(newDeleteCommand(), args)) })
	}
}
