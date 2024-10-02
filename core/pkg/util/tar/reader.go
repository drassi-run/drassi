package xtar

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/docker/docker/pkg/archive"
)

const defaultFilePerm int64 = 0o666

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

func FileEntryReader(entries ...*FileEntry) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	for _, entry := range entries {
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
