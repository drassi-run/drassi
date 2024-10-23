package incus

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/path"
	"drassi.run/core/util/tar"
	"github.com/gorilla/websocket"
	incusclient "github.com/lxc/incus/v6/client"
	incusapi "github.com/lxc/incus/v6/shared/api"
	"github.com/pkg/sftp"
)

const folderPerm fs.FileMode = 0o755

type sandbox struct {
	client       incusclient.InstanceServer
	sftpClient   *sftp.Client
	instanceName string

	layout   sandboxer.Layout
	path     string
	uid, gid uint32
}

func newSandbox(client incusclient.InstanceServer, inst string) (*sandbox, error) {
	dir := "/opt/drassi/"
	sb := &sandbox{
		client:       client,
		instanceName: inst,
		layout: sandboxer.Layout{
			Workspace: filepath.Join(dir, "workspace"),
			Temp:      filepath.Join(dir, "temp"),
			Actions:   filepath.Join(dir, "actions"),
			Tools:     filepath.Join(dir, "tools"),
		},
	}

	// retrieve path
	if path, err := getPath(client, inst); err != nil {
		return nil, err
	} else {
		sb.path = path
	}

	// init sftpClient
	if sftpClient, err := client.GetInstanceFileSFTP(inst); err != nil {
		return nil, err
	} else {
		sb.sftpClient = sftpClient
	}

	// init layout
	layout := &sb.layout
	dirs := map[string]fs.FileMode{
		dir:              folderPerm,
		layout.Workspace: folderPerm,
		layout.Temp:      fs.FileMode(0o777),
		layout.Actions:   folderPerm,
		layout.Tools:     folderPerm,
	}
	for d, pem := range dirs {
		fileArgs := incusclient.InstanceFileArgs{
			Type:      "directory",
			Mode:      int(pem),
			WriteMode: "overwrite",
		}
		if err := client.CreateInstanceFile(inst, d, fileArgs); err != nil {
			return nil, err
		}
	}

	return sb, nil
}

func (sb *sandbox) Layout() *sandboxer.Layout {
	return &sb.layout
}

func (sb *sandbox) ContainerInfo(ctx context.Context) (*sandboxer.ContainerInfo, error) {
	return nil, nil
}

func (sb *sandbox) Stat(_ context.Context, path string) (fs.FileInfo, error) {
	return sb.sftpClient.Stat(path)
}

func (sb *sandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	return xtar.Untar(ctx, reader, sb.newUntarHandler(dst))
}

func (sb *sandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	h := newTarHandler(tw)
	if err := sb.walk(src, "", h); err != nil {
		return nil, err
	}

	return io.NopCloser(buf), nil
}

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams sandboxer.Streams) error {
	// Prepare the command
	req := incusapi.InstanceExecPost{
		Command:     cmd,
		WaitForWS:   true,
		Interactive: false,
		Environment: env,
		User:        sb.uid,
		Group:       sb.gid,
	}

	// path
	if sb.path != "" {
		path = append(path, sb.path)
	}
	if len(path) > 0 {
		p := strings.Join(path, string(os.PathListSeparator))
		req.Environment["PATH"] = p
	}

	// workdir
	if workdir == "" {
		req.Cwd = sb.layout.Workspace
	} else {
		req.Cwd = xpath.Abs(workdir, sb.layout.Workspace)
	}

	execArgs := &incusclient.InstanceExecArgs{
		Stdin:    streams.In(),
		Stdout:   streams.Out(),
		Stderr:   streams.Err(),
		Control:  func(conn *websocket.Conn) {}, // TODO
		DataDone: make(chan bool),
	}

	// Run the command in the instance
	if op, err := sb.client.ExecInstance(sb.instanceName, req, execArgs); err != nil {
		return err
	} else if err = op.WaitContext(ctx); err != nil {
		return err
	} else if exitCode, ok := op.Get().Metadata["return"].(int); ok && exitCode != 0 {
		return fmt.Errorf("exitcode '%d': failure", exitCode)
	}
	return nil
}

func (sb *sandbox) Terminate(ctx context.Context) error {
	_ = sb.sftpClient.Close()

	if inst, _, err := sb.client.GetInstance(sb.instanceName); err != nil {
		return err
	} else if (inst.StatusCode != 0) && (inst.StatusCode != incusapi.Stopped) {
		req := incusapi.InstanceStatePut{
			Action: "stop",
			Force:  true,
		}
		if op, err := sb.client.UpdateInstanceState(sb.instanceName, req, ""); err != nil {
			return err
		} else if err = op.WaitContext(ctx); err != nil {
			return fmt.Errorf("stopping the instance %s failed: %s", sb.instanceName, err)
		}

		if inst.Ephemeral {
			return nil
		}
	}

	if op, err := sb.client.DeleteInstance(sb.instanceName); err != nil {
		return err
	} else if err = op.WaitContext(ctx); err != nil {
		return fmt.Errorf("failed deleting instance %s: %s", sb.instanceName, err)
	}
	return nil
}

type fileHandler = func(string, io.Reader, *incusclient.InstanceFileResponse) error

func (sb *sandbox) walk(root, name string, h fileHandler) error {
	path := filepath.Join(root, name)
	buf, resp, err := sb.client.GetInstanceFile(sb.instanceName, path)
	if err != nil {
		return err
	}
	if buf != nil {
		defer buf.Close()
	}

	if err = h(name, buf, resp); err != nil {
		return err
	}

	if resp.Type == "directory" {
		for _, ent := range resp.Entries {
			next := filepath.Join(name, ent)
			if err = sb.walk(root, next, h); err != nil {
				return err
			}
		}
	}
	return nil
}

func newTarHandler(tw *tar.Writer) fileHandler {
	h := func(name string, r io.Reader, resp *incusclient.InstanceFileResponse) error {
		var buf *bytes.Buffer
		hdr := &tar.Header{
			Name: name,
			Mode: int64(resp.Mode),
			Uid:  int(resp.UID),
			Gid:  int(resp.GID),
		}

		switch resp.Type {
		case "file":
			buf = new(bytes.Buffer)
			size, err := buf.ReadFrom(r)
			if err != nil {
				return err
			}

			hdr.Typeflag = tar.TypeReg
			hdr.Size = size
		case "directory":
			hdr.Typeflag = tar.TypeDir
			hdr.Name += "/"
		case "symlink":
			linkTarget, err := io.ReadAll(r) // BUG from incus, SHOULD not resolve symlink to absolute path
			if err != nil {
				return err
			}

			hdr.Typeflag = tar.TypeSymlink
			hdr.Linkname = string(linkTarget)
		default:
			return fmt.Errorf("file %s type %s is unsupported", name, resp.Type)
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if buf != nil {
			if _, err := buf.WriteTo(tw); err != nil {
				return err
			}
		}
		return nil
	}
	return h
}

func (sb *sandbox) newUntarHandler(root string) xtar.UntarHandler {
	h := func(hdr *tar.Header, r io.Reader) error {
		path := filepath.Join(root, hdr.Name)
		args := incusclient.InstanceFileArgs{
			UID:       int64(hdr.Uid),
			GID:       int64(hdr.Gid),
			Mode:      int(hdr.Mode),
			WriteMode: "overwrite",
		}

		switch hdr.Typeflag {
		case tar.TypeReg:
			args.Type = "file"
			if b, err := io.ReadAll(r); err != nil {
				return err
			} else {
				args.Content = bytes.NewReader(b)
			}
		case tar.TypeDir:
			args.Type = "directory"
		case tar.TypeSymlink:
			args.Type = "symlink" // BUG from incus, symlink always have uid=0 gid=0
			args.Content = strings.NewReader(hdr.Linkname)
		default:
			return fmt.Errorf("file %s type %b is unsupported", hdr.Name, hdr.Typeflag)
		}

		return sb.client.CreateInstanceFile(sb.instanceName, path, args)
	}
	return h
}
