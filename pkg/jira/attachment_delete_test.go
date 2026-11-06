package jira

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteAttachment(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   string
	}{
		{204, ""}, {401, "jira init"}, {403, "Delete Own Attachments"}, {404, "not found"}, {500, "HTTP 500"}, {302, "HTTP 302"},
	} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Equal(t, "/rest/api/3/attachment/91151", r.URL.Path)
				assert.Equal(t, "session=fake", r.Header.Get("Cookie"))
				w.Header().Set("Location", "/wrong")
				w.WriteHeader(tc.status)
				if tc.status != 204 {
					_, _ = io.WriteString(w, "secret-response")
				}
			}))
			defer server.Close()
			auth := AuthTypeCookie
			client := NewClient(Config{Server: server.URL, AuthType: &auth, Cookies: "session=fake", Debug: true})

			output, err := os.CreateTemp(t.TempDir(), "debug-log")
			require.NoError(t, err)
			defer func() { _ = output.Close() }()
			stdout, stderr := os.Stdout, os.Stderr
			os.Stdout, os.Stderr = output, output
			defer func() { os.Stdout, os.Stderr = stdout, stderr }()

			err = client.DeleteAttachment(context.Background(), "91151")

			os.Stdout, os.Stderr = stdout, stderr
			logs, logErr := os.ReadFile(output.Name())
			require.NoError(t, logErr)
			assert.Empty(t, string(logs), "debug must not log credentials or response bodies")

			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
			assert.NotContains(t, err.Error(), "secret")
		})
	}
}

func TestDeleteAttachmentRejectsInvalidID(t *testing.T) {
	client := NewClient(Config{Server: "http://127.0.0.1:1"})
	for _, id := range []string{"", "../issue/TEST-1", "0", "-1", "filename.txt"} {
		err := client.DeleteAttachment(context.Background(), id)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "attachment ID")
	}
}
