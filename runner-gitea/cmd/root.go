/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"drassi.run/gitea-runner/cmd/initialize"
	"drassi.run/gitea-runner/cmd/launch"
	"drassi.run/gitea-runner/cmd/register"
	"github.com/spf13/cobra"
)

func NewGiteaRunnerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitea-runner",
		Short: "Gitea runner",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		initialize.New(),
		register.New(),
		launch.New(),
	)

	return cmd
}
