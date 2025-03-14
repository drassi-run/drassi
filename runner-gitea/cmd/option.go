package cmd

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

type commonOptions struct {
	store     string
	configDir string
	name      string
}

func (o *commonOptions) SetFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.store, "store", "local", "Manifest store")

	flags.StringVar(&o.configDir, "config-dir", "", "Configuration directory")
	_ = cobra.MarkFlagDirname(flags, "config-dir")

	flags.StringVar(&o.name, "name", "", "Gitea instance name")
	_ = cobra.MarkFlagRequired(flags, "name")
}

func manifestStore(o *commonOptions) (manifest.Store, error) {
	s, err := newScheme()
	if err != nil {
		return nil, err
	}

	if o.store == "local" {
		if o.configDir == "" {
			return nil, fmt.Errorf("--config-dir is required")
		}

		absPath, err := filepath.Abs(o.configDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve config-dir path: %v", err)
		}
		store := filesystem.NewStore(absPath, s)
		return store, nil
	}

	return nil, fmt.Errorf("unknown manifest store: %s", o.store)
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
