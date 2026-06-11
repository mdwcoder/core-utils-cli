package main

import (
	"os"

	"github.com/mdwcoder/core-utils-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
