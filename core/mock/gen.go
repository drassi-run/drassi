//go:generate -command mockgen go tool mockgen

//go:generate mockgen -source=../pkg/container/engine.go -destination=container/engine.go -typed
//go:generate mockgen -source=../pkg/sandboxer/engine.go -destination=sandboxer/engine.go -typed
//go:generate mockgen -source=../pkg/sandboxer/sandbox.go -destination=sandboxer/sandbox.go -typed
//go:generate mockgen -source=../pkg/command/console.go -destination=command/console.go -typed
//go:generate mockgen -source=../pkg/command/file.go -destination=command/file.go -typed
//go:generate mockgen -source=../pkg/command/cmdtypes/types.go -destination=command/cmdtypes/types.go -typed
//go:generate mockgen -source=../pkg/command/cmdtypes/issue.go -destination=command/cmdtypes/issue.go -typed
//go:generate mockgen -source=../pkg/problem/matcher.go -destination=problem/matcher.go -typed
//go:generate mockgen -source=../pkg/secret/masker.go -destination=secret/masker.go -typed
//go:generate mockgen -source=../pkg/executor/job_executor.go -destination=executor/job_executor.go -typed
//go:generate mockgen -source=../pkg/executor/step_executor.go -destination=executor/step_executor.go -typed
//go:generate mockgen -source=../pkg/stream/handler.go -destination=stream/handler.go -typed
package mock
