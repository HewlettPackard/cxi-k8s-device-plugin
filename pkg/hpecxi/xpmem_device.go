package hpecxi

import "path/filepath"

const XpmemDevfsPath = "xpmem"

func GetXpmemDevPath() string {
	return filepath.Join(GetDevRoot(XpmemDevfsPath), XpmemDevfsPath)
}
