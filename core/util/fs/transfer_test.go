/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xfs

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/stretchr/testify/require"
)

func assertTar(t *testing.T, r io.Reader, expectedEntries, expectedSymlinks map[string]string) {
	t.Helper()
	entries := make(map[string]string)
	symlinks := make(map[string]string)
	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		if hdr.Typeflag == tar.TypeSymlink {
			symlinks[hdr.Name] = hdr.Linkname
		} else if hdr.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tr)
			require.NoError(t, err)
			entries[hdr.Name] = string(content)
		}
	}

	if expectedEntries != nil {
		require.Equal(t, expectedEntries, entries)
	} else {
		require.Empty(t, entries)
	}

	if expectedSymlinks != nil {
		require.Equal(t, expectedSymlinks, symlinks)
	} else {
		require.Empty(t, symlinks)
	}
}

func TestRead(t *testing.T) {
	t.Run("root directory", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0o644))
		require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "sub"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "sub", "file2.txt"), []byte("content2"), 0o644))
		require.NoError(t, os.Symlink("file1.txt", filepath.Join(tempDir, "link1")))

		fsys := osfs.New(tempDir)
		rc := Read(t.Context(), fsys, ".")
		defer rc.Close()

		entries := map[string]string{
			"file1.txt":     "content1",
			"sub/file2.txt": "content2",
		}
		symlinks := map[string]string{
			"link1": "file1.txt",
		}
		assertTar(t, rc, entries, symlinks)
	})

	t.Run("subpath directory preserves folder name", func(t *testing.T) {
		tempDir := t.TempDir()
		appConfigDir := filepath.Join(tempDir, "app", "config")
		require.NoError(t, os.MkdirAll(appConfigDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(appConfigDir, "settings.json"), []byte(`{"port":8080}`), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(appConfigDir, "app.conf"), []byte("key=val"), 0o644))

		fsys := osfs.New(tempDir)
		rc := Read(t.Context(), fsys, "app/config")
		defer rc.Close()

		entries := map[string]string{
			"config/settings.json": `{"port":8080}`,
			"config/app.conf":      "key=val",
		}
		assertTar(t, rc, entries, nil)
	})

	t.Run("single file subpath", func(t *testing.T) {
		tempDir := t.TempDir()
		appDir := filepath.Join(tempDir, "app")
		require.NoError(t, os.MkdirAll(appDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(appDir, "config.yaml"), []byte("env: test"), 0o644))

		fsys := osfs.New(tempDir)
		rc := Read(t.Context(), fsys, "app/config.yaml")
		defer rc.Close()

		entries := map[string]string{
			"config.yaml": "env: test",
		}
		assertTar(t, rc, entries, nil)
	})

	t.Run("context cancellation stops read", func(t *testing.T) {
		tempDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("data"), 0o644))

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		fsys := osfs.New(tempDir)
		rc := Read(ctx, fsys, ".")
		defer rc.Close()

		_, err := io.ReadAll(rc)
		require.Error(t, err)
		require.True(t, errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe))
	})
}

func TestWrite(t *testing.T) {
	t.Run("extracts directory structure regular files and symlinks", func(t *testing.T) {
		destDir := t.TempDir()
		destFsys := osfs.New(destDir)

		buf := new(bytes.Buffer)
		tw := tar.NewWriter(buf)

		// Directory entry
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name:     "extracted/sub/",
			Typeflag: tar.TypeDir,
			Mode:     0o755,
		}))

		// Regular file entry
		fileContent := []byte("hello extracted")
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name:     "extracted/sub/file.txt",
			Typeflag: tar.TypeReg,
			Mode:     0o644,
			Size:     int64(len(fileContent)),
		}))
		_, err := tw.Write(fileContent)
		require.NoError(t, err)

		// Symlink entry
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name:     "extracted/sub/link.txt",
			Typeflag: tar.TypeSymlink,
			Linkname: "file.txt",
		}))
		require.NoError(t, tw.Close())

		err = Write(t.Context(), destFsys, buf, "")
		require.NoError(t, err)

		// Verify extracted files on disk
		content, err := os.ReadFile(filepath.Join(destDir, "extracted", "sub", "file.txt"))
		require.NoError(t, err)
		require.Equal(t, "hello extracted", string(content))

		target, err := os.Readlink(filepath.Join(destDir, "extracted", "sub", "link.txt"))
		require.NoError(t, err)
		require.Equal(t, "file.txt", target)
	})

	t.Run("unsupported entry type returns error", func(t *testing.T) {
		destDir := t.TempDir()
		destFsys := osfs.New(destDir)

		buf := new(bytes.Buffer)
		tw := tar.NewWriter(buf)

		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name:     "fifo_pipe",
			Typeflag: tar.TypeFifo,
		}))
		require.NoError(t, tw.Close())

		err := Write(t.Context(), destFsys, buf, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported fifo file")
	})
}

func TestReadWriteRoundTrip(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "nested", "dir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "nested", "file.txt"), []byte("payload data"), 0o644))
	require.NoError(t, os.Symlink("file.txt", filepath.Join(srcDir, "nested", "symlink.txt")))

	srcFsys := osfs.New(srcDir)
	destFsys := osfs.New(destDir)

	rc := Read(t.Context(), srcFsys, "nested")
	defer rc.Close()

	err := Write(t.Context(), destFsys, rc, "target")
	require.NoError(t, err)

	extractedFile := filepath.Join(destDir, "target", "nested", "file.txt")
	content, err := os.ReadFile(extractedFile)
	require.NoError(t, err)
	require.Equal(t, "payload data", string(content))

	linkTarget, err := os.Readlink(filepath.Join(destDir, "target", "nested", "symlink.txt"))
	require.NoError(t, err)
	require.Equal(t, "file.txt", linkTarget)
}
