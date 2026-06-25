//go:generate -command mockgen go tool mockgen

//go:generate mockgen -typed -destination=sdk/io/io.go io Reader,Writer,ReadCloser,WriteCloser
//go:generate mockgen -typed -destination=container/engine.go -source=../pkg/container/engine.go
//go:generate mockgen -typed -destination=sandboxer/engine.go -source=../pkg/sandboxer/engine.go
//go:generate mockgen -typed -destination=sandboxer/sandbox.go -source=../pkg/sandboxer/sandbox.go
//go:generate mockgen -typed -destination=command/console.go -source=../pkg/command/console.go
//go:generate mockgen -typed -destination=command/file.go -source=../pkg/command/file.go
//go:generate mockgen -typed -destination=command/cmdtypes/types.go -source=../pkg/command/cmdtypes/types.go
//go:generate mockgen -typed -destination=command/cmdtypes/issue.go -source=../pkg/command/cmdtypes/issue.go
//go:generate mockgen -typed -destination=problem/matcher.go -source=../pkg/problem/matcher.go
//go:generate mockgen -typed -destination=secret/masker.go -source=../pkg/secret/masker.go
//go:generate mockgen -typed -destination=executor/job_executor.go -source=../pkg/executor/job_executor.go
//go:generate mockgen -typed -destination=executor/step_executor.go -source=../pkg/executor/step_executor.go
//go:generate mockgen -typed -destination=stream/handler.go -source=../pkg/stream/handler.go
package mock
