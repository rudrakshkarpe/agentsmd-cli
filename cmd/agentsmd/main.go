package main

import (
	"fmt"
	"os"

	"github.com/rudrakshkarpe/agentsmd-cli/cli"
)

func main() {
	if err := cli.New().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
