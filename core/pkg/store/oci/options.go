/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"io"

	"go.podman.io/image/v5/signature"
	"go.podman.io/image/v5/types"
)

type MountOption func(*mountOptions)
type mountOptions struct {
	Writable bool
}

func WithWritable(writable bool) MountOption {
	return func(o *mountOptions) {
		o.Writable = writable
	}
}

type PullOption func(*pullOptions)
type pullOptions struct {
	SystemContext *types.SystemContext
	PolicyContext *signature.PolicyContext
	ReportWriter  io.Writer
}

func (o *pullOptions) SysCtx() *types.SystemContext {
	if o.SystemContext == nil {
		o.SystemContext = new(types.SystemContext)
	}
	return o.SystemContext
}

func (o *pullOptions) PolicyCtx() (pc *signature.PolicyContext, ephemeral bool, err error) {
	if pc = o.PolicyContext; pc != nil {
		return
	}

	policy, err := signature.DefaultPolicy(o.SystemContext)
	if err != nil {
		policy = &signature.Policy{
			Default: []signature.PolicyRequirement{
				signature.NewPRInsecureAcceptAnything(),
			},
		}
	}
	pc, err = signature.NewPolicyContext(policy)
	ephemeral = true
	return
}

func WithAuth(username, password string) PullOption {
	return func(o *pullOptions) {
		o.SysCtx().DockerAuthConfig = &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}
	}
}

func WithAuthFile(path string) PullOption {
	return func(o *pullOptions) {
		o.SysCtx().AuthFilePath = path
	}
}

func WithReportWriter(w io.Writer) PullOption {
	return func(o *pullOptions) {
		o.ReportWriter = w
	}
}

func WithInsecureSkipTLSVerify(insecure bool) PullOption {
	return func(o *pullOptions) {
		o.SysCtx().DockerInsecureSkipTLSVerify = types.NewOptionalBool(insecure)
	}
}

func WithArchitecture(arch string) PullOption {
	return func(o *pullOptions) {
		o.SysCtx().ArchitectureChoice = arch
	}
}

func WithOS(os string) PullOption {
	return func(o *pullOptions) {
		o.SysCtx().OSChoice = os
	}
}

func WithPlatform(os, arch string) PullOption {
	return func(o *pullOptions) {
		sys := o.SysCtx()
		sys.OSChoice = os
		sys.ArchitectureChoice = arch
	}
}

func WithSystemContext(sysCtx *types.SystemContext) PullOption {
	return func(o *pullOptions) {
		o.SystemContext = sysCtx
	}
}

func WithPolicyContext(policyCtx *signature.PolicyContext) PullOption {
	return func(o *pullOptions) {
		o.PolicyContext = policyCtx
	}
}
