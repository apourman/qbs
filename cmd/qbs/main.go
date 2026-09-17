package main

import (
	"fmt"
	"os"

	"github.com/trues/qbs/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "qbs:", err)
		os.Exit(1)
	}
}
