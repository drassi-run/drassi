/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_command

import (
	exec "drassi.run/core/pkg/executor"
	cmd "drassi.run/core/pkg/executor/command"
)

type commandEnvProvider struct {
	fileMgr cmd.FileManager[exec.Milieu]
}

func NewCommandEnvProvider(fileMgr cmd.FileManager[exec.Milieu]) exec.EnvProvider {
	return &commandEnvProvider{fileMgr}
}

func (c *commandEnvProvider) Env(e exec.StepExecutor) map[string]string {
	res := exec.NewMilieu("", e)
	return c.fileMgr.Env(res)
}
