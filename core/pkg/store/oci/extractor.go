/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/go-archive/compression"
)

func extractLayer(ctx context.Context, r io.Reader, targetDir string) error {
	xr, err := compression.DecompressStream(r)
	if err != nil {
		return fmt.Errorf("decompress layer stream: %w", err)
	}
	defer xr.Close()

	tr := tar.NewReader(xr)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		cleanPath := filepath.Clean(hdr.Name)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			continue // skip dangerous path
		}

		destPath := filepath.Join(targetDir, cleanPath)
		baseName := filepath.Base(cleanPath)
		dirName := filepath.Dir(destPath)

		// Handle whiteouts
		if strings.HasPrefix(baseName, ".wh.") {
			targetName := strings.TrimPrefix(baseName, ".wh.")
			_ = os.RemoveAll(filepath.Join(dirName, targetName))
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(dirName, 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(dirName, 0755); err != nil {
				return err
			}
			_ = os.Remove(destPath)
			if err := os.Symlink(hdr.Linkname, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}
