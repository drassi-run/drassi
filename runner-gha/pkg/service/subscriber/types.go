package subscriber

import "drassi.run/gha-runner/pkg/log"

type Subscriber interface {
	Run(ch <-chan *log.Event)
	Wait()
}
