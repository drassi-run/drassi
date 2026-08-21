//go:build linux

/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"drassi.run/core/pkg/sandboxer/firecracker"
	"github.com/mdlayher/vsock"
)

func main() {
	port := flag.Uint("port", 1024, "vsock port to listen on")
	flag.Parse()

	prepareInit()

	var ln *vsock.Listener
	var err error
	for i := 0; i < 50; i++ {
		ln, err = vsock.Listen(uint32(*port), nil)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		log.Fatalf("listen vsock: %v", err)
	}
	defer ln.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := firecracker.Serve(ctx, ln); err != nil {
		log.Fatal(err)
	}
}

func prepareInit() {
	for _, dir := range []string{"/dev", "/proc", "/sys", "/tmp", "/opt"} {
		_ = os.MkdirAll(dir, 0o755)
	}
	_ = syscall.Mount("devtmpfs", "/dev", "devtmpfs", 0, "")
	_ = syscall.Mount("proc", "/proc", "proc", 0, "")
	_ = syscall.Mount("sysfs", "/sys", "sysfs", 0, "")
	_ = os.MkdirAll("/sys/fs/cgroup", 0o755)
	_ = syscall.Mount("cgroup2", "/sys/fs/cgroup", "cgroup2", 0, "")
	if os.Getenv("PATH") == "" {
		_ = os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	}
}
