package main

import (
	"os"

	"github.com/dungdm93/drassi/gitea-runner/cmd"
)

func main() {
	command := cmd.NewGiteaRunnerCommand()
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
