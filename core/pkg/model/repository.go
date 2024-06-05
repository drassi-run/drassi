package model

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

var (
	useRegex      = regexp.MustCompile(`^(?:(?P<scheme>[^:]+)://)?(?P<location>[^@]+)@(?P<ref>.+)$`)
	hostPortRegex = regexp.MustCompile(`^(?P<host>[^:\s]+)(?::(?P<port>\d{1,5}))?$`)
)

type Repository struct {
	Protocol string // e.g. https
	Endpoint string // e.g. github.com
	Repo     string // e.g. actions/checkout
	Path     string // e.g .github/actions/hello-world-action
	Ref      string // e.g. v3
}

func ParseRepository(s string) (*Repository, error) {
	match := useRegex.FindStringSubmatch(s)
	if match == nil {
		return nil, fmt.Errorf("invalid uses %q", s)
	}

	protocol := match[1]
	endpoint, repo, path := "", "", ""
	location := match[2]
	ref := match[3]

	locPart := strings.Split(location, "/")
	if len(locPart) < 2 {
		return nil, fmt.Errorf("invalid uses %q", s)
	}
	if isValidHostPort(locPart[0]) {
		if len(locPart) < 3 {
			return nil, fmt.Errorf("invalid uses %q", s)
		}
		endpoint = locPart[0]
		repo = strings.Join(locPart[1:3], "/")
		path = strings.Join(locPart[3:], "/")
	} else {
		if protocol != "" {
			return nil, fmt.Errorf("invalid uses %q", s)
		}
		repo = strings.Join(locPart[:2], "/")
		path = strings.Join(locPart[2:], "/")
	}

	r := Repository{
		Protocol: protocol,
		Endpoint: endpoint,
		Repo:     repo,
		Path:     path,
		Ref:      ref,
	}
	return &r, nil
}

func isValidHostPort(s string) bool {
	match := hostPortRegex.FindStringSubmatch(s)
	if match == nil {
		return false
	}
	host, port := match[1], match[2]
	return port != "" ||
		net.ParseIP(host) != nil || // check host is IPv4 or IPv6
		strings.Contains(host, ".") // simple way to check isDomain
}
