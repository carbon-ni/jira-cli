package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Download downloads a file attachment of an issue.`
	examples = `$ jira issue download ISSUE-1 10001
$ jira issue download ISSUE-1 design.png
$ jira issue download ISSUE-1 10001 -o /tmp/design.png
$ jira issue download ISSUE-1 10001 -o /tmp/designs`

	flagOutput = "output"

	configProject = "project.key"

	messageFetchingIssue   = "Fetching issue..."
	messageDownloadingFile = "Downloading attachment..."
)

// downloadResult is the structured envelope emitted on success.
type downloadResult struct {
	Issue        string `json:"issue" toon:"issue"`
	AttachmentID string `json:"attachmentId" toon:"attachmentId"`
	Filename     string `json:"filename" toon:"filename"`
	MimeType     string `json:"mimeType" toon:"mimeType"`
	Size         int64  `json:"size" toon:"size"`
	SavedTo      string `json:"savedTo" toon:"savedTo"`
}

// NewCmdDownload is a download command.
func NewCmdDownload() *cobra.Command {
	cmd := cobra.Command{
		Use:     "download ISSUE-KEY ATTACHMENT",
		Short:   "Download downloads a file attachment of an issue",
		Long:    helpText,
		Example: examples,
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\nATTACHMENT\tAttachment ID or filename, eg: 10001 or design.png",
		},
		Args: cobra.ExactArgs(2),
		Run:  download,
	}

	cmd.Flags().StringP(flagOutput, "o", "", "Output file path or directory (default: ./<filename>)")

	return &cmd
}

func download(cmd *cobra.Command, args []string) {
	format := cmdutil.OutputFormat(cmd)

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	key := jira.GetIssueKey(viper.GetString(configProject), args[0])
	target := args[1]

	client := cmdutil.NewJiraClient(debug)

	iss, err := func() (*jira.Issue, error) {
		s := cmdutil.Info(messageFetchingIssue)
		defer s.Stop()

		return api.ProxyGetIssue(client, key)
	}()
	if err != nil {
		exitError(format, "issue-fetch-failed", err.Error())
	}

	att, err := findAttachment(iss, target)
	if err != nil {
		exitError(format, "attachment-not-found", err.Error())
	}

	output, err := cmd.Flags().GetString(flagOutput)
	cmdutil.ExitIfError(err)

	dest, err := resolveDest(att.Filename, output)
	if err != nil {
		exitError(format, "invalid-output-path", err.Error())
	}

	if err := downloadAttachment(client, att, dest); err != nil {
		exitError(format, "attachment-download-failed", err.Error())
	}

	result := downloadResult{
		Issue:        iss.Key,
		AttachmentID: att.ID,
		Filename:     att.Filename,
		MimeType:     att.MimeType,
		Size:         att.Size,
		SavedTo:      dest,
	}

	if cmdutil.IsStructured(format) {
		if err := cmdutil.PrintStructured(result, format); err != nil {
			os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
				Error: cmdutil.ErrorBody{
					Code:    "attachment-result-render-failed",
					Message: fmt.Sprintf("Could not encode download result: %s", err),
				},
			}, format, false))
		}
		return
	}

	fmt.Printf("Downloaded %s (%s) to %s\n", att.Filename, cmdutil.FormatBytesHuman(att.Size), dest)
}

// exitError reports a failure in the requested format: a structured error
// envelope on stdout for toon/json, a plain stderr message otherwise.
func exitError(format, code, msg string) {
	if cmdutil.IsStructured(format) {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{Code: code, Message: msg},
		}, format, false))
	}
	cmdutil.Failed("%s", msg)
}

// findAttachment resolves the target attachment on an issue by exact ID first,
// then by exact filename. Duplicate filenames are ambiguous and reported.
func findAttachment(iss *jira.Issue, target string) (*jira.Attachment, error) {
	if len(iss.Fields.Attachments) == 0 {
		return nil, fmt.Errorf("issue %s has no attachments", iss.Key)
	}

	for i := range iss.Fields.Attachments {
		if iss.Fields.Attachments[i].ID == target {
			return &iss.Fields.Attachments[i], nil
		}
	}

	var matches []*jira.Attachment
	for i := range iss.Fields.Attachments {
		if iss.Fields.Attachments[i].Filename == target {
			matches = append(matches, &iss.Fields.Attachments[i])
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("attachment %q not found on issue %s (available: %s)",
			target, iss.Key, attachmentNames(iss.Fields.Attachments))
	default:
		return nil, fmt.Errorf("attachment filename %q is ambiguous (%d matches); use the attachment ID instead",
			target, len(matches))
	}
}

func attachmentNames(atts []jira.Attachment) string {
	names := make([]string, 0, len(atts))
	for _, att := range atts {
		names = append(names, fmt.Sprintf("%s (%s)", att.Filename, att.ID))
	}
	return strings.Join(names, ", ")
}

// resolveDest decides where the file is written. An empty -o defaults to
// <filename> in the working directory; -o pointing at an existing directory
// or ending with a path separator appends the filename; otherwise -o is used
// verbatim as the destination file path.
func resolveDest(filename, output string) (string, error) {
	if output == "" {
		return filename, nil
	}

	info, err := os.Stat(output)
	if err == nil {
		if info.IsDir() {
			return filepath.Join(output, filename), nil
		}
		return output, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	if strings.HasSuffix(output, string(os.PathSeparator)) {
		return filepath.Join(output, filename), nil
	}
	return output, nil
}

// downloadAttachment streams the attachment file to dest, creating parent
// directories as needed.
func downloadAttachment(client *jira.Client, att *jira.Attachment, dest string) error {
	s := cmdutil.Info(messageDownloadingFile)
	defer s.Stop()

	resp, err := api.ProxyDownloadAttachment(client, att)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download attachment: %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}
