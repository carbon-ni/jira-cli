package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	build "github.com/ankitpokhrel/jira-cli/internal/version"
)

func TestNewVersionOutput(t *testing.T) {
	out, err := newVersionOutput()
	require.NoError(t, err)

	assert.Equal(t, build.Version, out.Version)
	assert.Equal(t, build.GitCommit, out.GitCommit)
	assert.Equal(t, build.GoVersion, out.GoVersion)
	assert.Equal(t, build.Compiler, out.Compiler)
	assert.Equal(t, build.Platform, out.Platform)

	encoded, err := cmdutil.MarshalStructured(out, cmdutil.FormatTOON)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), "version: ")
	assert.Contains(t, string(encoded), "hint: jira serverinfo")
}

func TestNewVersionOutputRejectsInvalidSourceDateEpoch(t *testing.T) {
	old := build.SourceDateEpoch
	build.SourceDateEpoch = "invalid"
	t.Cleanup(func() { build.SourceDateEpoch = old })

	_, err := newVersionOutput()
	require.Error(t, err)
}
