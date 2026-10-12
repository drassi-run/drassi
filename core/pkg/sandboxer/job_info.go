/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"drassi.run/core/pkg/model/records"
	"drassi.run/core/util/fs"
)

func WithForge(forge *records.Forge) InfoFunc {
	return func() map[string]string {
		return forge.WellKnownLabels()
	}
}

func WithSandboxId(kind, id string) InfoFunc {
	return func() map[string]string {
		return map[string]string{
			records.LabelSandboxer:   kind,
			records.LabelSandboxerId: id,
		}
	}
}

type InfoFunc func() map[string]string
type Launcher[Req any] = func(ctx context.Context, req Req) (Sandbox, error)

func SetupJobDir[Req any](dir string, l Launcher[Req], fns ...InfoFunc) Launcher[Req] {
	if dir == "" {
		return l
	}

	return func(ctx context.Context, req Req) (Sandbox, error) {
		if err := os.MkdirAll(dir, xfs.DirPerm); err != nil {
			return nil, err
		}

		// launch
		sb, err := l(ctx, req)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, err
		}
		sb = PostTerminate(sb, func(ctx context.Context) error {
			return os.RemoveAll(dir)
		})

		// write job.info
		info := make(map[string]string)
		for _, fn := range fns {
			maps.Copy(info, fn())
		}
		if err = writeJobInfo(dir, info); err != nil {
			_ = sb.Terminate(ctx)
			return nil, err
		}

		return sb, nil
	}
}

func writeJobInfo(dir string, info map[string]string) error {
	keys := slices.Collect(maps.Keys(info))
	slices.Sort(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteRune('=')
		sb.WriteString(info[k])
		sb.WriteRune('\n')
	}

	file := filepath.Join(dir, "job.info")
	return os.WriteFile(file, []byte(sb.String()), xfs.FilePerm)
}
