/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"github.com/spf13/cobra"
)

func NewGiteaRunnerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitea-runner",
		Short: "Gitea runner",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		NewRegisterCommand(),
		NewLaunchCommand(),
	)

	return cmd
}
