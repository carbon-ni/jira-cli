package jira

import (
	"fmt"
	"strconv"
	"strings"
)

// GetIssueKey constructs an actual issue key from a project and key.
// If key is numeric, it prefixes with the project (e.g. "PROJ-123").
// Otherwise, returns the key uppercased.
func GetIssueKey(project, key string) string {
	if project == "" {
		return key
	}
	if _, err := strconv.Atoi(key); err != nil {
		return strings.ToUpper(key)
	}
	return fmt.Sprintf("%s-%s", project, key)
}
