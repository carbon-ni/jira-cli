package jira

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSubtaskHandle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		input     []*IssueType
		inputType string
		expected  string
	}{
		{
			name: "should get default issue type handle for sub-task",
			input: []*IssueType{
				{
					ID:      "123",
					Name:    "Story",
					Handle:  "story",
					Subtask: false,
				},
			},
			inputType: "Sub-task",
			expected:  "Sub-task",
		},
		{
			name: "should get valid sub-task handle",
			input: []*IssueType{
				{
					ID:      "123",
					Name:    "Story",
					Handle:  "story",
					Subtask: false,
				},
				{
					ID:      "234",
					Name:    "Sub-Task",
					Handle:  "Sub-Task",
					Subtask: true,
				},
			},
			inputType: "Sub-task",
			expected:  "Sub-Task",
		},
		{
			name: "should get sub-task name as handle",
			input: []*IssueType{
				{
					ID:      "123",
					Name:    "Story",
					Handle:  "story",
					Subtask: false,
				},
				{
					ID:      "234",
					Name:    "Subtask",
					Subtask: true,
				},
			},
			inputType: "Sub-task",
			expected:  "Subtask",
		},
		{
			name: "exact matches for a custom sub-task should take precedence",
			input: []*IssueType{
				{
					ID:      "123",
					Name:    "Story",
					Handle:  "story",
					Subtask: false,
				},
				{
					ID:      "234",
					Name:    "Sub-Task",
					Handle:  "Sub-Task",
					Subtask: true,
				},
				{
					ID:      "567",
					Name:    "Custom Sub-Task",
					Handle:  "Custom Sub-Task",
					Subtask: true,
				},
			},
			inputType: "Custom Sub-Task",
			expected:  "Custom Sub-Task",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, GetSubtaskHandle(tc.inputType, tc.input))
		})
	}
}
