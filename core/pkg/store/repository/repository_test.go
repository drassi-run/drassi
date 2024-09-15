package repository

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("scheme", testParseScheme)
	t.Run("endpoint", testEndpoint)
}

func testParseScheme(t *testing.T) {
	t.Run("success", func(tt *testing.T) {
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
			assert.NoError(tt, err, test.input)
			assert.Equal(tt, test.scheme, repo.Scheme, test.input)
			assert.Equal(tt, test.transport, repo.Transport, test.input)
		}
	})
	t.Run("failure", func(tt *testing.T) {
		tests := []string{
			"+://github.com/action/checkout@main",
			"git+://github.com/action/checkout@main",
			"git+https+abc://github.com/action/checkout@main",
		}
		for _, test := range tests {
			_, err := Parse(test)
			assert.Error(tt, err, test)
		}
	})
}

func testEndpoint(t *testing.T) {
	t.Run("success", func(tt *testing.T) {
		endpoints := []string{
			"",
			"gitserver.com", "gitserver.com:8080",
			"1.2.3.4", "1.2.3.4:8080",
			//"2002::1", "[2002::1]:8080",
		}
		paths := []string{"", "path/to/action"}
		for _, ep := range endpoints {
			for _, p := range paths {
				input := ep + "/action/checkout/" + p
				input = strings.Trim(input, "/") + "@main"
				repo, err := Parse(input)
				assert.NoError(tt, err, input)
				assert.Equal(tt, ep, repo.Endpoint, input)
				assert.Equal(tt, "action/checkout", repo.Name, input)
				assert.Equal(tt, p, repo.Path, input)
			}
		}
	})
	t.Run("failure", func(tt *testing.T) {
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
			assert.Error(tt, err, test)
		}
	})
}
