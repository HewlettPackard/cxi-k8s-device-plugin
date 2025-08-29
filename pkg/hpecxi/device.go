package hpecxi

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	HPEvendorID string = "0x17db"

	DevfsPath = "cxi"
	Sysfspath = "class/cxi"

	Vendor     = "hpe.com"
	Class      = "cxi"
	Kind       = Vendor + "/" + Class
	DriverName = Class + "." + Vendor

	UIDLength = len("0000-00-00-0-0x0000")
)

var (
	PciRegexp     = regexp.MustCompile(`[0-9a-f]{4}:[0-9a-f]{2}:[0-9a-f]{2}\.[0-7]$`)
	CardRegexp    = regexp.MustCompile(`^card[0-9]+$`)
	RenderdRegexp = regexp.MustCompile(`^renderD[0-9]+$`)
)

type DeviceInfo struct {
	Name         string `json:"name"`          // Name is the device name, e.g. "cxi0"
	UID          string `json:"uid"`           // UID is the device PCI Address , e.g. "0000-00-00-0-0x0000"
	DeviceId     uint64 `json:"deviceid"`      // device number (e.g. 0 for /dev/cxi0)
	PCIAddress   string `json:"pciaddress"`    // PCI address in Linux DBDF notation for use with sysfs, e.g. 0000:00:00.0
	LocalCPUs    string `json:"local_cpus"`    // list of local CPU cores, e.g. ["ffff0000","00000000","ffff0000","00000000"]
	LocalCPUList string `json:"local_cpulist"` // CPU list in Linux format, e.g. "0-3,8-11"
	NumaNode     string `json:"numa_node"`     // NUMA node number, e.g. 0
	Version      string `json:"version"`       // version of the HPE Cassini driver, e.g. "1.1"
	Speed        string `json:"speed"`         // speed of the HPE Cassini NIC in Mbps, e.g. 200000 for 1Gbps
}

// HPECXI check if a particular card is a HPE CXI NIC by checking the device's vendor ID
func HPECXI(device DeviceInfo, sysfspath string) bool {
	sysfsVendorPath := sysfspath + "/" + device.UID + "/vendor"
	contents, err := os.ReadFile(sysfsVendorPath)
	if err == nil {
		vendorID := strings.TrimSpace(string(contents))

		if vendorID == HPEvendorID {
			return true
		} else {
			klog.Infof("%s is not a HPE NIC.", device.UID)
		}
	} else {
		klog.Errorf("Error opening %s: %s", sysfsVendorPath, err)
	}
	return false
}

func (g *DeviceInfo) Clone() *DeviceInfo {
	di := *g
	return &di
}

type DevicesInfo map[string]*DeviceInfo

func (g *DevicesInfo) Clone() DevicesInfo {
	devicesInfoCopy := DevicesInfo{}
	for duid, device := range *g {
		devicesInfoCopy[duid] = device.Clone()
	}
	return devicesInfoCopy
}
func convertDeviceInfoToDeviceSpec(device DeviceInfo) *pluginapi.DeviceSpec {
	devicePath := GetDevPath() + strconv.FormatUint(device.DeviceId, 10)
	return &pluginapi.DeviceSpec{
		ContainerPath: devicePath,
		HostPath:      devicePath,
		Permissions:   "rw",
	}
}

func (g *DevicesInfo) ConvertToDeviceSpecs() []*pluginapi.DeviceSpec {
	var deviceSpecs []*pluginapi.DeviceSpec
	for _, device := range *g {
		deviceSpecs = append(deviceSpecs, convertDeviceInfoToDeviceSpec(*device))
	}
	return deviceSpecs
}

func GetDevPath() string {
	return filepath.Join(GetDevRoot(DevfsPath), DevfsPath)
}
