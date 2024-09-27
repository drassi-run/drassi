package utilreader

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"

	"drassi.run/core/pkg/util"
	"github.com/docker/docker/pkg/archive"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// FileEntry is a file to copy to a container
type FileEntry struct {
	Name    string // Name of file entry
	Mode    int64  // Permission and mode bits
	Content string
	Uid     int    // User ID of owner
	Gid     int    // Group ID of owner
	Uname   string // User name of owner
	Gname   string // Group name of owner
}

func FromFileEntries(ctx context.Context, entries ...*FileEntry) (io.Reader, error) {
	logger := util.Logger(ctx)
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	for _, entry := range entries {
		logger.Debugf("Writing entry to tarball %s len:%d", entry.Name, len(entry.Content))
		hdr := &tar.Header{
			Name:  entry.Name,
			Mode:  entry.Mode,
			Size:  int64(len(entry.Content)),
			Uid:   entry.Uid,
			Gid:   entry.Gid,
			Uname: entry.Uname,
			Gname: entry.Gname,
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

func FromJsonObject(m map[string]any, compact bool) (io.Reader, error) {
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

		hdr := &tar.Header{Name: file, Mode: 0o755, Size: int64(len(b))}
		if err = tw.WriteHeader(hdr); err != nil {
			return nil, err
		}

		if _, err = tw.Write(b); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func FromFilesystem(fs billy.Filesystem, path string, matcher gitignore.Matcher) (io.Reader, error) {
	//	TODO
	return nil, nil
}

type UntarHandler = func(*tar.Header, io.Reader) error

func Untar(ctx context.Context, r io.Reader, h UntarHandler) error {
	xr, err := archive.DecompressStream(r)
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
