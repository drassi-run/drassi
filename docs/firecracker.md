# Firecracker sandboxer

The Firecracker sandboxer runs each job in a microVM. The host starts
Firecracker, clones an ext4 rootfs, and talks to a guest agent over
[hybrid vsock](https://github.com/firecracker-microvm/firecracker/blob/main/docs/vsock.md).
Docker inside the VM is used through `docker system dial-stdio`.

## Prerequisites

- Linux host with `/dev/kvm` readable and writable by the runner user
- [`firecracker`](https://github.com/firecracker-microvm/firecracker) at an
  absolute path (set `bin`)
- Guest agent binary (`core/cmd/firecracker-agent`)
- Docker on the **host** to pull `oci://` kernels and convert `oci://` rootfs images
- Docker **in the guest** if jobs need `docker` / container actions
- Optional TAP device for guest outbound network (needed to pull images)

Build the guest agent:

```bash
cd core
CGO_ENABLED=0 go build -o firecracker-agent ./cmd/firecracker-agent
```

Put that binary on `PATH` as `firecracker-agent`, or set `agent` in the
sandboxer config.

## Configuration

Add this to the runner TOML. `use_sandboxer` must match the table name.

```toml
use_sandboxer = 'firecracker'

[sandboxers.firecracker]
provider = 'firecracker'
bin = '/usr/bin/firecracker'
kernel_path = 'oci://ghcr.io/edera-dev/zone-kernel:6.18.43'
rootfs_path = 'oci://ghcr.io/drassi-run/ubuntu:26.04'
agent = '/home/you/.cache/drassi/firecracker/firecracker-agent'
root_dir = '/tmp/drassi-fc-runner'
vcpu_count = 2
mem_size_mib = 2048
agent_wait_sec = 60
kernel_args = 'console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/usr/bin/tini -- /usr/sbin/drassi-init'
tap_device = 'tap-drassi'
guest_mac = 'AA:FC:00:00:00:01'
```

`provider` must be `firecracker`. Remote kernels (`http://`, `https://`,
`oci://`) and `oci://` rootfs images are cached under `{root_dir}/cache/<digest>/`.

## Settings

| Key | Default | Notes |
| --- | --- | --- |
| `bin` | `firecracker` | Absolute path to the Firecracker executable |
| `root_dir` | `/var/lib/drassi/firecracker` | Per-job VM dirs and image cache |
| `kernel_path` | *(required)* | `file://`, `http(s)://`, or `oci://` kernel |
| `rootfs_path` | `oci://ghcr.io/drassi-run/ubuntu:26.04` | `file://` ext4 or `oci://` image converted to ext4 |
| `agent` | `firecracker-agent` on `PATH` | Host path copied into the image as `/usr/sbin/drassi-agent`; `/usr/sbin/drassi-init` is installed alongside it |
| `rootfs_size_mib` | `2048` | Size of a converted ext4 image |
| `initrd` | | Optional initrd path |
| `kernel_args` | see below | Guest cmdline |
| `vcpu_count` | `2` | vCPUs |
| `mem_size_mib` | `2048` | Guest RAM |
| `agent_port` | `1024` | Guest vsock port the agent listens on |
| `tap_device` | | Host TAP name. Omit for no NIC |
| `guest_mac` | | MAC on the TAP NIC |
| `guest_ip` | next address on the TAP subnet | Guest CIDR on `eth0`, e.g. `172.16.0.2/24` |
| `guest_gateway` | TAP IPv4 | Default route in the guest |
| `guest_dns` | `1.1.1.1`, `8.8.8.8` | Written to `/etc/resolv.conf` |
| `guest_iface` | `eth0` | Guest interface to configure |
| `agent_wait_sec` | `30` | Seconds to wait for the guest agent after boot |

Default `kernel_args`:

```text
console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/usr/bin/tini -- /usr/sbin/drassi-init
```

That matches the default image: tini is PID 1, `drassi-init` starts dockerd
(if present) then execs the agent. Override it if your rootfs boots
differently (for example `init=/sbin/init` when the agent *is* PID 1).

## Kernel

`kernel_path` is a URI:

| Scheme | Example | Cache |
| --- | --- | --- |
| `file://` | `file:///boot/vmlinux` | none if the file is already ELF; otherwise extracted ELF is cached by content sha256 |
| `http://` / `https://` | `https://example.com/vmlinux` | content sha256; URL → digest index skips re-download |
| `oci://` | `oci://ghcr.io/edera-dev/zone-kernel:6.18.43` | image digest; blob taken from `/kernel/image` or `/kernel/vmlinuz` |

A path with no scheme is treated as `file://`. Cached files live at
`{root_dir}/cache/<digest>/kernel`.

On x86_64 Firecracker **main** accepts uncompressed ELF (`vmlinux`) and
`bzImage`. Firecracker **1.16** is ELF-only, so Drassi extracts a `vmlinux`
payload (gzip or zstd) from a `bzImage` before caching. On aarch64
Firecracker wants a PE `Image`.
See [Firecracker rootfs and kernel setup](https://github.com/firecracker-microvm/firecracker/blob/main/docs/rootfs-and-kernel-setup.md).

[Edera zone-kernel](https://github.com/edera-dev/linux-kernel-oci) images
(`ghcr.io/edera-dev/zone-kernel:6.18.43`) ship `/kernel/image` as a
`bzImage` plus `config.gz`, `addons.squashfs`, and `metadata`. Only the
kernel blob is used; modules in `addons.squashfs` are not attached.

The kernel needs virtio-blk, virtio-net, vsock, ext4, overlayfs, and
virtio-mmio (this sandboxer boots with `pci=off`).

## Rootfs

`rootfs_path` is the guest root. Schemes:

| Scheme | Behavior | Cache |
| --- | --- | --- |
| `file://` (or a bare path) | local ext4 used in place | none |
| `oci://` | docker pull, export, install agent and init, `mkfs.ext4` | image digest |

For `oci://`, Drassi resolves a content digest (from `@sha256:…` if pinned,
otherwise `docker pull` + inspect). The ext4 is cached at
`{root_dir}/cache/<digest>/rootfs.ext4` (`:` in the digest becomes `-`).
On a cache miss it exports the image, installs the agent at
`/usr/sbin/drassi-agent`, writes `/usr/sbin/drassi-init`, and runs
`mkfs.ext4`. Host Docker is required to resolve a tag and on the first
conversion. A local ext4 is used as-is; bake the agent and init script
in yourself.

Each job clones that file (`cp --reflink=auto --sparse=always`).

The default image is a Ubuntu 26.04 DinD-style rootfs: tini, dockerd, no
systemd. Overlayfs in the guest is required for Docker.

## Network

Firecracker attaches at most one TAP (`tap_device`) as guest `eth0`. The TAP
must already exist on the host. After the guest agent is up, Drassi brings
the interface up, assigns `guest_ip` (or the next IPv4 on the TAP subnet),
installs a default route via `guest_gateway` (or the TAP address), and writes
`guest_dns` to `/etc/resolv.conf`. Workflows do not need a network setup step.

Example host TAP and NAT (one VM at a time on this TAP):

```bash
sudo ip tuntap add tap-drassi mode tap user "$USER"
sudo ip addr add 172.16.0.1/24 dev tap-drassi
sudo ip link set tap-drassi up
sudo sysctl -w net.ipv4.ip_forward=1
sudo iptables -t nat -A POSTROUTING -s 172.16.0.0/24 -j MASQUERADE
```

A TAP often shows `DOWN` until a VM attaches; that is expected. One shared
TAP cannot serve overlapping VMs — keep `MaxParallelism` at 1, or give each
concurrent job its own TAP.

Without `tap_device`, the agent still works over vsock, but the guest has no
IP network (no image pulls).

## Operations

Each job directory under `root_dir` contains `vm.json`, `rootfs.ext4`,
`vsock.sock`, and `firecracker.log`. These VMs are started with Firecracker
directly (`pgrep -a firecracker`); they do not show up in Ignite/Incus.

If a job is cancelled and Firecracker does not exit, the listener can block
and GitHub will mark the runner offline. Kill leftover `firecracker`
processes, then restart `gha-runner`.
