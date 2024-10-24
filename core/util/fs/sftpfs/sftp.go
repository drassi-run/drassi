package sftpfs

import (
	"os"

	"drassi.run/core/util/fs"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	"github.com/pkg/sftp"
)

type SftpFS struct {
	*sftp.Client
}

func New(client *sftp.Client) *SftpFS {
	return &SftpFS{Client: client}
}

func (fs *SftpFS) Create(name string) (billy.File, error) {
	f, err := fs.Client.Create(name)
	return newFile(f), err
}

func (fs *SftpFS) Open(name string) (billy.File, error) {
	f, err := fs.Client.Open(name)
	return newFile(f), err
}

func (fs *SftpFS) OpenFile(name string, flag int, perm os.FileMode) (billy.File, error) {
	f, err := fs.Client.OpenFile(name, flag)
	return newFile(f), err
}

func (fs *SftpFS) TempFile(dir, prefix string) (billy.File, error) {
	return util.TempFile(fs, dir, prefix)
}

func (fs *SftpFS) Mkdir(name string, perm os.FileMode) error {
	if err := fs.Client.Mkdir(name); err != nil || perm == xfs.DirPerm {
		return err
	}
	return fs.Client.Chmod(name, perm)
}

func (fs *SftpFS) MkdirAll(name string, perm os.FileMode) error {
	if err := fs.Client.MkdirAll(name); err != nil || perm == xfs.DirPerm {
		return err
	}
	return fs.Client.Chmod(name, perm)
}

func (fs *SftpFS) Readlink(link string) (string, error) {
	return fs.Client.ReadLink(link)
}

func (fs *SftpFS) Chroot(string) (billy.Filesystem, error) {
	return nil, billy.ErrNotSupported
}

func (fs *SftpFS) Root() string {
	return "/"
}

type file struct {
	*sftp.File
}

func newFile(f *sftp.File) billy.File {
	return file{File: f}
}

func (f file) Lock() error   { return nil }
func (f file) Unlock() error { return nil }
