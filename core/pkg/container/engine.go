/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"io"
	"io/fs"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/stream"
)

type Engine interface {
	io.Closer
	Address() string

	ImagePull(ctx context.Context, ref string, opts *PullOptions) error
	ImageBuild(ctx context.Context, context io.Reader, opts *BuildOptions) error

	ContainerRun(ctx context.Context, spec *types.ContainerSpec, opts *RunOptions) (string, error)
	ContainerExec(ctx context.Context, id string, opts *ExecOptions) (string, error)
	ContainerRemove(ctx context.Context, opts *RemoveOptions) error
	ContainerInspect(ctx context.Context, id string) (*types.ContainerSpec, error)

	Stat(ctx context.Context, id string, path string) (fs.FileInfo, error)
	CopyIn(ctx context.Context, id string, opts *CopyInOptions) error
	CopyOut(ctx context.Context, id string, opts *CopyOutOptions) (io.ReadCloser, error)

	NetworkCreate(ctx context.Context, spec *types.NetworkSpec) (string, error)
	NetworkRemove(ctx context.Context, opts *RemoveOptions) error

	VolumeCreate(ctx context.Context, spec *types.VolumeSpec) (string, error)
	VolumeRemove(ctx context.Context, opts *RemoveOptions) error
}

type PullOptions struct {
	PullPolicy   string // "always", "missing" (default), "never"
	RegistryAuth RegistryAuth
	Streams      stream.Streams
}

type BuildOptions struct {
	Tags    []string
	Streams stream.Streams
}

type RunOptions struct {
	Stdio   *types.Stdio
	Streams stream.Streams
}

type ExecOptions struct {
	Cmd     []string
	Env     map[string]string
	Workdir string
	Stdio   *types.Stdio
	Streams stream.Streams
}

type RemoveOptions struct {
	Id     string
	Labels map[string]string
}

type CopyInOptions struct {
	Reader          io.Reader
	DestinationPath string
}

type CopyOutOptions struct {
	SourcePath string
}

type RegistryAuth interface {
	// Credential returns the base64 encoded credentials for the registry
	Credential() string
}

func NewBasicAuth(username, password string) RegistryAuth {
	return &basicAuth{
		username: username,
		password: password,
	}
}

type basicAuth struct {
	username, password string
}

func (auth *basicAuth) Credential() string {
	return auth.username + ":" + auth.password
}
