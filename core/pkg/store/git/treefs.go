/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gitstore

import (
	"cmp"
	"errors"
	"io"
	"io/fs"
	"slices"
	"time"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/storer"
)

var (
	_ fs.FS          = (*treeFS)(nil)
	_ fs.ReadDirFS   = (*treeFS)(nil)
	_ fs.StatFS      = (*treeFS)(nil)
	_ fs.ReadFileFS  = (*treeFS)(nil)
	_ fs.SubFS       = (*treeFS)(nil)
	_ fs.ReadDirFile = (*treeDirFile)(nil)
)

type treeFS struct {
	tree    *object.Tree
	storer  storer.EncodedObjectStorer
	modTime time.Time
}

func newTreeFS(tree *object.Tree, storer storer.EncodedObjectStorer, modTime time.Time) *treeFS {
	return &treeFS{
		tree:    tree,
		storer:  storer,
		modTime: modTime,
	}
}

func (tfs *treeFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	if name == "." {
		entries, err := tfs.treeEntriesToDirEntries(tfs.tree.Entries)
		if err != nil {
			return nil, &fs.PathError{Op: "open", Path: name, Err: err}
		}
		info := &treeFileInfo{
			name:    ".",
			size:    0,
			mode:    fs.ModeDir | 0o755,
			modTime: tfs.modTime,
			isDir:   true,
		}
		return &treeDirFile{
			name:    ".",
			entries: entries,
			info:    info,
		}, nil
	}

	entry, err := tfs.tree.FindEntry(name)
	if err != nil {
		if notFoundErr(err) {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	osMode, err := entry.Mode.ToOSFileMode()
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	if entry.Mode == filemode.Dir {
		subTree, err := tfs.tree.Tree(name)
		if err != nil {
			if notFoundErr(err) {
				return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
			}
			return nil, &fs.PathError{Op: "open", Path: name, Err: err}
		}
		entries, err := tfs.treeEntriesToDirEntries(subTree.Entries)
		if err != nil {
			return nil, &fs.PathError{Op: "open", Path: name, Err: err}
		}
		info := &treeFileInfo{
			name:    entry.Name,
			size:    0,
			mode:    osMode,
			modTime: tfs.modTime,
			isDir:   true,
			sys:     entry,
		}
		return &treeDirFile{
			name:    name,
			entries: entries,
			info:    info,
		}, nil
	}

	blob, err := object.GetBlob(tfs.storer, entry.Hash)
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	rc, err := blob.Reader()
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	info := &treeFileInfo{
		name:    entry.Name,
		size:    blob.Size,
		mode:    osMode,
		modTime: tfs.modTime,
		isDir:   false,
		sys:     entry,
	}

	return &treeFile{
		name: name,
		rc:   rc,
		info: info,
	}, nil
}

func (tfs *treeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}

	if name == "." {
		return tfs.treeEntriesToDirEntries(tfs.tree.Entries)
	}

	entry, err := tfs.tree.FindEntry(name)
	if err != nil {
		if notFoundErr(err) {
			return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}

	if entry.Mode != filemode.Dir {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: errors.New("not a directory")}
	}

	subTree, err := tfs.tree.Tree(name)
	if err != nil {
		if notFoundErr(err) {
			return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: err}
	}

	return tfs.treeEntriesToDirEntries(subTree.Entries)
}

func (tfs *treeFS) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}

	if name == "." {
		return &treeFileInfo{
			name:    ".",
			size:    0,
			mode:    fs.ModeDir | 0o755,
			modTime: tfs.modTime,
			isDir:   true,
		}, nil
	}

	entry, err := tfs.tree.FindEntry(name)
	if err != nil {
		if notFoundErr(err) {
			return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}

	osMode, err := entry.Mode.ToOSFileMode()
	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}

	var size int64
	if entry.Mode != filemode.Dir && tfs.storer != nil {
		size, err = tfs.storer.EncodedObjectSize(entry.Hash)
		if err != nil && !errors.Is(err, plumbing.ErrObjectNotFound) {
			return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
		}
	}

	return &treeFileInfo{
		name:    entry.Name,
		size:    size,
		mode:    osMode,
		modTime: tfs.modTime,
		isDir:   entry.Mode == filemode.Dir,
		sys:     entry,
	}, nil
}

func (tfs *treeFS) ReadFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrInvalid}
	}

	if name == "." {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: errors.New("is a directory")}
	}

	entry, err := tfs.tree.FindEntry(name)
	if err != nil {
		if notFoundErr(err) {
			return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: err}
	}

	if entry.Mode == filemode.Dir {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: errors.New("is a directory")}
	}

	blob, err := object.GetBlob(tfs.storer, entry.Hash)
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			return nil, &fs.PathError{Op: "readfile", Path: name, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: err}
	}

	rc, err := blob.Reader()
	if err != nil {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: err}
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, &fs.PathError{Op: "readfile", Path: name, Err: err}
	}

	return data, nil
}

func (tfs *treeFS) Sub(dir string) (fs.FS, error) {
	if !fs.ValidPath(dir) {
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrInvalid}
	}

	if dir == "." {
		return tfs, nil
	}

	entry, err := tfs.tree.FindEntry(dir)
	if err != nil {
		if notFoundErr(err) {
			return nil, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: err}
	}

	if entry.Mode != filemode.Dir {
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: errors.New("not a directory")}
	}

	subTree, err := tfs.tree.Tree(dir)
	if err != nil {
		if notFoundErr(err) {
			return nil, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrNotExist}
		}
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: err}
	}

	return newTreeFS(subTree, tfs.storer, tfs.modTime), nil
}

func (tfs *treeFS) treeEntriesToDirEntries(entries []object.TreeEntry) ([]fs.DirEntry, error) {
	result := make([]fs.DirEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, &treeDirEntry{
			TreeEntry: e,
			storer:    tfs.storer,
			modTime:   tfs.modTime,
		})
	}
	slices.SortFunc(result, func(a, b fs.DirEntry) int {
		return cmp.Compare(a.Name(), b.Name())
	})
	return result, nil
}

type treeDirFile struct {
	name    string
	entries []fs.DirEntry
	offset  int
	info    fs.FileInfo
}

func (d *treeDirFile) Stat() (fs.FileInfo, error) {
	return d.info, nil
}

func (d *treeDirFile) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: d.name, Err: errors.New("is a directory")}
}

func (d *treeDirFile) Close() error {
	return nil
}

func (d *treeDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.offset >= len(d.entries) && n > 0 {
		return nil, io.EOF
	}

	if n <= 0 {
		rest := d.entries[d.offset:]
		d.offset = len(d.entries)
		return rest, nil
	}

	avail := len(d.entries) - d.offset
	if avail == 0 {
		return nil, io.EOF
	}
	if avail > n {
		avail = n
	}
	res := d.entries[d.offset : d.offset+avail]
	d.offset += avail
	return res, nil
}

type treeFile struct {
	name   string
	rc     io.ReadCloser
	info   fs.FileInfo
	closed bool
}

func (f *treeFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *treeFile) Read(b []byte) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}
	return f.rc.Read(b)
}

func (f *treeFile) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	return f.rc.Close()
}

type treeDirEntry struct {
	object.TreeEntry
	storer  storer.EncodedObjectStorer
	modTime time.Time
}

func (e *treeDirEntry) Name() string {
	return e.TreeEntry.Name
}

func (e *treeDirEntry) IsDir() bool {
	return e.Mode == filemode.Dir
}

func (e *treeDirEntry) Type() fs.FileMode {
	osMode, err := e.Mode.ToOSFileMode()
	if err != nil {
		return 0
	}
	return osMode.Type()
}

func (e *treeDirEntry) Info() (fs.FileInfo, error) {
	osMode, err := e.Mode.ToOSFileMode()
	if err != nil {
		return nil, err
	}

	var size int64
	if e.Mode != filemode.Dir && e.storer != nil {
		size, err = e.storer.EncodedObjectSize(e.Hash)
		if err != nil && !errors.Is(err, plumbing.ErrObjectNotFound) {
			return nil, err
		}
	}

	info := &treeFileInfo{
		name:    e.TreeEntry.Name,
		size:    size,
		mode:    osMode,
		modTime: e.modTime,
		isDir:   e.Mode == filemode.Dir,
		sys:     e.TreeEntry,
	}
	return info, nil
}

type treeFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

func (i *treeFileInfo) Name() string       { return i.name }
func (i *treeFileInfo) Size() int64        { return i.size }
func (i *treeFileInfo) Mode() fs.FileMode  { return i.mode }
func (i *treeFileInfo) ModTime() time.Time { return i.modTime }
func (i *treeFileInfo) IsDir() bool        { return i.isDir }
func (i *treeFileInfo) Sys() any           { return i.sys }
