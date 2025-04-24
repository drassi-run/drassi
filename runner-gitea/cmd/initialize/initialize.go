/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package initialize

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"drassi.run/gitea-runner/cmd/common"
	"github.com/charmbracelet/huh"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var opts common.Options

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init default objects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runInit(ctx, &opts)
		},
	}

	flags := cmd.Flags()
	opts.RegisterFlags(flags)

	return cmd
}

func runInit(ctx context.Context, opts *common.Options) error {
	switch opts.Store {
	case "", "local":
		return runInitLocal(ctx, opts)
	case "k8s":
		return runInitK8s(ctx, opts)
	default:
		return fmt.Errorf("unknown store: %s", opts.Store)
	}
}

func runInitK8s(ctx context.Context, opts *common.Options) error {
	return nil
}

//go:embed default.yaml
var defaultManifest string

func runInitLocal(ctx context.Context, opts *common.Options) error {
	dir := opts.ConfigDir
	if dir == "" {
		return fmt.Errorf("--config-dir is required")
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve config-dir path: %v", err)
	}

	file := filepath.Join(dir, "default.yaml")
	fi, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return os.WriteFile(file, []byte(defaultManifest), 0644)
	}

	if !fi.Mode().IsRegular() {
		return fmt.Errorf("%s is not a file", file)
	}
	var c string
	inquiry := huh.NewSelect[string]().
		Title(fmt.Sprintf("File %s already exists", file)).
		Value(&c).
		Options(
			huh.NewOption("Show Diff", "diff"),
			huh.NewOption("Overwrite", "overwrite"),
		)
	if err = inquiry.Run(); err != nil {
		return err
	}

	var overwrite bool
	switch c {
	case "diff":
		if overwrite, err = showDiff(file, defaultManifest); err != nil {
			return err
		}
	case "overwrite":
		overwrite = true
	default:
		return fmt.Errorf("unknown option: %s", c)
	}

	if overwrite {
		return os.WriteFile(file, []byte(defaultManifest), 0644)
	}
	return nil
}

func showDiff(file string, manifest string) (bool, error) {
	var fileContent string
	if b, err := os.ReadFile(file); err != nil {
		return false, err
	} else {
		fileContent = string(b)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(fileContent, manifest, false)
	if !anyChanges(diffs) {
		fmt.Println("config file is up-to-date")
		return false, nil
	}
	fmt.Print(dmp.DiffPrettyText(diffs))

	var c bool
	confirm := huh.NewConfirm().
		Title("Overwrite?").
		Inline(true).
		Value(&c)
	if err := confirm.Run(); err != nil {
		return false, err
	}
	return c, nil
}

func anyChanges(diffs []diffmatchpatch.Diff) bool {
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			return true
		}
	}
	return false
}
