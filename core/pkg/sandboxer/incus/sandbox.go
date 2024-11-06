package incus

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/fs/sftpfs"
	"drassi.run/core/util/io"
	"drassi.run/core/util/path"
	"github.com/gorilla/websocket"
	incusclient "github.com/lxc/incus/v6/client"
	incusapi "github.com/lxc/incus/v6/shared/api"
)

type sandbox struct {
	client       incusclient.InstanceServer
	fsys         *sftpfs.SftpFS
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

	// init fsys
	if sftpClient, err := client.GetInstanceFileSFTP(inst); err != nil {
		return nil, err
	} else {
		sb.fsys = sftpfs.New(sftpClient)
	}

	// init layout
	layout := &sb.layout
	if err := sb.fsys.MkdirAll(dir, xfs.DirPerm); err != nil {
		return nil, err
	}
	dirs := map[string]fs.FileMode{
		layout.Workspace: xfs.DirPerm,
		layout.Temp:      xfs.AllPerm,
		layout.Actions:   xfs.DirPerm,
		layout.Tools:     xfs.DirPerm,
	}
	for d, pem := range dirs {
		if err := sb.fsys.Mkdir(d, pem); err != nil {
			return nil, err
		}
	}

	return sb, nil
}

func (sb *sandbox) Layout() *sandboxer.Layout {
	return &sb.layout
}

func (sb *sandbox) Stat(_ context.Context, path string) (fs.FileInfo, error) {
	return sb.fsys.Stat(path)
}

func (sb *sandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	return xfs.Write(ctx, sb.fsys, reader, dst)
}

func (sb *sandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	r := xfs.Read(ctx, sb.fsys, src)
	return r, nil
}

func (sb *sandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams sandboxer.Streams) error {
	if op, err := sb.execute(ctx, cmd, path, env, workdir, streams); err != nil {
		return err
	} else if err = op.WaitContext(ctx); err != nil {
		return err
	} else if exitCode, ok := op.Get().Metadata["return"]; ok && exitCode != float64(0) {
		return fmt.Errorf("exitcode '%v': failure", exitCode)
	}
	return nil
}

func (sb *sandbox) execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams sandboxer.Streams) (incusclient.Operation, error) {
	// Prepare the command
	req := incusapi.InstanceExecPost{
		Command:   cmd,
		WaitForWS: true,
		// TODO: Interactive=true if both stdin AND stdout are terminals (stderr is ignored).
		// See: https://github.com/lxc/incus/blob/v6.6.0/cmd/incus/exec.go#L141-L157
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
		if req.Environment == nil {
			req.Environment = make(map[string]string)
		}
		req.Environment["PATH"] = p
	}

	// workdir
	if workdir == "" {
		req.Cwd = sb.layout.Workspace
	} else {
		req.Cwd = xpath.Abs(workdir, sb.layout.Workspace)
	}

	// incus streams stdin/out/err not respect ctx
	// => So we need to wrap them in ContextReader/Writer
	// NOTE: executing command will exit when all streams terminated
	var (
		stdin  = streams.In()
		stdout = streams.Out()
		stderr = streams.Err()
	)
	if stdin != nil {
		stdin = xio.NewContextReader(ctx, stdin)
	}
	if stdout != nil {
		stdout = xio.NewContextWriter(ctx, stdout)
	}
	if stderr != nil {
		stderr = xio.NewContextWriter(ctx, stderr)
	}

	execArgs := &incusclient.InstanceExecArgs{
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Control: sb.controlSocketHandler,
	}

	// Run the command in the instance
	return sb.client.ExecInstance(sb.instanceName, req, execArgs)
}

// TODO: implement resize terminal & forward signal
// See: https://github.com/lxc/incus/blob/v6.6.0/cmd/incus/exec_unix.go#L20
func (sb *sandbox) controlSocketHandler(control *websocket.Conn) {
}

func (sb *sandbox) Terminate(ctx context.Context) error {
	_ = sb.fsys.Close()

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
