package incus

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/archive"
	"github.com/dungdm93/drasi/pkg/sandboxer"
	"github.com/gorilla/websocket"
	incusclient "github.com/lxc/incus/client"
	incusapi "github.com/lxc/incus/shared/api"
)

type incus struct {
	config *Incus
	client incusclient.InstanceServer
}

func NewIncusSandboxer(config *Incus) sandboxer.Sandboxer {
	i := &incus{
		config: config,
	}
	return i
}

// see: cliconfig.GetInstanceServer
func (i *incus) Connect(ctx context.Context) error {
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

func (i *incus) Close() error {
	i.client.Disconnect()
	return nil
}

func (i *incus) LaunchSandbox(ctx context.Context, request sandboxer.LaunchSandboxRequest) (sandboxer.LaunchSandboxResponse, error) {
	name := "foobar"
	res := sandboxer.LaunchSandboxResponse{}

	err := i.createInstance(ctx, name, request)
	if err != nil {
		return res, err
	}

	res.Sandbox = &incusSandbox{
		sandboxId: name,
		sandboxer: i,
	}
	return res, nil
}

func (i *incus) createInstance(ctx context.Context, name string, request sandboxer.LaunchSandboxRequest) error {
	template := i.config.Spec.Template
	req := incusapi.InstancesPost{
		Name:         name,
		Start:        true,
		Source:       template.Source,
		Type:         template.Type,
		InstanceType: template.InstanceSize,
		InstancePut: incusapi.InstancePut{
			Architecture: template.Architecture,
			Config:       template.Config,
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

func (i *incus) TerminateSandbox(ctx context.Context, request sandboxer.TerminateSandboxRequest) (sandboxer.TerminateSandboxResponse, error) {
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

func (i *incus) ExecuteSandbox(ctx context.Context, request sandboxer.ExecuteSandboxRequest) (sandboxer.ExecuteSandboxResponse, error) {
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
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
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

func (i *incus) CopyFromSandbox(ctx context.Context, request sandboxer.CopyFromSandboxRequest) (sandboxer.CopyFromSandboxResponse, error) {
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

func (i *incus) CopyToSandbox(ctx context.Context, request sandboxer.CopyToSandboxRequest) (sandboxer.CopyToSandboxResponse, error) {
	res := sandboxer.CopyToSandboxResponse{}

	handler := newUntarHandler(i.client, request.SandboxId, request.DestinationPath)
	err := untar(request.Content, handler)
	return res, err
}

type tarHandler = func(string, io.Reader, *incusclient.InstanceFileResponse) error
type untarHandler = func(*tar.Header, io.Reader) error

func (i *incus) walk(inst, root, name string, h tarHandler) error {
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

func untar(r io.Reader, h untarHandler) error {
	xr, err := archive.DecompressStream(r)
	if err != nil {
		return err
	}
	defer xr.Close()
	tr := tar.NewReader(xr)

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break // end of tar archive
			}
			return err
		}

		if err := h(hdr, tr); err != nil {
			return err
		}
	}

	return nil
}

func newTarHandler(tw *tar.Writer) tarHandler {
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

func newUntarHandler(client incusclient.InstanceServer, inst, root string) untarHandler {
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
