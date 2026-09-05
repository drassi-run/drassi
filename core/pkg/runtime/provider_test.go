/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"testing"

	"drassi.run/core/config"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestProviderSuite(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}

type ProviderTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	mockSb *mock_sandboxer.MockSandbox
}

func (s *ProviderTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockSb = mock_sandboxer.NewMockSandbox(s.ctrl)
	s.mockSb.EXPECT().Layout().Return(&sandboxer.Layout{
		Runtimes: "/opt/drassi/runtimes",
	}).AnyTimes()
}

func (s *ProviderTestSuite) TestGet() {
	s.mockSb.EXPECT().Execute(
		gomock.Any(),
		[]string{"/opt/drassi/runtimes/node/bin/node", "/workspace/app.js"},
		[]string{"/extra/bin"},
		map[string]string{"NODE_ENV": "production"},
		"/workspace",
		nil,
	).Return(nil)

	s.mockSb.EXPECT().Execute(
		gomock.Any(),
		[]string{"/opt/drassi/runtimes/python/python", "/workspace/main.py"},
		nil,
		nil,
		"/workspace",
		nil,
	).Return(nil)

	runtimes := map[string]*config.Runtime{
		"node": {
			Image:      "drassi/node:24",
			Alias:      []string{"node20", "node22"},
			Executable: "./bin/node",
			Cmd:        []string{"{0}"},
			Paths:      []string{"bin"},
		},
		"python": {
			Image:      "drassi/python:3.12",
			Alias:      []string{"py3", "python3"},
			Executable: "python",
			Cmd:        nil,
			Paths:      nil,
		},
	}

	p, err := NewProvider(s.mockSb, runtimes)
	s.Require().NoError(err)
	s.Require().NotNil(p)

	// Direct name
	rt, err := p.Get("node")
	s.Require().NoError(err)
	s.Require().NotNil(rt)

	// Alias name returns same pre-initialized instance
	rt20, err := p.Get("node20")
	s.Require().NoError(err)
	s.Require().Same(rt, rt20)

	// Unknown runtime
	_, err = p.Get("unknown-rt")
	s.Require().Error(err)
	s.Assert().ErrorContains(err, "unsupported runtime")

	// Execute bound runtime
	err = rt.Run(s.T().Context(), "/workspace/app.js", []string{"/extra/bin"}, map[string]string{"NODE_ENV": "production"}, "/workspace", nil)
	s.Require().NoError(err)

	// Execute bound runtime with empty Cmd (fallback to scriptPath appended)
	rtPy, err := p.Get("py3")
	s.Require().NoError(err)
	s.Require().NotNil(rtPy)

	err = rtPy.Run(s.T().Context(), "/workspace/main.py", nil, nil, "/workspace", nil)
	s.Require().NoError(err)
}

func (s *ProviderTestSuite) TestNew() {
	tests := map[string]struct {
		runtimes map[string]*config.Runtime
		errMsg   string
	}{
		"duplicate alias": {
			runtimes: map[string]*config.Runtime{
				"node": {
					Image: "drassi/node:24",
					Alias: []string{"node20"},
				},
				"node-alt": {
					Image: "drassi/node:24-alt",
					Alias: []string{"node20"},
				},
			},
			errMsg: "duplicate runtime alias",
		},
		"alias collides with runtime name": {
			runtimes: map[string]*config.Runtime{
				"node": {
					Image: "drassi/node:24",
				},
				"python": {
					Image: "drassi/python:3.12",
					Alias: []string{"node"},
				},
			},
			errMsg: "conflicts with existing runtime name",
		},
		"cmd missing placeholder": {
			runtimes: map[string]*config.Runtime{
				"node": {
					Image:      "drassi/node:24",
					Executable: "node",
					Cmd:        []string{"--version"},
				},
			},
			errMsg: "must contain \"{0}\" placeholder",
		},
	}

	for name, tt := range tests {
		s.Run(name, func() {
			_, err := NewProvider(s.mockSb, tt.runtimes)
			s.Require().Error(err)
			s.Assert().ErrorContains(err, tt.errMsg)
		})
	}
}
