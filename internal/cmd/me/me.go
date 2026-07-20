package me

import (
	"fmt"
	"os"

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

func me(cmd *cobra.Command, _ []string) {
	format := cmdutil.OutputFormat(cmd)
	if cmdutil.IsStructured(format) {
		meStructured(format)
		return
	}

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

// profile is the format-independent output shape for `jira me`. The explicit
// json/toon tags keep compatibility output and agent output semantically equal.
type profile struct {
	Login     string `json:"name" toon:"name"`
	AccountID string `json:"accountId" toon:"accountId"`
	Name      string `json:"displayName" toon:"displayName"`
	Email     string `json:"emailAddress" toon:"emailAddress"`
	Timezone  string `json:"timeZone" toon:"timeZone"`
}

// meStructured emits the full jira.Me profile as deterministic structured
// output (TOON by default, JSON for compatibility). It always hits the API so
// agents receive the canonical account (id, email, timezone) rather than the
// configured-login shortcut used by the human-oriented path.
func meStructured(format string) {
	user, err := api.Client(jira.Config{}).Me()
	if err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:     "me-fetch-failed",
				Message:  fmt.Sprintf("Could not fetch the current user: %s", err),
				Recovery: "Check authentication and server reachability, then retry.",
			},
		}, format, false))
	}

	out := profile{
		Login:     user.Login,
		AccountID: user.AccountID,
		Name:      user.Name,
		Email:     user.Email,
		Timezone:  user.Timezone,
	}
	if err := cmdutil.PrintStructured(out, format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "me-render-failed",
				Message: fmt.Sprintf("Could not encode the user: %s", err),
			},
		}, format, false))
	}
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
