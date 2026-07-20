package serverinfo

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

type serverInfoOutput struct {
	Version        string `json:"version" toon:"version"`
	VersionNumbers []int  `json:"versionNumbers" toon:"versionNumbers"`
	DeploymentType string `json:"deploymentType" toon:"deploymentType"`
	BuildNumber    int    `json:"buildNumber" toon:"buildNumber"`
	Locale         string `json:"locale" toon:"locale"`
	Hint           string `json:"hint" toon:"hint"`
}

func newServerInfoOutput(info *jira.ServerInfo) serverInfoOutput {
	return serverInfoOutput{
		Version:        info.Version,
		VersionNumbers: info.VersionNumbers,
		DeploymentType: info.DeploymentType,
		BuildNumber:    info.BuildNumber,
		Locale:         info.DefaultLocale.Locale,
		Hint:           "jira project list",
	}
}

func renderStructured(info *jira.ServerInfo, format string) {
	if err := cmdutil.PrintStructured(newServerInfoOutput(info), format); err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "server-info-render-failed",
				Message: fmt.Sprintf("Could not encode server information: %s", err),
			},
		}, format, false))
	}
}
