package me

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// NewCmdMe is a me command.
func NewCmdMe() *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Displays configured jira user",
		Long:  "Displays configured jira user.",
		Run:   me,
	}
}

func me(*cobra.Command, []string) {
	login := viper.GetString("login")
	if login != "" {
		fmt.Println(login)
		return
	}

	currentUser, err := api.Client(jira.Config{}).Me()
	if err != nil {
		cmdutil.Failed("Error: %s", err)
		return
	}

	fmt.Println(currentUserIdentifier(currentUser))
}

func currentUserIdentifier(currentUser *jira.Me) string {
	if currentUser.Login != "" {
		return currentUser.Login
	}
	if currentUser.Email != "" {
		return currentUser.Email
	}
	if currentUser.Name != "" {
		return currentUser.Name
	}
	return currentUser.AccountID
}
