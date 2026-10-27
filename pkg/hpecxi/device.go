// Package hpecxi provides utilities and data structures for managing HPE CXI devices
// in Kubernetes device plugin environments. It includes device identification,
// cloning, and conversion to Kubernetes device specifications.
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
	// HPEvendorID is the PCI vendor ID for HPE devices
	HPEvendorID string = "0x17db"

	// DevfsPath is the device name for CXI devices
	DevfsPath = "cxi"
	// Sysfspath is the default sysfs path for CXI device class
	Sysfspath = "class/cxi"

	// Vendor identifies the device vendor namespace
	Vendor = "hpe.com"
	// Class identifies the device class
	Class = "cxi"
	// Kind combines vendor and class for Kubernetes resource identification
	Kind = Vendor + "/" + Class
	// DriverName combines class and vendor for driver identification
	DriverName = Class + "." + Vendor

	// UIDLength defines the expected length of device UID strings
	UIDLength = len("0000-00-00-0-0x0000")
)

var (
	// PciRegexp matches PCI addresses in DBDF format (domain:bus:device.function)
	PciRegexp = regexp.MustCompile(`[0-9a-f]{4}:[0-9a-f]{2}:[0-9a-f]{2}\.[0-7]$`)
	// CardRegexp matches card device names like "card0", "card1", etc.
	CardRegexp = regexp.MustCompile(`^card[0-9]+$`)
	// RenderdRegexp matches render device names like "renderD128", "renderD129", etc.
	RenderdRegexp = regexp.MustCompile(`^renderD[0-9]+$`)
)

// DeviceInfo holds metadata about a CXI device for use in device plugin operations.
type DeviceInfo struct {
	Name         string `json:"name"`          // Device name, e.g. "cxi0"
	UID          string `json:"uid"`           // PCI Address, e.g. "0000-00-00-0-0x0000"
	DeviceId     uint64 `json:"deviceid"`      // Device number (e.g. 0 for /dev/cxi0)
	PCIAddress   string `json:"pciaddress"`    // PCI address in Linux DBDF notation, e.g. 0000:00:00.0
	LocalCPUs    string `json:"local_cpus"`    // Hex mask of local CPU cores, e.g. "ffff0000"
	LocalCPUList string `json:"local_cpulist"` // Comma-separated Linux CPU list, e.g. "0-3,8-11"
	NumaNode     string `json:"numa_node"`     // NUMA node number, e.g. "0"
	Version      string `json:"version"`       // HPE Cassini driver version, e.g. "1.1"
	Speed        string `json:"speed"`         // NIC speed in Mbps, e.g. "200000" for 200Gbps
}

// HPECXI checks if a particular card is a HPE CXI NIC by verifying the device's vendor ID in sysfs.
// Returns true if the vendor ID matches HPEvendorID, otherwise false.
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

// Clone creates a deep copy of the DeviceInfo instance.
func (g *DeviceInfo) Clone() *DeviceInfo {
	di := *g
	return &di
}

// DevicesInfo is a map of device UID to DeviceInfo pointer.
type DevicesInfo map[string]*DeviceInfo

// Clone creates a deep copy of the DevicesInfo map and its DeviceInfo values.
func (g *DevicesInfo) Clone() DevicesInfo {
	devicesInfoCopy := DevicesInfo{}
	for duid, device := range *g {
		devicesInfoCopy[duid] = device.Clone()
	}
	return devicesInfoCopy
}

// convertDeviceInfoToDeviceSpec converts a DeviceInfo to a Kubernetes DeviceSpec for device plugin API.
func convertDeviceInfoToDeviceSpec(device DeviceInfo) *pluginapi.DeviceSpec {
	devicePath := GetDevPath() + strconv.FormatUint(device.DeviceId, 10)
	return &pluginapi.DeviceSpec{
		ContainerPath: devicePath,
		HostPath:      devicePath,
		Permissions:   "rw",
	}
}

// ConvertToDeviceSpecs converts all DeviceInfo entries in DevicesInfo to DeviceSpec slices for Kubernetes device plugin API.
func (g *DevicesInfo) ConvertToDeviceSpecs() []*pluginapi.DeviceSpec {
	var deviceSpecs []*pluginapi.DeviceSpec
	for _, device := range *g {
		deviceSpecs = append(deviceSpecs, convertDeviceInfoToDeviceSpec(*device))
	}
	return deviceSpecs
}

// GetDevPath returns the full device path for CXI devices in /dev.
func GetDevPath() string {
	return filepath.Join(GetDevRoot(DevfsPath), DevfsPath)
}
