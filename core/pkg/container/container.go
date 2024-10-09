package container

import "github.com/compose-spec/compose-go/v2/types"

//// [x]: supported     [~]: renamed     [-]: not supported     [ ]: WIP
//type ServiceConfig struct {
//	[ ] Name     string   `yaml:"name,omitempty" json:"-"`
//	[ ] Profiles []string `yaml:"profiles,omitempty" json:"profiles,omitempty"`
//
//	[x] Annotations  Mapping        `yaml:"annotations,omitempty" json:"annotations,omitempty"`
//	[ ] Attach       *bool          `yaml:"attach,omitempty" json:"attach,omitempty"`
//	[ ] Build        *BuildConfig   `yaml:"build,omitempty" json:"build,omitempty"`
//	[ ] Develop      *DevelopConfig `yaml:"develop,omitempty" json:"develop,omitempty"`
//	[x] BlkioConfig  *BlkioConfig   `yaml:"blkio_config,omitempty" json:"blkio_config,omitempty"`
//	[x] CapAdd       []string       `yaml:"cap_add,omitempty" json:"cap_add,omitempty"`
//	[x] CapDrop      []string       `yaml:"cap_drop,omitempty" json:"cap_drop,omitempty"`
//	[x] CgroupParent string         `yaml:"cgroup_parent,omitempty" json:"cgroup_parent,omitempty"`
//	[x] Cgroup       string         `yaml:"cgroup,omitempty" json:"cgroup,omitempty"`
//	[x] CPUCount     int64          `yaml:"cpu_count,omitempty" json:"cpu_count,omitempty"`
//	[x] CPUPercent   float32        `yaml:"cpu_percent,omitempty" json:"cpu_percent,omitempty"`
//	[x] CPUPeriod    int64          `yaml:"cpu_period,omitempty" json:"cpu_period,omitempty"`
//	[x] CPUQuota     int64          `yaml:"cpu_quota,omitempty" json:"cpu_quota,omitempty"`
//	[x] CPURTPeriod  int64          `yaml:"cpu_rt_period,omitempty" json:"cpu_rt_period,omitempty"`
//	[x] CPURTRuntime int64          `yaml:"cpu_rt_runtime,omitempty" json:"cpu_rt_runtime,omitempty"`
//	[x] CPUS         float32        `yaml:"cpus,omitempty" json:"cpus,omitempty"`
//	[~] CPUSet       string         `yaml:"cpuset,omitempty" json:"cpuset,omitempty"` // renamed to `cpuset_cpus`
//	[x] CPUShares    int64          `yaml:"cpu_shares,omitempty" json:"cpu_shares,omitempty"`
//
//	// Command for the service containers.
//	// If set, overrides COMMAND from the image.
//	//
//	// Set to `[]` or an empty string to clear the command from the image.
//	[x] Command ShellCommand `yaml:"command,omitempty" json:"command"` // NOTE: we can NOT omitempty for JSON! see ShellCommand type for details.
//
//	[ ] Configs           []ServiceConfigObjConfig `yaml:"configs,omitempty" json:"configs,omitempty"`
//	[v] ContainerName     string                   `yaml:"container_name,omitempty" json:"container_name,omitempty"`
//	[ ] CredentialSpec    *CredentialSpecConfig    `yaml:"credential_spec,omitempty" json:"credential_spec,omitempty"`
//	[ ] DependsOn         DependsOnConfig          `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
//	[ ] Deploy            *DeployConfig            `yaml:"deploy,omitempty" json:"deploy,omitempty"`
//	[x] DeviceCgroupRules []string                 `yaml:"device_cgroup_rules,omitempty" json:"device_cgroup_rules,omitempty"`
//	[x] Devices           []string                 `yaml:"devices,omitempty" json:"devices,omitempty"`
//	[x] DNS               StringList               `yaml:"dns,omitempty" json:"dns,omitempty"`
//	[x] DNSOpts           []string                 `yaml:"dns_opt,omitempty" json:"dns_opt,omitempty"`
//	[x] DNSSearch         StringList               `yaml:"dns_search,omitempty" json:"dns_search,omitempty"`
//	[ ] Dockerfile        string                   `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
//	[x] DomainName        string                   `yaml:"domainname,omitempty" json:"domainname,omitempty"`
//
//	// Entrypoint for the service containers.
//	// If set, overrides ENTRYPOINT from the image.
//	//
//	// Set to `[]` or an empty string to clear the entrypoint from the image.
//	[x] Entrypoint ShellCommand `yaml:"entrypoint,omitempty" json:"entrypoint"` // NOTE: we can NOT omitempty for JSON! see ShellCommand type for details.
//
//	[x] Environment     MappingWithEquals                `yaml:"environment,omitempty" json:"environment,omitempty"`
//	[x] EnvFiles        []EnvFile                        `yaml:"env_file,omitempty" json:"env_file,omitempty"`
//	[x] Expose          StringOrNumberList               `yaml:"expose,omitempty" json:"expose,omitempty"`
//	[ ] Extends         *ExtendsConfig                   `yaml:"extends,omitempty" json:"extends,omitempty"`
//	[ ] ExternalLinks   []string                         `yaml:"external_links,omitempty" json:"external_links,omitempty"`
//	[x] ExtraHosts      HostsList                        `yaml:"extra_hosts,omitempty" json:"extra_hosts,omitempty"`
//	[x] GroupAdd        []string                         `yaml:"group_add,omitempty" json:"group_add,omitempty"`
//	[x] Hostname        string                           `yaml:"hostname,omitempty" json:"hostname,omitempty"`
//	[x] HealthCheck     *HealthCheckConfig               `yaml:"healthcheck,omitempty" json:"healthcheck,omitempty"`
//	[x] Image           string                           `yaml:"image,omitempty" json:"image,omitempty"`
//	[ ] Init            *bool                            `yaml:"init,omitempty" json:"init,omitempty"`
//	[x] Ipc             string                           `yaml:"ipc,omitempty" json:"ipc,omitempty"`
//	[x] Isolation       string                           `yaml:"isolation,omitempty" json:"isolation,omitempty"`
//	[x] Labels          Labels                           `yaml:"labels,omitempty" json:"labels,omitempty"`
//	[ ] CustomLabels    Labels                           `yaml:"-" json:"-"`
//	[ ] Links           []string                         `yaml:"links,omitempty" json:"links,omitempty"`
//	[x] Logging         *LoggingConfig                   `yaml:"logging,omitempty" json:"logging,omitempty"`
//	[-] LogDriver       string                           `yaml:"log_driver,omitempty" json:"log_driver,omitempty"` // use Logging.Driver instead
//	[-] LogOpt          map[string]string                `yaml:"log_opt,omitempty" json:"log_opt,omitempty"`	// use Logging.Options instead
//	[~] MemLimit        UnitBytes                        `yaml:"mem_limit,omitempty" json:"mem_limit,omitempty"` // renamed to `memory`
//	[x] MemReservation  UnitBytes                        `yaml:"mem_reservation,omitempty" json:"mem_reservation,omitempty"`
//	[x] MemSwapLimit    UnitBytes                        `yaml:"memswap_limit,omitempty" json:"memswap_limit,omitempty"`
//	[x] MemSwappiness   UnitBytes                        `yaml:"mem_swappiness,omitempty" json:"mem_swappiness,omitempty"`
//	[ ] MacAddress      string                           `yaml:"mac_address,omitempty" json:"mac_address,omitempty"`
//	[-] Net             string                           `yaml:"net,omitempty" json:"net,omitempty"`
//	[x] NetworkMode     string                           `yaml:"network_mode,omitempty" json:"network_mode,omitempty"`
//	[x] Networks        map[string]*ServiceNetworkConfig `yaml:"networks,omitempty" json:"networks,omitempty"`
//	[x] OomKillDisable  bool                             `yaml:"oom_kill_disable,omitempty" json:"oom_kill_disable,omitempty"`
//	[x] OomScoreAdj     int64                            `yaml:"oom_score_adj,omitempty" json:"oom_score_adj,omitempty"`
//	[x] Pid             string                           `yaml:"pid,omitempty" json:"pid,omitempty"`
//	[x] PidsLimit       int64                            `yaml:"pids_limit,omitempty" json:"pids_limit,omitempty"`
//	[x] Platform        string                           `yaml:"platform,omitempty" json:"platform,omitempty"`
//	[x] Ports           []ServicePortConfig              `yaml:"ports,omitempty" json:"ports,omitempty"`
//	[x] Privileged      bool                             `yaml:"privileged,omitempty" json:"privileged,omitempty"`
//	[x] PullPolicy      string                           `yaml:"pull_policy,omitempty" json:"pull_policy,omitempty"`
//	[ ] ReadOnly        bool                             `yaml:"read_only,omitempty" json:"read_only,omitempty"`
//	[ ] Restart         string                           `yaml:"restart,omitempty" json:"restart,omitempty"`
//	[x] Runtime         string                           `yaml:"runtime,omitempty" json:"runtime,omitempty"`
//	[ ] Scale           *int                             `yaml:"scale,omitempty" json:"scale,omitempty"`
//	[ ] Secrets         []ServiceSecretConfig            `yaml:"secrets,omitempty" json:"secrets,omitempty"`
//	[x] SecurityOpt     []string                         `yaml:"security_opt,omitempty" json:"security_opt,omitempty"`
//	[x] ShmSize         UnitBytes                        `yaml:"shm_size,omitempty" json:"shm_size,omitempty"`
//	[ ] StdinOpen       bool                             `yaml:"stdin_open,omitempty" json:"stdin_open,omitempty"`
//	[x] StopGracePeriod *Duration                        `yaml:"stop_grace_period,omitempty" json:"stop_grace_period,omitempty"`
//	[x] StopSignal      string                           `yaml:"stop_signal,omitempty" json:"stop_signal,omitempty"`
//	[x] StorageOpt      map[string]string                `yaml:"storage_opt,omitempty" json:"storage_opt,omitempty"`
//	[x] Sysctls         Mapping                          `yaml:"sysctls,omitempty" json:"sysctls,omitempty"`
//	[x] Tmpfs           StringList                       `yaml:"tmpfs,omitempty" json:"tmpfs,omitempty"`
//	[ ] Tty             bool                             `yaml:"tty,omitempty" json:"tty,omitempty"`
//	[x] Ulimits         map[string]*UlimitsConfig        `yaml:"ulimits,omitempty" json:"ulimits,omitempty"`
//	[x] User            string                           `yaml:"user,omitempty" json:"user,omitempty"`
//	[x] UserNSMode      string                           `yaml:"userns_mode,omitempty" json:"userns_mode,omitempty"`
//	[x] Uts             string                           `yaml:"uts,omitempty" json:"uts,omitempty"`
//	[x] VolumeDriver    string                           `yaml:"volume_driver,omitempty" json:"volume_driver,omitempty"`
//	[x] Volumes         []ServiceVolumeConfig            `yaml:"volumes,omitempty" json:"volumes,omitempty"`
//	[x] VolumesFrom     []string                         `yaml:"volumes_from,omitempty" json:"volumes_from,omitempty"`
//	[x] WorkingDir      string                           `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
//
//	Extensions Extensions `yaml:"#extensions,inline,omitempty" json:"-"`
//}

type ContainerSpec struct {
	Annotations   types.Mapping `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Labels        types.Labels  `yaml:"labels,omitempty" json:"labels,omitempty"`
	ContainerName string        `yaml:"container_name,omitempty" json:"container_name,omitempty"`
	Image         string        `yaml:"image,omitempty" json:"image,omitempty"`
	PullPolicy    string        `yaml:"pull_policy,omitempty" json:"pull_policy,omitempty"`
	// NOTE: we can NOT omitempty for JSON! see ShellCommand type for details.
	Command     types.ShellCommand      `yaml:"command,omitempty" json:"command"`
	Entrypoint  types.ShellCommand      `yaml:"entrypoint,omitempty" json:"entrypoint"`
	Environment types.MappingWithEquals `yaml:"environment,omitempty" json:"environment,omitempty"`
	EnvFiles    []types.EnvFile         `yaml:"env_file,omitempty" json:"env_file,omitempty"`
	WorkingDir  string                  `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`

	// Networking
	Networks   map[string]*types.ServiceNetworkConfig `yaml:"networks,omitempty" json:"networks,omitempty"`
	Ports      []types.ServicePortConfig              `yaml:"ports,omitempty" json:"ports,omitempty"`
	Expose     types.StringOrNumberList               `yaml:"expose,omitempty" json:"expose,omitempty"`
	DNS        types.StringList                       `yaml:"dns,omitempty" json:"dns,omitempty"`
	DNSOpts    []string                               `yaml:"dns_opt,omitempty" json:"dns_opt,omitempty"`
	DNSSearch  types.StringList                       `yaml:"dns_search,omitempty" json:"dns_search,omitempty"`
	DomainName string                                 `yaml:"domainname,omitempty" json:"domainname,omitempty"`
	Hostname   string                                 `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	ExtraHosts types.HostsList                        `yaml:"extra_hosts,omitempty" json:"extra_hosts,omitempty"`

	// Storage & Device
	Volumes           []types.ServiceVolumeConfig `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	VolumeDriver      string                      `yaml:"volume_driver,omitempty" json:"volume_driver,omitempty"`
	VolumesFrom       []string                    `yaml:"volumes_from,omitempty" json:"volumes_from,omitempty"`
	Tmpfs             types.StringList            `yaml:"tmpfs,omitempty" json:"tmpfs,omitempty"`
	StorageOpt        map[string]string           `yaml:"storage_opt,omitempty" json:"storage_opt,omitempty"`
	Devices           []string                    `yaml:"devices,omitempty" json:"devices,omitempty"`
	DeviceCgroupRules []string                    `yaml:"device_cgroup_rules,omitempty" json:"device_cgroup_rules,omitempty"`

	// Runtime
	Runtime         string                   `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Platform        string                   `yaml:"platform,omitempty" json:"platform,omitempty"`
	Isolation       string                   `yaml:"isolation,omitempty" json:"isolation,omitempty"` // Windows only
	StopSignal      string                   `yaml:"stop_signal,omitempty" json:"stop_signal,omitempty"`
	StopGracePeriod *types.Duration          `yaml:"stop_grace_period,omitempty" json:"stop_grace_period,omitempty"`
	Logging         *types.LoggingConfig     `yaml:"logging,omitempty" json:"logging,omitempty"`
	HealthCheck     *types.HealthCheckConfig `yaml:"healthcheck,omitempty" json:"healthcheck,omitempty"`

	// Resources
	//// Applicable to all platforms
	CPUShares int64           `yaml:"cpu_shares,omitempty" json:"cpu_shares,omitempty"`
	CPUS      string          `yaml:"cpus,omitempty" json:"cpus,omitempty"`     // TODO: allow unmarshal from number
	Memory    types.UnitBytes `yaml:"memory,omitempty" json:"memory,omitempty"` // named `mem_limit` (deprecated) in compose.yaml
	//// Applicable to Windows
	CPUCount   int64   `yaml:"cpu_count,omitempty" json:"cpu_count,omitempty"`
	CPUPercent float32 `yaml:"cpu_percent,omitempty" json:"cpu_percent,omitempty"`
	//// Applicable to UNIX
	CPUPeriod      int64           `yaml:"cpu_period,omitempty" json:"cpu_period,omitempty"`
	CPUQuota       int64           `yaml:"cpu_quota,omitempty" json:"cpu_quota,omitempty"`
	CPURTPeriod    int64           `yaml:"cpu_rt_period,omitempty" json:"cpu_rt_period,omitempty"`
	CPURTRuntime   int64           `yaml:"cpu_rt_runtime,omitempty" json:"cpu_rt_runtime,omitempty"`
	CpusetCpus     string          `yaml:"cpuset_cpus,omitempty" json:"cpuset_cpus,omitempty"` // named `cpuset` in compose.yaml
	CpusetMems     string          `yaml:"cpuset_mems,omitempty" json:"cpuset_mems,omitempty"` // not yet available in compose.yaml
	MemReservation types.UnitBytes `yaml:"mem_reservation,omitempty" json:"mem_reservation,omitempty"`
	MemSwapLimit   types.UnitBytes `yaml:"memswap_limit,omitempty" json:"memswap_limit,omitempty"`
	MemSwappiness  types.UnitBytes `yaml:"mem_swappiness,omitempty" json:"mem_swappiness,omitempty"`
	ShmSize        types.UnitBytes `yaml:"shm_size,omitempty" json:"shm_size,omitempty"`
	OomKillDisable bool            `yaml:"oom_kill_disable,omitempty" json:"oom_kill_disable,omitempty"`
	OomScoreAdj    int64           `yaml:"oom_score_adj,omitempty" json:"oom_score_adj,omitempty"`
	PidsLimit      int64           `yaml:"pids_limit,omitempty" json:"pids_limit,omitempty"`

	BlkioConfig *types.BlkioConfig              `yaml:"blkio_config,omitempty" json:"blkio_config,omitempty"`
	Ulimits     map[string]*types.UlimitsConfig `yaml:"ulimits,omitempty" json:"ulimits,omitempty"`

	// Namespace & CGroup
	NetworkMode  string `yaml:"network_mode,omitempty" json:"network_mode,omitempty"`
	IpcMode      string `yaml:"ipc,omitempty" json:"ipc,omitempty"`
	PidMode      string `yaml:"pid,omitempty" json:"pid,omitempty"`
	UTSMode      string `yaml:"uts,omitempty" json:"uts,omitempty"`
	UserMode     string `yaml:"userns_mode,omitempty" json:"userns_mode,omitempty"`
	CgroupMode   string `yaml:"cgroup,omitempty" json:"cgroup,omitempty"`
	CgroupParent string `yaml:"cgroup_parent,omitempty" json:"cgroup_parent,omitempty"`

	// Security
	User        string        `yaml:"user,omitempty" json:"user,omitempty"`
	GroupAdd    []string      `yaml:"group_add,omitempty" json:"group_add,omitempty"`
	CapAdd      []string      `yaml:"cap_add,omitempty" json:"cap_add,omitempty"`
	CapDrop     []string      `yaml:"cap_drop,omitempty" json:"cap_drop,omitempty"`
	Privileged  bool          `yaml:"privileged,omitempty" json:"privileged,omitempty"`
	SecurityOpt []string      `yaml:"security_opt,omitempty" json:"security_opt,omitempty"`
	Sysctls     types.Mapping `yaml:"sysctls,omitempty" json:"sysctls,omitempty"`
}
