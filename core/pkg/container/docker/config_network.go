/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"net"
	"net/netip"
	"strconv"
	"strings"

	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/types"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockernetwork "github.com/moby/moby/api/types/network"
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
	c.ExposedPorts = make(dockernetwork.PortSet)
	for _, e := range exposes {
		if port, err := dockernetwork.ParsePort(e.String()); err == nil {
			c.ExposedPorts[port] = empty{}
		}
	}
}

func (cc *containerConfig) setPublish(publishes []*types.PortBinding) {
	c, hc := cc.Config, cc.HostConfig
	hc.PortBindings = make(dockernetwork.PortMap)
	for _, p := range publishes {
		s := strconv.Itoa(int(p.ContainerPort))
		if p.Protocol != "" {
			s += "/" + p.Protocol
		}
		port, err := dockernetwork.ParsePort(s)
		if err != nil {
			continue
		}
		c.ExposedPorts[port] = empty{} // publish a port also expose it

		var hostIP netip.Addr
		if p.HostIP != "" {
			hostIP, _ = netip.ParseAddr(p.HostIP)
		}
		binding := dockernetwork.PortBinding{HostIP: hostIP}
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
			DriverOpts: ep.Options,
			MacAddress: dockernetwork.HardwareAddr(ep.MacAddress),
		}
		if ep.IPv4Address.IsValid() || ep.IPv6Address.IsValid() || len(ep.LinkLocalIPs) > 0 {
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

func (cs *containerSpec) setNetwork(info dockercontainer.InspectResponse) error {
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

func (cs *containerSpec) setExpose(exposes dockernetwork.PortSet) error {
	if len(exposes) == 0 {
		return nil
	}

	for e := range exposes {
		expose := &types.Port{Number: e.Num(), Protocol: string(e.Proto())}
		cs.Spec.Exposes = append(cs.Spec.Exposes, expose)
	}

	return nil
}

func (cs *containerSpec) setPublish(publishes dockernetwork.PortMap) error {
	if len(publishes) == 0 {
		return nil
	}

	for k, v := range publishes {
		publish := types.PortBinding{
			ContainerPort: k.Num(),
			Protocol:      string(k.Proto()),
		}

		for _, h := range v {
			pub := publish // cloned
			if h.HostPort != "" {
				if port, err := strconv.ParseUint(h.HostPort, 10, 16); err != nil {
					return err
				} else {
					pub.HostPort = uint16(port)
				}
			}
			if h.HostIP.IsValid() {
				pub.HostIP = h.HostIP.String()
			}
			cs.Spec.Publish = append(cs.Spec.Publish, &pub)
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
			MacAddress:  net.HardwareAddr(endpoint.MacAddress),
			IPv4Address: endpoint.IPAddress,
			IPv6Address: endpoint.GlobalIPv6Address,
		}
		if endpoint.IPAMConfig != nil {
			ep.LinkLocalIPs = endpoint.IPAMConfig.LinkLocalIPs
		}
		ne = append(ne, ep)
	}
	cs.Spec.Endpoints = ne
}
