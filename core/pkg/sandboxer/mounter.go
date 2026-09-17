/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import "drassi.run/core/pkg/container/types"

// Mounter resolves container mounts for overlays and runtime containers.
type Mounter interface {
	// OverlayMounts returns the container bind/volume mounts for an overlay sandbox (the JobContainer).
	// When coarse is true and both source and target layouts are compatible (e.g. StandardLayout),
	// it returns a single mount for the entire job directory instead of discrete mounts for each individual subpath.
	OverlayMounts(target Layout, coarse bool) []*types.Mount

	// RuntimeMounts returns mounts for runtime.Container (Docker actions).
	// The sandboxer decides which folders to expose (e.g. Workspace, Temp, Actions).
	RuntimeMounts(target Layout) []*types.Mount
}
