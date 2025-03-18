/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package filesystem

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"drassi.run/core/pkg/store/cache/storage"
	"drassi.run/core/pkg/store/cache/types"
	"drassi.run/core/util/io"
	"drassi.run/core/util/path"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

const (
	fileName   string      = "cache.tar"
	filePerm   fs.FileMode = 0o600
	folderPerm fs.FileMode = 0o755
)

type fsStorage struct {
	fsys billy.Filesystem
}

func New(rootDir string) (storage.Storage, error) {
	if d, err := xpath.ResolveDir(rootDir); err != nil {
		return nil, err
	} else {
		rootDir = d
	}

	if err := os.MkdirAll(rootDir, folderPerm); err != nil {
		return nil, err
	}

	s := &fsStorage{
		fsys: osfs.New(rootDir),
	}
	return s, nil
}

func (s *fsStorage) InitObject(_ context.Context, cache *types.Cache) error {
	path := s.path(cache)
	if file, err := s.fsys.Create(path); err != nil {
		return err
	} else {
		return file.Truncate(cache.Size)
	}
}

func (s *fsStorage) WriteObject(ctx context.Context, cache *types.Cache, r io.Reader, offset, length int64) error {
	path := s.path(cache)
	file, err := s.fsys.OpenFile(path, os.O_WRONLY, filePerm)
	if err != nil {
		return err
	}

	defer file.Close()
	if _, err = file.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	r = xio.NewContextReader(ctx, r)
	_, err = io.CopyN(file, r, length)
	return err
}

func (s *fsStorage) CommitObject(_ context.Context, _ *types.Cache) error {
	// do nothing
	return nil
}

func (s *fsStorage) ObjectLocation(_ context.Context, _ *types.Cache) string {
	// not support external direct access
	return ""
}

func (s *fsStorage) ReadObject(ctx context.Context, cache *types.Cache, w io.Writer, offset, length int64) error {
	path := s.path(cache)
	file, err := s.fsys.OpenFile(path, os.O_RDONLY, filePerm)
	if err != nil {
		return err
	}

	defer file.Close()
	if _, err = file.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	w = xio.NewContextWriter(ctx, w)
	if length > 0 {
		_, err = io.CopyN(w, file, length)
	} else {
		_, err = io.Copy(w, file)
	}
	return err
}

func (s *fsStorage) path(cache *types.Cache) string {
	return filepath.Join(cache.Namespace, strconv.FormatUint(cache.ID, 10), fileName)
}
