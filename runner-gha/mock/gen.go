//go:generate -command mockgen go tool mockgen

//go:generate mockgen -source=../pkg/log/chunker.go -destination=log/chunker.go -typed
//go:generate mockgen -source=../pkg/log/types.go -destination=log/types.go -typed
//go:generate mockgen -source=../pkg/report/types/appender.go -destination=report/types/appender.go -typed
//go:generate mockgen -source=../pkg/report/types/conveyor.go -destination=report/types/conveyor.go -typed
//go:generate mockgen -source=../pkg/report/types/uploader.go -destination=report/types/uploader.go -typed
//go:generate mockgen -source=../pkg/report/types/types.go -destination=report/types/types.go -typed
package mock
