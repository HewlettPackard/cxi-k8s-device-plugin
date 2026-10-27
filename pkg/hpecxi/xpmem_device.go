package hpecxi

import "path/filepath"

// XpmemDevfsPath is the device filesystem path for XPMEM devices.
// XPMEM is a kernel module that enables efficient memory sharing between processes.
const XpmemDevfsPath = "xpmem"

// GetXpmemDevPath returns the full device path for XPMEM devices in /dev.
// It constructs the path by joining the device root directory with the XPMEM device path.
// This is used for accessing XPMEM character devices that facilitate cross-memory attach operations.
func GetXpmemDevPath() string {
	return filepath.Join(GetDevRoot(XpmemDevfsPath), XpmemDevfsPath)
}
