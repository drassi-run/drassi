package container

// https://github.com/moby/moby/blob/v27.3.1/api/types/network/network.go#L33-L45
// https://github.com/containers/common/blob/v0.60.4/libnetwork/types/network.go#L53-L88
type NetworkSpec struct {
	Name   string
	Labels map[string]string

	Driver  string
	Options map[string]string

	IPAMDriver  string
	IPAMOptions map[string]string
}
