// hago is a command-line interface for interacting with the Home Assistant REST API.
package main

import "github.com/rmrfslashbin/hago/cmd/hago/cmd"

// Build information set via ldflags.
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Set version info
	cmd.Version = version
	cmd.GitCommit = gitCommit
	cmd.BuildTime = buildTime

	cmd.Execute()
}
