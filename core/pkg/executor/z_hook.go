package executor

import "context"

type Hook[R any] interface {
	Hook(context.Context, R) error
}

type Hooks[R any] []Hook[R]

func (m Hooks[R]) Hook(ctx context.Context, res R) error {
	for _, h := range m {
		if err := h.Hook(ctx, res); err != nil {
			return err
		}
	}
	return nil
}
