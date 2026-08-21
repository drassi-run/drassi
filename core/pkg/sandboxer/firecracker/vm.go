/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type vm struct {
	dir     string
	cmd     *exec.Cmd
	waitErr chan error
}

type fcConfig struct {
	BootSource        fcBootSource         `json:"boot-source"`
	Drives            []fcDrive            `json:"drives"`
	MachineConfig     fcMachineConfig      `json:"machine-config"`
	Vsock             fcVsock              `json:"vsock"`
	NetworkInterfaces []fcNetworkInterface `json:"network-interfaces,omitempty"`
}

type fcBootSource struct {
	KernelImagePath string `json:"kernel_image_path"`
	BootArgs        string `json:"boot_args,omitempty"`
	InitrdPath      string `json:"initrd_path,omitempty"`
}

type fcDrive struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
}

type fcMachineConfig struct {
	VcpuCount  int64 `json:"vcpu_count"`
	MemSizeMiB int64 `json:"mem_size_mib"`
	SMT        bool  `json:"smt"`
}

type fcVsock struct {
	GuestCID uint32 `json:"guest_cid"`
	UdsPath  string `json:"uds_path"`
}

type fcNetworkInterface struct {
	IfaceID     string `json:"iface_id"`
	HostDevName string `json:"host_dev_name"`
	GuestMAC    string `json:"guest_mac,omitempty"`
}

func startVM(cfg *Config, dir string) (*vm, error) {
	rootfs := filepath.Join(dir, "rootfs.ext4")
	if err := cloneFile(cfg.rootfs, rootfs); err != nil {
		return nil, fmt.Errorf("clone rootfs: %w", err)
	}

	apiSock := filepath.Join(dir, "firecracker.sock")
	vsockPath := filepath.Join(dir, "vsock.sock")
	configPath := filepath.Join(dir, "vm.json")
	_ = os.Remove(apiSock)
	_ = os.Remove(vsockPath)

	fc := fcConfig{
		BootSource: fcBootSource{
			KernelImagePath: cfg.kernel,
			BootArgs:        cfg.KernelArgs,
			InitrdPath:      cfg.Initrd,
		},
		Drives: []fcDrive{{
			DriveID:      "rootfs",
			PathOnHost:   rootfs,
			IsRootDevice: true,
			IsReadOnly:   false,
		}},
		MachineConfig: fcMachineConfig{
			VcpuCount:  cfg.VcpuCount,
			MemSizeMiB: cfg.MemSizeMiB,
		},
		Vsock: fcVsock{
			GuestCID: defaultGuestCID,
			UdsPath:  vsockPath,
		},
	}
	if cfg.TapDevice != "" {
		fc.NetworkInterfaces = []fcNetworkInterface{{
			IfaceID:     "eth0",
			HostDevName: cfg.TapDevice,
			GuestMAC:    cfg.GuestMAC,
		}}
	}

	body, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(configPath, body, 0o644); err != nil {
		return nil, err
	}

	logFile, err := os.Create(filepath.Join(dir, "firecracker.log"))
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cfg.Bin,
		"--api-sock", apiSock,
		"--config-file", configPath,
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err = cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, fmt.Errorf("start firecracker: %w", err)
	}

	machine := &vm{
		dir:     dir,
		cmd:     cmd,
		waitErr: make(chan error, 1),
	}
	go func() {
		err := cmd.Wait()
		_ = logFile.Close()
		machine.waitErr <- err
	}()
	return machine, nil
}

func (m *vm) vsockPath() string {
	return filepath.Join(m.dir, "vsock.sock")
}

func (m *vm) stop(ctx context.Context) error {
	if m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
	}
	select {
	case <-m.waitErr:
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
	}
	return os.RemoveAll(m.dir)
}

func cloneFile(src, dst string) error {
	cmd := exec.Command("cp", "--reflink=auto", "--sparse=always", src, dst)
	if err := cmd.Run(); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
