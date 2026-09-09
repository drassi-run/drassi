/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gitstore

import (
	"archive/tar"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"drassi.run/core/util/string"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"
	"k8s.io/apimachinery/pkg/util/rand"
)

type Manager interface {
	Fetch(ctx context.Context, repo *RepoReference, opts ...FetchOption) (rev string, err error)
	Read(ctx context.Context, repo *RepoReference, rev string, opts ...ReadOption) (io.ReadCloser, error)
	Close() error
}

const remoteName string = "anonymous"

func New(rootDir string, opts ...Option) (Manager, error) {
	if d, err := xpath.ResolveDir(rootDir); err != nil {
		return nil, err
	} else {
		rootDir = d
	}

	if err := os.MkdirAll(rootDir, xfs.DirPerm); err != nil {
		return nil, err
	}

	opt := &options{
		size: defaultSize,
		ttl:  defaultTTL,
	}
	for _, o := range opts {
		o(opt)
	}

	m := &manager{
		// Using `billy.Filesystem` instead of `rootDir` to
		// abstract from file system implementations and simplify testing.
		fsys: osfs.New(rootDir),
	}
	m.repos = expirable.NewLRU[string, *git.Repository](opt.size, m.onEvict, opt.ttl)
	return m, nil
}

type manager struct {
	fsys  billy.Filesystem
	sf    singleflight.Group
	repos *expirable.LRU[string, *git.Repository]
}

func (m *manager) onEvict(_ string, repo *git.Repository) {
	if closer, ok := repo.Storer.(io.Closer); ok {
		_ = closer.Close()
	}
}

func (m *manager) Fetch(ctx context.Context, repo *RepoReference, opts ...FetchOption) (string, error) {
	key := Location(repo)
	v, err, _ := m.sf.Do(key, func() (any, error) {
		repoPath, err := m.ensureDir(repo)
		if err != nil {
			return "", err
		}

		gitRepo, err := m.ensureRepo(repoPath, repo)
		if err != nil {
			return "", err
		}

		tmpBranch := rand.String(12)
		defer gitRepo.DeleteBranch(tmpBranch)

		err = m.fetch(ctx, gitRepo, repo, tmpBranch, opts...)
		if err != nil {
			return "", err
		}

		hash, err := gitRepo.ResolveRevision(plumbing.Revision(tmpBranch))
		if err != nil {
			return "", err
		}
		return hash.String(), nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (m *manager) Read(ctx context.Context, repo *RepoReference, rev string, opts ...ReadOption) (io.ReadCloser, error) {
	ro := new(readOptions)
	for _, opt := range opts {
		opt(ro)
	}

	if ro.file != "" && ro.subpath != "" {
		return nil, errors.New("cannot specify both file and subpath")
	}

	gitRepo, err := m.getRepo(repo)
	if err != nil {
		return nil, err
	}

	commit, err := gitRepo.CommitObject(plumbing.NewHash(rev))
	if err != nil {
		return nil, err
	}

	if ro.file != "" {
		return m.readFile(commit, ro.file)
	}
	return m.readArchive(ctx, commit, ro.subpath)
}

func (m *manager) readFile(commit *object.Commit, filePath string) (io.ReadCloser, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	cleanPath := strings.TrimPrefix(path.Clean(filePath), "/")
	entry, err := tree.FindEntry(cleanPath)
	if err != nil {
		if notFoundErr(err) {
			return nil, fs.ErrNotExist
		}
		return nil, err
	}
	if !entry.Mode.IsFile() {
		return nil, fmt.Errorf("%q is not a (regular) file", filePath)
	}

	file, err := tree.TreeEntryFile(entry)
	if err != nil {
		return nil, err
	}
	return file.Reader()
}

func (m *manager) readArchive(ctx context.Context, commit *object.Commit, subpath string) (io.ReadCloser, error) {
	files, err := commit.Files()
	if err != nil {
		return nil, err
	}

	reader, writer := io.Pipe()
	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = writer.CloseWithError(ctx.Err())
		case <-ch:
		}
	}()
	go func() {
		defer files.Close()
		defer close(ch)

		tw := tar.NewWriter(writer)
		handler := newTarHandler(tw, subpath)

		err := files.ForEach(handler)
		err = cmp.Or(err, tw.Close())
		_ = writer.CloseWithError(err)
	}()
	return reader, nil
}

func (m *manager) getRepo(repo *RepoReference) (*git.Repository, error) {
	id := FullName(repo)
	if gitRepo, ok := m.repos.Get(id); ok {
		return gitRepo, nil
	}

	repoPath := xstring.EnsureSuffix(id, ".git")
	dot, err := m.fsys.Chroot(repoPath)
	if err != nil {
		return nil, fmt.Errorf("repo %q not found: %w", id, err)
	}

	storer := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
	gitRepo, err := git.Open(storer, nil)
	if err != nil {
		return nil, fmt.Errorf("repo %q not found: %w", id, err)
	}

	m.repos.Add(id, gitRepo)
	return gitRepo, nil
}

func (m *manager) fetch(ctx context.Context, gitRepo *git.Repository, repo *RepoReference, branch string, opts ...FetchOption) error {
	fo := new(fetchOptions)
	for _, opt := range opts {
		opt(fo)
	}

	remoteConfig := &config.RemoteConfig{
		Name: remoteName,
		URLs: []string{Url(repo)},
	}
	remote, err := gitRepo.CreateRemoteAnonymous(remoteConfig)
	if err != nil {
		return err
	}

	// TODO: using treeless clone when go-git implement it
	// https://github.blog/2020-12-21-get-up-to-speed-with-partial-clone-and-shallow-clone/
	gfo := &git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+%s:refs/heads/%s", repo.Ref, branch)),
		},

		Auth:  fo.auth,
		Tags:  git.NoTags,
		Force: true,
		Prune: true,
	}

	return remote.FetchContext(ctx, gfo)
}

func (m *manager) ensureDir(repo *RepoReference) (string, error) {
	id := FullName(repo)
	repoPath := xstring.EnsureSuffix(id, ".git")
	fileInfo, err := m.fsys.Stat(repoPath)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return repoPath, m.fsys.MkdirAll(repoPath, xfs.DirPerm)
		}
		return "", err
	}

	if !fileInfo.IsDir() {
		if err = util.RemoveAll(m.fsys, repoPath); err != nil {
			return "", err
		}
		return repoPath, m.fsys.MkdirAll(repoPath, xfs.DirPerm)
	}

	return repoPath, nil
}

func (m *manager) ensureRepo(path string, repo *RepoReference) (*git.Repository, error) {
	id := FullName(repo)
	if gitRepo, ok := m.repos.Get(id); ok {
		return gitRepo, nil
	}

	var storer storage.Storer
	if dot, err := m.fsys.Chroot(path); err != nil {
		return nil, err
	} else {
		storer = filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
	}

	gitRepo, err := git.Init(storer, nil)
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		gitRepo, err = git.Open(storer, nil)
	}
	if err != nil {
		return nil, err
	}

	if gitRepo != nil {
		m.repos.Add(id, gitRepo)
	}
	return gitRepo, err
}

type tarHandler = func(f *object.File) error

func newTarHandler(tw *tar.Writer, dir string) tarHandler {
	h := func(f *object.File) error {
		name := path.Join(dir, f.Name)
		mode, err := f.Mode.ToOSFileMode()
		if err != nil {
			return err
		}

		hdr := &tar.Header{
			Name: name,
			Mode: int64(mode),
		}

		if mode&fs.ModeSymlink != 0 {
			hdr.Typeflag = tar.TypeSymlink
			content, err := f.Contents()
			if err != nil {
				return err
			}

			hdr.Linkname = content
			return tw.WriteHeader(hdr)
		}

		hdr.Typeflag = tar.TypeReg
		hdr.Size = f.Size
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		reader, err := f.Reader()
		if err != nil {
			return err
		}
		defer reader.Close()

		_, err = io.Copy(tw, reader)
		return err
	}
	return h
}

func (m *manager) Close() error {
	m.repos.Purge()
	return nil
}

func notFoundErr(err error) bool {
	return errors.Is(err, object.ErrFileNotFound) ||
		errors.Is(err, object.ErrEntryNotFound) ||
		errors.Is(err, object.ErrDirectoryNotFound)
}
