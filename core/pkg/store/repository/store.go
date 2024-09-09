package repository

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/model"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"k8s.io/apimachinery/pkg/util/rand"
)

type Repositorial interface {
	Repository() *model.Repository
}

type Repository interface {
	// The fullname of the repository. e.g. github.com/octocat/hello-world
	Name() string

	// The Git URL to the repository. e.g. git://github.com/octocat/hello-world.git.
	Url() string

	// The reference of the repository. It cound be a branch, a tag or a commit SHA
	Reference() string
}

type Store interface {
	Fetch(ctx context.Context, repo Repository, token string) (rev string, err error)
	Read(ctx context.Context, repo Repository, rev string) (io.ReadCloser, error)
	File(ctx context.Context, repo Repository, rev, path string) (io.ReadCloser, error)
}

func NewStore(rootDir string) (Store, error) {
	if err := os.MkdirAll(rootDir, 0x755); err != nil {
		return nil, err
	}

	rs := &store{
		rootDir: rootDir,
	}
	return rs, nil
}

type store struct {
	rootDir string
	repos   map[string]*git.Repository
}

func (s *store) Fetch(ctx context.Context, repo Repository, token string) (string, error) {
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

func (s *store) Read(ctx context.Context, repo Repository, rev string) (io.ReadCloser, error) {
	gitRepo, ok := s.repos[repo.Name()]
	if !ok {
		return nil, fmt.Errorf("repo %s not found", repo.Name())
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
		handler := newTarHandler(tw)

		err := files.ForEach(handler)
		_ = writer.CloseWithError(err)
	}()
	return reader, nil
}

func (s *store) File(ctx context.Context, repo Repository, rev, path string) (io.ReadCloser, error) {
	gitRepo, ok := s.repos[repo.Name()]
	if !ok {
		return nil, fmt.Errorf("repo %s not found", repo.Name())
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
		return nil, fmt.Errorf("%s is not a (regular) file", path)
	}

	return file.Reader()
}

func (s *store) fetch(ctx context.Context, repo Repository, token, branch string) error {
	gitRepo := s.repos[repo.Name()]

	var auth transport.AuthMethod
	if token != "" {
		auth = &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	}
	// TODO: using treeless clone when go-git implement it
	// https://github.blog/2020-12-21-get-up-to-speed-with-partial-clone-and-shallow-clone/
	return gitRepo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "anonymous",
		RemoteURL:  repo.Url(),
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+%s:%s", repo.Reference(), branch)),
		},

		Auth:  auth,
		Tags:  git.NoTags,
		Force: true,
		Prune: true,
	})
}

func (s *store) ensureDir(repo Repository) (string, error) {
	path := filepath.Join(s.rootDir, ensureSuffix(repo.Name(), ".git"))
	fileInfo, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			return path, os.MkdirAll(path, 0o755)
		}
		return "", err
	}

	if !fileInfo.IsDir() {
		if err := os.RemoveAll(path); err != nil {
			return "", err
		}
		return path, os.MkdirAll(path, 0o755)
	}

	return path, nil
}

func (s *store) ensureRepo(path string, repo Repository) (*git.Repository, error) {
	id := repo.Name()
	gitRepo, ok := s.repos[id]
	if ok {
		return gitRepo, nil
	}

	gitRepo, err := git.PlainInit(path, true)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			gitRepo, err = git.PlainOpen(path)
		}

		if err != nil {
			return nil, err
		}
	} else {
		_, err = gitRepo.CreateRemoteAnonymous(&config.RemoteConfig{
			Name: "anonymous",
			URLs: []string{repo.Url()},
		})
		if err != nil {
			return nil, err
		}
	}

	s.repos[id] = gitRepo
	return gitRepo, nil
}

func ensureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}

type tarHandler = func(f *object.File) error

func newTarHandler(tw *tar.Writer) tarHandler {
	h := func(f *object.File) error {
		name := f.Name
		mode, err := f.Mode.ToOSFileMode()
		if err != nil {
			return err
		}

		hdr := &tar.Header{
			Name: name,
			Mode: int64(mode),
		}

		if mode&fs.ModeSymlink == fs.ModeSymlink {
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
