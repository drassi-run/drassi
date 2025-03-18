/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	incusclient "github.com/lxc/incus/v6/client"
	incusapi "github.com/lxc/incus/v6/shared/api"
)

func getPath(client incusclient.InstanceServer, inst string) (string, error) {
	req := incusapi.InstanceExecPost{
		Command:     []string{"true"},
		WaitForWS:   true,
		Interactive: false,
	}
	op, err := client.ExecInstance(inst, req, nil)
	if err != nil {
		return "", err
	}

	metadata := op.Get().Metadata
	if env, ok := metadata["environment"]; ok {
		if m, ok := env.(map[string]string); ok {
			return m["PATH"], nil
		}
		if m, ok := env.(map[string]any); ok {
			if path, ok := m["PATH"].(string); ok {
				return path, nil
			}
		}
	}
	_ = op.Cancel()

	return "", nil
}
