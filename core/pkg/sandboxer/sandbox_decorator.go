package sandboxer

import (
	"context"
	"errors"
)

type Cleanup func(ctx context.Context) error

type decoratedSandbox struct {
	Sandbox

	beforeCleanup []Cleanup
	afterCleanup  []Cleanup
}

func (s *decoratedSandbox) Terminate(ctx context.Context) error {
	var errs []error

	for _, fn := range s.beforeCleanup {
		errs = append(errs, fn(ctx))
	}

	errs = append(errs, s.Sandbox.Terminate(ctx))

	for _, fn := range s.afterCleanup {
		errs = append(errs, fn(ctx))
	}

	return errors.Join(errs...)
}

func AddBeforeCleanup(sb Sandbox, fns ...Cleanup) Sandbox {
	if len(fns) == 0 {
		return sb
	}

	if s, ok := sb.(*decoratedSandbox); ok {
		// LIFO, fns must be cleaned up first
		s.beforeCleanup = append(fns, s.beforeCleanup...)
		return s
	}

	return &decoratedSandbox{
		Sandbox:       sb,
		beforeCleanup: fns,
	}
}

func AddAfterCleanup(sb Sandbox, fns ...Cleanup) Sandbox {
	if len(fns) == 0 {
		return sb
	}

	if s, ok := sb.(*decoratedSandbox); ok {
		s.afterCleanup = append(s.afterCleanup, fns...)
		return s
	}

	return &decoratedSandbox{
		Sandbox:      sb,
		afterCleanup: fns,
	}
}
