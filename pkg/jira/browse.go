package jira

import "fmt"

// BrowseURL constructs a Jira browse URL for a given issue key.
func BrowseURL(server, key string) string {
	return fmt.Sprintf("%s/browse/%s", server, key)
}
