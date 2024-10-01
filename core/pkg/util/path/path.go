package xpath

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func ResolveDir(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		var u *user.User
		var err error
		idx := strings.IndexByte(path, os.PathSeparator)
		if idx < 0 {
			idx = len(path)
		}
		if username := path[1:idx]; username == "" {
			u, err = user.Current()
		} else {
			u, err = user.Lookup(username)
		}
		if err != nil {
			return "", err
		}
		path = filepath.Join(u.HomeDir, path[idx:])
	}

	return filepath.Abs(path)
}
