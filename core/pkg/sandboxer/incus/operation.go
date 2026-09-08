/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"fmt"
	"path/filepath"

	"drassi.run/core/pkg/runtime/provision"
)

type addDiskDeviceOp struct {
	provision.Noop
	template *Template
}

func AddDiskDevice(template *Template) provision.Operation {
	return &addDiskDeviceOp{template: template}
}

func (op *addDiskDeviceOp) Name() string { return "incus/disk-device" }

func (op *addDiskDeviceOp) PreLaunch(pctx *provision.Context) error {
	hostMountDir, ok := pctx.Get(provision.KeyHostMountDir)
	if !ok {
		return fmt.Errorf("host mount directory not set in context")
	}

	sourcePath := hostMountDir
	if pctx.Config.Subpath != "" {
		sourcePath = filepath.Join(sourcePath, pctx.Config.Subpath)
	}

	if op.template.Devices == nil {
		op.template.Devices = make(map[string]map[string]string)
	}
	op.template.Devices["runtime-"+pctx.RuntimeName] = map[string]string{
		"type":   "disk",
		"source": sourcePath,
		"path":   pctx.TargetDir,
	}
	return nil
}
