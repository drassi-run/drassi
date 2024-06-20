package incus

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net/url"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	utilreader "drassi.run/core/pkg/util/reader"
	"github.com/gorilla/websocket"
	incusclient "github.com/lxc/incus/client"
	incusapi "github.com/lxc/incus/shared/api"
)

type incusSandboxRuntime struct {
	config *Incus
	client incusclient.InstanceServer
}

func NewSandboxRuntime(config *Incus) sandboxer.SandboxRuntime {
	i := &incusSandboxRuntime{
		config: config,
	}
	return i
}

// see: cliconfig.GetInstanceServer
func (i *incusSandboxRuntime) Connect(ctx context.Context) error {
	u, err := url.Parse(i.config.Spec.Endpoint)
	if err != nil {
		return err
	}
	if u.Scheme != "unix" {
		return fmt.Errorf("only unix endpoint is supported at the moment")
	}
	i.client, err = incusclient.ConnectIncusUnixWithContext(ctx, u.Path, nil)
	return err
}

func (i *incusSandboxRuntime) Close() error {
	i.client.Disconnect()
	return nil
}

func (i *incusSandboxRuntime) LaunchSandbox(ctx context.Context, request sandboxer.LaunchSandboxRequest) (sandboxer.LaunchSandboxResponse, error) {
	name := "foobar"
	res := sandboxer.LaunchSandboxResponse{}

	err := i.createInstance(ctx, name, request)
	if err != nil {
		return res, err
	}

	sandbox := &incusSandbox{
		sandboxId:      name,
		sandboxRuntime: i,
	}

	if err := i.setupWellKnownDirectories(sandbox); err != nil {
		return res, err
	}

	res.Sandbox = sandbox
	return res, nil
}

func (i *incusSandboxRuntime) createInstance(ctx context.Context, name string, request sandboxer.LaunchSandboxRequest) error {
	template := i.config.Spec.Template

	config := maps.Clone(template.Config)
	if config == nil {
		config = make(map[string]string)
	}
	for k, v := range request.JobEnv {
		k = fmt.Sprintf("environment.%s", k)
		config[k] = v
	}
	for k, v := range i.getPredefinedEnv() {
		k = fmt.Sprintf("environment.%s", k)
		config[k] = v
	}

	req := incusapi.InstancesPost{
		Name:         name,
		Start:        true,
		Source:       template.Source,
		Type:         template.Type,
		InstanceType: template.InstanceSize,
		InstancePut: incusapi.InstancePut{
			Architecture: template.Architecture,
			Config:       config,
			Devices:      template.Devices,
			Ephemeral:    template.Ephemeral,
			Profiles:     template.Profiles,
			Restore:      template.Restore,
			Stateful:     template.Stateful,
			Description:  template.Description,
		},
	}
	if op, err := i.client.CreateInstance(req); err != nil {
		return err
	} else if err = op.WaitContext(ctx); err != nil {
		return err
	}
	return nil
}

func (i *incusSandboxRuntime) setupWellKnownDirectories(sandbox *incusSandbox) error {
	args := incusclient.InstanceFileArgs{
		Type: "directory",
	}

	rootDir := "/opt/drassi"
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, rootDir, args); err != nil {
		return err
	}

	sandbox.workspaceDir = filepath.Join(rootDir, "workspace")
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, sandbox.workspaceDir, args); err != nil {
		return err
	}

	sandbox.actionsDir = filepath.Join(rootDir, "actions")
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, sandbox.actionsDir, args); err != nil {
		return err
	}

	sandbox.toolsDir = filepath.Join(rootDir, "tools")
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, sandbox.toolsDir, args); err != nil {
		return err
	}

	sandbox.tempDir = filepath.Join(rootDir, "tmp")
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, sandbox.tempDir, args); err != nil {
		return err
	}
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, filepath.Join(sandbox.tempDir, "file_commands"), args); err != nil {
		return err
	}
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, filepath.Join(sandbox.tempDir, "workflow"), args); err != nil {
		return err
	}
	if err := i.client.CreateInstanceFile(sandbox.sandboxId, filepath.Join(sandbox.tempDir, "scripts"), args); err != nil {
		return err
	}

	return nil
}

func (i *incusSandboxRuntime) getPredefinedEnv() map[string]string {
	env := map[string]string{
		"GITHUB_STEP_SUMMARY": "/opt/drassi/tmp/file_commands/GITHUB_STEP_SUMMARY",
		"GITHUB_OUTPUT":       "/opt/drassi/tmp/file_commands/GITHUB_OUTPUT",
		"GITHUB_STATE":        "/opt/drassi/tmp/file_commands/GITHUB_STATE",
		"GITHUB_PATH":         "/opt/drassi/tmp/file_commands/GITHUB_PATH",
		"GITHUB_ENV":          "/opt/drassi/tmp/file_commands/GITHUB_ENV",
	}
	return env
}

func (i *incusSandboxRuntime) TerminateSandbox(ctx context.Context, request sandboxer.TerminateSandboxRequest) (sandboxer.TerminateSandboxResponse, error) {
	res := sandboxer.TerminateSandboxResponse{}
	sandbox, ok := request.Sandbox.(*incusSandbox)
	if !ok {
		return res, fmt.Errorf("unsupport sandbox type %T", request.Sandbox)
	}

	name := sandbox.sandboxId
	ct, _, err := i.client.GetInstance(name)
	if err != nil {
		return res, err
	}

	if (ct.StatusCode != 0) && (ct.StatusCode != incusapi.Stopped) {
		req := incusapi.InstanceStatePut{
			Action: "stop",
			Force:  true,
		}
		if request.Timeout != nil {
			req.Timeout = *request.Timeout
		}
		if op, err := i.client.UpdateInstanceState(name, req, ""); err != nil {
			return res, err
		} else if err := op.WaitContext(ctx); err != nil {
			return res, fmt.Errorf("stopping the instance %s failed: %s", name, err)
		}

		if ct.Ephemeral {
			return res, nil
		}
	}

	if op, err := i.client.DeleteInstance(name); err != nil {
		return res, err
	} else if err := op.WaitContext(ctx); err != nil {
		return res, fmt.Errorf("failed deleting instance %s: %s", name, err)
	}
	return res, nil
}

func (i *incusSandboxRuntime) ExecuteSandbox(ctx context.Context, request sandboxer.ExecuteSandboxRequest) (sandboxer.ExecuteSandboxResponse, error) {
	res := sandboxer.ExecuteSandboxResponse{}

	// Prepare the command
	req := incusapi.InstanceExecPost{
		Command:     request.Cmd,
		WaitForWS:   true,
		Interactive: false,
		Environment: request.Env,
		Cwd:         request.Workdir,
		//User:        request.User,
		//Group:       request.Group,
		Width:  0,
		Height: 0,
	}

	execArgs := incusclient.InstanceExecArgs{
		Stdin:    request.Streams.In,
		Stdout:   request.Streams.Out,
		Stderr:   request.Streams.Err,
		Control:  func(conn *websocket.Conn) {}, // TODO
		DataDone: make(chan bool),
	}

	// Run the command in the instance
	if op, err := i.client.ExecInstance(request.SandboxId, req, &execArgs); err != nil {
		return res, err
	} else if err := op.WaitContext(ctx); err != nil {
		return res, err
	} else {
		opAPI := op.Get()
		if exitCode, ok := opAPI.Metadata["return"].(int); ok {
			res.ExitCode = exitCode
		}
	}

	if res.ExitCode != 0 {
		return res, fmt.Errorf("exitcode '%d': failure", res.ExitCode)
	} else {
		return res, nil
	}
}

func (i *incusSandboxRuntime) CopyFromSandbox(ctx context.Context, request sandboxer.CopyFromSandboxRequest) (sandboxer.CopyFromSandboxResponse, error) {
	res := sandboxer.CopyFromSandboxResponse{}
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	h := newTarHandler(tw)
	if err := i.walk(request.SandboxId, request.SourcePath, "", h); err != nil {
		return res, err
	}

	res.Reader = io.NopCloser(buf)
	return res, nil
}

func (i *incusSandboxRuntime) CopyToSandbox(ctx context.Context, request sandboxer.CopyToSandboxRequest) (sandboxer.CopyToSandboxResponse, error) {
	res := sandboxer.CopyToSandboxResponse{}

	handler := newUntarHandler(i.client, request.SandboxId, request.DestinationPath)
	err := utilreader.Untar(request.Content, handler)
	return res, err
}

type fileHandler = func(string, io.Reader, *incusclient.InstanceFileResponse) error

func (i *incusSandboxRuntime) walk(inst, root, name string, h fileHandler) error {
	path := filepath.Join(root, name)
	buf, resp, err := i.client.GetInstanceFile(inst, path)
	if err != nil {
		return err
	}
	if buf != nil {
		defer buf.Close()
	}

	if err := h(name, buf, resp); err != nil {
		return err
	}

	if resp.Type == "directory" {
		for _, ent := range resp.Entries {
			next := filepath.Join(name, ent)
			if err := i.walk(inst, root, next, h); err != nil {
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

func newUntarHandler(client incusclient.InstanceServer, inst, root string) utilreader.UntarHandler {
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

		return client.CreateInstanceFile(inst, path, args)
	}
	return h
}
