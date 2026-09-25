/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"context"
	"io"
	"io/fs"
	"strings"

	"github.com/tonistiigi/fsutil"
)

type fsAdapter struct {
	fsys fs.FS
}

func newFSAdapter(fsys fs.FS) fsutil.FS {
	return &fsAdapter{fsys: fsys}
}

func (a *fsAdapter) Walk(ctx context.Context, target string, fn fs.WalkDirFunc) error {
	root := "."
	if target != "" && target != "/" {
		root = strings.TrimPrefix(target, "/")
	}
	return fs.WalkDir(a.fsys, root, fn)
}

func (a *fsAdapter) Open(name string) (io.ReadCloser, error) {
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		name = "."
	}
	return a.fsys.Open(name)
}
