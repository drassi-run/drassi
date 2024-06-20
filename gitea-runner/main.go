package main

import (
	"os"

	"drassi.run/gitea-runner/cmd"
)

func main() {
	command := cmd.NewGiteaRunnerCommand()
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
