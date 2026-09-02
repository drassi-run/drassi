/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"drassi.run/gha-runner/pkg/dotnet"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

type options struct {
	Dir       string
	Output    string
	Sandboxer string
}

type migrator struct {
	options
}

func New() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:   "migrate [path]",
		Short: "Migrate actions/runner configuration to gha-runner configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Dir = args[0]
			} else {
				opts.Dir = "."
			}

			m := &migrator{options: opts}
			return m.Run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Output, "output", "o", "config.toml", "Output configuration file path")
	flags.StringVar(&opts.Sandboxer, "sandboxer", "host", "Sandboxer name to use")

	return cmd
}

func (m *migrator) Run(_ context.Context) error {
	dotNetCfg, err := dotnet.LoadConfiguration(m.Dir)
	if err != nil {
		return fmt.Errorf("load actions-runner configuration from %q: %w", m.Dir, err)
	}

	cfg, err := dotNetCfg.ToConfig(m.Sandboxer)
	if err != nil {
		return fmt.Errorf("convert configuration: %w", err)
	}

	var buf bytes.Buffer
	if b, err := toml.Marshal(cfg); err != nil {
		return fmt.Errorf("marshal config to toml: %w", err)
	} else {
		buf.Write(b)
	}

	if err := os.WriteFile(m.Output, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write configuration to %q: %w", m.Output, err)
	}

	fmt.Printf("Successfully migrated actions-runner configuration to %s\n", m.Output)
	return nil
}
