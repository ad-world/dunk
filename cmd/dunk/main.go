package main

import (
	"fmt"
	"os"

	"dunk/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "dunk:", err)
		os.Exit(1)
	}
}
