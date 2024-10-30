package types

type ContainerSecurity struct {
	// Namespace & CGroup
	NetworkMode  string
	IpcMode      string
	PidMode      string
	UTSMode      string
	UserMode     string
	CgroupMode   string
	CgroupParent string

	// Security
	User        string
	GroupAdd    []string
	CapAdd      []string
	CapDrop     []string
	Privileged  bool
	SecurityOpt []string
	Sysctls     map[string]string
}
