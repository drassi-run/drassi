/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

const (
	Stdin = 1 << iota
	Stdout
	Stderr
	None = 0 // detach mode
)

// See also: https://github.com/docker/cli/blob/v27.3.1/cli/command/container/run.go#L122-L135
type Stdio struct {
	Tty         bool
	Attach      int
	Interactive bool
}

func (s *Stdio) AttachStdin() bool {
	return s.Attach&Stdin != 0
}

func (s *Stdio) AttachStdout() bool {
	return s.Attach&Stdout != 0
}

func (s *Stdio) AttachStderr() bool {
	return s.Attach&Stderr != 0
}

func (s *Stdio) Detach() bool {
	return s.Attach == None
}
