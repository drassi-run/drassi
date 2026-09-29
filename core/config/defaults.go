/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"path/filepath"
)

func RootDir() string {
	if os.Geteuid() == 0 {
		return "/var/lib/drassi"
	}

	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "drassi")
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "share", "drassi")
	}

	return "/tmp/drassi"
}

func DefaultConfig[R any]() *Config[R] {
	return &Config[R]{
		Sandboxer: &Sandboxer{
			Provider: "host",
		},
		Runtimes: map[string]*Runtime{
			"node": {
				Alias:      []string{"node20", "node22", "node24"},
				Image:      "ghcr.io/drassi-run/runtimes/node:24",
				Executable: "./bin/node",
				Cmd:        []string{"{0}"},
				Paths:      []string{"bin"},
			},
		},
		OciStore: DefaultOciStoreConfig(),
	}
}

func DefaultOciStoreConfig() *OciStore {
	return &OciStore{
		Backend:    "containerd",
		Containerd: DefaultContainerdStoreConfig(),
	}
}

func DefaultContainerdStoreConfig() *ContainerdStoreConfig {
	return &ContainerdStoreConfig{
		Address:       "/run/containerd/containerd.sock",
		Namespace:     "moby",
		FsSnapshotter: "overlayfs",
	}
}
