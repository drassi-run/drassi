//go:generate -command mockgen go tool mockgen

//go:generate mockgen -source=../pkg/container/engine.go -destination=container/engine.go -typed
//go:generate mockgen -source=../pkg/sandboxer/engine.go -destination=sandboxer/engine.go -typed
//go:generate mockgen -source=../pkg/sandboxer/sandbox.go -destination=sandboxer/sandbox.go -typed
//go:generate mockgen -source=../pkg/executor/command/console.go -destination=executor/command/console.go -typed
//go:generate mockgen -source=../pkg/executor/command/file.go -destination=executor/command/file.go -typed
//go:generate mockgen -source=../pkg/executor/command/cmdhandler/types.go -destination=executor/command/cmdhandler/types.go -typed
//go:generate mockgen -source=../pkg/executor/problem/matcher.go -destination=executor/problem/matcher.go -typed
//go:generate mockgen -source=../pkg/executor/secret/masker.go -destination=executor/secret/masker.go -typed
//go:generate mockgen -source=../pkg/executor/support/tracker.go -destination=executor/support/tracker.go -typed
//go:generate mockgen -source=../pkg/executor/job_executor.go -destination=executor/job_executor.go -typed
//go:generate mockgen -source=../pkg/executor/step_executor.go -destination=executor/step_executor.go -typed
//go:generate mockgen -source=../pkg/stream/handler.go -destination=stream/handler.go -typed
package mock
