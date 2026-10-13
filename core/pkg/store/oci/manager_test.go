/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore_test

import (
	"testing"

	"drassi.run/core/config"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
)

func TestNew_OrasBackend(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.OciStore{
		Backend: "oras",
		Oras: &config.OrasStoreConfig{
			StoreDir: tempDir,
		},
	}
	mgr, err := ocistore.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, mgr)
	require.NoError(t, mgr.Close())
}
