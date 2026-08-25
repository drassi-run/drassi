/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xtar

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"github.com/moby/go-archive/compression"
)

const defaultFilePerm int64 = 0o666

// FileEntry is a file to copy to a container
type FileEntry struct {
	Name    string      // Name of file entry
	Mode    fs.FileMode // Permission and mode bits
	Content string
	Uid     int    // User ID of owner
	Gid     int    // Group ID of owner
	Uname   string // User name of owner
	Gname   string // Group name of owner
}

func FileEntryReader(entries ...*FileEntry) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	for _, entry := range entries {
		hdr := &tar.Header{
			Name:  entry.Name,
			Mode:  int64(entry.Mode),
			Size:  int64(len(entry.Content)),
			Uid:   entry.Uid,
			Gid:   entry.Gid,
			Uname: entry.Uname,
			Gname: entry.Gname,
		}
		if typ, err := fileType(entry.Mode); err != nil {
			return nil, err
		} else {
			hdr.Typeflag = typ
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := io.WriteString(tw, entry.Content); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

// see [archive/tar.FileInfoHeader]
func fileType(mod fs.FileMode) (byte, error) {
	var typ byte
	switch m := mod.Type(); m {
	case 0:
		typ = tar.TypeReg
	case fs.ModeDir:
		typ = tar.TypeDir
	case fs.ModeSymlink:
		typ = tar.TypeSymlink
	case fs.ModeNamedPipe:
		typ = tar.TypeFifo
	case fs.ModeDevice:
		typ = tar.TypeBlock
	case fs.ModeCharDevice:
		typ = tar.TypeChar
	default: // fs.ModeIrregular, fs.ModeSocket
		return 0, fmt.Errorf("unknown filetype %q", m)
	}
	return typ, nil
}

func JsonObjectReader(m map[string]any, compact bool) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	for file, obj := range m {
		var b []byte
		var err error
		if compact {
			b, err = json.Marshal(obj)
		} else {
			b, err = json.MarshalIndent(obj, "", "  ")
		}
		if err != nil {
			return nil, err
		}

		hdr := &tar.Header{Name: file, Mode: defaultFilePerm, Size: int64(len(b))}
		if err = tw.WriteHeader(hdr); err != nil {
			return nil, err
		}

		if _, err = tw.Write(b); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func ContentReader(m map[string]string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	for file, content := range m {
		hdr := &tar.Header{Name: file, Mode: defaultFilePerm, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}

		if _, err := io.WriteString(tw, content); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

type UntarHandler = func(*tar.Header, io.Reader) error

func Untar(ctx context.Context, r io.Reader, h UntarHandler) error {
	xr, err := compression.DecompressStream(r)
	if err != nil {
		return err
	}
	defer xr.Close()

	tr := tar.NewReader(xr)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if hdr, err := tr.Next(); err != nil {
			if err == io.EOF {
				break // end of tar archive
			}
			return err
		} else if err = h(hdr, tr); err != nil {
			return err
		}
	}

	return nil
}

func FileType(typ byte) string {
	switch typ {
	//nolint:staticcheck // SA1019 - TypeRegA has been deprecated
	case tar.TypeReg, tar.TypeRegA:
		return "regular"
	case tar.TypeLink:
		return "hardlink"
	case tar.TypeSymlink:
		return "symlink"
	case tar.TypeChar:
		return "character device"
	case tar.TypeBlock:
		return "block device"
	case tar.TypeDir:
		return "directory"
	case tar.TypeFifo:
		return "fifo"
	default:
		return "unknown"
	}
}

func IsRegular(hdr *tar.Header) bool {
	//nolint:staticcheck // SA1019 - TypeRegA has been deprecated
	return hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA
}

func IsDir(hdr *tar.Header) bool {
	return hdr.Typeflag == tar.TypeDir
}
