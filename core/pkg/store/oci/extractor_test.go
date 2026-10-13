/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractLayer_Basic(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add directory
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "etc/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}))

	// Add file
	content := []byte("hello oras")
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "etc/test.conf",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	}))
	_, err := tw.Write(content)
	require.NoError(t, err)

	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	destDir := t.TempDir()
	err = extractLayer(context.Background(), &buf, destDir)
	require.NoError(t, err)

	readContent, err := os.ReadFile(filepath.Join(destDir, "etc", "test.conf"))
	require.NoError(t, err)
	require.Equal(t, content, readContent)
}

func TestExtractLayer_Whiteout(t *testing.T) {
	destDir := t.TempDir()

	// Setup pre-existing file and directory to be whited out
	etcDir := filepath.Join(destDir, "etc")
	require.NoError(t, os.MkdirAll(etcDir, 0755))
	oldFile := filepath.Join(etcDir, "old.conf")
	require.NoError(t, os.WriteFile(oldFile, []byte("to be deleted"), 0644))

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add whiteout entry: etc/.wh.old.conf
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "etc/.wh.old.conf",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     0,
	}))

	// Add new file: etc/new.conf
	newContent := []byte("new config")
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "etc/new.conf",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(newContent)),
	}))
	_, err := tw.Write(newContent)
	require.NoError(t, err)

	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	err = extractLayer(context.Background(), &buf, destDir)
	require.NoError(t, err)

	// Old file should be gone
	_, err = os.Stat(oldFile)
	require.True(t, os.IsNotExist(err))

	// New file should exist
	content, err := os.ReadFile(filepath.Join(etcDir, "new.conf"))
	require.NoError(t, err)
	require.Equal(t, newContent, content)
}

func TestExtractLayer_Symlink(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add target file
	content := []byte("symlink target")
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "file.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	}))
	_, err := tw.Write(content)
	require.NoError(t, err)

	// Add symlink
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "link.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "file.txt",
	}))

	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	destDir := t.TempDir()
	err = extractLayer(context.Background(), &buf, destDir)
	require.NoError(t, err)

	linkTarget, err := os.Readlink(filepath.Join(destDir, "link.txt"))
	require.NoError(t, err)
	require.Equal(t, "file.txt", linkTarget)
}

func TestExtractLayer_PathTraversal(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add path with .. traversal
	badContent := []byte("escape")
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "../escape.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(badContent)),
	}))
	_, err := tw.Write(badContent)
	require.NoError(t, err)

	// Add path with absolute path
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "/etc/passwd",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(badContent)),
	}))
	_, err = tw.Write(badContent)
	require.NoError(t, err)

	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	destDir := t.TempDir()
	err = extractLayer(context.Background(), &buf, destDir)
	require.NoError(t, err)

	// Ensure files were skipped and not written outside destDir
	parentDir := filepath.Dir(destDir)
	escapedFile := filepath.Join(parentDir, "escape.txt")
	_, err = os.Stat(escapedFile)
	require.True(t, os.IsNotExist(err))
}
