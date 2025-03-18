/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"drassi.run/core/pkg/container/types"
	"github.com/stretchr/testify/assert"
	"iter"
	"strings"
	"testing"
)

// https://github.com/containers/podman/blob/v5.2.5/pkg/specgenutil/util_test.go#L8
func TestParseExpose(t *testing.T) {
	t.Run("success", testParseExposeSuccess)
	t.Run("failure", testParseExposeFailure)
}

type rangedPort struct {
	types.Port
	Range uint16
}

func testParseExposeSuccess(t *testing.T) {
	tc := map[string]*rangedPort{
		"99":         {Port: types.Port{Number: 99, Protocol: "tcp"}, Range: 1},
		"99-100":     {Port: types.Port{Number: 99, Protocol: "tcp"}, Range: 2},
		"99/udp":     {Port: types.Port{Number: 99, Protocol: "udp"}, Range: 1},
		"99-100/udp": {Port: types.Port{Number: 99, Protocol: "udp"}, Range: 2},
	}
	for input, expected := range tc {
		actualPort, actualLength, err := ParseExpose(input)
		assert.NoError(t, err, input)
		assert.EqualValues(t, &expected.Port, actualPort, input)
		assert.EqualValues(t, expected.Range, actualLength, input)
	}
}

func testParseExposeFailure(t *testing.T) {
	tc := []string{
		"100-99",
		"99/tcp-100/tcp",
		"-1234",
		"1234-",
		"999999999",
	}
	for _, input := range tc {
		_, _, err := ParseExpose(input)
		assert.ErrorContains(t, err, "invalid", input)
	}
}

func TestParsePublish(t *testing.T) {
	t.Run("success", testParsePublishSuccess)
	t.Run("failure", testParsePublishFailure)
}

type rangedPortBinding struct {
	types.PortBinding
	Range uint16
}

func testParsePublishSuccess(t *testing.T) {
	hostIp := []string{"", "127.0.0.1", "[2001:db8::]"} // IPv6 must be encased in a bracket
	hostPort := []string{"", "80", "80-82"}
	containerPort := []string{"8080", "8080-8082"}
	protocol := []string{"", "tcp", "udp"}

	var tcGen iter.Seq2[string, rangedPortBinding] = func(yield func(string, rangedPortBinding) bool) {
		for _, cPort := range containerPort {
			in, pm, length := cPort, types.PortBinding{}, uint16(0)
			pm.ContainerPort, length, _ = ParsePortRange(cPort)

			for _, proto := range protocol {
				in, pm := in, pm
				if proto != "" {
					in += "/" + proto
					pm.Protocol = proto
				} else {
					pm.Protocol = "tcp"
				}

				for _, hPort := range hostPort {
					in, pm, length := in, pm, length
					if hPort != "" {
						in = hPort + ":" + in
						port, portRange, _ := ParsePortRange(hPort)
						pm.HostPort = port
						if portRange > 1 {
							length = portRange
						}
					}

					for _, hIP := range hostIp {
						in, pm := in, pm
						if hIP != "" {
							if hPort != "" {
								in = hIP + ":" + in
							} else {
								in = hIP + "::" + in
							}
							pm.HostIP = strings.Trim(hIP, "[]")
						}

						if !yield(in, rangedPortBinding{pm, length}) {
							return
						}
					}
				}
			}
		}
	}

	for input, expected := range tcGen {
		actualPortBinding, actualLength, err := ParsePublish(input)
		assert.NoError(t, err, input)
		assert.EqualValues(t, &expected.PortBinding, actualPortBinding, input)
		assert.EqualValues(t, expected.Range, actualLength, input)
	}
}

func testParsePublishFailure(t *testing.T) {
	tc := []string{
		"192.168.1.100", // missing Container port
		"192.168.1.100:-1:-1",
		"80-90:8000-9000", // port-range mismatch
	}
	for _, input := range tc {
		_, _, err := ParsePublish(input)
		assert.ErrorContains(t, err, "invalid", input)
	}
}

func TestSplitProto(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := map[string][2]string{
			"80":     {"80", "tcp"},
			"80/":    {"80", "tcp"},
			"80/udp": {"80", "udp"},
		}
		for input, expected := range tc {
			remains, proto, err := SplitProto(input)
			assert.NoError(t, err, input)
			assert.Equal(t, expected[0], remains, input)
			assert.Equal(t, expected[1], proto, input)
		}
	})
	t.Run("failure", func(t *testing.T) {
		tc := []string{
			"80/tcp/udp",
		}
		for _, input := range tc {
			_, _, err := SplitProto(input)
			assert.ErrorContains(t, err, "invalid protocol", input)
		}
	})
}

// https://github.com/docker/go-connections/blob/master/nat/parse_test.go
func TestParsePortRange(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tc := map[string][2]uint16{
			"8080":      {8080, 1},
			"8080-8088": {8080, 9},
		}
		for input, expected := range tc {
			start, length, err := ParsePortRange(input)
			assert.NoError(t, err, input)
			assert.EqualValues(t, expected[0], start, input)
			assert.EqualValues(t, expected[1], length, input)
		}
	})

	t.Run("failure", func(t *testing.T) {
		tc := []string{
			"",
			"8000-8000",
			"9000-8080",
			"8000-a",
			"8000-30a",
			"a-8000",
			"30a-8000",
			"8080-",
			"-8080",
			"-8000-",
		}
		for _, input := range tc {
			_, _, err := ParsePortRange(input)
			assert.ErrorContains(t, err, "invalid", input)
		}
	})
}
