/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"drassi.run/gha-runner/cmd/launch"
	"drassi.run/gha-runner/cmd/migrate"
	"drassi.run/gha-runner/cmd/register"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gha-runner",
		Short: "GitHub Actions runner rewrite in Go",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		register.New(),
		launch.New(),
		migrate.New(),
	)

	return cmd
}
