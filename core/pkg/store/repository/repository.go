package repository

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

type Repositorial interface {
	Repository() *Repository
}

type Repository struct {
	Scheme    string // e.g. git
	Transport string // e.g. http, https, ssh
	Endpoint  string // e.g. github.com
	Name      string // e.g. actions/checkout
	Path      string // e.g .github/actions/hello-world-action
	Ref       string // e.g. v3
}

func (r *Repository) Key() string {
	return fmt.Sprintf("%s/%s", r.Endpoint, r.Name)
}

func Parse(s string) (*Repository, error) {
	match := useRegex.FindStringSubmatch(s)
	if match == nil {
		return nil, fmt.Errorf("invalid action %q", s)
	}

	scheme, transport := "", ""
	protocol := match[1]
	endpoint, name, path := "", "", ""
	location := match[2]
	ref := match[3]

	proParts := strings.SplitN(protocol, "+", 3)
	switch len(proParts) {
	case 2:
		scheme, transport = proParts[0], proParts[1]
		if scheme == "" || transport == "" {
			return nil, fmt.Errorf("invalid action scheme %q", protocol)
		}
	case 1:
		if protocol == "http" || protocol == "https" || protocol == "ssh" {
			scheme, transport = "git", protocol
		} else {
			scheme = protocol
		}
	default:
		return nil, fmt.Errorf("invalid action scheme %q", protocol)
	}

	locParts := strings.Split(location, "/")
	if len(locParts) < 2 {
		return nil, fmt.Errorf("invalid action name %q", s)
	}
	if isValidHostPort(locParts[0]) {
		if len(locParts) < 3 {
			return nil, fmt.Errorf("invalid action name %q", s)
		}
		endpoint = locParts[0]
		name = strings.Join(locParts[1:3], "/")
		path = strings.Join(locParts[3:], "/")
	} else {
		if protocol != "" {
			return nil, fmt.Errorf("invalid action name %q", s)
		}
		name = strings.Join(locParts[:2], "/")
		path = strings.Join(locParts[2:], "/")
	}

	r := Repository{
		Scheme:    scheme,
		Transport: transport,
		Endpoint:  endpoint,
		Name:      name,
		Path:      path,
		Ref:       ref,
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
