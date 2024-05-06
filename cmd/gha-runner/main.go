package main

import (
	"os"

	"github.com/dungdm93/drasi/cmd/gha-runner/command"
)

func main() {
	cmd := command.NewGHARunnerCommand()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
