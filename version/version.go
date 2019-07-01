package version

import "time"

var (
	gitVersion   = "dev"
	gitTreeState = "unknown"
	gitCommit    = "?"

	buildDate = "1970-01-01T00:00:00Z"
)

// GitVersion will return the version of the current app binary.
func GitVersion() string { return gitVersion }

// GitCommit will return the commit hash of the current app binary.
func GitCommit() string { return gitCommit }

// GitTreeState will indicate the state of the working directory when the app was built.
func GitTreeState() string { return gitTreeState }

// BuildDate returns a the timestamp of when the current binary was built.
func BuildDate() time.Time {
	t, _ := time.Parse(time.RFC3339, buildDate)
	return t
}
