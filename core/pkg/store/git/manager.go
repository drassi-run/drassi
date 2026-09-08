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
	"sync"

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
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"golang.org/x/sync/singleflight"
	"k8s.io/apimachinery/pkg/util/rand"
)

type Manager interface {
	Fetch(ctx context.Context, repo *RepoReference, token string) (rev string, err error)
	Read(ctx context.Context, repo *RepoReference, rev, dir string) (io.ReadCloser, error)
	File(ctx context.Context, repo *RepoReference, rev, path string) (io.ReadCloser, error)
}

const remoteName string = "anonymous"

func New(rootDir string) (Manager, error) {
	if d, err := xpath.ResolveDir(rootDir); err != nil {
		return nil, err
	} else {
		rootDir = d
	}

	if err := os.MkdirAll(rootDir, xfs.DirPerm); err != nil {
		return nil, err
	}

	m := &manager{
		// Using `billy.Filesystem` instead of `rootDir` to
		// abstract from file system implementations and simplify testing.
		fsys:  osfs.New(rootDir),
		repos: make(map[string]*git.Repository),
	}
	return m, nil
}

type manager struct {
	fsys  billy.Filesystem
	sf    singleflight.Group
	mu    sync.RWMutex
	repos map[string]*git.Repository
}

func (m *manager) Fetch(ctx context.Context, repo *RepoReference, token string) (string, error) {
	key := Location(repo)
	v, err, _ := m.sf.Do(key, func() (any, error) {
		path, err := m.ensureDir(repo)
		if err != nil {
			return "", err
		}

		gitRepo, err := m.ensureRepo(path, repo)
		if err != nil {
			return "", err
		}

		tmpBranch := rand.String(12)
		defer gitRepo.DeleteBranch(tmpBranch)

		err = m.fetch(ctx, gitRepo, repo, token, tmpBranch)
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

func (m *manager) Read(ctx context.Context, repo *RepoReference, rev string, dir string) (io.ReadCloser, error) {
	gitRepo, err := m.getRepo(repo)
	if err != nil {
		return nil, err
	}

	commit, err := gitRepo.CommitObject(plumbing.NewHash(rev))
	if err != nil {
		return nil, err
	}
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
		handler := newTarHandler(tw, dir)

		err := files.ForEach(handler)
		err = cmp.Or(err, tw.Close())
		_ = writer.CloseWithError(err)
	}()
	return reader, nil
}

func (m *manager) File(ctx context.Context, repo *RepoReference, rev, filePath string) (io.ReadCloser, error) {
	gitRepo, err := m.getRepo(repo)
	if err != nil {
		return nil, err
	}

	commit, err := gitRepo.CommitObject(plumbing.NewHash(rev))
	if err != nil {
		return nil, err
	}

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

	if file, err := tree.TreeEntryFile(entry); err != nil {
		return nil, err
	} else {
		return file.Reader()
	}
}

func (m *manager) getRepo(repo *RepoReference) (*git.Repository, error) {
	id := FullName(repo)
	m.mu.RLock()
	gitRepo, ok := m.repos[id]
	m.mu.RUnlock()
	if ok {
		return gitRepo, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if gitRepo, ok := m.repos[id]; ok {
		return gitRepo, nil
	}

	path := xstring.EnsureSuffix(id, ".git")
	dot, err := m.fsys.Chroot(path)
	if err != nil {
		return nil, fmt.Errorf("repo %q not found: %w", id, err)
	}

	storer := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
	gitRepo, err = git.Open(storer, nil)
	if err != nil {
		return nil, fmt.Errorf("repo %q not found: %w", id, err)
	}

	m.repos[id] = gitRepo
	return gitRepo, nil
}

func (m *manager) fetch(ctx context.Context, gitRepo *git.Repository, repo *RepoReference, token, branch string) error {
	var auth transport.AuthMethod
	if token != "" {
		auth = &http.BasicAuth{
			Username: "token",
			Password: token,
		}
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
	fetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+%s:refs/heads/%s", repo.Ref, branch)),
		},

		Auth:  auth,
		Tags:  git.NoTags,
		Force: true,
		Prune: true,
	}

	return remote.FetchContext(ctx, fetchOptions)
}

func (m *manager) ensureDir(repo *RepoReference) (string, error) {
	path := FullName(repo)
	path = xstring.EnsureSuffix(path, ".git")
	fileInfo, err := m.fsys.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return path, m.fsys.MkdirAll(path, xfs.DirPerm)
		}
		return "", err
	}

	if !fileInfo.IsDir() {
		if err = util.RemoveAll(m.fsys, path); err != nil {
			return "", err
		}
		return path, m.fsys.MkdirAll(path, xfs.DirPerm)
	}

	return path, nil
}

func (m *manager) ensureRepo(path string, repo *RepoReference) (*git.Repository, error) {
	id := FullName(repo)
	m.mu.Lock()
	defer m.mu.Unlock()

	if gitRepo, ok := m.repos[id]; ok {
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
		m.repos[id] = gitRepo
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

func notFoundErr(err error) bool {
	return errors.Is(err, object.ErrFileNotFound) ||
		errors.Is(err, object.ErrEntryNotFound) ||
		errors.Is(err, object.ErrDirectoryNotFound)
}
