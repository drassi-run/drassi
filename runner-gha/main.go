package main

import (
	"os"

	"drassi.run/gha-runner/cmd"
)

func main() {
	command := cmd.NewGHARunnerCommand()
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
