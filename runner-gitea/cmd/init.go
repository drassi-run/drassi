/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	var opts commonOptions

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

func runInit(ctx context.Context, opts *commonOptions) error {
	switch opts.store {
	case "", "local":
		return runInitLocal(ctx, opts)
	case "k8s":
		return runInitK8s(ctx, opts)
	default:
		return fmt.Errorf("unknown store: %s", opts.store)
	}
}

func runInitK8s(ctx context.Context, opts *commonOptions) error {
	return nil
}

//go:embed default.yaml
var defaultManifest string

func runInitLocal(ctx context.Context, opts *commonOptions) error {
	dir := opts.configDir
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
	err = huh.NewSelect[string]().
		Title(fmt.Sprintf("File %s already exists", file)).
		Value(&c).
		Options(
			huh.NewOption("Show Diff", "diff"),
			huh.NewOption("Overwrite", "overwrite"),
		).Run()
	if err != nil {
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
	err := huh.NewConfirm().
		Title("Overwrite?").
		Inline(true).
		Value(&c).
		Run()
	if err != nil {
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
