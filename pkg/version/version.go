package version

import (
	"github.com/jaxxstorm/vers"
)

// Version is the version of this tool.
var Version string

// GetVersion returns the version using the same logic as the version command
func GetVersion() string {
	v := Version
	// If we haven't set a version with linker flags, calculate from git
	if v == "" {
		repo, err := vers.OpenRepository(".")
		if err != nil {
			return "unknown"
		}

		opts := vers.Options{
			Repository: repo,
			Commitish:  "HEAD",
		}

		versions, err := vers.Calculate(opts)
		if err != nil {
			return "unknown"
		}
		v = versions.SemVer
	}
	return v
}
