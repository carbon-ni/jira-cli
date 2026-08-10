package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestProxyDownloadAttachmentStreamsContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "cloud.session=abc", r.Header.Get("Cookie"))
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("PNGDATA"))
	}))
	defer server.Close()

	ResetClient()

	at := jira.AuthTypeCookie
	client := jira.NewClient(jira.Config{
		Server:   "https://jira.example.com",
		Cookies:  "cloud.session=abc",
		AuthType: &at,
	}, jira.WithTimeout(3*time.Second))

	att := &jira.Attachment{ID: "87891", Content: server.URL}

	resp, err := ProxyDownloadAttachment(client, att)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "PNGDATA", string(body))
}
