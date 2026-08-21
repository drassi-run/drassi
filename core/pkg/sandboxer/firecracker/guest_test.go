/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	fcKernelBase = "https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/20260819-0a745def42dd-0"
	fcKernelVer  = "vmlinux-6.1.177"
)

const echoMain = `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	nl := true
	if len(args) > 0 && args[0] == "-n" {
		nl = false
		args = args[1:]
	}
	fmt.Print(strings.Join(args, " "))
	if nl {
		fmt.Println()
	}
}
`

const falseMain = `package main

import "os"

func main() { os.Exit(1) }
`

var guestOnce struct {
	sync.Once
	kernel string
	rootfs string
	err    error
}

func requireFirecracker(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping firecracker vm test in short mode")
	}
	if _, err := exec.LookPath("firecracker"); err != nil {
		t.Skip("firecracker not installed")
	}
	f, err := os.OpenFile("/dev/kvm", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("/dev/kvm not accessible: %v", err)
	}
	_ = f.Close()
}

func guestImages(t *testing.T) (kernel, rootfs string) {
	t.Helper()
	guestOnce.Do(func() {
		guestOnce.err = prepareGuestImages()
	})
	if guestOnce.err != nil {
		t.Fatalf("guest images: %v", guestOnce.err)
	}
	return guestOnce.kernel, guestOnce.rootfs
}

func prepareGuestImages() error {
	cache, err := testdataDir()
	if err != nil {
		return err
	}

	kernel := filepath.Join(cache, fcKernelVer)
	if err = downloadFile(kernelURL(), kernel); err != nil {
		return fmt.Errorf("kernel: %w", err)
	}

	agent := filepath.Join(cache, "firecracker-agent")
	if err = buildAgent(agent); err != nil {
		return fmt.Errorf("agent: %w", err)
	}
	echoBin := filepath.Join(cache, "echo")
	if err = buildGuestCmd(echoBin, echoMain); err != nil {
		return fmt.Errorf("echo: %w", err)
	}
	falseBin := filepath.Join(cache, "false")
	if err = buildGuestCmd(falseBin, falseMain); err != nil {
		return fmt.Errorf("false: %w", err)
	}

	rootfs := filepath.Join(cache, "rootfs.ext4")
	if err = buildRootfs(rootfs, agent, echoBin, falseBin); err != nil {
		return fmt.Errorf("rootfs: %w", err)
	}

	guestOnce.kernel = kernel
	guestOnce.rootfs = rootfs
	return nil
}

func testdataDir() (string, error) {
	dir := os.Getenv("DRASSI_FC_TESTDATA")
	if dir == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			cache = os.TempDir()
		}
		dir = filepath.Join(cache, "drassi", "firecracker")
	}
	return dir, os.MkdirAll(dir, 0o755)
}

func kernelURL() string {
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	if arch == "arm64" {
		arch = "aarch64"
	}
	return fcKernelBase + "/" + arch + "/" + fcKernelVer
}

func downloadFile(url, dest string) error {
	if st, err := os.Stat(dest); err == nil && st.Size() > 0 {
		return nil
	}

	tmp := dest + ".tmp"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	cerr := f.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if cerr != nil {
		_ = os.Remove(tmp)
		return cerr
	}
	return os.Rename(tmp, dest)
}

func moduleRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../..")), nil
}

func buildAgent(dest string) error {
	modRoot, err := moduleRoot()
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", dest, "./cmd/firecracker-agent")
	cmd.Dir = modRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return os.Chmod(dest, 0o755)
}

func buildGuestCmd(dest, src string) error {
	dir, err := os.MkdirTemp("", "drassi-fc-cmd-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err = os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module guestcmd\n\ngo 1.26\n"), 0o644); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", dest, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return os.Chmod(dest, 0o755)
}

func buildRootfs(image, agent, echoBin, falseBin string) error {
	dir, err := os.MkdirTemp("", "drassi-fc-rootfs-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	for _, d := range []string{"dev", "proc", "sys", "tmp", "opt", "bin", "sbin"} {
		if err = os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			return err
		}
	}
	files := map[string]string{
		filepath.Join(dir, "sbin", "init"): agent,
		filepath.Join(dir, "bin", "echo"):  echoBin,
		filepath.Join(dir, "bin", "false"): falseBin,
	}
	for dst, src := range files {
		if err = copyFile(src, dst); err != nil {
			return err
		}
		if err = os.Chmod(dst, 0o755); err != nil {
			return err
		}
	}

	_ = os.Remove(image)
	f, err := os.Create(image)
	if err != nil {
		return err
	}
	if err = f.Truncate(64 << 20); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()

	cmd := exec.Command("mkfs.ext4", "-q", "-F", "-d", dir, image)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mkfs.ext4: %w: %s", err, out)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func dumpVMLog(t *testing.T, dir string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "firecracker.log"))
	if err != nil {
		return
	}
	t.Logf("firecracker.log:\n%s", b)
}
