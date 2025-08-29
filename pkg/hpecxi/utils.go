package hpecxi

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	SysfsEnvVarName  = "SYSFS_ROOT"
	sysfsDefaultRoot = "/sys"

	DevfsEnvVarName  = "DEVFS_ROOT"
	devfsDefaultRoot = "/dev"

	LibfabricEnvVarName  = "OFI_ROOT"
	libfabricDefaultRoot = "/opt/cray/lib64"
	libfabricName        = "libfabric.so"

	LibcxiEnvVarName  = "CXI_ROOT"
	libcxiDefaultRoot = "/usr/lib64"
	libcxiName        = "libcxi.so"

	PCIAddressLength = len("0000:00:00.0")

	virtualDevicesEnvVarName   = "CXI_VIRTUAL_DEVICES"
	virtualDevicesDefaultValue = "0"
)

func GetSysfsRoot(sysfsPath string) string {
	sysfsRoot, found := os.LookupEnv(SysfsEnvVarName)

	if found {
		if _, err := os.Stat(path.Join(sysfsRoot, sysfsPath)); err == nil {
			klog.V(4).Infof("using custom sysfs location: %v\n", sysfsRoot)
			return sysfsRoot
		} else {
			klog.V(4).Infof("could not find sysfs at '%v' from %v env var: %v\n", sysfsPath, SysfsEnvVarName, err)
		}
	}

	klog.V(4).Infof("using default sysfs location: %v\n", sysfsDefaultRoot)
	return sysfsDefaultRoot
}

func GetDevRoot(devPath string) string {
	devfsRoot, found := os.LookupEnv(DevfsEnvVarName)

	if found {
		if _, err := os.Stat(path.Join(devfsRoot, devPath)); err == nil {
			klog.V(4).Infof("using custom devfs location: %v\n", devfsRoot)
			return devfsRoot
		} else {
			klog.V(4).Infof("could not find devfs at '%v' from %v env var: %v\n", devPath, DevfsEnvVarName, err)
		}
	}

	klog.V(4).Infof("using default devfs root: %v\n", devfsDefaultRoot)
	return devfsDefaultRoot
}

func GetLibfabricRoot() (string, error) {
	libfabricRoot, found := os.LookupEnv(LibfabricEnvVarName)
	if found {
		exists := false
		err := existInPath(libfabricName, libfabricRoot, &exists)
		if err != nil {
			return "", err
		}
		if exists {
			klog.V(4).Infof("using custom Libfabric location: %v\n", libfabricRoot)
			return libfabricRoot, nil
		}
	}
	exists := false
	err := existInPath(libfabricName, libfabricDefaultRoot, &exists)
	if err != nil {
		return "", err
	}
	if exists {
		klog.V(4).Infof("using default Libfabric root: %v\n", libfabricDefaultRoot)
		return libfabricDefaultRoot, nil
	}
	return "", fmt.Errorf("no Libfabric found")
}

func GetLibcxiRoot() (string, error) {
	libcxiRoot, found := os.LookupEnv(LibcxiEnvVarName)
	if found {
		exists := false
		err := existInPath(libcxiName, libcxiRoot, &exists)
		if err != nil {
			return "", err
		}
		if exists {
			klog.V(4).Infof("using custom libcxi location: %v\n", libcxiRoot)
			return libcxiRoot, nil
		}
	}
	exists := false
	err := existInPath(libcxiName, libcxiDefaultRoot, &exists)
	if err != nil {
		return "", err
	}
	if exists {
		klog.V(4).Infof("using default libcxi root: %v\n", libcxiDefaultRoot)
		return libcxiDefaultRoot, nil
	}
	return "", fmt.Errorf("no libcxi found")
}

func existInPath(libName, libPath string, exists *bool) error {
	fileInfos, err := os.ReadDir(libPath)
	if err != nil {
		klog.Errorf("Error checking for %s in %s: %v", libName, libPath, err)
		return err
	}
	*exists = false
	for _, fileInfo := range fileInfos {
		if strings.HasPrefix(fileInfo.Name(), libName) {
			*exists = true
		}
	}
	return nil
}

func GetVirtualDevicesCount() int {
	virtualDevicesPerPhysical, found := os.LookupEnv(virtualDevicesEnvVarName)
	if !found {
		virtualDevicesPerPhysical = virtualDevicesDefaultValue
	}
	count, err := strconv.Atoi(virtualDevicesPerPhysical)
	if err != nil {
		klog.Errorf("Error parsing %s: %v", virtualDevicesEnvVarName, err)
		return 0
	}
	return count
}

func PciInfoFromDeviceUID(deviceUID string) (string, string) {
	// 0000-00-01-0-0x0000 -> 0000:00:01.0, 0x0000
	rfc1123PCIaddress := deviceUID[:PCIAddressLength]
	pciAddress := strings.Replace(strings.Replace(rfc1123PCIaddress, "-", ":", 2), "-", ".", 1)
	deviceId := deviceUID[PCIAddressLength+1:]

	return pciAddress, deviceId
}

func DeviceUIDFromPCIinfo(pciAddress string, pciid string) string {
	// 0000:00:01.0, 0x0000 -> 0000-00-01-0-0x0000
	// Replace colons and the dot in PCI address with hyphens.
	rfc1123PCIaddress := strings.ReplaceAll(strings.ReplaceAll(pciAddress, ":", "-"), ".", "-")
	newUID := fmt.Sprintf("%v-%v", rfc1123PCIaddress, pciid)

	return newUID
}

// ExtractCXINumber extracts the integer after '/dev/cxi' from a device path
func ExtractCXINumber(devicePath string) (int, error) {
	re := regexp.MustCompile(`/dev/cxi(\d+)`)
	matches := re.FindStringSubmatch(devicePath)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no CXI number found in device path: %s", devicePath)
	}
	return strconv.Atoi(matches[1])
}
