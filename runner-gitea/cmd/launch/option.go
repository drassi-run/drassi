/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package launch

import (
	"fmt"
	"os"

	giteaconfig "drassi.run/gitea-runner/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type options struct {
	Config string
}

func (o *options) RegisterFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Config, "config", "", "TOML configuration file")
	_ = cobra.MarkFlagFilename(flags, "config", "toml")
	_ = cobra.MarkFlagRequired(flags, "config")
}

func LoadConfig(o *options) (*giteaconfig.Config, error) {
	if o.Config == "" {
		return nil, fmt.Errorf("--config is required")
	}

	f, err := os.Open(o.Config)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	config := giteaconfig.DefaultConfig()
	dec := toml.NewDecoder(f).EnableUnmarshalerInterface()
	if err = dec.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}
