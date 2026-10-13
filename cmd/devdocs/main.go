package main

import (
	"fmt"
	"github.com/hugojanuario/devdocs/cli/commands"
	"os"
)

func main() {
	if err := commands.New().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
