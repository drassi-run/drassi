/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextIPv4(t *testing.T) {
	_, n, err := net.ParseCIDR("172.16.0.1/24")
	require.NoError(t, err)
	next, err := nextIPv4(net.ParseIP("172.16.0.1"), n)
	require.NoError(t, err)
	assert.Equal(t, "172.16.0.2", next.String())

	_, n, err = net.ParseCIDR("10.0.0.254/24")
	require.NoError(t, err)
	_, err = nextIPv4(net.ParseIP("10.0.0.254"), n)
	require.Error(t, err)
}

func TestGuestNetExplicitIP(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TapDevice = "tap-missing"
	cfg.GuestIP = "172.16.0.2/24"
	cfg.GuestGateway = "172.16.0.1"
	spec, err := cfg.guestNet()
	require.NoError(t, err)
	require.NotNil(t, spec)
	assert.Equal(t, "eth0", spec.Iface)
	assert.Equal(t, "172.16.0.2/24", spec.Addr)
	assert.Equal(t, "172.16.0.1", spec.Gateway)
	assert.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, spec.DNS)
}

func TestGuestNetDisabledWithoutTAP(t *testing.T) {
	cfg := DefaultConfig()
	spec, err := cfg.guestNet()
	require.NoError(t, err)
	assert.Nil(t, spec)
}

func TestApplyGuestNetRequiresAddr(t *testing.T) {
	err := applyGuestNet(request{Op: opNet, Iface: "eth0"})
	require.Error(t, err)
}
