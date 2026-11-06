package jira

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// DeleteAttachment removes only the attachment with the given ID, not its issue.
func (c *Client) DeleteAttachment(ctx context.Context, id string) error {
	return c.deleteAttachment(ctx, baseURLv3, id)
}

// DeleteAttachmentV2 removes an attachment using the Server API.
func (c *Client) DeleteAttachmentV2(ctx context.Context, id string) error {
	return c.deleteAttachment(ctx, baseURLv2, id)
}

func (c *Client) deleteAttachment(ctx context.Context, baseURL, id string) error {
	value, err := strconv.ParseUint(id, 10, 64)
	if err != nil || value == 0 {
		return fmt.Errorf("attachment ID must be a positive decimal number")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.server+baseURL+"/attachment/"+id, nil)
	if err != nil {
		return fmt.Errorf("create attachment deletion request: %w", err)
	}
	c.applyAuth(req, http.MethodDelete)
	// Do not dump credentials in debug mode or follow destructive redirects.
	client := &http.Client{
		Transport: c.transport, Timeout: c.timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("attachment deletion request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("attachment deletion authentication failed (HTTP 401); run 'jira init' to configure authentication")
	case http.StatusForbidden:
		return fmt.Errorf("attachment deletion forbidden (HTTP 403); check Delete Own Attachments or Delete All Attachments permission")
	case http.StatusNotFound:
		return fmt.Errorf("attachment not found (HTTP 404); verify the ID with 'jira issue view ISSUE-KEY'")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("attachment deletion failed (HTTP %d)", resp.StatusCode)
	}
	return nil
}
