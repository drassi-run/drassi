/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xfs

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"drassi.run/core/util/io"
	"drassi.run/core/util/string"
	"drassi.run/core/util/tar"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
)

func Read(ctx context.Context, fsys billy.Filesystem, src string) io.ReadCloser {
	reader, writer := io.Pipe()
	go readPipe(ctx, fsys, src, writer)

	return reader
}

func readPipe(ctx context.Context, fsys billy.Filesystem, src string, writer *io.PipeWriter) {
	tw := tar.NewWriter(writer)
	defer tw.Close()

	src = path.Clean(src)
	dir := path.Dir(src)
	if dir == "." {
		dir = ""
	} else {
		dir = xstring.EnsureSuffix(dir, "/")
	}

	err := util.Walk(fsys, src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}

		mode := info.Mode()
		var link string
		if mode&fs.ModeSymlink != 0 {
			if link, err = fsys.Readlink(path); err != nil {
				return err
			}
		}

		var hdr *tar.Header
		if hdr, err = tar.FileInfoHeader(info, link); err != nil {
			return err
		} else {
			// info only contains file's base, but we want the path
			hdr.Name = strings.TrimPrefix(path, dir)
		}
		if err = tw.WriteHeader(hdr); err != nil {
			return err
		}

		if mode.IsRegular() {
			f, err := fsys.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			// os.File implemented io.WriterTo, but fast path only used when writer is also a file
			// tar.Writer implemented io.ReaderFrom, but it's disabled for now https://github.com/golang/go/issues/22735
			// => ctx is added to the reader
			if _, err = io.Copy(tw, xio.NewContextReader(ctx, f)); err != nil {
				return err
			}
		}
		return nil
	})

	_ = writer.CloseWithError(err)
}

func Write(ctx context.Context, fsys billy.Filesystem, reader io.Reader, dest string) error {
	h := writeHandler(ctx, fsys, dest)
	return xtar.Untar(ctx, reader, h)
}

func writeHandler(ctx context.Context, fsys billy.Filesystem, dest string) xtar.UntarHandler {
	return func(hdr *tar.Header, r io.Reader) error {
		p := fsys.Join(dest, hdr.Name)
		// ensure directory existed
		if err := fsys.MkdirAll(path.Dir(p), DirPerm); err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			return fsys.MkdirAll(p, fs.FileMode(hdr.Mode))
		case tar.TypeSymlink:
			return fsys.Symlink(hdr.Linkname, p)
		case tar.TypeReg:
			// Same as os.Create(path), but with custom mode
			f, err := fsys.Create(p)
			if err != nil {
				return err
			}
			defer f.Close()

			// os.File implemented io.ReaderFrom, but fast path only used when reader is also a file
			// tar.Reader implemented io.WriterTo, but it's disabled for now https://github.com/golang/go/issues/22735
			// => ctx is added to the writer
			_, err = io.Copy(xio.NewContextWriter(ctx, f), r)
			return err
		default:
			return fmt.Errorf("unsupported %s file", xtar.FileType(hdr.Typeflag))
		}
	}
}
