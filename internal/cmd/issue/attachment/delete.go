package attachment

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete ISSUE-KEY ATTACHMENT-ID",
		Short: "Delete one attachment, not the issue",
		Long: `Permanently delete one attachment by ID. The issue and other attachments remain unchanged.
The command first verifies that the attachment belongs to the specified issue.
Deletion is immediate, with no confirmation prompt.

Prerequisites: configure authentication with 'jira init'. Your account needs
Delete Own Attachments (for your files) or Delete All Attachments permission.
Use 'jira issue view ISSUE-KEY' to find attachment IDs; filenames are not accepted.`,
		Example: `# Inspect attachments, then delete only the selected file
$ jira issue view WPB-29891
$ jira issue attachment delete WPB-29891 91151

# Machine-readable result
$ jira issue attachment delete WPB-29891 91151 --format json`,
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue containing the attachment\nATTACHMENT-ID\tExact attachment ID, not a filename",
		},
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := deleteAttachment(cmd, args); err != nil {
				format := cmdutil.OutputFormat(cmd)
				if cmdutil.IsStructured(format) {
					os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
						Error: cmdutil.ErrorBody{Code: "attachment-delete-failed", Message: err.Error()},
					}, format, false))
				}
				cmdutil.Failed("%s", err)
			}
		},
	}
}

type deleteResult struct {
	Issue        string `json:"issue" toon:"issue"`
	AttachmentID string `json:"attachmentId" toon:"attachmentId"`
	Filename     string `json:"filename" toon:"filename"`
	Deleted      bool   `json:"deleted" toon:"deleted"`
}

func deleteAttachment(cmd *cobra.Command, args []string) error {
	key := jira.GetIssueKey(viper.GetString("project.key"), args[0])
	// Keep debug dumps off for the prerequisite issue fetch too: it uses the
	// general request path, which otherwise dumps authentication headers.
	client := cmdutil.NewJiraClient(false)
	issue, err := api.ProxyGetIssue(client, key)
	if err != nil {
		return fmt.Errorf("could not verify attachment on issue %s; run 'jira issue view %s' to check access", key, key)
	}
	var selected *jira.Attachment
	for i := range issue.Fields.Attachments {
		if issue.Fields.Attachments[i].ID == args[1] {
			selected = &issue.Fields.Attachments[i]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("attachment ID %q not found on issue %s; run 'jira issue view %s' to find attachment IDs", args[1], key, key)
	}
	if err := api.ProxyDeleteAttachment(cmd.Context(), client, selected.ID); err != nil {
		return err
	}
	result := deleteResult{Issue: key, AttachmentID: selected.ID, Filename: selected.Filename, Deleted: true}
	format := cmdutil.OutputFormat(cmd)
	if cmdutil.IsStructured(format) {
		return cmdutil.PrintStructured(result, format)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted attachment %s (%s) from %s; issue unchanged\n", selected.ID, selected.Filename, key)
	return err
}
