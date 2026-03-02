//go:generate -command mockgen go tool mockgen

//go:generate mockgen -source=../pkg/log/chunker.go -destination=log/chunker.go -typed
//go:generate mockgen -source=../pkg/log/types.go -destination=log/types.go -typed
package mock
