package worker

import (
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/report/subscriber"
	"go.uber.org/dig"
)

const LogSubscribers = "log-subscribers"

type logSubscriberParams struct {
	dig.In

	lm   *log.Manager
	subs []subscriber.Subscriber `group:"log-subscribers"`
}

func (w *Worker) wireLogSubscribers(p logSubscriberParams) {
	for _, sub := range p.subs {
		ch := p.lm.Subscribe()
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
