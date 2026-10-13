/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"context"
	"errors"
)

type Cleanup = func(ctx context.Context) error

type hookedSandbox struct {
	Sandbox

	preTerminates  []Cleanup
	postTerminates []Cleanup
}

func (s *hookedSandbox) Terminate(ctx context.Context) error {
	var errs []error

	for _, fn := range s.preTerminates {
		errs = append(errs, fn(ctx))
	}

	errs = append(errs, s.Sandbox.Terminate(ctx))

	for _, fn := range s.postTerminates {
		errs = append(errs, fn(ctx))
	}

	return errors.Join(errs...)
}

func PreTerminate(sb Sandbox, fns ...Cleanup) Sandbox {
	if len(fns) == 0 {
		return sb
	}

	if s, ok := sb.(*hookedSandbox); ok {
		// LIFO, fns must be cleaned up first
		s.preTerminates = append(fns, s.preTerminates...)
		return s
	}

	return &hookedSandbox{
		Sandbox:       sb,
		preTerminates: fns,
	}
}

func PostTerminate(sb Sandbox, fns ...Cleanup) Sandbox {
	if len(fns) == 0 {
		return sb
	}

	if s, ok := sb.(*hookedSandbox); ok {
		s.postTerminates = append(s.postTerminates, fns...)
		return s
	}

	return &hookedSandbox{
		Sandbox:        sb,
		postTerminates: fns,
	}
}

func (s *hookedSandbox) Unwrap() Sandbox {
	return s.Sandbox
}

// Unwrap recursively unwraps Sandbox wrappers until reaching the base.
func Unwrap(sb Sandbox) Sandbox {
	for {
		if u, ok := sb.(interface{ Unwrap() Sandbox }); ok {
			sb = u.Unwrap()
			continue
		}
		return sb
	}
}
