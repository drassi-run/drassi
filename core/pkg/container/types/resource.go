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
	CPUShares int64
	CPUS      string
	Memory    int64

	//// Applicable to Windows
	CPUCount           int64
	CPUPercent         float32
	IOMaximumIOps      uint64
	IOMaximumBandwidth uint64

	//// Applicable to UNIX
	CPUPeriod      int64
	CPUQuota       int64
	CPURTPeriod    int64
	CPURTRuntime   int64
	CpusetCpus     string
	CpusetMems     string
	MemReservation int64
	MemSwapLimit   int64
	MemSwappiness  int64
	ShmSize        int64
	OomKillDisable bool
	OomScoreAdj    int64
	PidsLimit      int64

	BlkioConfig *BlkioConfig
	Ulimits     []*units.Ulimit
}

// BlkioConfig define blkio config
//   - [github.com/compose-spec/compose-go/v2/types.BlkioConfig]
type BlkioConfig struct {
	Weight          uint16
	WeightDevice    []WeightDevice
	DeviceReadBps   []ThrottleDevice
	DeviceReadIOps  []ThrottleDevice
	DeviceWriteBps  []ThrottleDevice
	DeviceWriteIOps []ThrottleDevice
}

// WeightDevice is a structure that holds device:weight pair
//   - [github.com/compose-spec/compose-go/v2/types.WeightDevice]
//   - [github.com/moby/moby/api/types/blkiodev.WeightDevice]
type WeightDevice struct {
	Path   string
	Weight uint16
}

func (w *WeightDevice) String() string {
	return fmt.Sprintf("%s:%d", w.Path, w.Weight)
}

// ThrottleDevice is a structure that holds device:rate_per_second pair
//   - [github.com/compose-spec/compose-go/v2/types.ThrottleDevice]
//   - [github.com/moby/moby/api/types/blkiodev.ThrottleDevice]
type ThrottleDevice struct {
	Path string
	Rate uint64
}

func (t *ThrottleDevice) String() string {
	return fmt.Sprintf("%s:%d", t.Path, t.Rate)
}
