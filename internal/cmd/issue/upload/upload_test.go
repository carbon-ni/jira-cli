package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestUploadCommand(t *testing.T) {
	for _, tc := range []struct{ installation, format string }{
		{jira.InstallationTypeCloud, "json"},
		{jira.InstallationTypeLocal, "json"},
		{jira.InstallationTypeCloud, "toon"},
		{jira.InstallationTypeCloud, "auto"},
	} {
		installation := tc.installation
		t.Run(installation+"/"+tc.format, func(t *testing.T) {
			viper.Reset()
			api.ResetClient()
			t.Cleanup(func() { viper.Reset(); api.ResetClient() })
			version := "3"
			if installation == jira.InstallationTypeLocal {
				version = "2"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/rest/api/"+version+"/issue/TEST-1/attachments", r.URL.Path)
				assert.Equal(t, "session=fake", r.Header.Get("Cookie"))
				reader, err := r.MultipartReader()
				if !assert.NoError(t, err) {
					return
				}
				part, err := reader.NextPart()
				if !assert.NoError(t, err) {
					return
				}
				data, err := io.ReadAll(part)
				assert.NoError(t, err)
				assert.Equal(t, "example-data", string(data))
				_, _ = io.WriteString(w, `[{"id":"91150","filename":"demo.sh"}]`)
			}))
			defer server.Close()
			viper.Set("server", server.URL)
			viper.Set("installation", installation)
			viper.Set("auth_type", "cookie")
			viper.Set("cookies", "session=fake")
			viper.Set("project.key", "TEST")
			path := filepath.Join(t.TempDir(), "demo.sh")
			require.NoError(t, os.WriteFile(path, []byte("example-data"), 0600))
			cmd := NewCmdUpload()
			cmd.SetContext(context.Background())
			cmd.Flags().String("format", tc.format, "")
			var output bytes.Buffer
			previous := os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = w
			defer func() { os.Stdout = previous; _ = r.Close(); _ = w.Close() }()

			cmd.SetArgs([]string{"1", path})
			err = cmd.Execute()

			require.NoError(t, err)
			require.NoError(t, w.Close())
			os.Stdout = previous
			_, err = io.Copy(&output, r)
			require.NoError(t, err)
			if tc.format == "json" {
				var result map[string]string
				require.NoError(t, json.Unmarshal(output.Bytes(), &result))
				assert.Equal(t, "TEST-1", result["issue"])
				assert.Equal(t, "91150", result["attachmentId"])
				assert.Equal(t, "demo.sh", result["filename"])
				return
			}
			if tc.format == "toon" {
				assert.Contains(t, output.String(), "issue: TEST-1")
				assert.Contains(t, output.String(), `attachmentId: "91150"`)
				assert.Contains(t, output.String(), "filename: demo.sh")
				return
			}
			assert.Contains(t, output.String(), "Uploaded demo.sh (attachment ID: 91150) to TEST-1")
		})
	}
}

func TestUploadRejectsInvalidFilesBeforeRequest(t *testing.T) {
	dir := t.TempDir()
	for _, path := range []string{filepath.Join(dir, "missing"), dir} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			err := upload(NewCmdUpload(), []string{"TEST-1", path})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "file")
		})
	}
}

func TestUploadRequiresTwoArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"TEST-1"}, {"TEST-1", "a", "b"}} {
		assert.Error(t, NewCmdUpload().Args(NewCmdUpload(), args))
	}
}
