/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xfs

import "io/fs"

const (
	FilePerm fs.FileMode = 0o644
	DirPerm  fs.FileMode = 0o755
	AllPerm  fs.FileMode = 0o777
)
