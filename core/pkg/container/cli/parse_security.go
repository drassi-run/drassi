/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	docker "github.com/docker/docker/api/types/container"
)

func (fm *flagMapper) mapSecurity(copts *containerOptions) error {
	spec := fm.Spec

	// Namespace & CGroup
	if pidMode := docker.PidMode(copts.pidMode); !pidMode.Valid() {
		return fmt.Errorf("--pid: invalid PID mode")
	} else {
		spec.PidMode = string(pidMode)
	}

	if utsMode := docker.UTSMode(copts.utsMode); !utsMode.Valid() {
		return fmt.Errorf("--uts: invalid UTS mode")
	} else {
		spec.UTSMode = string(utsMode)
	}

	if usernsMode := docker.UsernsMode(copts.usernsMode); !usernsMode.Valid() {
		return fmt.Errorf("--userns: invalid USER mode")
	} else {
		spec.UserMode = string(usernsMode)
	}

	if cgroupnsMode := docker.CgroupnsMode(copts.cgroupnsMode); !cgroupnsMode.Valid() {
		return fmt.Errorf("--cgroupns: invalid CGROUP mode")
	} else {
		spec.CgroupMode = string(cgroupnsMode)
	}
	spec.CgroupParent = copts.cgroupParent
	spec.NetworkMode = copts.netMode.NetworkMode()

	// Security
	spec.User = copts.user
	spec.GroupAdd = copts.groupAdd.GetAll()
	spec.CapAdd = copts.capAdd.GetAll()
	spec.CapDrop = copts.capDrop.GetAll()
	spec.Privileged = copts.privileged
	spec.SecurityOpt = copts.securityOpt.GetAll()

	if securityOpts, err := parseSecurityOpts(copts.securityOpt.GetAll()); err != nil {
		return err
	} else {
		spec.SecurityOpt = securityOpts // TODO: parseSystemPaths https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L542
	}

	spec.Sysctls = copts.sysctls.GetAll()

	return nil
}

const (
	// seccompProfileDefault is the built-in default seccomp profile.
	seccompProfileDefault = "builtin"
	// seccompProfileUnconfined is a special profile name for seccomp to use an
	// "unconfined" seccomp profile.
	seccompProfileUnconfined = "unconfined"
)

// takes a local seccomp daemon, reads the file contents for sending to the daemon
// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L921-L952
func parseSecurityOpts(securityOpts []string) ([]string, error) {
	for key, opt := range securityOpts {
		k, v, ok := strings.Cut(opt, "=")
		if !ok && k != "no-new-privileges" {
			k, v, ok = strings.Cut(opt, ":")
		}
		if (!ok || v == "") && k != "no-new-privileges" {
			// "no-new-privileges" is the only option that does not require a value.
			return securityOpts, fmt.Errorf("invalid --security-opt: %q", opt)
		}
		if k == "seccomp" {
			switch v {
			case seccompProfileDefault, seccompProfileUnconfined:
				// known special names for built-in profiles, nothing to do.
			default:
				// value may be a filename, in which case we send the profile's
				// content if it's valid JSON.
				f, err := os.ReadFile(v)
				if err != nil {
					return securityOpts, fmt.Errorf("opening seccomp profile (%s) failed: %w", v, err)
				}
				b := bytes.NewBuffer(nil)
				if err := json.Compact(b, f); err != nil {
					return securityOpts, fmt.Errorf("compacting json for seccomp profile (%s) failed: %w", v, err)
				}
				securityOpts[key] = fmt.Sprintf("seccomp=%s", b.Bytes())
			}
		}
	}

	return securityOpts, nil
}
