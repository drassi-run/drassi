/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

type Architecture string

const (
	Arm32 Architecture = "ARM32"
	Arm64 Architecture = "ARM64"
	X64   Architecture = "X64"
	X86   Architecture = "X86"
)

type Machine string

const (
	Linux   Machine = "Linux"
	MacOS   Machine = "macOS"
	Windows Machine = "Windows"
)

func (m *Machine) FileSeparator() rune {
	switch *m {
	case Linux, MacOS:
		return '/'
	case Windows:
		return '\\'
	default:
		return 0 // null character
	}
}

func (m *Machine) PathSeparator() rune {
	switch *m {
	case Linux, MacOS:
		return ':'
	case Windows:
		return ';'
	default:
		return 0 // null character
	}
}

func (m *Machine) LineSeparator() string {
	switch *m {
	case Linux, MacOS:
		return "\n"
	case Windows:
		return "\r\n"
	default:
		return "\x00" // string of a null character
	}
}
