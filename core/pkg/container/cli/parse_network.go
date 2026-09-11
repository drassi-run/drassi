/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"net"
	"net/netip"

	"drassi.run/core/pkg/container/parser"
	"drassi.run/core/pkg/container/types"
)

func (fm *flagMapper) mapNetwork(copts *containerOptions) error {
	if err := fm.mapExposes(copts); err != nil {
		return err
	}
	if err := fm.mapPublish(copts); err != nil {
		return err
	}
	fm.Spec.PublishAll = copts.publishAll

	if err := fm.mapEndpoints(copts); err != nil {
		return err
	}

	if err := fm.mapDNS(copts); err != nil {
		return err
	}

	return nil
}

func (fm *flagMapper) mapExposes(copts *containerOptions) error {
	for _, opt := range copts.expose.GetAllOrEmpty() {
		if p, length, err := parser.ParseExpose(opt); err != nil {
			return err
		} else {
			for i := range length {
				port := &types.Port{Number: p.Number + i, Protocol: p.Protocol}
				fm.Spec.Exposes = append(fm.Spec.Exposes, port)
			}
		}
	}
	return nil
}

func (fm *flagMapper) mapPublish(copts *containerOptions) error {
	for _, opt := range copts.publish.GetAllOrEmpty() {
		if p, length, err := parser.ParsePublish(opt); err != nil {
			return err
		} else {
			for i := range length {
				binding := &types.PortBinding{HostIP: p.HostIP, ContainerPort: p.ContainerPort + i, Protocol: p.Protocol}
				if p.HostPort != 0 {
					binding.HostPort = p.HostPort + i
				}
				fm.Spec.Publish = append(fm.Spec.Publish, binding)
			}
		}
	}
	return nil
}

// - [github.com/docker/cli/opts.NetworkOpt]
// - https://github.com/docker/cli/blob/v29.7.2/cli/command/container/opts.go#L750
func (fm *flagMapper) mapEndpoints(copts *containerOptions) error {
	if len(copts.netMode.Value()) > 0 {
		return fmt.Errorf("the --network option is not supported")
	}

	userDefined := false
	ep := new(types.Endpoint)
	if copts.ipv4Address != "" {
		ip, err := netip.ParseAddr(copts.ipv4Address)
		if err != nil {
			return fmt.Errorf("invalid IPv4 address %s: %w", copts.ipv4Address, err)
		}
		userDefined, ep.IPv4Address = true, ip
	}
	if copts.ipv6Address != "" {
		ip, err := netip.ParseAddr(copts.ipv6Address)
		if err != nil {
			return fmt.Errorf("invalid IPv6 address %s: %w", copts.ipv6Address, err)
		}
		userDefined, ep.IPv6Address = true, ip
	}
	if copts.macAddress != "" {
		mac, err := net.ParseMAC(copts.macAddress)
		if err != nil {
			return fmt.Errorf("invalid MAC address %s: %w", copts.macAddress, err)
		}
		userDefined, ep.MacAddress = true, mac
	}
	if copts.linkLocalIPs.Len() > 0 {
		ips := make([]netip.Addr, 0, copts.linkLocalIPs.Len())
		for _, s := range copts.linkLocalIPs.GetAllOrEmpty() {
			ip, err := netip.ParseAddr(s)
			if err != nil {
				return fmt.Errorf("invalid link-local IP address %s: %w", s, err)
			}
			ips = append(ips, ip)
		}
		userDefined, ep.LinkLocalIPs = true, ips
	}
	if copts.aliases.Len() > 0 {
		userDefined, ep.Aliases = true, copts.aliases.GetAllOrEmpty()
	}
	if copts.links.Len() > 0 {
		userDefined, ep.Links = true, copts.links.GetAllOrEmpty()
	}
	if userDefined {
		fm.Spec.Endpoints = append(fm.Spec.Endpoints, ep)
	}
	return nil
}

func (fm *flagMapper) mapDNS(copts *containerOptions) error {
	dns := fm.Spec.DNS
	if copts.dns.Len() > 0 {
		servers := make([]netip.Addr, 0, copts.dns.Len())
		for _, s := range copts.dns.GetAllOrEmpty() {
			ip, err := netip.ParseAddr(s)
			if err != nil {
				return fmt.Errorf("invalid DNS IP address %s: %w", s, err)
			}
			servers = append(servers, ip)
		}
		dns.Servers = servers
	}
	dns.Options = copts.dnsOptions.GetAllOrEmpty()
	dns.Search = copts.dnsSearch.GetAllOrEmpty()
	dns.HostName = copts.hostname
	dns.DomainName = copts.domainname

	if copts.extraHosts.Len() > 0 {
		dns.HostAdd = make(map[string][]string)
	}
	for _, h := range copts.extraHosts.GetAllOrEmpty() {
		if host, ips, err := parser.ParseHost(h); err != nil {
			return err
		} else if exist, ok := dns.HostAdd[host]; ok {
			dns.HostAdd[host] = append(exist, ips...)
		} else {
			dns.HostAdd[host] = ips
		}
	}

	return nil
}
