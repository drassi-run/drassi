package sandboxer

import (
	"context"
	"errors"
)

type Cleanup func(ctx context.Context) error
type Informer func(*ContainerInfo)

type decoratedSandbox struct {
	Sandbox

	informers     []Informer
	beforeCleanup []Cleanup
	afterCleanup  []Cleanup
}

func (s *decoratedSandbox) ContainerInfo(ctx context.Context) (*ContainerInfo, error) {
	info, err := s.Sandbox.ContainerInfo(ctx)
	if err != nil {
		return nil, err
	}

	for _, fn := range s.informers {
		fn(info)
	}
	return info, nil
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
		s.beforeCleanup = append(s.beforeCleanup, fns...)
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

func ProvideInfo(sb Sandbox, fns ...Informer) Sandbox {
	if len(fns) == 0 {
		return sb
	}

	if s, ok := sb.(*decoratedSandbox); ok {
		s.informers = append(s.informers, fns...)
		return s
	}

	return &decoratedSandbox{
		Sandbox:   sb,
		informers: fns,
	}
}
