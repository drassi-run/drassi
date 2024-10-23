package xsftp

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"drassi.run/core/util/io"
	"drassi.run/core/util/tar"
	"github.com/pkg/sftp"
)

func Read(ctx context.Context, client *sftp.Client, src string) io.ReadCloser {
	reader, writer := io.Pipe()
	go readPipe(ctx, client, src, writer)

	return reader
}

func readPipe(ctx context.Context, client *sftp.Client, src string, writer *io.PipeWriter) {
	defer writer.Close()
	tw := tar.NewWriter(writer)
	defer tw.Close()
	w := xio.NewContextWriter(ctx, tw)
	base := filepath.Dir(src)

	for walker := client.Walk(src); walker.Step(); {
		if err := walker.Err(); err != nil {
			_ = writer.CloseWithError(err)
			return
		}

		fi := walker.Stat()
		hdr := &tar.Header{
			Mode:    int64(fi.Mode().Perm()),
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
		}
		path := walker.Path()
		if name, err := filepath.Rel(base, path); err != nil {
			_ = writer.CloseWithError(err)
			return
		} else {
			hdr.Name = name
		}

		mode := fi.Mode()
		switch {
		case mode.IsDir():
			hdr.Typeflag = tar.TypeDir
		case mode.IsRegular():
			hdr.Typeflag = tar.TypeReg
		case mode&fs.ModeSymlink != 0:
			hdr.Typeflag = tar.TypeSymlink
			if target, err := client.ReadLink(path); err != nil {
				_ = writer.CloseWithError(err)
				return
			} else {
				hdr.Linkname = target
			}
		default:
			err := fmt.Errorf("unsupported file mode: %s", mode)
			_ = writer.CloseWithError(err)
			return
		}

		if err := tw.WriteHeader(hdr); err != nil {
			_ = writer.CloseWithError(err)
			return
		}
		if !mode.IsRegular() {
			continue
		}

		if r, err := client.Open(path); err != nil {
			_ = writer.CloseWithError(err)
			return
		} else if _, err = io.Copy(w, r); err != nil {
			_ = writer.CloseWithError(err)
			return
		}
	}
}

func Write(ctx context.Context, client *sftp.Client, reader io.Reader, dest string) error {
	h := writeHandler(ctx, client, dest)
	return xtar.Untar(ctx, reader, h)
}

func writeHandler(ctx context.Context, client *sftp.Client, dest string) xtar.UntarHandler {
	return func(hdr *tar.Header, r io.Reader) error {
		path := filepath.Join(dest, hdr.Name)
		// ensure directory existed
		if err := client.MkdirAll(filepath.Dir(path)); err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			return client.Mkdir(path)
		case tar.TypeSymlink:
			return client.Symlink(hdr.Linkname, path)
		case tar.TypeReg:
			// Same as os.Create(path), but with custom mode
			f, err := client.Create(path)
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
			return fmt.Errorf("unsupported file type %v", hdr.Typeflag)
		}
	}
}
