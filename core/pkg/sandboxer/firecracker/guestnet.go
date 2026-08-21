/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type netSpec struct {
	Iface   string
	Addr    string
	Gateway string
	DNS     []string
}

func (c *Config) guestNet() (*netSpec, error) {
	if c.TapDevice == "" {
		return nil, nil
	}

	iface := c.GuestIface
	if iface == "" {
		iface = "eth0"
	}
	dns := c.GuestDNS
	if len(dns) == 0 {
		dns = []string{"1.1.1.1", "8.8.8.8"}
	}

	spec := &netSpec{Iface: iface, DNS: dns}
	if c.GuestIP != "" {
		spec.Addr = c.GuestIP
		if !strings.Contains(spec.Addr, "/") {
			spec.Addr += "/24"
		}
		spec.Gateway = c.GuestGateway
		if spec.Gateway == "" {
			if ip, _, err := ifaceIPv4(c.TapDevice); err == nil {
				spec.Gateway = ip.String()
			}
		}
		return spec, nil
	}

	ip, ipnet, err := ifaceIPv4(c.TapDevice)
	if err != nil {
		return nil, fmt.Errorf("guest net: %w (set guest_ip)", err)
	}
	next, err := nextIPv4(ip, ipnet)
	if err != nil {
		return nil, fmt.Errorf("guest net: %w", err)
	}
	ones, _ := ipnet.Mask.Size()
	spec.Addr = fmt.Sprintf("%s/%d", next, ones)
	spec.Gateway = c.GuestGateway
	if spec.Gateway == "" {
		spec.Gateway = ip.String()
	}
	return spec, nil
}

func ifaceIPv4(name string) (net.IP, *net.IPNet, error) {
	ifi, err := net.InterfaceByName(name)
	if err != nil {
		return nil, nil, fmt.Errorf("tap %q: %w", name, err)
	}
	addrs, err := ifi.Addrs()
	if err != nil {
		return nil, nil, fmt.Errorf("tap %q: %w", name, err)
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipnet.IP.To4()
		if ip == nil || ip.IsUnspecified() {
			continue
		}
		return ip, ipnet, nil
	}
	return nil, nil, fmt.Errorf("tap %q has no IPv4 address", name)
}

func nextIPv4(ip net.IP, ipnet *net.IPNet) (net.IP, error) {
	v4 := ip.To4()
	if v4 == nil {
		return nil, fmt.Errorf("not an IPv4 address")
	}
	mask := maskTo4(ipnet.Mask)
	cur := binary.BigEndian.Uint32(v4)
	next := cur + 1
	n := make(net.IP, 4)
	binary.BigEndian.PutUint32(n, next)
	if !ipnet.Contains(n) {
		return nil, fmt.Errorf("no address in %s after %s", ipnet, ip)
	}
	ones, bits := mask.Size()
	if bits == 32 && ones <= 30 {
		m := binary.BigEndian.Uint32(mask)
		bcast := (cur & m) | ^m
		if next == bcast {
			return nil, fmt.Errorf("next address after %s is broadcast in %s", ip, ipnet)
		}
	}
	return n, nil
}

func maskTo4(mask net.IPMask) net.IPMask {
	if len(mask) == net.IPv6len {
		return mask[12:]
	}
	return mask
}
