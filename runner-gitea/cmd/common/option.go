/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"path/filepath"

	"drassi.run/core/pkg/manifest"
	"drassi.run/core/pkg/manifest/filesystem"
	sandboxerv1a1 "drassi.run/core/pkg/sandboxer/apis/v1alpha1"
	giteav1a1 "drassi.run/gitea-runner/pkg/apis/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

type Options struct {
	Store     string
	ConfigDir string
}

func (o *Options) RegisterFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Store, "store", "local", "Manifest store")

	flags.StringVar(&o.ConfigDir, "config-dir", "", "Configuration directory")
	_ = cobra.MarkFlagDirname(flags, "config-dir")
}

func ManifestStore(o *Options) (manifest.Store, error) {
	s, err := newScheme()
	if err != nil {
		return nil, err
	}

	if o.Store == "local" {
		if o.ConfigDir == "" {
			return nil, fmt.Errorf("--config-dir is required")
		}

		absPath, err := filepath.Abs(o.ConfigDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve config-dir path: %v", err)
		}
		store := filesystem.NewStore(absPath, s)
		return store, nil
	}

	return nil, fmt.Errorf("unknown manifest store: %s", o.Store)
}

func newScheme() (*runtime.Scheme, error) {
	s := runtime.NewScheme()

	if err := scheme.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := sandboxerv1a1.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := giteav1a1.AddToScheme(s); err != nil {
		return nil, err
	}
	return s, nil
}
