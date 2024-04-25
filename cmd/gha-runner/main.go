package main

import (
	"github.com/dungdm93/drasi/cmd/gha-runner/root"
	"os"
)

func main() {
	cmd := root.NewGHARunnerCommand()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
