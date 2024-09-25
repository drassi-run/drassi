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

	"drassi.run/core/pkg/sandboxer"
	utilio "drassi.run/core/pkg/util/io"
	utilreader "drassi.run/core/pkg/util/reader"
)

type hostSandboxRuntime struct {
	config *Host
}

func NewSandboxRuntime(config *Host) sandboxer.SandboxRuntime {
	h := &hostSandboxRuntime{
		config: config,
	}
	return h
}

func (h *hostSandboxRuntime) Close() error {
	return nil
}

func (h *hostSandboxRuntime) Connect(ctx context.Context) error {
	return nil
}

func (h *hostSandboxRuntime) LaunchSandbox(ctx context.Context, request sandboxer.LaunchSandboxRequest) (sandboxer.LaunchSandboxResponse, error) {
	res := sandboxer.LaunchSandboxResponse{}

	spath := filepath.Join(h.config.Spec.RootDir, request.JobId)
	if err := os.RemoveAll(spath); err != nil {
		return res, err
	}
	sb := &hostSandbox{
		sandboxId:      spath,
		sandboxRuntime: h,
	}
	if err := os.MkdirAll(spath, 0o755); err != nil {
		return res, err
	}
	sb.workspaceDir = filepath.Join(spath, "workspace")
	if err := os.Mkdir(sb.workspaceDir, 0o755); err != nil {
		return res, err
	}
	sb.actionsDir = filepath.Join(spath, "actions")
	if err := os.Mkdir(sb.actionsDir, 0o755); err != nil {
		return res, err
	}
	sb.toolsDir = filepath.Join(spath, "tools")
	if err := os.Mkdir(sb.toolsDir, 0o755); err != nil {
		return res, err
	}
	sb.tempDir = filepath.Join(spath, "tmp")
	if err := os.Mkdir(sb.tempDir, 0o755); err != nil {
		return res, err
	}
	if err := os.Mkdir(filepath.Join(sb.tempDir, "file_commands"), 0o755); err != nil {
		return res, err
	}
	if err := os.Mkdir(filepath.Join(sb.tempDir, "workflow"), 0o755); err != nil {
		return res, err
	}
	if err := os.Mkdir(filepath.Join(sb.tempDir, "scripts"), 0o755); err != nil {
		return res, err
	}
	path := h.config.Spec.Path
	if path == "" {
		path = os.Getenv("PATH")
	}
	res.Sandbox = sb
	res.Env = map[string]string{
		"PATH":              path,
		"RUNNER_TEMP":       sb.tempDir,
		"RUNNER_TOOL_CACHE": sb.toolsDir,
		"RUNNER_WORKSPACE":  sb.workspaceDir,
		"GITHUB_WORKSPACE":  sb.workspaceDir, // TODO
	}
	return res, nil
}

func (h *hostSandboxRuntime) TerminateSandbox(ctx context.Context, request sandboxer.TerminateSandboxRequest) (sandboxer.TerminateSandboxResponse, error) {
	res := sandboxer.TerminateSandboxResponse{}
	sandbox, ok := request.Sandbox.(*hostSandbox)
	if !ok {
		return res, fmt.Errorf("unsupport sandbox type %T", request.Sandbox)
	}

	return res, os.RemoveAll(sandbox.sandboxId)
}

func (h *hostSandboxRuntime) ExecuteSandbox(ctx context.Context, request sandboxer.ExecuteSandboxRequest) (sandboxer.ExecuteSandboxResponse, error) {
	// TODO lookup entrypoint under custom PATH
	cmd := exec.CommandContext(ctx, request.Cmd[0], request.Cmd[1:]...)
	cmd.Dir = request.Workdir
	env := make([]string, 0, len(request.Env))
	for k, v := range request.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	if _, ok := request.Env["PATH"]; !ok {
		path := os.Getenv("PATH")
		env = append(env, fmt.Sprintf("PATH=%s", path))
	}
	cmd.Env = env

	cmd.Stdin = request.Streams.In
	cmd.Stdout = request.Streams.Out
	cmd.Stderr = request.Streams.Err

	err := cmd.Run()
	res := sandboxer.ExecuteSandboxResponse{
		ExitCode: cmd.ProcessState.ExitCode(),
	}
	return res, err
}

func (h *hostSandboxRuntime) CopyFromSandbox(ctx context.Context, request sandboxer.CopyFromSandboxRequest) (sandboxer.CopyFromSandboxResponse, error) {
	// tar > gzip > buf
	buf := new(bytes.Buffer)
	zw := gzip.NewWriter(buf)
	tw := tar.NewWriter(zw)
	defer zw.Close()
	defer tw.Close()

	res := sandboxer.CopyFromSandboxResponse{}
	fsys := os.DirFS("/")
	root, err := filepath.Rel("/", request.SourcePath)
	if err != nil {
		return res, err
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
			if _, err := io.Copy(tw, utilio.NewContextReader(ctx, f)); err != nil {
				return err
			}
		}
		return nil
	})

	if err == nil {
		res.Reader = io.NopCloser(buf)
	}
	return res, err
}

func (h *hostSandboxRuntime) CopyToSandbox(ctx context.Context, request sandboxer.CopyToSandboxRequest) (sandboxer.CopyToSandboxResponse, error) {
	res := sandboxer.CopyToSandboxResponse{}
	err := utilreader.Untar(ctx, request.Content, func(hdr *tar.Header, r io.Reader) error {
		path := filepath.Join(request.DestinationPath, hdr.Name)
		// ensure directory existed
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			return os.Mkdir(path, hdr.FileInfo().Mode())
		case tar.TypeSymlink:
			return os.Symlink(hdr.Linkname, path)
		case tar.TypeReg:
			// Same as os.Create(path), but with custom
			f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			// os.File implemented io.ReaderFrom, but fast path only used when reader is also a file
			// tar.Reader implemented io.WriterTo, but it's disabled for now https://github.com/golang/go/issues/22735
			// => ctx is added to the writer
			_, err = io.Copy(utilio.NewContextWriter(ctx, f), r)
			return err
		default:
			return fmt.Errorf("unsupported file type %v", hdr.Typeflag)
		}
	})
	return res, err
}
