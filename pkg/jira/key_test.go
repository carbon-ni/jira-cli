package jira

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIssueKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		project  string
		input    string
		expected string
	}{
		{
			name:     "full key on same project",
			project:  "ANK",
			input:    "ANK-11",
			expected: "ANK-11",
		},
		{
			name:     "full key on different project",
			project:  "POK",
			input:    "ANK-11",
			expected: "ANK-11",
		},
		{
			name:     "key number only",
			project:  "ANK",
			input:    "11",
			expected: "ANK-11",
		},
		{
			name:     "text only key",
			project:  "POK",
			input:    "ANK",
			expected: "ANK",
		},
		{
			name:     "invalid key format",
			project:  "POK",
			input:    "ANK-",
			expected: "ANK-",
		},
		{
			name:     "empty project and numeric key",
			project:  "",
			input:    "11",
			expected: "11",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, GetIssueKey(tc.project, tc.input))
		})
	}
}
