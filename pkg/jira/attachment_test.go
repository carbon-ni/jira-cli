package jira

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueFieldsUnmarshalAttachments(t *testing.T) {
	const raw = `{
		"key": "TEST-1",
		"fields": {
			"summary": "Has a design",
			"attachment": [
				{
					"id": "87891",
					"filename": "design.png",
					"mimeType": "image/png",
					"size": 35748,
					"content": "https://jira.example.com/rest/api/3/attachment/content/87891",
					"created": "2026-01-02T10:00:00.000+0000",
					"author": {"displayName": "Ada", "name": "ada"}
				},
				{
					"id": "87892",
					"filename": "notes.md",
					"mimeType": "text/markdown",
					"size": 2048,
					"content": "https://jira.example.com/rest/api/3/attachment/content/87892"
				}
			]
		}
	}`

	var iss Issue
	require.NoError(t, json.Unmarshal([]byte(raw), &iss))

	require.Len(t, iss.Fields.Attachments, 2)

	a := iss.Fields.Attachments[0]
	assert.Equal(t, "87891", a.ID)
	assert.Equal(t, "design.png", a.Filename)
	assert.Equal(t, "image/png", a.MimeType)
	assert.Equal(t, int64(35748), a.Size)
	assert.Equal(t, "https://jira.example.com/rest/api/3/attachment/content/87891", a.Content)
	assert.Equal(t, "Ada", a.Author.DisplayName)

	assert.Equal(t, "notes.md", iss.Fields.Attachments[1].Filename)
	assert.Equal(t, "https://jira.example.com/rest/api/3/attachment/content/87892", iss.Fields.Attachments[1].Content)
}

func TestIssueWithoutAttachmentsHasEmptySlice(t *testing.T) {
	const raw = `{"key": "TEST-1", "fields": {"summary": "No files"}}`

	var iss Issue
	require.NoError(t, json.Unmarshal([]byte(raw), &iss))

	assert.Empty(t, iss.Fields.Attachments)
}

func TestGetAttachmentContentDownloadsFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PNGDATA"))
	}))
	defer server.Close()

	at := AuthTypeCookie
	client := NewClient(Config{
		Server:   "https://jira.example.com",
		Cookies:  "cloud.session=abc",
		AuthType: &at,
	}, WithTimeout(3*time.Second))

	resp, err := client.GetAttachmentContent(context.Background(), server.URL+"/content")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "PNGDATA", string(body))
}

func TestGetAttachmentContentFollowsRedirectAndPreservesCookie(t *testing.T) {
	var gotCookie string

	// The file lives on a different host than the API server (Jira CDN case).
	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		_, _ = w.Write([]byte("FILEDATA"))
	}))
	defer fileServer.Close()

	contentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fileServer.URL, http.StatusFound)
	}))
	defer contentServer.Close()

	at := AuthTypeCookie
	client := NewClient(Config{
		Server:   "https://jira.example.com",
		Cookies:  "cloud.session=abc",
		AuthType: &at,
	}, WithTimeout(3*time.Second))

	resp, err := client.GetAttachmentContent(context.Background(), contentServer.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "cloud.session=abc", gotCookie, "cookie must survive cross-host redirect")
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "FILEDATA", string(body))
}

func TestGetAttachmentContentReturnsNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	resp, err := client.GetAttachmentContent(context.Background(), server.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
