package xfs

import "io/fs"

const (
	FilePerm fs.FileMode = 0o644
	DirPerm  fs.FileMode = 0o755
	AllPerm  fs.FileMode = 0o777
)
