/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"errors"
	"fmt"
	"path/filepath"

	"drassi.run/core/pkg/runtime/provision"
)

type addDiskDeviceOp struct {
	provision.Noop[*Template]
}

// AddDiskDevice returns an Operation that injects a runtime disk device into Template.Devices.
func AddDiskDevice() provision.Operation[*Template] {
	return &addDiskDeviceOp{}
}

func (op *addDiskDeviceOp) Name() string { return "incus/disk-device" }

func (op *addDiskDeviceOp) PreLaunch(pctx *provision.Context, tmpl *Template) (*Template, error) {
	if tmpl == nil {
		return nil, errors.New("template cannot be nil")
	}
	hostMountDir, ok := pctx.Get(provision.KeyHostMountDir)
	if !ok {
		return tmpl, fmt.Errorf("host mount directory not set in context")
	}

	sourcePath := hostMountDir
	if pctx.Config.Subpath != "" {
		sourcePath = filepath.Join(sourcePath, pctx.Config.Subpath)
	}

	if tmpl.Devices == nil {
		tmpl.Devices = make(map[string]map[string]string)
	}
	tmpl.Devices["runtime-"+pctx.RuntimeName] = map[string]string{
		"type":   "disk",
		"source": sourcePath,
		"path":   pctx.TargetDir,
	}
	return tmpl, nil
}
