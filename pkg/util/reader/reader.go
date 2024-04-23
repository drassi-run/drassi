package utilreader

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/pkg/archive"
	"github.com/dungdm93/drasi/pkg/util"
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
		if _, err := tw.Write([]byte(entry.Content)); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

func FromFilesystem(fs billy.Filesystem, path string, matcher gitignore.Matcher) (io.Reader, error) {
	//	TODO
	return nil, nil
}

func ReadLine(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l := scanner.Text()
		if l != "" && !strings.HasPrefix(l, "#") {
			lines = append(lines, l)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L342-L403
func ParseEnvVars(reader io.Reader) (map[string]string, error) {
	env := make(map[string]string)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		equalsIndex := strings.Index(line, "=")
		heredocIndex := strings.Index(line, "<<")

		// Normal style NAME=VALUE
		if 0 <= equalsIndex && (heredocIndex < 0 || equalsIndex < heredocIndex) {
			key, value := line[:equalsIndex], line[equalsIndex+1:]
			if key == "" {
				return nil, fmt.Errorf("invalid nil key in line: %s", line)
			}
			env[key] = value
			continue
		}

		// Heredoc style NAME<<EOF
		if 0 <= heredocIndex && (equalsIndex < 0 || heredocIndex < equalsIndex) {
			key, delimiter := line[:heredocIndex], line[heredocIndex+2:]
			if key == "" || delimiter == "" {
				return nil, fmt.Errorf("invalid format '%s'. key and delimiter MUST NOT be empty", line)
			}
			value, finish := make([]string, 0), false
			for scanner.Scan() {
				doc := scanner.Text()
				if doc == delimiter {
					finish = true
					break
				}
				value = append(value, doc)
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			if !finish {
				return nil, fmt.Errorf("invalid value. EOF marker missing new line")
			}

			env[key] = strings.Join(value, "\n")
			continue
		}

		return nil, fmt.Errorf("invalid format: %s", line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return env, nil
}

type UntarHandler = func(*tar.Header, io.Reader) error

func Untar(r io.Reader, h UntarHandler) error {
	xr, err := archive.DecompressStream(r)
	if err != nil {
		return err
	}
	defer xr.Close()

	tr := tar.NewReader(xr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break // end of tar archive
			}
			return err
		}

		if err := h(hdr, tr); err != nil {
			return err
		}
	}

	return nil
}
