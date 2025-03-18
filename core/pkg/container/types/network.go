/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import "strconv"

type ContainerNetwork struct {
	Exposes    []*Port        // informs Docker that the container listens on the specified network ports at runtime
	Publish    []*PortBinding // Publish a container's port, or range of ports, to the host.
	PublishAll bool           // Publish all exposed ports to random ports on the host interfaces.

	DNS       DNS
	Endpoints []*Endpoint
}

// PortBinding define the port mappings between the host machine and the container.
//   - [github.com/docker/go-connections/nat.PortMap]
//   - [github.com/containers/common/libnetwork/types.PortMapping]
//   - [github.com/compose-spec/compose-go/v2/types.ServicePortConfig]
type PortBinding struct {
	HostIP        string
	HostPort      uint16
	ContainerPort uint16
	Protocol      string
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
	Number   uint16
	Protocol string
}

func (e *Port) String() string {
	s := strconv.Itoa(int(e.Number))
	if e.Protocol != "" {
		s += "/" + e.Protocol
	}
	return s
}

// Endpoint represents the container's networking configuration for each of its interfaces
//   - [github.com/docker/cli/opts.NetworkAttachmentOpts]
//   - [github.com/docker/docker/api/types/network.NetworkingConfig]
//   - [github.com/containers/common/libnetwork/types.PerNetworkOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceNetworkConfig]
type Endpoint struct {
	Target  string
	Options map[string]string // driver options

	IPv4Address  string
	IPv6Address  string
	MacAddress   string
	LinkLocalIPs []string
	Aliases      []string
	Links        []string
}

type DNS struct {
	Servers    []string
	Options    []string
	Search     []string
	HostName   string
	DomainName string
	HostAdd    map[string][]string
}

// https://github.com/moby/moby/blob/v27.3.1/api/types/network/network.go#L33-L45
// https://github.com/containers/common/blob/v0.60.4/libnetwork/types/network.go#L53-L88
type NetworkSpec struct {
	Name   string
	Labels map[string]string

	Driver  string
	Options map[string]string

	IPAMDriver  string
	IPAMOptions map[string]string
}
