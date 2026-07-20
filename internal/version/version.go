package version

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

// Build information is populated at build-time.
var (
	Version         = "v1.0.0-dev"
	GitCommit       = ""
	SourceDateEpoch = "-1"
	GoVersion       = runtime.Version()
	Compiler        = runtime.Compiler
	Platform        = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// CommitDate returns the build commit date in RFC3339 form, or an empty string
// when the source date epoch was not set.
func CommitDate() (string, error) {
	i, err := strconv.ParseInt(SourceDateEpoch, 10, 64) //nolint:gomnd
	if err != nil {
		return "", err
	}
	if i < 0 {
		return "", nil
	}
	// https://pkg.go.dev/time#Time.Format
	return time.Unix(i, 0).UTC().Format("2006-01-02T15:04:05-07:00"), nil
}

// Info returns version and build information.
func Info() string {
	commitDate, err := CommitDate()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(
		"(Version=%q, GitCommit=%q, CommitDate=%q, GoVersion=%q, Compiler=%q, Platform=%q)",
		Version, GitCommit, commitDate, GoVersion, Compiler, Platform,
	)
}
