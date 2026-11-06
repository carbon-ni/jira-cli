package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
)

// UploadAttachment uploads one file using the cloud API without debug body dumps.
func (c *Client) UploadAttachment(ctx context.Context, key, filename string, content io.Reader) ([]Attachment, error) {
	return c.uploadAttachment(ctx, baseURLv3, key, filename, content)
}

// UploadAttachmentV2 uploads one file using the server API.
func (c *Client) UploadAttachmentV2(ctx context.Context, key, filename string, content io.Reader) ([]Attachment, error) {
	return c.uploadAttachment(ctx, baseURLv2, key, filename, content)
}

func (c *Client) uploadAttachment(ctx context.Context, baseURL, key, filename string, content io.Reader) ([]Attachment, error) {
	// Buffer only multipart framing, not file contents. This avoids a producer
	// goroutine that could remain blocked when the server rejects an upload early.
	var framing bytes.Buffer
	writer := multipart.NewWriter(&framing)
	if _, err := writer.CreateFormFile("file", filepath.Base(filename)); err != nil {
		return nil, fmt.Errorf("create attachment multipart header: %w", err)
	}
	prefix := append([]byte(nil), framing.Bytes()...)
	framing.Reset()
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close attachment multipart header: %w", err)
	}
	body := io.MultiReader(bytes.NewReader(prefix), content, bytes.NewReader(framing.Bytes()))
	endpoint := c.server + baseURL + "/issue/" + url.PathEscape(key) + "/attachments"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create attachment upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	c.applyAuth(req, http.MethodPost)
	req.Header.Set("X-Atlassian-Token", "no-check")

	// Reuse configured auth and TLS, but never dump headers, request or response
	// bodies. Do not replay uploaded files or credentials through redirects.
	client := &http.Client{
		Transport:     c.transport,
		Timeout:       c.timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("attachment upload request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("attachment upload authentication failed (HTTP 401); run 'jira init' to configure authentication")
	case http.StatusForbidden:
		return nil, fmt.Errorf("attachment upload forbidden (HTTP 403); check Create Attachments permission and that attachments are enabled")
	case http.StatusRequestEntityTooLarge:
		return nil, fmt.Errorf("attachment exceeds Jira attachment size limit (HTTP 413); use a smaller file or ask an administrator to raise the limit")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("attachment upload failed (HTTP %d)", resp.StatusCode)
	}
	var attachments []Attachment
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("invalid attachment response; upload may have succeeded, check issue before retrying")
	}
	if len(attachments) == 0 {
		return nil, fmt.Errorf("upload returned no attachment; check issue before retrying")
	}
	return attachments, nil
}
