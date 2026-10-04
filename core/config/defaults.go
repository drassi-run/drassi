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
