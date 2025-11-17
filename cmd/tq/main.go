package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ssccio/tq/pkg/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := cli.Execute(version, commit, date); err != nil {
		if errors.Is(err, cli.ErrExitWithStatus) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
