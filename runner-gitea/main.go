package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"drassi.run/gitea-runner/cmd"
)

func execute() error {
	// SIGINT (Signal interrupt)
	// A signal that requests a process to terminate gracefully, usually initiated by the user.
	// For example, pressing Ctrl+C in Bash, but on some systems, the "delete" character or "break" key can be used.
	//
	// SIGTERM (Signal terminate)
	// A signal that requests a process to terminate gracefully, which can be sent by other processes or the system itself.
	// SIGTERM is the default signal when using the kill command. It allows the process to send information to its parent
	// and child processes, and if the program has a handler for SIGTERM, it can clean up and terminate in an orderly fashion.
	//
	// SIGQUIT (Signal quit)
	// Like SIGTERM, SIGQUIT signal is meant to terminate the process.
	// However, SIGQUIT also generates a core dump before exiting.
	//
	// SIGKILL (Signal kill)
	// A signal that forcefully terminates a process immediately without allowing it to perform any cleanup operations.
	// In contrast to SIGTERM and SIGINT, this signal cannot be caught or ignored, and the receiving process
	// cannot perform any clean-up upon receiving this signal.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	command := cmd.NewGiteaRunnerCommand()
	return command.ExecuteContext(ctx)
}

func main() {
	if err := execute(); err != nil {
		// Exit causes the current program terminates immediately; deferred functions are not run.
		os.Exit(1)
	}
}
