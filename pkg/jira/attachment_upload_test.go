package jira

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadAttachmentMultipartAndAuth(t *testing.T) {
	for _, auth := range []AuthType{AuthTypeCookie, AuthTypeBasic, AuthTypeBearer} {
		t.Run(string(auth), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/rest/api/3/issue/TEST-1/attachments", r.URL.Path)
				assert.Equal(t, "no-check", r.Header.Get("X-Atlassian-Token"))
				switch auth {
				case AuthTypeCookie:
					assert.Equal(t, "session=test-cookie", r.Header.Get("Cookie"))
				case AuthTypeBearer:
					assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				default:
					login, token, ok := r.BasicAuth()
					assert.True(t, ok)
					assert.Equal(t, "test-user", login)
					assert.Equal(t, "test-token", token)
				}
				reader, err := r.MultipartReader()
				if !assert.NoError(t, err) {
					return
				}
				part, err := reader.NextPart()
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, "file", part.FormName())
				assert.Equal(t, "demo.sh", part.FileName())
				data, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "#!/bin/sh\n\x00demo", string(data))
				_, err = reader.NextPart()
				assert.ErrorIs(t, err, io.EOF)
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, `[{"id":"91150","filename":"demo.sh"}]`)
			}))
			defer server.Close()
			client := NewClient(Config{Server: server.URL, AuthType: &auth, Login: "test-user", APIToken: "test-token", Cookies: "session=test-cookie", Debug: true})

			logFile, err := os.CreateTemp(t.TempDir(), "debug-log")
			require.NoError(t, err)
			defer func() { _ = logFile.Close() }()
			stdout, stderr := os.Stdout, os.Stderr
			os.Stdout, os.Stderr = logFile, logFile
			defer func() { os.Stdout, os.Stderr = stdout, stderr }()

			result, err := client.UploadAttachment(context.Background(), "TEST-1", "demo.sh", strings.NewReader("#!/bin/sh\n\x00demo"))

			os.Stdout, os.Stderr = stdout, stderr
			logs, logErr := os.ReadFile(logFile.Name())
			require.NoError(t, logErr)
			assert.Empty(t, string(logs), "debug mode must not log upload headers or bodies")

			require.NoError(t, err)
			require.Len(t, result, 1)
			assert.Equal(t, "91150", result[0].ID)
			assert.Equal(t, "demo.sh", result[0].Filename)
		})
	}
}

func TestUploadAttachmentFailures(t *testing.T) {
	for _, tc := range []struct {
		status     int
		body, want string
	}{
		{401, "secret-response", "authentication failed"},
		{403, "secret-response", "Create Attachments permission"},
		{413, "secret-response", "attachment size limit"},
		{500, "secret-response", "HTTP 500"},
		{302, "secret-response", "HTTP 302"},
		{200, "not json secret-response", "invalid attachment response"},
		{200, "[]", "no attachment"},
	} {
		t.Run(fmt.Sprint(tc.status, tc.body), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				w.Header().Set("Location", "/unexpected")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			client := NewClient(Config{Server: server.URL, Debug: true})

			_, err := client.UploadAttachment(context.Background(), "TEST-1", "demo.sh", strings.NewReader("file-secret"))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
			assert.NotContains(t, err.Error(), "secret")
		})
	}
}

func TestUploadAttachmentCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := NewClient(Config{Server: "http://127.0.0.1:1"})
	_, err := client.UploadAttachment(ctx, "TEST-1", "demo.sh", strings.NewReader("data"))
	require.Error(t, err)
}
