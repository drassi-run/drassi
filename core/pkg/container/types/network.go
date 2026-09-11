/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
)

type ContainerNetwork struct {
	Exposes    []*Port        `json:"expose,omitempty"`      // informs Docker that the container listens on the specified network ports at runtime
	Publish    []*PortBinding `json:"ports,omitempty"`       // Publish a container's port, or range of ports, to the host.
	PublishAll bool           `json:"publish_all,omitempty"` // Publish all exposed ports to random ports on the host interfaces.

	DNS       *DNS        `json:"dns,omitempty"`
	Endpoints []*Endpoint `json:"endpoints,omitempty"`
}

// PortBinding define the port mappings between the host machine and the container.
//   - [github.com/docker/go-connections/nat.PortMap]
//   - [github.com/containers/common/libnetwork/types.PortMapping]
//   - [github.com/compose-spec/compose-go/v2/types.ServicePortConfig]
type PortBinding struct {
	HostIP        string `json:"host_ip,omitempty"`
	HostPort      uint16 `json:"host_port,omitempty"`
	ContainerPort uint16 `json:"container_port,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}

var ParsePublish func(str string) (*PortBinding, uint16, error)

func (pb *PortBinding) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	case jsontext.KindString:
		var s string
		if err := json.UnmarshalDecode(d, &s); err != nil {
			return err
		}
		if ParsePublish == nil {
			return errors.New("types: publish parser not registered (import _ \"drassi.run/core/pkg/container/parser\")")
		}
		parsed, _, err := ParsePublish(s)
		if err != nil {
			return err
		}
		*pb = *parsed
		return nil
	case jsontext.KindBeginObject:
		type alias PortBinding
		return json.UnmarshalDecode(d, (*alias)(pb))
	default:
		return fmt.Errorf("expected string or object for PortBinding, got %v", k)
	}
}

func (pb *PortBinding) String() string {
	containerPart := strconv.Itoa(int(pb.ContainerPort))
	if pb.Protocol != "" {
		containerPart += "/" + pb.Protocol
	}

	hostPart := pb.HostIP
	if hostPart != "" {
		hostPart += ":"
	}
	if pb.HostPort != 0 {
		hostPart += strconv.Itoa(int(pb.HostPort))
	}
	if hostPart != "" {
		return hostPart + ":" + containerPart
	} else {
		return containerPart
	}
}

// Port defines the (incoming) port and protocol
type Port struct {
	Number   uint16 `json:"number,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

var _ json.UnmarshalerFrom = (*Port)(nil)

func (e *Port) String() string {
	s := strconv.Itoa(int(e.Number))
	if e.Protocol != "" {
		s += "/" + e.Protocol
	}
	return s
}

var ParseExpose func(str string) (*Port, uint16, error)

func (p *Port) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	case jsontext.KindString:
		var s string
		if err := json.UnmarshalDecode(d, &s); err != nil {
			return err
		}
		if ParseExpose == nil {
			return errors.New("types: expose parser not registered (import _ \"drassi.run/core/pkg/container/parser\")")
		}
		parsed, _, err := ParseExpose(s)
		if err != nil {
			return err
		}
		*p = *parsed
		return nil
	case jsontext.KindNumber:
		var num uint16
		if err := json.UnmarshalDecode(d, &num); err != nil {
			return err
		}
		*p = Port{Number: num, Protocol: "tcp"}
		return nil
	case jsontext.KindBeginObject:
		type alias Port
		return json.UnmarshalDecode(d, (*alias)(p))
	default:
		return fmt.Errorf("expected string, number, or object for Port, got %v", k)
	}
}

// Endpoint represents the container's networking configuration for each of its interfaces
//   - [github.com/docker/cli/opts.NetworkAttachmentOpts]
//   - [github.com/moby/moby/api/types/network.NetworkingConfig]
//   - [github.com/containers/common/libnetwork/types.PerNetworkOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceNetworkConfig]
type Endpoint struct {
	Target  string            `json:"target,omitempty"`
	Options map[string]string `json:"options,omitempty"` // driver options

	IPv4Address  netip.Addr       `json:"ipv4_address,omitzero"`
	IPv6Address  netip.Addr       `json:"ipv6_address,omitzero"`
	MacAddress   net.HardwareAddr `json:"mac_address,omitempty"`
	LinkLocalIPs []netip.Addr     `json:"link_local_ips,omitempty"`
	Aliases      []string         `json:"aliases,omitempty"`
	Links        []string         `json:"links,omitempty"`
}

type DNS struct {
	Servers    []netip.Addr        `json:"servers,omitempty"`
	Options    []string            `json:"options,omitempty"`
	Search     []string            `json:"search,omitempty"`
	HostName   string              `json:"hostname,omitempty"`
	DomainName string              `json:"domainname,omitempty"`
	HostAdd    map[string][]string `json:"extra_hosts,omitempty"`
}

// https://github.com/moby/moby/blob/docker-v29.7.2/api/types/network/network.go
// https://github.com/containers/common/blob/v0.60.4/libnetwork/types/network.go#L53-L88
type NetworkSpec struct {
	Name   string            `json:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`

	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`

	IPAMDriver  string            `json:"ipam_driver,omitempty"`
	IPAMOptions map[string]string `json:"ipam_options,omitempty"`
}
