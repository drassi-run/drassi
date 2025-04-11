package chunk

import "time"

type Waiter interface {
	Wait()
}

type DurationWaiter int64

func (w DurationWaiter) Wait() {
	time.Sleep(time.Duration(w))
}
