package main

import "github.com/azemoning/omni-5e/internal/cli"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersionInfo(version, commit, date)
	cli.Execute()
}
