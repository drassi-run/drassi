package main

import (
	"os"

	"github.com/dungdm93/drassi/gha-runner/cmd"
)

func main() {
	command := cmd.NewGHARunnerCommand()
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
