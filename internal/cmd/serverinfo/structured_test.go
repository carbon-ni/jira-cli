package serverinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestNewServerInfoOutput(t *testing.T) {
	info := &jira.ServerInfo{
		Version:        "10.5.0",
		VersionNumbers: []int{10, 5, 0},
		DeploymentType: "Cloud",
		BuildNumber:    10500,
	}
	info.DefaultLocale.Locale = "en_US"

	encoded, err := cmdutil.MarshalStructured(newServerInfoOutput(info), cmdutil.FormatTOON)
	require.NoError(t, err)

	want := "version: 10.5.0\n" +
		"versionNumbers[3]: 10,5,0\n" +
		"deploymentType: Cloud\n" +
		"buildNumber: 10500\n" +
		"locale: en_US\n" +
		"hint: jira project list"
	assert.Equal(t, want, string(encoded))
}
