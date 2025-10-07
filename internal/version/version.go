package version

import "fmt"

// These variables are set at build time using ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// String returns a formatted version string including version, git commit, and build date
func String() string {
	return fmt.Sprintf("%s (commit: %s, date: %s)", Version, GitCommit, BuildDate)
}