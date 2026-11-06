package upload

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// NewCmdUpload uploads a local file as an issue attachment.
func NewCmdUpload() *cobra.Command {
	return &cobra.Command{
		Use:   "upload ISSUE-KEY FILE",
		Short: "Upload a local file as an issue attachment",
		Long: `Upload one regular local file as an issue attachment using configured Jira authentication.

Prerequisites: run 'jira init' to configure Jira and authentication. Attachments
must be enabled, and your account needs Create Attachments permission on the issue.
The file must be readable and within your Jira site's attachment size limit.
Uses existing CLI credentials; no upload-specific authentication is needed.`,
		Example: `# Configure Jira once, then upload a local file
$ jira init
$ jira issue upload WPB-29891 ./demo.sh

# Return attachment ID and filename as JSON
$ jira issue upload WPB-29891 ./demo.sh --format json`,
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: WPB-29891\nFILE\tReadable regular local file, eg: ./demo.sh",
		},
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := upload(cmd, args); err != nil {
				format := cmdutil.OutputFormat(cmd)
				if cmdutil.IsStructured(format) {
					os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
						Error: cmdutil.ErrorBody{Code: "attachment-upload-failed", Message: err.Error()},
					}, format, false))
				}
				cmdutil.Failed("%s", err)
			}
		},
	}
}

type uploadResult struct {
	Issue        string `json:"issue" toon:"issue"`
	AttachmentID string `json:"attachmentId" toon:"attachmentId"`
	Filename     string `json:"filename" toon:"filename"`
}

func upload(cmd *cobra.Command, args []string) error {
	// Check before opening so directories and special files (such as FIFOs) do
	// not trigger a request or block waiting for input. Recheck the opened file.
	info, err := os.Stat(args[1])
	if err != nil {
		return fmt.Errorf("cannot access upload file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("upload file must be a regular file: %s", args[1])
	}
	file, err := os.Open(args[1])
	if err != nil {
		return fmt.Errorf("cannot open upload file: %w", err)
	}
	defer func() { _ = file.Close() }()
	info, err = file.Stat()
	if err != nil {
		return fmt.Errorf("cannot inspect upload file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("upload file must be a regular file: %s", args[1])
	}

	debug, _ := cmd.Flags().GetBool("debug")
	key := jira.GetIssueKey(viper.GetString("project.key"), args[0])
	client := cmdutil.NewJiraClient(debug)
	attachments, err := api.ProxyUploadAttachment(cmd.Context(), client, key, filepath.Base(file.Name()), file)
	if err != nil {
		return err
	}
	att := attachments[0]
	result := uploadResult{Issue: key, AttachmentID: att.ID, Filename: att.Filename}
	format := cmdutil.OutputFormat(cmd)
	if cmdutil.IsStructured(format) {
		return cmdutil.PrintStructured(result, format)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Uploaded %s (attachment ID: %s) to %s\n", att.Filename, att.ID, key)
	return err
}
