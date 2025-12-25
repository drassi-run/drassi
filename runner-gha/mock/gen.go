//go:generate -command mockgen go tool mockgen

//go:generate mockgen -source=../pkg/log/batcher.go -destination=log/batcher.go -typed
//go:generate mockgen -source=../pkg/log/chunker.go -destination=log/chunker.go -typed
//go:generate mockgen -source=../pkg/report/job_service.go -destination=report/job_service.go -typed
//go:generate mockgen -source=../pkg/report/result_service.go -destination=report/result_service.go -typed
//go:generate mockgen -source=../pkg/report/types/appender.go -destination=report/types/appender.go -typed
//go:generate mockgen -source=../pkg/report/types/conveyor.go -destination=report/types/conveyor.go -typed
//go:generate mockgen -source=../pkg/report/types/uploader.go -destination=report/types/uploader.go -typed
//go:generate mockgen -source=../pkg/report/types/types.go -destination=report/types/types.go -typed
package mock
