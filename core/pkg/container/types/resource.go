/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"fmt"

	"github.com/docker/go-units"
)

type ContainerResource struct {
	//// Applicable to all platforms
	CPUShares int64  `json:"cpu_shares,omitempty"`
	CPUS      string `json:"cpus,omitempty"`
	Memory    int64  `json:"memory,omitempty"`

	//// Applicable to Windows
	CPUCount           int64   `json:"cpu_count,omitempty"`
	CPUPercent         float32 `json:"cpu_percent,omitempty"`
	IOMaximumIOps      uint64  `json:"io_max_iops,omitempty"`
	IOMaximumBandwidth uint64  `json:"io_max_bandwidth,omitempty"`

	//// Applicable to UNIX
	CPUPeriod      int64  `json:"cpu_period,omitempty"`
	CPUQuota       int64  `json:"cpu_quota,omitempty"`
	CPURTPeriod    int64  `json:"cpu_rt_period,omitempty"`
	CPURTRuntime   int64  `json:"cpu_rt_runtime,omitempty"`
	CpusetCpus     string `json:"cpuset_cpus,omitempty"`
	CpusetMems     string `json:"cpuset_mems,omitempty"`
	MemReservation int64  `json:"mem_reservation,omitempty"`
	MemSwapLimit   int64  `json:"mem_swap_limit,omitempty"`
	MemSwappiness  int64  `json:"mem_swappiness,omitempty"`
	ShmSize        int64  `json:"shm_size,omitempty"`
	OomKillDisable bool   `json:"oom_kill_disable,omitempty"`
	OomScoreAdj    int64  `json:"oom_score_adj,omitempty"`
	PidsLimit      int64  `json:"pids_limit,omitempty"`

	BlkioConfig *BlkioConfig    `json:"blkio_config,omitempty"`
	Ulimits     []*units.Ulimit `json:"ulimits,omitempty"`
}

// BlkioConfig define blkio config
//   - [github.com/compose-spec/compose-go/v2/types.BlkioConfig]
type BlkioConfig struct {
	Weight          uint16           `json:"weight,omitempty"`
	WeightDevice    []WeightDevice   `json:"weight_device,omitempty"`
	DeviceReadBps   []ThrottleDevice `json:"device_read_bps,omitempty"`
	DeviceReadIOps  []ThrottleDevice `json:"device_read_iops,omitempty"`
	DeviceWriteBps  []ThrottleDevice `json:"device_write_bps,omitempty"`
	DeviceWriteIOps []ThrottleDevice `json:"device_write_iops,omitempty"`
}

// WeightDevice is a structure that holds device:weight pair
//   - [github.com/compose-spec/compose-go/v2/types.WeightDevice]
//   - [github.com/moby/moby/api/types/blkiodev.WeightDevice]
type WeightDevice struct {
	Path   string `json:"path,omitempty"`
	Weight uint16 `json:"weight,omitempty"`
}

func (w *WeightDevice) String() string {
	return fmt.Sprintf("%s:%d", w.Path, w.Weight)
}

// ThrottleDevice is a structure that holds device:rate_per_second pair
//   - [github.com/compose-spec/compose-go/v2/types.ThrottleDevice]
//   - [github.com/moby/moby/api/types/blkiodev.ThrottleDevice]
type ThrottleDevice struct {
	Path string `json:"path,omitempty"`
	Rate uint64 `json:"rate,omitempty"`
}

func (t *ThrottleDevice) String() string {
	return fmt.Sprintf("%s:%d", t.Path, t.Rate)
}
