package jira

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeJiraError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Received an error with message",
			input:    "Error:\n- The request reported 404.",
			expected: "The request reported 404.",
		},
		{
			name:     "no error",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, NormalizeJiraError(tc.input))
		})
	}
}
