package version

import (
	"fmt"
	"os"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	build "github.com/ankitpokhrel/jira-cli/internal/version"
)

type versionOutput struct {
	Version    string `json:"version" toon:"version"`
	GitCommit  string `json:"gitCommit" toon:"gitCommit"`
	CommitDate string `json:"commitDate" toon:"commitDate"`
	GoVersion  string `json:"goVersion" toon:"goVersion"`
	Compiler   string `json:"compiler" toon:"compiler"`
	Platform   string `json:"platform" toon:"platform"`
	Hint       string `json:"hint" toon:"hint"`
}

func newVersionOutput() (versionOutput, error) {
	commitDate, err := build.CommitDate()
	if err != nil {
		return versionOutput{}, err
	}
	return versionOutput{
		Version:    build.Version,
		GitCommit:  build.GitCommit,
		CommitDate: commitDate,
		GoVersion:  build.GoVersion,
		Compiler:   build.Compiler,
		Platform:   build.Platform,
		Hint:       "jira serverinfo",
	}, nil
}

func renderStructured(format string) {
	out, err := newVersionOutput()
	if err == nil {
		err = cmdutil.PrintStructured(out, format)
	}
	if err != nil {
		os.Exit(cmdutil.PrintStructuredError(cmdutil.ErrorEnvelope{
			Error: cmdutil.ErrorBody{
				Code:    "version-render-failed",
				Message: fmt.Sprintf("Could not encode version information: %s", err),
			},
		}, format, false))
	}
}
