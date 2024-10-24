package gitstore

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"drassi.run/core/pkg/store/repository"
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
	"k8s.io/apimachinery/pkg/util/rand"
)

type Store interface {
	Fetch(ctx context.Context, repo *repository.Repository, token string) (rev string, err error)
	Read(ctx context.Context, repo *repository.Repository, rev, dir string) (io.ReadCloser, error)
	File(ctx context.Context, repo *repository.Repository, rev, path string) (io.ReadCloser, error)
}

const remoteName string = "anonymous"

func New(rootDir string) (Store, error) {
	if d, err := xpath.ResolveDir(rootDir); err != nil {
		return nil, err
	} else {
		rootDir = d
	}

	if err := os.MkdirAll(rootDir, xfs.DirPerm); err != nil {
		return nil, err
	}

	rs := &store{
		// Using `billy.Filesystem` instead of `rootDir` to
		// abstract from file system implementations and simplify testing.
		fsys:  osfs.New(rootDir),
		repos: make(map[string]*git.Repository),
	}
	return rs, nil
}

type store struct {
	fsys  billy.Filesystem
	repos map[string]*git.Repository
}

func (s *store) Fetch(ctx context.Context, repo *repository.Repository, token string) (string, error) {
	path, err := s.ensureDir(repo)
	if err != nil {
		return "", err
	}

	gitRepo, err := s.ensureRepo(path, repo)
	if err != nil {
		return "", err
	}

	tmpBranch := rand.String(12)
	defer gitRepo.DeleteBranch(tmpBranch)

	err = s.fetch(ctx, repo, token, tmpBranch)
	if err != nil {
		return "", err
	}

	hash, err := gitRepo.ResolveRevision(plumbing.Revision(tmpBranch))
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}

func (s *store) Read(ctx context.Context, repo *repository.Repository, rev string, dir string) (io.ReadCloser, error) {
	id := repository.FullName(repo)
	gitRepo, ok := s.repos[id]
	if !ok {
		return nil, fmt.Errorf("repo %q not found", id)
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
	ch := make(chan int, 1)
	go func() {
		select {
		case <-ctx.Done():
			_ = writer.CloseWithError(ctx.Err())
		case <-ch:
		}
	}()
	go func() {
		defer files.Close()
		defer writer.Close()
		defer close(ch)
		tw := tar.NewWriter(writer)
		defer tw.Close()
		handler := newTarHandler(tw, dir)

		err := files.ForEach(handler)
		_ = writer.CloseWithError(err)
	}()
	return reader, nil
}

func (s *store) File(ctx context.Context, repo *repository.Repository, rev, path string) (io.ReadCloser, error) {
	id := repository.FullName(repo)
	gitRepo, ok := s.repos[id]
	if !ok {
		return nil, fmt.Errorf("repo %q not found", id)
	}

	commit, err := gitRepo.CommitObject(plumbing.NewHash(rev))
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	file, err := tree.File(path)
	if err != nil {
		return nil, err
	}
	if !file.Mode.IsFile() {
		return nil, fmt.Errorf("%q is not a (regular) file", path)
	}

	return file.Reader()
}

func (s *store) fetch(ctx context.Context, repo *repository.Repository, token, branch string) error {
	gitRepo := s.repos[repository.FullName(repo)]

	var auth transport.AuthMethod
	if token != "" {
		auth = &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	}

	remoteConfig := &config.RemoteConfig{
		Name: remoteName,
		URLs: []string{repository.Url(repo)},
	}
	remote, err := gitRepo.CreateRemoteAnonymous(remoteConfig)
	if err != nil {
		return err
	}

	// TODO: using treeless clone when go-git implement it
	// https://github.blog/2020-12-21-get-up-to-speed-with-partial-clone-and-shallow-clone/
	fetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+%s:%s", repo.Ref, branch)),
		},

		Auth:  auth,
		Tags:  git.NoTags,
		Force: true,
		Prune: true,
	}

	return remote.FetchContext(ctx, fetchOptions)
}

func (s *store) ensureDir(repo *repository.Repository) (string, error) {
	path := repository.FullName(repo)
	path = xstring.EnsureSuffix(path, ".git")
	fileInfo, err := s.fsys.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return path, s.fsys.MkdirAll(path, xfs.DirPerm)
		}
		return "", err
	}

	if !fileInfo.IsDir() {
		if err = util.RemoveAll(s.fsys, path); err != nil {
			return "", err
		}
		return path, s.fsys.MkdirAll(path, xfs.DirPerm)
	}

	return path, nil
}

func (s *store) ensureRepo(path string, repo *repository.Repository) (*git.Repository, error) {
	id := repository.FullName(repo)
	if gitRepo, ok := s.repos[id]; ok {
		return gitRepo, nil
	}

	var storer storage.Storer
	if dot, err := s.fsys.Chroot(path); err != nil {
		return nil, err
	} else {
		storer = filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
	}

	//gitRepo, err := git.PlainInit(path, true)
	gitRepo, err := git.Init(storer, nil)
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		//gitRepo, err = git.PlainOpen(path)
		gitRepo, err = git.Open(storer, nil)
	}
	if gitRepo != nil {
		s.repos[id] = gitRepo
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
			content, err := f.Contents()
			if err != nil {
				return err
			}

			hdr.Linkname = content
			return tw.WriteHeader(hdr)
		}

		hdr.Size = f.Size
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if reader, err := f.Reader(); err != nil {
			return err
		} else {
			defer reader.Close()
			_, err = io.Copy(tw, reader)
			return err
		}
	}
	return h
}
