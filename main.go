package main

import (
	"os"

	"github.com/vibeyang/multitab/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
