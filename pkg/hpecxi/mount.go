package hpecxi

var DefaultOptions = []string{"ro", "nosuid", "nodev", "bind", "relatime"}

type MountInfo struct {
	Name          string   `json:"name"`           // mount name, e.g. "libfabric" or "libcxi"
	HostPath      string   `json:"host_path"`      // mount path, e.g. /usr/lib64/libcxi.so
	ContainerPath string   `json:"container_path"` // container mount point, e.g. /usr/lib64/libcxi.so
	Options       []string `json:"options"`        // mount options, e.g. ["ro", "nosuid", "nodev", "bind", "relatime"]
	Type          string   `json:"type"`           // mount type, e.g. "-", "d", "b", "s", "c"
}

func (g *MountInfo) Clone() *MountInfo {
	di := *g
	return &di
}

type MountsInfo map[string]*MountInfo

func (g *MountsInfo) Clone() MountsInfo {
	MountsInfoCopy := MountsInfo{}
	for mPath, mountInfo := range *g {
		MountsInfoCopy[mPath] = mountInfo.Clone()
	}
	return MountsInfoCopy
}
