package jira

import "testing"

func TestErrUnexpectedResponseIncludesStatusWhenBodyIsEmpty(t *testing.T) {
	err := &ErrUnexpectedResponse{Status: "401 Unauthorized", StatusCode: 401}

	if got := err.Error(); got != "unexpected response: 401 Unauthorized" {
		t.Fatalf("expected status fallback, got %q", got)
	}
}
