/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"drassi.run/core/util/fs"
	"github.com/klauspost/compress/zstd"
)

var kernelBlobPaths = []string{"/kernel/image", "/kernel/vmlinuz"}

func (c *Config) ensureKernel() error {
	if c.kernel != "" {
		if _, err := os.Stat(c.kernel); err != nil {
			return fmt.Errorf("kernel %q: %w", c.kernel, err)
		}
		return nil
	}
	if c.KernelPath == "" {
		return fmt.Errorf("firecracker kernel_path is required")
	}
	path, err := c.materializeKernel()
	if err != nil {
		return err
	}
	c.kernel = path
	return nil
}

func (c *Config) materializeKernel() (string, error) {
	scheme, loc, err := parseKernelRef(c.KernelPath)
	if err != nil {
		return "", err
	}
	switch scheme {
	case "file":
		return c.materializeFileKernel(loc)
	case "http", "https":
		return c.materializeHTTPKernel(loc)
	case "oci":
		return c.materializeOCIKernel(loc)
	default:
		return "", fmt.Errorf("unsupported kernel_path scheme %q", scheme)
	}
}

func parseURIRef(field, ref string) (scheme, loc string, err error) {
	if ref == "" {
		return "", "", fmt.Errorf("%s is empty", field)
	}
	u, err := url.Parse(ref)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", field, err)
	}
	switch strings.ToLower(u.Scheme) {
	case "file":
		p := u.Path
		if p == "" {
			p = u.Opaque
		}
		if u.Host != "" && u.Host != "localhost" {
			return "", "", fmt.Errorf("file:// host %q is not supported", u.Host)
		}
		if p == "" {
			return "", "", fmt.Errorf("file:// %s is empty", field)
		}
		return "file", p, nil
	case "http", "https":
		return strings.ToLower(u.Scheme), ref, nil
	case "oci":
		loc = strings.TrimPrefix(ref, u.Scheme+":")
		loc = strings.TrimPrefix(loc, "//")
		if loc == "" {
			return "", "", fmt.Errorf("oci:// %s is empty", field)
		}
		return "oci", loc, nil
	case "":
		return "file", ref, nil
	default:
		return "", "", fmt.Errorf("unsupported %s scheme %q", field, u.Scheme)
	}
}

func parseKernelRef(ref string) (scheme, loc string, err error) {
	scheme, loc, err = parseURIRef("kernel_path", ref)
	if err != nil {
		return "", "", err
	}
	switch scheme {
	case "file", "http", "https", "oci":
		return scheme, loc, nil
	default:
		return "", "", fmt.Errorf("unsupported kernel_path scheme %q (want file, http, https, or oci)", scheme)
	}
}

func parseRootfsRef(ref string) (scheme, loc string, err error) {
	scheme, loc, err = parseURIRef("rootfs_path", ref)
	if err != nil {
		return "", "", err
	}
	switch scheme {
	case "file", "oci":
		return scheme, loc, nil
	default:
		return "", "", fmt.Errorf("unsupported rootfs_path scheme %q (want file or oci)", scheme)
	}
}

func (c *Config) materializeFileKernel(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	ok, err := isELF(abs)
	if err != nil {
		return "", fmt.Errorf("kernel %q: %w", abs, err)
	}
	if ok {
		return abs, nil
	}
	f, err := os.Open(abs)
	if err != nil {
		return "", fmt.Errorf("kernel %q: %w", abs, err)
	}
	defer f.Close()
	return c.storeKernel(f, "")
}

func (c *Config) materializeHTTPKernel(rawURL string) (string, error) {
	if out, ok := c.cachedKernelForURL(rawURL); ok {
		if err := ensureELFKernel(out); err != nil {
			return "", err
		}
		return out, nil
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download kernel %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download kernel %s: %s", rawURL, resp.Status)
	}
	return c.storeKernel(resp.Body, rawURL)
}

func (c *Config) materializeOCIKernel(image string) (string, error) {
	digest, err := resolveImageDigest(image)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(c.RootDir, "cache", digestCacheDir(digest))
	out := filepath.Join(dir, "kernel")
	if st, err := os.Stat(out); err == nil && st.Size() > 0 {
		if err = ensureELFKernel(out); err != nil {
			return "", err
		}
		return out, nil
	}
	if err = os.MkdirAll(dir, xfs.DirPerm); err != nil {
		return "", err
	}

	tmp := out + ".tmp"
	_ = os.Remove(tmp)
	if err = exportKernel(image, tmp); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err = os.Rename(tmp, out); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err = ensureELFKernel(out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Config) storeKernel(r io.Reader, sourceURL string) (string, error) {
	cacheRoot := filepath.Join(c.RootDir, "cache")
	if err := os.MkdirAll(cacheRoot, xfs.DirPerm); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(cacheRoot, "kernel-*.tmp")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(tmp, h), r)
	cerr := tmp.Close()
	if err != nil {
		_ = os.Remove(tmpName)
		return "", err
	}
	if cerr != nil {
		_ = os.Remove(tmpName)
		return "", cerr
	}

	digest := "sha256:" + hex.EncodeToString(h.Sum(nil))
	dir := filepath.Join(cacheRoot, digestCacheDir(digest))
	out := filepath.Join(dir, "kernel")
	if st, err := os.Stat(out); err == nil && st.Size() > 0 {
		_ = os.Remove(tmpName)
	} else {
		if err = os.MkdirAll(dir, xfs.DirPerm); err != nil {
			_ = os.Remove(tmpName)
			return "", err
		}
		if err = os.Rename(tmpName, out); err != nil {
			_ = os.Remove(tmpName)
			return "", err
		}
	}
	if sourceURL != "" {
		if err = c.rememberURL(sourceURL, digest); err != nil {
			return "", err
		}
	}
	if err = ensureELFKernel(out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Config) cachedKernelForURL(rawURL string) (string, bool) {
	b, err := os.ReadFile(c.urlIndexPath(rawURL))
	if err != nil {
		return "", false
	}
	digest := strings.TrimSpace(string(b))
	if digest == "" {
		return "", false
	}
	out := filepath.Join(c.RootDir, "cache", digestCacheDir(digest), "kernel")
	st, err := os.Stat(out)
	if err != nil || st.Size() == 0 {
		return "", false
	}
	return out, true
}

func (c *Config) rememberURL(rawURL, digest string) error {
	path := c.urlIndexPath(rawURL)
	if err := os.MkdirAll(filepath.Dir(path), xfs.DirPerm); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(digest+"\n"), 0o644)
}

func (c *Config) urlIndexPath(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return filepath.Join(c.RootDir, "cache", "urls", hex.EncodeToString(sum[:]))
}

func exportKernel(image, dest string) error {
	if err := dockerPull(image); err != nil {
		return err
	}

	var cid string
	var err error
	for _, blob := range kernelBlobPaths {
		create := exec.Command("docker", "create", "--entrypoint", blob, image)
		var cidb []byte
		cidb, err = create.Output()
		if err == nil {
			cid = strings.TrimSpace(string(cidb))
			break
		}
	}
	if cid == "" {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("docker create %s: %w: %s", image, err, ee.Stderr)
		}
		return fmt.Errorf("docker create %s: %w", image, err)
	}
	defer func() {
		_ = exec.Command("docker", "rm", cid).Run()
	}()

	var last error
	for _, blob := range kernelBlobPaths {
		cp := exec.Command("docker", "cp", cid+":"+blob, dest)
		if out, err := cp.CombinedOutput(); err == nil {
			st, err := os.Stat(dest)
			if err == nil && st.Size() > 0 {
				return nil
			}
			last = err
			continue
		} else {
			last = fmt.Errorf("docker cp %s:%s: %w: %s", image, blob, err, out)
		}
	}
	if last == nil {
		last = fmt.Errorf("kernel blob not found in %s (tried %s)", image, strings.Join(kernelBlobPaths, ", "))
	}
	return last
}

func ensureELFKernel(path string) error {
	ok, err := isELF(path)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	elf, err := extractELFKernel(raw)
	if err != nil {
		return fmt.Errorf("kernel %q: %w", path, err)
	}
	tmp := path + ".elf"
	if err = os.WriteFile(tmp, elf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func extractELFKernel(img []byte) ([]byte, error) {
	if isELFBytes(img) {
		return img, nil
	}
	if out := tryDecompress(img, []byte{0x28, 0xb5, 0x2f, 0xfd}, decompressZstd); out != nil {
		return out, nil
	}
	if out := tryDecompress(img, []byte{0x1f, 0x8b}, decompressGzip); out != nil {
		return out, nil
	}
	return nil, fmt.Errorf("not ELF and no gzip/zstd vmlinux payload (Firecracker 1.16 cannot boot bzImage)")
}

func tryDecompress(img, magic []byte, dec func([]byte) ([]byte, error)) []byte {
	for off := 0; ; off++ {
		i := bytes.Index(img[off:], magic)
		if i < 0 {
			return nil
		}
		off += i
		out, _ := dec(img[off:])
		if isELFBytes(out) {
			return out
		}
	}
}

func decompressZstd(b []byte) ([]byte, error) {
	d, err := zstd.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer d.Close()
	return io.ReadAll(d)
}

func decompressGzip(b []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func isELF(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	var mag [4]byte
	if _, err = io.ReadFull(f, mag[:]); err != nil {
		return false, err
	}
	return isELFBytes(mag[:]), nil
}

func isELFBytes(b []byte) bool {
	return len(b) >= 4 && b[0] == 0x7f && b[1] == 'E' && b[2] == 'L' && b[3] == 'F'
}
