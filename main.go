// Command missionctl works with Mission Control projects and time entries.
package main

import (
	"os"

	"github.com/kmarsh/missionctl/internal/cli"
)

// version is set by GoReleaser at build time.
var version = "dev"

func main() {
	os.Exit(cli.Run(version, os.Args[1:]))
}
