package provision

import (
	"fmt"

	"drassi.run/core/pkg/sandboxer"
	ocistore "drassi.run/core/pkg/store/oci"
)

const (
	KeyHostMountDir = StateKey[string]("host_mount_dir")
	KeyMountID      = StateKey[string]("mount_id")
	KeyImage        = StateKey[*ocistore.Image]("image")
)

type Operation interface {
	Name() string
	PreLaunch(pctx *Context) error
	PostLaunch(pctx *Context, sb sandboxer.Sandbox) error
}

type Noop struct{}

func (n Noop) Name() string                                     { return "noop" }
func (n Noop) PreLaunch(_ *Context) error                       { return nil }
func (n Noop) PostLaunch(_ *Context, _ sandboxer.Sandbox) error { return nil }

type pullOp struct {
	Noop
	mgr ocistore.Manager
}

// Pull returns an Operation that checks if the configured image is locally available,
// and pulls it using the provided ocistore.Manager if missing.
func Pull(mgr ocistore.Manager) Operation {
	return &pullOp{mgr: mgr}
}

func (op *pullOp) Name() string { return "pull" }

func (op *pullOp) PreLaunch(pctx *Context) error {
	img, err := op.mgr.Image(pctx, pctx.Config.Image)
	if err != nil {
		return fmt.Errorf("check image %q: %w", pctx.Config.Image, err)
	}
	if img == nil {
		if img, err = op.mgr.Pull(pctx, pctx.Config.Image); err != nil {
			return fmt.Errorf("pull image %q: %w", pctx.Config.Image, err)
		}
	}
	pctx.Set(KeyImage, img)
	return nil
}

type mountOp struct {
	Noop
	mgr  ocistore.Manager
	opts []ocistore.MountOption
}

// Mount returns an Operation that mounts the configured runtime image with the given MountOptions,
// and records KeyHostMountDir and KeyMountID in the Context.
func Mount(mgr ocistore.Manager, opts ...ocistore.MountOption) Operation {
	return &mountOp{mgr: mgr, opts: opts}
}

func (op *mountOp) Name() string { return "mount" }

func (op *mountOp) PreLaunch(pctx *Context) error {
	img, ok := pctx.Get(KeyImage)
	if !ok || img == nil {
		return fmt.Errorf("image %q not found in context: pull operation must be used first", pctx.Config.Image)
	}
	mountDir, id, err := op.mgr.Mount(pctx, img, op.opts...)
	if err != nil {
		return fmt.Errorf("mount image %q: %w", pctx.Config.Image, err)
	}
	pctx.Set(KeyHostMountDir, mountDir)
	pctx.Set(KeyMountID, id)
	return nil
}
