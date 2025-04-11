package chunk

import "sync/atomic"

type Option func(*option)

type option struct {
	bufferSize  int
	softLimit   int
	lineSafe    bool
	followInput atomic.Bool
	waiter      Waiter
}

func WithBufferSize(size int) Option {
	return func(o *option) {
		o.bufferSize = size
	}
}

func WithSoftLimit(size int) Option {
	return func(o *option) {
		o.softLimit = size
	}
}

// WithLineSafety enable line-safety, so line won't be split across chunks (middle-line cut)
func WithLineSafety(b bool) Option {
	return func(o *option) {
		o.lineSafe = b
	}
}

// WithFollowInput continuous read from Reader until got Complete signal
func WithFollowInput(b bool) Option {
	return func(o *option) {
		o.followInput.Store(b)
	}
}

func WithWaiter(waiter Waiter) Option {
	return func(o *option) {
		o.waiter = waiter
	}
}
