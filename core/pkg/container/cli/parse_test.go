/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEmpty(t *testing.T) {
	spec, stdio, err := Parse("")
	require.NoError(t, err)
	require.NotNil(t, spec)
	require.NotNil(t, stdio)
	require.Nil(t, spec.DNS)
}

func TestParseDNS(t *testing.T) {
	spec, _, err := Parse("--dns 8.8.8.8 --dns-search example.com --hostname myhost")
	require.NoError(t, err)
	require.NotNil(t, spec.DNS)
	require.Len(t, spec.DNS.Servers, 1)
	require.Equal(t, "8.8.8.8", spec.DNS.Servers[0].String())
	require.Equal(t, []string{"example.com"}, spec.DNS.Search)
	require.Equal(t, "myhost", spec.DNS.HostName)
}
