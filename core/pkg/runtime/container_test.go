/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"context"
	"testing"

	mock_container "drassi.run/core/mock/container"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/specdef"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/stream"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestContainerRuntimeSuite(t *testing.T) {
	suite.Run(t, new(ContainerRuntimeTestSuite))
}

type ContainerRuntimeTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	engine *mock_container.MockEngine
}

func (s *ContainerRuntimeTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.engine = mock_container.NewMockEngine(s.ctrl)
}

func (s *ContainerRuntimeTestSuite) TestMountDuplicated() {
	dupSandboxPath := []*types.Mount{
		{Type: "bind", Source: "/abc", Target: "/unique1"},
		{Type: "bind", Source: "/abc/", Target: "/unique2"},
	}
	_, err := NewContainerRuntime(nil, specdef.AddMount(dupSandboxPath...))
	s.ErrorContains(err, "found duplicate sandbox mount at")

	dupContainerPath := []*types.Mount{
		{Type: "bind", Source: "/unique1", Target: "/abc"},
		{Type: "bind", Source: "/unique2", Target: "/abc/"},
	}
	_, err = NewContainerRuntime(nil, specdef.AddMount(dupContainerPath...))
	s.ErrorContains(err, "found duplicate container mount at")
}

var testMounts = []*types.Mount{
	{
		Type:   "bind",
		Source: "/path/to/",
		Target: "/mnt/third",
	},
	{
		Type:   "bind",
		Source: "/path/to/bar",
		Target: "/mnt/second",
	},
	{
		Type:   "bind",
		Source: "/a/new/path",
		Target: "/mnt/second/fourth",
	},
	{
		Type:   "bind",
		Source: "/path/to/foo/",
		Target: "/mnt/second/first",
	},
}

func (s *ContainerRuntimeTestSuite) TestMountSorted() {
	r, err := NewContainerRuntime(nil, specdef.AddMount(testMounts...))
	s.Require().NoError(err)

	rt := r.(*containerRuntime)
	s.Require().Len(rt.mountMap, 4)
	for i := 1; i < len(rt.mountMap); i++ {
		s.True(rt.mountMap[i-1][0] > rt.mountMap[i][0])
	}
	s.Require().Len(rt.pathMap, 4)
	for i := 1; i < len(rt.pathMap); i++ {
		s.True(rt.pathMap[i-1][0] > rt.pathMap[i][0])
	}
}

func (s *ContainerRuntimeTestSuite) TestTranslatePath() {
	rt, err := NewContainerRuntime(nil, specdef.AddMount(testMounts...))
	s.Require().NoError(err)

	s.Run("exact-match", func() {
		m := map[string]string{
			"/mnt/second/first":  "/path/to/foo/",
			"/mnt/second/fourth": "/a/new/path",
			"/mnt/second":        "/path/to/bar",
			"/mnt/third":         "/path/to/",
		}
		for k, v := range m {
			sbPath, ok := rt.TranslatePath(k)
			s.True(ok)
			s.Equal(v, sbPath)

			sbPath, ok = rt.TranslatePath(k + "/")
			s.True(ok)
			s.Equal(v, sbPath)
		}
	})

	s.Run("subpath", func() {
		m := map[string]string{
			"/mnt/second/first/xxx":  "/path/to/foo/xxx",
			"/mnt/second/foobar":     "/path/to/bar/foobar",
			"/mnt/second/f":          "/path/to/bar/f",
			"/mnt/second/firstttttt": "/path/to/bar/firstttttt",
			"/mnt/third/abcxyz":      "/path/to/abcxyz",
		}
		for k, v := range m {
			sbPath, ok := rt.TranslatePath(k)
			s.True(ok)
			s.Equal(v, sbPath)
		}
	})

	s.Run("not-match", func() {
		p := []string{"/mnt/", "/", "/foobar"}
		for _, v := range p {
			_, ok := rt.TranslatePath(v)
			s.False(ok)
		}
	})
}

func (s *ContainerRuntimeTestSuite) TestRun() {
	ctx := context.Background()

	labels := map[string]string{"label": "value"}
	workdir := "/path/to/workdir"
	network := "net01"
	rt, err := NewContainerRuntime(s.engine,
		specdef.SetLabels(labels),
		specdef.SetWorkdir(workdir),
		specdef.SetNetwork(network),
		specdef.AddMount(testMounts...),
	)
	s.Require().NoError(err)

	image := "drassi.run/docker/image"
	env := map[string]string{
		"A_SANDBOX_PATH": "/path/to/foobar",
		"A_NORMAL_ENV":   "hello-world",
	}
	entrypoint := []string{"/path/to/entrypoint.sh"}
	cmd := []string{"--flag", "with", "some", "arg"}
	s.engine.EXPECT().ContainerRun(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, spec *types.ContainerSpec, _ *container.RunOptions) (string, error) {
			s.EqualValues(labels, spec.Labels)
			s.Equal(workdir, spec.WorkingDir)
			s.Equal(network, spec.Endpoints[0].Target)
			s.Equal(image, spec.Image)
			s.Equal(entrypoint, spec.Entrypoint)
			s.Equal(cmd, spec.Command)
			s.True(spec.AutoRemove)

			e := map[string]string{
				"A_NORMAL_ENV":   "hello-world",
				"A_SANDBOX_PATH": "/mnt/third/foobar",
			}
			s.EqualValues(e, spec.Environment)

			expectedMounts := make(map[string]*types.Mount)
			for _, m := range testMounts {
				expectedMounts[m.Target] = m
			}
			actualMounts := make(map[string]*types.Mount)
			for _, m := range spec.Mounts {
				actualMounts[m.Target] = m
			}
			s.Equal(expectedMounts, actualMounts)

			return "container_id", nil
		})

	err = rt.Run(ctx, image, entrypoint, cmd, env, new(stream.Streams))
	s.Require().NoError(err)
}
