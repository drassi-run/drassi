//go:generate -command mockgen go tool mockgen

//go:generate mockgen -source=../pkg/log/batcher.go -destination=log/batcher.go -typed
//go:generate mockgen -source=../pkg/log/chunker.go -destination=log/chunker.go -typed
//go:generate mockgen -source=../pkg/log/logtypes/types.go -destination=log/logtypes/types.go -typed
//go:generate mockgen -source=../pkg/log/logtypes/appender.go -destination=log/logtypes/appender.go -typed
//go:generate mockgen -source=../pkg/log/logtypes/conveyor.go -destination=log/logtypes/conveyor.go -typed
//go:generate mockgen -source=../pkg/log/logtypes/uploader.go -destination=log/logtypes/uploader.go -typed
//go:generate mockgen -source=../pkg/service/job.go -destination=service/job.go -typed
//go:generate mockgen -source=../pkg/service/result.go -destination=service/result.go -typed
package mock
