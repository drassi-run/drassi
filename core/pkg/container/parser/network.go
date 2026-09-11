/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"drassi.run/core/pkg/container/types"
)

// ParsePublish parses user-provided publish definitions into types.PortBinding format
//   - [github.com/compose-spec/compose-go/v2/types.ParsePortConfig]
//   - [github.com/docker/go-connections/nat.ParsePortSpec]
//   - [github.com/containers/podman/v5/pkg/specgenutil.CreatePortBindings]
func ParsePublish(str string) (*types.PortBinding, uint16, error) {
	remains := str
	var hostIP, hostPort, containerPort, proto string

	if r, p, err := SplitProto(remains); err != nil {
		return nil, 0, err
	} else {
		remains, proto = r, p
	}

	if idx := strings.LastIndexByte(remains, ':'); idx != -1 {
		remains, containerPort = remains[:idx], remains[idx+1:]
	} else {
		remains, containerPort = "", remains
	}

	if remains != "" {
		if !strings.ContainsRune(remains, ':') {
			remains, hostPort = "", remains
		} else if host, port, err := net.SplitHostPort(remains); err != nil {
			return nil, 0, fmt.Errorf("invalid publish: %s - %s", str, err)
		} else {
			remains, hostIP, hostPort = "", host, port
		}
	}

	length := uint16(0)
	publish := &types.PortBinding{
		HostIP:   hostIP,
		Protocol: proto,
	}

	if port, portRange, err := ParsePortRange(containerPort); err != nil {
		return nil, 0, err
	} else {
		publish.ContainerPort = port
		length = portRange
	}

	if hostPort == "" {
		return publish, length, nil
	}

	if port, portRange, err := ParsePortRange(hostPort); err != nil {
		return nil, 0, err
	} else {
		publish.HostPort = port
		if portRange > 1 {
			if length > 1 && length != portRange {
				return nil, 0, fmt.Errorf("invalid publish %q : port-range mismatch", str)
			}
			length = portRange
		}
	}

	return publish, length, nil
}

// ParseExpose parses user-provided exposed port definitions into types.Port format
//   - [github.com/containers/podman/v5/pkg/specgenutil.CreateExpose]
func ParseExpose(str string) (*types.Port, uint16, error) {
	remains, expose := str, new(types.Port)

	if r, p, err := SplitProto(remains); err != nil {
		return nil, 0, err
	} else {
		remains, expose.Protocol = r, p
	}

	if port, length, err := ParsePortRange(remains); err != nil {
		return nil, 0, err
	} else {
		expose.Number = port
		return expose, length, nil
	}
}

var hostListSeparators = []string{"=", ":"}

// ParseHost parses user-provided additional host into hostname and list of IPs
//   - [github.com/compose-spec/compose-go/v2/types.NewHostsList]
func ParseHost(s string) (string, []string, error) {
	for _, sep := range hostListSeparators {
		host, ip, ok := strings.Cut(s, sep)
		if ok {
			return host, strings.Split(ip, ","), nil
		}
	}

	return "", nil, fmt.Errorf("invalid additional host, missing IP: %s", s)
}

func SplitProto(s string) (string, string, error) {
	splits := strings.SplitN(s, "/", 3)
	if len(splits) > 2 {
		return "", "", fmt.Errorf("invalid protocol: %s - multiple protocols", s)
	} else if len(splits) == 2 {
		remains, proto := splits[0], splits[1]
		if proto == "" {
			proto = "tcp"
		}
		return remains, proto, nil
	}
	return s, "tcp", nil
}

// ParsePortRange parses specified string as a port-range
//   - https://github.com/containers/podman/blob/v5.2.5/pkg/specgenutil/util.go#L216
//   - [github.com/docker/go-connections/nat.ParsePortRange]
func ParsePortRange(portRange string) (uint16, uint16, error) {
	var (
		port    string  = portRange
		endPort *string = nil
	)

	if splits := strings.SplitN(portRange, "-", 3); len(splits) > 2 {
		return 0, 0, fmt.Errorf("invalid portRange: %s - too many parts", portRange)
	} else if len(splits) == 2 {
		port, endPort = splits[0], &splits[1]
	}

	var portNum uint16
	if num, err := ParsePort(port); err != nil {
		return 0, 0, err
	} else {
		portNum = num
	}

	if endPort != nil {
		if num, err := ParsePort(*endPort); err != nil {
			return 0, 0, err
		} else if portNum >= num {
			return 0, 0, fmt.Errorf("invalid portRange: %s - startPort >= endPort", portRange)
		} else {
			length := num - portNum + 1
			return portNum, length, nil
		}
	}

	return portNum, 1, nil
}

// ParsePort turn a single string into a valid U16 port.
//   - https://github.com/containers/podman/blob/v5.2.5/pkg/specgenutil/util.go#L253-L262
//   - [github.com/docker/go-connections/nat.ParsePort]
func ParsePort(port string) (uint16, error) {
	num, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return 0, fmt.Errorf("invalid port: %s - must be in range [1, 65535]", port)
		}
		return 0, fmt.Errorf("invalid port: %s - %w", port, err)
	}
	return uint16(num), nil
}
