/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"strconv"
	"strings"

	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/types"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

type empty = struct{}

func (cc *containerConfig) setNetwork(conf *types.ContainerNetwork) {
	cc.setExpose(conf.Exposes)
	cc.setPublish(conf.Publish)
	cc.HostConfig.PublishAllPorts = conf.PublishAll
	cc.setDNS(&conf.DNS)
	cc.setNetworkEndpoints(conf.Endpoints)
}

func (cc *containerConfig) setExpose(exposes []*types.Port) {
	c := cc.Config
	c.ExposedPorts = make(nat.PortSet)
	for _, e := range exposes {
		port := nat.Port(e.String())
		c.ExposedPorts[port] = empty{}
	}
}

func (cc *containerConfig) setPublish(publishes []*types.PortBinding) {
	c, hc := cc.Config, cc.HostConfig
	hc.PortBindings = make(nat.PortMap)
	for _, p := range publishes {
		s := strconv.Itoa(int(p.ContainerPort))
		if p.Protocol != "" {
			s += "/" + p.Protocol
		}
		port := nat.Port(s)
		c.ExposedPorts[port] = empty{} // publish a port also expose it

		binding := nat.PortBinding{HostIP: p.HostIP}
		if p.HostPort != 0 {
			binding.HostPort = strconv.Itoa(int(p.HostPort))
		}

		hc.PortBindings[port] = append(hc.PortBindings[port], binding)
	}
}

func (cc *containerConfig) setDNS(dns *types.DNS) {
	c, hc := cc.Config, cc.HostConfig

	hc.DNS = dns.Servers
	hc.DNSSearch = dns.Search
	hc.DNSOptions = dns.Options
	for k, v := range dns.HostAdd {
		val := strings.Join(v, ",")
		hc.ExtraHosts = append(hc.ExtraHosts, k+"="+val)
	}

	c.Hostname = dns.HostName
	c.Domainname = dns.DomainName
}

func (cc *containerConfig) setNetworkEndpoints(endpoints []*types.Endpoint) {
	if len(endpoints) == 0 {
		return
	}

	conf := make(map[string]*dockernetwork.EndpointSettings, len(endpoints))
	for _, ep := range endpoints {
		endpoint := &dockernetwork.EndpointSettings{
			Links:      ep.Links,
			Aliases:    ep.Aliases,
			MacAddress: ep.MacAddress,
			DriverOpts: ep.Options,
		}
		if ep.IPv4Address != "" || ep.IPv6Address != "" || len(ep.LinkLocalIPs) > 0 {
			endpoint.IPAMConfig = &dockernetwork.EndpointIPAMConfig{
				IPv4Address:  ep.IPv4Address,
				IPv6Address:  ep.IPv6Address,
				LinkLocalIPs: ep.LinkLocalIPs,
			}
		}
		target := ep.Target
		if target == "" {
			target = dockernetwork.NetworkDefault
		}
		conf[target] = endpoint
	}

	cc.NetworkingConfig = &dockernetwork.NetworkingConfig{
		EndpointsConfig: conf,
	}
}

func (cs *containerSpec) setNetwork(info dockertypes.ContainerJSON) error {
	c, hc, net := info.Config, info.HostConfig, info.NetworkSettings

	if err := cs.setExpose(c.ExposedPorts); err != nil {
		return err
	}
	// hc.PortBindings is not resolved HostPort (if we use random one)
	// So, use net.Ports instead
	if err := cs.setPublish(net.Ports); err != nil {
		return err
	}
	cs.Spec.PublishAll = hc.PublishAllPorts
	if err := cs.setDNS(c, hc); err != nil {
		return err
	}
	cs.setNetworkEndpoints(net.Networks)

	return nil
}

func (cs *containerSpec) setExpose(exposes nat.PortSet) error {
	if len(exposes) == 0 {
		return nil
	}

	for e := range exposes {
		// nat.Port resolved portRange into single port
		if port, proto, err := parsePortProto(string(e)); err != nil {
			return err
		} else {
			expose := &types.Port{Number: port, Protocol: proto}
			cs.Spec.Exposes = append(cs.Spec.Exposes, expose)
		}
	}

	return nil
}

func (cs *containerSpec) setPublish(publishes nat.PortMap) error {
	if len(publishes) == 0 {
		return nil
	}

	for k, v := range publishes {
		var publish types.PortBinding

		// nat.Port resolved portRange into single port
		if port, proto, err := parsePortProto(string(k)); err != nil {
			return err
		} else {
			publish.ContainerPort = port
			publish.Protocol = proto
		}

		for _, h := range v {
			publish := publish // cloned
			// nat.PortBinding.HostPort is single port as well
			if port, err := strconv.ParseUint(h.HostPort, 10, 16); err != nil {
				return err
			} else {
				publish.HostIP = h.HostIP
				publish.HostPort = uint16(port)
			}
			cs.Spec.Publish = append(cs.Spec.Publish, &publish)
		}
	}

	return nil
}

func (cs *containerSpec) setDNS(c *dockercontainer.Config, hc *dockercontainer.HostConfig) error {
	cs.Spec.DNS = types.DNS{
		Servers:    hc.DNS,
		Options:    hc.DNSOptions,
		Search:     hc.DNSSearch,
		HostName:   c.Hostname,
		DomainName: c.Domainname,
	}
	if len(hc.ExtraHosts) == 0 {
		return nil
	}

	extraHost := make(map[string][]string)
	for _, h := range hc.ExtraHosts {
		if host, ips, err := cli.ParseHost(h); err != nil {
			return err
		} else if exist, ok := extraHost[host]; ok {
			extraHost[host] = append(exist, ips...)
		} else {
			extraHost[host] = ips
		}
	}
	cs.Spec.DNS.HostAdd = extraHost

	return nil
}

func (cs *containerSpec) setNetworkEndpoints(endpoints map[string]*dockernetwork.EndpointSettings) {
	if len(endpoints) == 0 {
		return
	}

	ne := make([]*types.Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		ep := &types.Endpoint{
			Target:      endpoint.NetworkID,
			Options:     endpoint.DriverOpts,
			Links:       endpoint.Links,
			Aliases:     endpoint.Aliases,
			MacAddress:  endpoint.MacAddress,
			IPv4Address: endpoint.IPAddress,
			IPv6Address: endpoint.GlobalIPv6Address,
		}
		ne = append(ne, ep)
	}
	cs.Spec.Endpoints = ne
}

func parsePortProto(s string) (uint16, string, error) {
	if port, proto, err := cli.SplitProto(s); err != nil {
		return 0, "", err
	} else if num, err := cli.ParsePort(port); err != nil {
		return 0, "", err
	} else {
		return num, proto, nil
	}
}
