package attachment

import "github.com/spf13/cobra"

// NewCmdAttachment groups attachment-specific operations, separate from issue deletion.
func NewCmdAttachment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Manage issue attachments without deleting the issue",
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newDeleteCommand())
	return cmd
}
