/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
)

// AddDiskDevice returns an Operation that injects a runtime disk device into Template.Devices.
func AddDiskDevice(name, sourceDir, targetDir string, cfg config.Runtime) provision.Operation[*Template] {
	return &addDiskDeviceOp{
		name:      name,
		sourceDir: sourceDir,
		targetDir: targetDir,
		config:    cfg,
	}
}

type addDiskDeviceOp struct {
	provision.Noop[*Template]
	name      string
	sourceDir string
	targetDir string
	config    config.Runtime
}

func (op *addDiskDeviceOp) Name() string { return "incus/disk-device" }

func (op *addDiskDeviceOp) PreLaunch(_ context.Context, tmpl *Template) (*Template, error) {
	if tmpl == nil {
		return nil, errors.New("template cannot be nil")
	}

	sourcePath := op.sourceDir
	if op.config.Subpath != "" {
		sourcePath = filepath.Join(sourcePath, op.config.Subpath)
	}

	if tmpl.Devices == nil {
		tmpl.Devices = make(map[string]map[string]string)
	}
	tmpl.Devices["runtime-"+op.name] = map[string]string{
		"type":     "disk",
		"source":   sourcePath,
		"path":     op.targetDir,
		"readonly": strconv.FormatBool(op.config.ReadOnly),
	}
	return tmpl, nil
}
