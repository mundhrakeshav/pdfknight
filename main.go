package main

import "pdfdarkmode/cmd"

// Version info - set via ldflags at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	cmd.SetVersionInfo(Version, BuildTime, GitCommit)
	cmd.Execute()
}
