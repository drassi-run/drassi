/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sftpfs

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"drassi.run/core/util/fs"
	"github.com/go-git/go-billy/v6"
	"github.com/go-git/go-billy/v6/util"
	"github.com/pkg/sftp"
)

func New(client *sftp.Client) *SftpFS {
	return &SftpFS{Client: client}
}

var _ billy.Filesystem = (*SftpFS)(nil)

type SftpFS struct {
	*sftp.Client
}

func (fsys *SftpFS) Create(name string) (billy.File, error) {
	return fsys.Client.Create(name)
}

func (fsys *SftpFS) Open(name string) (billy.File, error) {
	return fsys.Client.Open(name)
}

func (fsys *SftpFS) OpenFile(name string, flag int, perm os.FileMode) (billy.File, error) {
	return fsys.Client.OpenFile(name, flag)
}

func (fsys *SftpFS) TempFile(dir, prefix string) (billy.File, error) {
	return util.TempFile(fsys, dir, prefix)
}

func (fsys *SftpFS) Mkdir(name string, perm os.FileMode) error {
	if err := fsys.Client.Mkdir(name); err != nil || perm == xfs.DirPerm {
		return normaliseError(err)
	}
	return fsys.Client.Chmod(name, perm) //nolint:staticcheck
}

func (fsys *SftpFS) MkdirAll(name string, perm os.FileMode) error {
	if err := fsys.Client.MkdirAll(name); err != nil || perm == xfs.DirPerm {
		return normaliseError(err)
	}
	return fsys.Client.Chmod(name, perm) //nolint:staticcheck
}

func (fsys *SftpFS) Readlink(link string) (string, error) {
	return fsys.Client.ReadLink(link) //nolint:staticcheck
}

func (fsys *SftpFS) ReadDir(path string) ([]fs.DirEntry, error) {
	infos, err := fsys.Client.ReadDir(path)
	if err != nil {
		return nil, normaliseError(err)
	}
	entries := make([]fs.DirEntry, len(infos))
	for i, info := range infos {
		entries[i] = fs.FileInfoToDirEntry(info)
	}
	return entries, nil
}

func (fsys *SftpFS) Chroot(string) (billy.Filesystem, error) {
	return nil, billy.ErrNotSupported
}

func (fsys *SftpFS) Root() string {
	return "/"
}

func normaliseError(err error) error {
	//goland:noinspection GoTypeAssertionOnErrors
	switch e := err.(type) {
	case *sftp.StatusError:
		if errors.Is(e.FxCode(), sftp.ErrSSHFxFailure) &&
			strings.Contains(e.Error(), "file exists") {
			return fs.ErrExist
		}
		return err
	default:
		return err
	}
}
