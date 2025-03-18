/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gha

import (
	"fmt"
)

type UserAgentInfo struct {
	// Version is the version of the controller
	Version string
	// CommitSHA is the git commit SHA of the controller
	CommitSHA string
	// HasProxy is true if the controller is running behind a proxy
	HasProxy bool
}

func (u UserAgentInfo) String() string {
	proxy := "Proxy/disabled"
	if u.HasProxy {
		proxy = "Proxy/enabled"
	}

	return fmt.Sprintf("gha-runner/%s (%s) (%s)", u.Version, u.CommitSHA, proxy)
}
