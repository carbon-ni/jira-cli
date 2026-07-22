package jira

import "strings"

// NormalizeJiraError normalizes error messages received from Jira.
func NormalizeJiraError(msg string) string {
	msg = strings.TrimSpace(strings.Replace(msg, "Error:\n", "", 1))
	msg = strings.Replace(msg, "- ", "", 1)
	return msg
}
