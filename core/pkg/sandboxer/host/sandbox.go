package host

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/io"
	"drassi.run/core/util/path"
	"drassi.run/core/util/tar"
)

type sandbox struct {
	layout sandboxer.Layout
}

func newSandbox(dir string) (*sandbox, error) {
	// - if dir is not absolute it will be joined with the cwd
	// - clean the result
	if d, err := filepath.Abs(dir); err != nil {
		return nil, err
	} else {
		dir = d
	}

	layout := sandboxer.Layout{
		Workspace: filepath.Join(dir, "workspace"),
		Temp:      filepath.Join(dir, "temp"),
		Actions:   filepath.Join(dir, "actions"),
		Tools:     filepath.Join(dir, "tools"),
	}

	dirs := []string{
		layout.Workspace,
		layout.Actions,
		layout.Tools,
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, folderPerm); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(layout.Temp, 0o777); err != nil {
		return nil, err
	}

	return &sandbox{layout: layout}, nil
}

func (sb *sandbox) Layout() *sandboxer.Layout {
	return &sb.layout
}

func (sb *sandbox) ContainerInfo(context.Context) (*sandboxer.ContainerInfo, error) {
	return nil, nil
}

func (sb *sandbox) Stat(_ context.Context, path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func (sb *sandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	return xtar.Untar(ctx, reader, func(hdr *tar.Header, r io.Reader) error {
		path := filepath.Join(dst, hdr.Name)
		// ensure directory existed
		if err := os.MkdirAll(filepath.Dir(path), folderPerm); err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			return os.Mkdir(path, hdr.FileInfo().Mode())
		case tar.TypeSymlink:
			return os.Symlink(hdr.Linkname, path)
		case tar.TypeReg:
			// Same as os.Create(path), but with custom mode
			f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			// os.File implemented io.ReaderFrom, but fast path only used when reader is also a file
			// tar.Reader implemented io.WriterTo, but it's disabled for now https://github.com/golang/go/issues/22735
			// => ctx is added to the writer
			_, err = io.Copy(xio.NewContextWriter(ctx, f), r)
			return err
		default:
			return fmt.Errorf("unsupported file type %v", hdr.Typeflag)
		}
	})
}

func (sb *sandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	// tar > gzip > buf
	buf := new(bytes.Buffer)
	zw := gzip.NewWriter(buf)
	tw := tar.NewWriter(zw)
	defer zw.Close()
	defer tw.Close()

	fsys := os.DirFS("/")
	root, err := filepath.Rel("/", src)
	if err != nil {
		return nil, err
	}

	err = fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		mode := info.Mode()
		link := ""
		if mode&os.ModeSymlink != 0 {
			if link, err = os.Readlink(path); err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		} else if rpath, err := filepath.Rel(root, path); err != nil {
			return err
		} else if rpath == "." {
			hdr.Name = ""
		} else {
			hdr.Name = rpath
		}

		if err = tw.WriteHeader(hdr); err != nil {
			return err
		}
		if mode.IsRegular() {
			f, err := fsys.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			// os.File implemented io.WriterTo, but fast path only used when writer is also a file
			// tar.Writer implemented io.ReaderFrom, but it's disabled for now https://github.com/golang/go/issues/22735
			// => ctx is added to the reader
			if _, err := io.Copy(tw, xio.NewContextReader(ctx, f)); err != nil {
				return err
			}
		}
		return nil
	})

	if err == nil {
		return io.NopCloser(buf), nil
	}
	return nil, err
}

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams sandboxer.Streams) error {
	// TODO lookup entrypoint under custom PATH
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)

	// env
	c.Env = make([]string, 0, len(env))
	for k, v := range env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// path
	if p := os.Getenv("PATH"); p != "" {
		path = append(path, p)
	}
	if len(path) > 0 {
		p := strings.Join(path, string(os.PathListSeparator))
		c.Env = append(c.Env, "PATH="+p)
	}

	// workdir
	if workdir == "" {
		c.Dir = sb.layout.Workspace
	} else {
		c.Dir = xpath.Abs(workdir, sb.layout.Workspace)
	}

	// streams
	c.Stdin = streams.In()
	c.Stdout = streams.Out()
	c.Stderr = streams.Err()

	return c.Run()
}

func (sb *sandbox) Terminate(context.Context) error {
	jobDir := filepath.Dir(sb.layout.Workspace)
	return os.RemoveAll(jobDir)
}
