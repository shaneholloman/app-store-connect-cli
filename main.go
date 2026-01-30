package main

import (
	"fmt"
	"os"

	"github.com/rudrankriyam/App-Store-Connect-CLI/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	versionInfo := fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date)
	os.Exit(cmd.Run(os.Args[1:], versionInfo))
}
