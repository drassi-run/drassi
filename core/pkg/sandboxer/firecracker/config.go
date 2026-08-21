/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

type Config struct {
	Bin           string `toml:"bin" json:"bin,omitempty"`
	RootDir       string `toml:"root_dir" json:"rootDir,omitempty"`
	KernelPath    string `toml:"kernel_path,omitempty" json:"kernelPath,omitempty"`
	RootfsPath    string `toml:"rootfs_path,omitempty" json:"rootfsPath,omitempty"`
	Agent         string `toml:"agent,omitempty" json:"agent,omitempty"`
	RootfsSizeMiB int64  `toml:"rootfs_size_mib,omitempty" json:"rootfsSizeMib,omitempty"`
	Initrd        string `toml:"initrd,omitempty" json:"initrd,omitempty"`
	KernelArgs    string `toml:"kernel_args,omitempty" json:"kernelArgs,omitempty"`
	VcpuCount     int64  `toml:"vcpu_count,omitempty" json:"vcpuCount,omitempty"`
	MemSizeMiB    int64  `toml:"mem_size_mib,omitempty" json:"memSizeMib,omitempty"`
	AgentPort     uint32 `toml:"agent_port,omitempty" json:"agentPort,omitempty"`
	TapDevice     string `toml:"tap_device,omitempty" json:"tapDevice,omitempty"`
	GuestMAC      string `toml:"guest_mac,omitempty" json:"guestMac,omitempty"`
	AgentWait     int    `toml:"agent_wait_sec,omitempty" json:"agentWaitSec,omitempty"`

	// kernel and rootfs are materialized host paths, filled by ensureKernel/ensureRootfs.
	kernel string
	rootfs string
}

func DefaultConfig() *Config {
	return &Config{
		Bin:         "firecracker",
		RootDir:     "/var/lib/drassi/firecracker",
		RootfsPath:  "oci://" + defaultRootfsImage,
		KernelArgs:  "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=" + guestTiniPath + " -- " + guestAgentPath,
		VcpuCount:   2,
		MemSizeMiB:  2048,
		AgentPort:   defaultAgentPort,
		AgentWait:   30,
	}
}
