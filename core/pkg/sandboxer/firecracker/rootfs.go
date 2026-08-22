/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"drassi.run/core/util/fs"
)

const (
	defaultRootfsImage = "ghcr.io/drassi-run/ubuntu:26.04"
	defaultRootfsMiB   = 2048
	guestAgentPath     = "/usr/sbin/drassi-agent"
	guestInitPath      = "/usr/sbin/drassi-init"
	guestTiniPath      = "/usr/bin/tini"
)

// guestInitScript is PID 1's child under tini. It starts dockerd if present,
// then execs the guest agent so both are in tini's process tree.
const guestInitScript = `#!/bin/sh
export PATH="${PATH:-/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin}"

for d in /dev /proc /sys /run /tmp /opt /var/log /var/lib/docker; do
	mkdir -p "$d"
done

mount -t devtmpfs devtmpfs /dev 2>/dev/null || true
mount -t proc proc /proc 2>/dev/null || true
mount -t sysfs sysfs /sys 2>/dev/null || true
mkdir -p /sys/fs/cgroup
mount -t cgroup2 cgroup2 /sys/fs/cgroup 2>/dev/null || true

if command -v dockerd >/dev/null 2>&1; then
	dockerd >/var/log/dockerd.log 2>&1 &
	i=0
	while [ "$i" -lt 30 ]; do
		if docker info >/dev/null 2>&1; then
			break
		fi
		i=$((i + 1))
		sleep 1
	done
fi

exec ` + guestAgentPath + ` "$@"
`

func (c *Config) ensureRootfs() error {
	if c.rootfs != "" {
		if _, err := os.Stat(c.rootfs); err != nil {
			return fmt.Errorf("rootfs %q: %w", c.rootfs, err)
		}
		return nil
	}
	if c.RootfsPath == "" {
		return fmt.Errorf("firecracker rootfs_path is required")
	}
	path, err := c.materializeRootfs()
	if err != nil {
		return err
	}
	c.rootfs = path
	return nil
}

func (c *Config) materializeRootfs() (string, error) {
	scheme, loc, err := parseRootfsRef(c.RootfsPath)
	if err != nil {
		return "", err
	}
	switch scheme {
	case "file":
		return c.materializeFileRootfs(loc)
	case "oci":
		return c.materializeOCIRootfs(loc)
	default:
		return "", fmt.Errorf("unsupported rootfs_path scheme %q", scheme)
	}
}

func (c *Config) materializeFileRootfs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if _, err = os.Stat(abs); err != nil {
		return "", fmt.Errorf("rootfs %q: %w", abs, err)
	}
	return abs, nil
}

func (c *Config) materializeOCIRootfs(image string) (string, error) {
	digest, err := resolveImageDigest(image)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(c.RootDir, "cache", digestCacheDir(digest))
	out := filepath.Join(dir, "rootfs.ext4")
	if st, err := os.Stat(out); err == nil && st.Size() > 0 {
		return out, nil
	}
	if err = os.MkdirAll(dir, xfs.DirPerm); err != nil {
		return "", err
	}

	agent, err := resolveAgent(c.Agent)
	if err != nil {
		return "", err
	}

	staging, err := os.MkdirTemp(filepath.Join(c.RootDir, "cache"), "rootfs-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(staging)

	if err = exportImage(image, staging); err != nil {
		return "", err
	}
	if err = installGuestAgent(staging, agent); err != nil {
		return "", err
	}
	if err = installGuestInit(staging); err != nil {
		return "", err
	}

	sizeMiB := c.RootfsSizeMiB
	if sizeMiB <= 0 {
		sizeMiB = defaultRootfsMiB
	}
	tmp := out + ".tmp"
	_ = os.Remove(tmp)
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	if err = f.Truncate(int64(sizeMiB) << 20); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return "", err
	}
	_ = f.Close()

	cmd := exec.Command("mkfs.ext4", "-q", "-F", "-d", staging, tmp)
	if outb, err := cmd.CombinedOutput(); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("mkfs.ext4: %w: %s", err, outb)
	}
	if err = os.Rename(tmp, out); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return out, nil
}

func resolveImageDigest(image string) (string, error) {
	if d := namedDigest(image); d != "" {
		return d, nil
	}
	if err := dockerPull(image); err != nil {
		return "", err
	}
	return inspectDigest(image)
}

func namedDigest(ref string) string {
	i := strings.LastIndex(ref, "@")
	if i < 0 {
		return ""
	}
	d := ref[i+1:]
	algo, hex, ok := strings.Cut(d, ":")
	if !ok || algo == "" || hex == "" {
		return ""
	}
	return d
}

func digestCacheDir(digest string) string {
	return strings.ReplaceAll(digest, ":", "-")
}

type dockerImageInspect struct {
	ID          string   `json:"Id"`
	RepoDigests []string `json:"RepoDigests"`
}

func inspectDigest(image string) (string, error) {
	cmd := exec.Command("docker", "image", "inspect", "--format", "{{json .}}", image)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("docker inspect %s: %w: %s", image, err, ee.Stderr)
		}
		return "", fmt.Errorf("docker inspect %s: %w", image, err)
	}

	var img dockerImageInspect
	if err = json.Unmarshal(out, &img); err != nil {
		return "", fmt.Errorf("docker inspect %s: %w", image, err)
	}
	d := digestFromInspect(image, img)
	if d == "" {
		return "", fmt.Errorf("docker inspect %s: missing image digest", image)
	}
	return d, nil
}

func digestFromInspect(image string, img dockerImageInspect) string {
	want := imageRefName(image)
	for _, rd := range img.RepoDigests {
		name, d, ok := strings.Cut(rd, "@")
		if !ok || d == "" {
			continue
		}
		if name == want {
			return d
		}
	}
	if len(img.RepoDigests) > 0 {
		if _, d, ok := strings.Cut(img.RepoDigests[0], "@"); ok && d != "" {
			return d
		}
	}
	return img.ID
}

func imageRefName(ref string) string {
	if i := strings.LastIndex(ref, "@"); i >= 0 {
		ref = ref[:i]
	}
	slash := strings.LastIndex(ref, "/")
	colon := strings.LastIndex(ref, ":")
	if colon > slash {
		return ref[:colon]
	}
	return ref
}

func dockerPull(image string) error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker is required to convert image %q: %w", image, err)
	}
	pull := exec.Command("docker", "pull", image)
	if out, err := pull.CombinedOutput(); err != nil {
		return fmt.Errorf("docker pull %s: %w: %s", image, err, out)
	}
	return nil
}

func exportImage(image, dest string) error {
	if err := dockerPull(image); err != nil {
		return err
	}

	create := exec.Command("docker", "create", image)
	cidb, err := create.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("docker create %s: %w: %s", image, err, ee.Stderr)
		}
		return fmt.Errorf("docker create %s: %w", image, err)
	}
	cid := strings.TrimSpace(string(cidb))
	defer func() {
		_ = exec.Command("docker", "rm", cid).Run()
	}()

	for _, d := range []string{"dev", "proc", "sys", "tmp", "opt", "sbin"} {
		if err = os.MkdirAll(filepath.Join(dest, d), 0o755); err != nil {
			return err
		}
	}

	export := exec.Command("docker", "export", cid)
	untar := exec.Command("tar", "-C", dest, "--no-same-owner",
		"--exclude=./dev", "--exclude=./proc", "--exclude=./sys", "-xf", "-")
	untar.Stdin, err = export.StdoutPipe()
	if err != nil {
		return err
	}
	untar.Stderr = os.Stderr
	if err = untar.Start(); err != nil {
		return err
	}
	if err = export.Run(); err != nil {
		_ = untar.Wait()
		return fmt.Errorf("docker export %s: %w", image, err)
	}
	if err = untar.Wait(); err != nil {
		return fmt.Errorf("unpack %s: %w", image, err)
	}
	return nil
}

func installGuestAgent(root, agent string) error {
	dst := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(guestAgentPath, "/")))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(agent)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err = out.ReadFrom(in); err != nil {
		_ = out.Close()
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	return nil
}

func installGuestInit(root string) error {
	dst := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(guestInitPath, "/")))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(guestInitScript), 0o755)
}

func resolveAgent(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("agent %q: %w", explicit, err)
		}
		return explicit, nil
	}
	if p, err := exec.LookPath("firecracker-agent"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("guest agent binary is required (set agent = \"/path/to/firecracker-agent\")")
}
