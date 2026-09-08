/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gitstore

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

type RepositoryTestSuite struct {
	suite.Suite
}

func (s *RepositoryTestSuite) TestParseScheme() {
	s.Run("success", func() {
		tests := []struct {
			input     string
			scheme    string
			transport string
		}{
			{input: "action/checkout@main"},
			{input: "github.com/action/checkout@main"},

			{input: "http://github.com/action/checkout@main", scheme: "git", transport: "http"},
			{input: "https://github.com/action/checkout@main", scheme: "git", transport: "https"},
			{input: "ssh://github.com/action/checkout@main", scheme: "git", transport: "ssh"},

			{input: "git://github.com/action/checkout@main", scheme: "git"},
			{input: "git+http://github.com/action/checkout@main", scheme: "git", transport: "http"},
			{input: "git+https://github.com/action/checkout@main", scheme: "git", transport: "https"},
			{input: "git+ssh://github.com/action/checkout@main", scheme: "git", transport: "ssh"},

			{input: "foobar://github.com/action/checkout@main", scheme: "foobar"},
			{input: "foobar+trans://github.com/action/checkout@main", scheme: "foobar", transport: "trans"},
		}
		for _, test := range tests {
			repo, err := Parse(test.input)
			s.Require().NoError(err, test.input)
			s.Assert().Equal(test.scheme, repo.Scheme, test.input)
			s.Assert().Equal(test.transport, repo.Transport, test.input)
		}
	})

	s.Run("failure", func() {
		tests := []string{
			"+://github.com/action/checkout@main",
			"git+://github.com/action/checkout@main",
			"git+https+abc://github.com/action/checkout@main",
		}
		for _, test := range tests {
			_, err := Parse(test)
			s.Assert().Error(err, test)
		}
	})
}

func (s *RepositoryTestSuite) TestEndpoint() {
	s.Run("success", func() {
		endpoints := []string{
			"",
			"gitserver.com", "gitserver.com:8080",
			"1.2.3.4", "1.2.3.4:8080",
		}
		paths := []string{"", "path/to/action"}
		for _, ep := range endpoints {
			for _, p := range paths {
				input := ep + "/action/checkout/" + p
				input = strings.Trim(input, "/") + "@main"
				repo, err := Parse(input)
				s.Require().NoError(err, input)
				s.Assert().Equal(ep, repo.Endpoint, input)
				s.Assert().Equal("action/checkout", repo.Name, input)
				s.Assert().Equal(p, repo.Path, input)
			}
		}
	})

	s.Run("failure", func() {
		tests := []string{
			"actions@main",
			"gitserver.com/actions@main",
			"gitserver.com:8080/actions@main",
			"1.2.3.4/actions@main",
			"1.2.3.4:8080/actions@main",
			//"2002::1/actions@main",
			//"[2002::1]:8080/actions@main",
		}
		for _, test := range tests {
			_, err := Parse(test)
			s.Assert().Error(err, test)
		}
	})
}

func (s *RepositoryTestSuite) TestHelpers() {
	ref := &RepoReference{
		Endpoint:  "custom.git.org",
		Name:      "actions/checkout",
		Path:      "action.yml",
		Ref:       "v3",
		Transport: "https",
	}

	s.Assert().Equal("custom.git.org", Endpoint(ref))
	s.Assert().Equal("custom.git.org/actions/checkout", FullName(ref))
	s.Assert().Equal("https://custom.git.org/actions/checkout", Url(ref))
	s.Assert().Equal("custom.git.org/actions/checkout@v3/action.yml", Location(ref))

	defaultRef := &RepoReference{
		Name: "actions/checkout",
		Ref:  "main",
	}
	s.Assert().Equal("github.com", Endpoint(defaultRef))
	s.Assert().Equal("github.com/actions/checkout", FullName(defaultRef))
	s.Assert().Equal("https://github.com/actions/checkout", Url(defaultRef))
}
