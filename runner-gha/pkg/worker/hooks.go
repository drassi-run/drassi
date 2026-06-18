/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"encoding/json/v2"
	"fmt"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/gha-runner/pkg/messages"
)

func printSecretSource[R any](gh *records.Github) executor.Hook[R] {
	return executor.HookFunc[R](func(ctx context.Context, r R) error {
		scribe.Writef(ctx, "Secret source: %s", gh.SecretSource)
		return nil
	})
}

func printTokenPermissions[R any](req *messages.PipelineAgentJobRequest) executor.Hook[R] {
	sysVar := req.Variables

	return executor.HookFunc[R](func(ctx context.Context, r R) error {
		var perm map[string]string
		if v, ok := sysVar["system.github.token.permissions"]; !ok {
			return nil
		} else if err := json.Unmarshal([]byte(v.Value), &perm); err != nil {
			return fmt.Errorf("unmarshal permissions: %w", err)
		}

		if len(perm) > 0 {
			s := scribe.FromContext(ctx)
			end := s.Groupf("Token Permissions")
			defer end()

			for k, v := range perm {
				s.Writef("%s: %s", k, v)
			}
		}
		return nil
	})
}
