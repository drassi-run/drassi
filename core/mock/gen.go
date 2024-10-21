//go:generate mockgen -source=../pkg/sandboxer/engine.go -destination=sandboxer/engine.go -typed
//go:generate mockgen -source=../pkg/sandboxer/sandbox.go -destination=sandboxer/sandbox.go -typed
//go:generate mockgen -source=../pkg/executor/command/console.go -destination=executor/command/console.go -typed
//go:generate mockgen -source=../pkg/executor/command/file.go -destination=executor/command/file.go -typed
//go:generate mockgen -source=../pkg/executor/problem/matcher.go -destination=executor/problem/matcher.go -typed
//go:generate mockgen -source=../pkg/executor/secret/masker.go -destination=executor/secret/masker.go -typed
//go:generate mockgen -source=../pkg/executor/logging/logger.go -destination=executor/logging/logger.go -typed
//go:generate mockgen -source=../pkg/executor/executor_job.go -destination=executor/executor_job.go -typed
//go:generate mockgen -source=../pkg/executor/executor_step.go -destination=executor/executor_step.go -typed
//go:generate mockgen -source=../pkg/executor/supervisor.go -destination=executor/supervisor.go -typed
//go:generate mockgen -source=../pkg/executor/reporter/reporter.go -destination=executor/reporter/reporter.go -typed
package mock
