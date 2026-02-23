package worker

import (
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/report/subscriber"
	"go.uber.org/dig"
)

const LogSubscribers = "log-subscribers"

type logSubscriberParams struct {
	dig.In

	LogManager  *log.Manager
	Subscribers []subscriber.Subscriber `group:"log-subscriberscribers"`
}

func (w *Worker) wireLogSubscribers(p logSubscriberParams) {
	for _, sub := range p.Subscribers {
		ch := p.LogManager.Subscribe()
		go sub.Run(ch)
		w.waiters = append(w.waiters, sub)
	}
}

func provideJobServiceSubscribers(scope *dig.Scope) error {
	return scope.Provide(subscriber.NewJobServiceLogSubscriber, dig.Group(LogSubscribers))
}

func provideResultServiceSubscribers(scope *dig.Scope) error {
	if err := scope.Provide(subscriber.NewResultServiceStepLogSubscriber, dig.Group(LogSubscribers)); err != nil {
		return err
	}

	return scope.Provide(subscriber.NewResultServiceJobLogSubscriber, dig.Group(LogSubscribers))
}
