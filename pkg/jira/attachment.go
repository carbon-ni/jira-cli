package jira

import (
	"context"
	"net/http"
)

// GetAttachmentContent streams a file attachment from its absolute content
// URL. Jira answers with a redirect to the actual file, which may live on a
// different host (a CDN), so redirects are followed while re-applying the
// configured auth headers — Go's default client strips sensitive headers on
// cross-host redirects.
func (c *Client) GetAttachmentContent(ctx context.Context, contentURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, contentURL, nil)
	if err != nil {
		return nil, err
	}

	c.applyAuth(req, http.MethodGet)

	httpClient := &http.Client{
		Transport: c.transport,
		CheckRedirect: func(next *http.Request, via []*http.Request) error {
			if len(via) == 0 {
				return nil
			}
			for _, key := range []string{http.CanonicalHeaderKey("Authorization"), http.CanonicalHeaderKey("Cookie")} {
				if vv, ok := via[0].Header[key]; ok {
					next.Header[key] = append([]string(nil), vv...)
				}
			}
			return nil
		},
	}

	return httpClient.Do(req.WithContext(ctx))
}
