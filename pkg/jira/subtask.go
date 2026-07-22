package jira

import "strings"

// GetSubtaskHandle returns the subtask handle for a given issue type.
// The returned value can be either a handle or name based on the Jira version.
func GetSubtaskHandle(issueType string, issueTypes []*IssueType) string {
	get := func(it *IssueType) string {
		if it.Handle != "" {
			return it.Handle
		}
		return it.Name
	}

	var fallback string

	for _, it := range issueTypes {
		if it.Subtask {
			if strings.EqualFold(issueType, it.Name) {
				return get(it)
			}
			if fallback == "" {
				fallback = get(it)
			}
		}
	}

	if strings.EqualFold(issueType, IssueTypeSubTask) && fallback == "" {
		fallback = IssueTypeSubTask
	}

	return fallback
}
