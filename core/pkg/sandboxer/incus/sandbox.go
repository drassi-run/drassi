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
	"drassi.run/core/util/path"
	"drassi.run/core/util/sftp"
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
	return xsftp.Write(ctx, sb.sftpClient, reader, dst)
}

func (sb *sandbox) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	r := xsftp.Read(ctx, sb.sftpClient, src)
	return r, nil
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
